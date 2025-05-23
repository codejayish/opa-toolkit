package lint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/styrainc/regal/pkg/linter"
)

// Finding represents a single lint violation.
type Finding struct {
	File    string
	Rule    string
	Message string
}

// Run concurrently lints all .rego files under the given paths.
func Run(ctx context.Context, paths []string) ([]Finding, error) {
	// Step 1: Discover all .rego files under the provided paths
	var allFiles []string
	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".rego") {
				allFiles = append(allFiles, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error scanning directory: %w", err)
		}
	}

	// Step 2: Lint files in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allFindings []Finding
	var firstErr error

	semaphore := make(chan struct{}, 8) // Limit concurrency (e.g., 8 workers)

	for _, file := range allFiles {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			semaphore <- struct{}{}        // acquire slot
			defer func() { <-semaphore }() // release slot

			l := linter.NewLinter().WithInputPaths([]string{filePath})
			rep, err := l.Lint(ctx)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("lint error in %s: %w", filePath, err)
				}
				mu.Unlock()
				return
			}

			var localFindings []Finding
			for _, v := range rep.Violations {
				localFindings = append(localFindings, Finding{
					File:    v.Location.File,
					Rule:    v.Title,
					Message: v.Description,
				})
			}

			mu.Lock()
			allFindings = append(allFindings, localFindings...)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	return allFindings, firstErr
}
