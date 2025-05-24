package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Queries         []string
	Paths           []string
	InputFile       string
	MaxWorkers      int
	TimeoutPerQuery time.Duration
	WarmupRuns      int
	OnQueryComplete func(query string, result BenchmarkResult)
}

type BenchmarkResult struct {
	Output   string
	Stats    BenchmarkStats
	Duration time.Duration
	Error    error
}

type BenchmarkStats struct {
	Query      string  `json:"query"`
	Iterations int     `json:"iterations"`
	MeanNs     float64 `json:"mean_ns"`
	P99Ns      float64 `json:"p99_ns"`
	MemoryKB   int     `json:"memory_kb"`
}

// Run executes multiple benchmark queries concurrently and returns structured results.
func Run(ctx context.Context, cfg Config) (map[string]BenchmarkResult, error) {
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 4
	}
	if cfg.TimeoutPerQuery <= 0 {
		cfg.TimeoutPerQuery = 15 * time.Second
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		results  = make(map[string]BenchmarkResult)
		firstErr error
		sem      = make(chan struct{}, cfg.MaxWorkers)
	)

	for _, query := range cfg.Queries {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			localCtx, cancel := context.WithTimeout(ctx, cfg.TimeoutPerQuery)
			defer cancel()

			start := time.Now()
			output, err := runSingle(localCtx, q, cfg.Paths, cfg.InputFile, cfg.WarmupRuns)
			duration := time.Since(start)

			result := BenchmarkResult{
				Output:   output,
				Duration: duration,
				Error:    err, // still store error if exists
			}

			// Always try to parse the output, even if error occurred
			stats, parseErr := parseOPAOutput(output, q)
			if parseErr == nil {
				result.Stats = stats
			}

			mu.Lock()
			results[q] = result
			if err != nil && firstErr == nil {
				firstErr = err
			}
			mu.Unlock()

			if cfg.OnQueryComplete != nil {
				cfg.OnQueryComplete(q, result)
			}
		}(query)
	}

	wg.Wait()
	return results, firstErr
}

// runSingle executes `opa bench` for one query and returns its output.
func runSingle(ctx context.Context, query string, paths []string, inputFile string, warmup int) (string, error) {
	args := []string{"bench", query, "--format=json"}
	if inputFile != "" {
		args = append(args, "-i", inputFile)
	}
	for _, path := range paths {
		args = append(args, "-d", path)
	}

	cmd := exec.CommandContext(ctx, "opa", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err // return output even if there's an error
}

// parseOPAOutput extracts benchmark stats from JSON output.
func parseOPAOutput(output string, query string) (BenchmarkStats, error) {
	var flat struct {
		N         int            `json:"N"`
		Extra     map[string]any `json:"Extra"`
		MemBytes  int            `json:"MemBytes"`
		MemAllocs int            `json:"MemAllocs"`
	}

	if err := json.Unmarshal([]byte(output), &flat); err != nil {
		return BenchmarkStats{}, fmt.Errorf("failed to parse flat OPA benchmark JSON: %w", err)
	}

	stats := BenchmarkStats{
		Query:      query,
		Iterations: flat.N,
		MemoryKB:   flat.MemBytes / 1024,
	}

	// Pull known fields from Extra if available
	if v, ok := flat.Extra["histogram_timer_rego_query_eval_ns_mean"].(float64); ok {
		stats.MeanNs = v
	}
	if v, ok := flat.Extra["histogram_timer_rego_query_eval_ns_99%"].(float64); ok {
		stats.P99Ns = v
	}

	return stats, nil
}
