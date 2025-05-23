package format

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/ast"
	opaformat "github.com/open-policy-agent/opa/format"
)

// Format takes raw Rego source bytes, parses them, and returns formatted output.
func Format(input []byte) ([]byte, error) {
	mod, err := ast.ParseModule("", string(input))
	if err != nil {
		return nil, err
	}

	formatted, err := opaformat.Ast(mod)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if _, err := buf.Write(formatted); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// FormatAll concurrently formats all .rego files under the given paths.
// Returns a map[filePath]formattedContent and an error if any.
func FormatAll(ctx context.Context, paths []string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	sem := make(chan struct{}, 8) // limit to 8 concurrent formatters

	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".rego") {
				wg.Add(1)
				go func(p string) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					data, err := os.ReadFile(p)
					if err != nil {
						mu.Lock()
						if firstErr == nil {
							firstErr = fmt.Errorf("error reading %s: %w", p, err)
						}
						mu.Unlock()
						return
					}

					formatted, err := Format(data)
					if err != nil {
						mu.Lock()
						if firstErr == nil {
							firstErr = fmt.Errorf("formatting error in %s: %w", p, err)
						}
						mu.Unlock()
						return
					}

					mu.Lock()
					result[p] = formatted
					mu.Unlock()
				}(path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("directory walk error: %w", err)
		}
	}

	wg.Wait()
	return result, firstErr
}
