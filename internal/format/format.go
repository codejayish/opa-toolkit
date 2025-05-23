package format

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/ast"
	opaformat "github.com/open-policy-agent/opa/format"
)

// Config holds formatting options and behavior hooks.
type Config struct {
	MaxWorkers      int               // Number of concurrent workers
	Write           bool              // If true, write files back to disk
	OnFileFormatted func(path string) // Optional progress callback
}

// Format parses and formats the input Rego source.
func Format(input []byte) ([]byte, error) {
	mod, err := ast.ParseModule("", string(input))
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	formatted, err := opaformat.Ast(mod)
	if err != nil {
		return nil, fmt.Errorf("format error: %w", err)
	}

	return formatted, nil
}

// FormatAll discovers, formats, and optionally writes .rego files.
func FormatAll(ctx context.Context, paths []string, cfg Config) (map[string][]byte, error) {
	if cfg.MaxWorkers <= 0 {
		return nil, fmt.Errorf("MaxWorkers must be > 0")
	}

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		result = make(map[string][]byte)
		errors []error
	)

	sem := make(chan struct{}, cfg.MaxWorkers)

	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".rego") {
				wg.Add(1)
				go func(p string, mode os.FileMode) {
					defer wg.Done()

					select {
					case <-ctx.Done():
						return
					case sem <- struct{}{}:
						defer func() { <-sem }()
					}

					data, err := os.ReadFile(p)
					if err != nil {
						mu.Lock()
						errors = append(errors, fmt.Errorf("read %s: %w", p, err))
						mu.Unlock()
						return
					}

					formatted, err := Format(data)
					if err != nil {
						mu.Lock()
						errors = append(errors, fmt.Errorf("format %s: %w", p, err))
						mu.Unlock()
						return
					}

					mu.Lock()
					result[p] = formatted
					mu.Unlock()

					if cfg.Write {
						if err := atomicWrite(p, formatted, mode); err != nil {
							mu.Lock()
							errors = append(errors, fmt.Errorf("write %s: %w", p, err))
							mu.Unlock()
							return
						}
					}

					if cfg.OnFileFormatted != nil {
						cfg.OnFileFormatted(p)
					}
				}(path, info.Mode())
			}
			return nil
		})
		if err != nil {
			errors = append(errors, fmt.Errorf("walk %s: %w", root, err))
		}
	}

	wg.Wait()

	if len(errors) > 0 {
		return result, fmt.Errorf("encountered %d errors (first: %v)", len(errors), errors[0])
	}
	return result, nil
}

// atomicWrite writes content to a temporary file and renames it atomically.
func atomicWrite(path string, content []byte, mode os.FileMode) error {
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, content, mode); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// WriteAll provides an alternate way to write pre-formatted content to files.
func WriteAll(formatted map[string][]byte) error {
	for path, content := range formatted {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		if err := atomicWrite(path, content, info.Mode()); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}
