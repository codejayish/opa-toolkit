package bench

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// RunSingle executes a single `opa bench` query
func Run(ctx context.Context, query string, paths []string, inputFile string) (string, error) {
	args := []string{"bench", query}

	if inputFile != "" {
		args = append(args, "-i", inputFile)
	}

	for _, path := range paths {
		args = append(args, "-d", path)
	}

	cmd := exec.CommandContext(ctx, "opa", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("benchmark failed for query %q: %w", query, err)
	}

	return out.String(), nil
}

// RunMany concurrently runs `opa bench` for multiple queries, aggregating results.
func RunMany(ctx context.Context, queries []string, paths []string, inputFile string) (map[string]string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make(map[string]string)
	var firstErr error

	sem := make(chan struct{}, 4) // Control parallelism

	for _, query := range queries {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Scoped context per query
			localCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			output, err := Run(localCtx, q, paths, inputFile)
			mu.Lock()
			defer mu.Unlock()

			if err != nil && firstErr == nil {
				firstErr = err
			}
			results[q] = output
		}(query)
	}

	wg.Wait()
	return results, firstErr
}
