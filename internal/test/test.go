package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Timeout        time.Duration
	MaxWorkers     int
	TestFlags      []string                // e.g., ["--verbose"]
	InputFile      string                  // Optional path to input.json
	OnTestComplete func(result TestResult) // Optional callback for progress reporting
}

type TestResult struct {
	Dir      string
	Output   string
	Passed   bool
	Coverage json.RawMessage
	Summary  CoverageSummary
}

type CoverageSummary struct {
	FileCount    int
	TotalRules   int
	CoveredRules int
	Percent      float64
}

// Run executes tests on all .rego/.test.rego files in the given directories.
func Run(ctx context.Context, rootDirs []string, cfg Config) ([]TestResult, error) {
	if cfg.MaxWorkers <= 0 {
		return nil, fmt.Errorf("MaxWorkers must be > 0")
	}

	dirs, err := findTestDirs(rootDirs)
	if err != nil {
		return nil, err
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []TestResult
		errors  []error
	)

	sem := make(chan struct{}, cfg.MaxWorkers)

	for _, dir := range dirs {
		wg.Add(1)
		go func(testDir string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			localCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
			defer cancel()

			args := []string{"test", testDir, "--format=json", "--coverage"}
			if cfg.InputFile != "" {
				args = append(args, "-i", cfg.InputFile)
			}
			args = append(args, cfg.TestFlags...)

			cmd := exec.CommandContext(localCtx, "opa", args...)
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			err := cmd.Run()

			res := TestResult{
				Dir:    testDir,
				Output: out.String(),
				Passed: err == nil,
			}

			// Always attempt to parse coverage, even on failed tests
			res.Coverage = out.Bytes()

			summary, parseErr := summarizeCoverage(res.Coverage)
			if parseErr == nil {
				res.Summary = summary
			}

			mu.Lock()
			results = append(results, res)
			if err != nil {
				errors = append(errors, fmt.Errorf("test failed for %s: %w", testDir, err))
			}
			mu.Unlock()

			if cfg.OnTestComplete != nil {
				cfg.OnTestComplete(res)
			}
		}(dir)
	}

	wg.Wait()

	if len(errors) > 0 {
		return results, fmt.Errorf("%d test failures (first: %v)", len(errors), errors[0])
	}
	return results, nil
}

// summarizeCoverage safely extracts rule-level coverage info.
func summarizeCoverage(raw json.RawMessage) (CoverageSummary, error) {
	var report struct {
		Files map[string]struct {
			Rules map[string]bool `json:"rules"`
		} `json:"files"`
	}

	if err := json.Unmarshal(raw, &report); err != nil {
		return CoverageSummary{}, err
	}

	summary := CoverageSummary{FileCount: len(report.Files)}
	for _, file := range report.Files {
		summary.TotalRules += len(file.Rules)
		for _, covered := range file.Rules {
			if covered {
				summary.CoveredRules++
			}
		}
	}

	if summary.TotalRules > 0 {
		summary.Percent = (float64(summary.CoveredRules) / float64(summary.TotalRules)) * 100
	}

	return summary, nil
}

// findTestDirs returns directories containing at least one .rego file.
func findTestDirs(rootDirs []string) ([]string, error) {
	dirSet := make(map[string]struct{})
	for _, root := range rootDirs {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".rego") {
				dirSet[filepath.Dir(path)] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking %s: %w", root, err)
		}
	}

	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	return dirs, nil
}
