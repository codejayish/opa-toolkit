package test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Run discovers subdirs with .rego files, runs opa test per dir, and aggregates results
func Run(ctx context.Context, rootDirs []string) (string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []string
	var allErr error

	// Discover all unique testable dirs
	dirSet := make(map[string]struct{})
	for _, root := range rootDirs {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".rego") {
				dir := filepath.Dir(path)
				dirSet[dir] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("error walking directories: %w", err)
		}
	}

	// Convert map to list of dirs
	var dirs []string
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}

	// Run tests on each dir in parallel
	for _, dir := range dirs {
		wg.Add(1)

		go func(testDir string) {
			defer wg.Done()

			// Use a child context with timeout
			localCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(localCtx, "opa", "test", testDir, "--format", "json")
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			if err := cmd.Run(); err != nil {
				mu.Lock()
				allErr = fmt.Errorf("error testing %s: %w", testDir, err)
				mu.Unlock()
			}

			mu.Lock()
			results = append(results, fmt.Sprintf("=== %s ===\n%s", testDir, out.String()))
			mu.Unlock()
		}(dir)
	}

	wg.Wait()

	return strings.Join(results, "\n\n"), allErr
}
