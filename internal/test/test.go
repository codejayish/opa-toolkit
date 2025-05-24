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
	TestFlags      []string
	InputFile      string
	OnTestComplete func(result TestResult)
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

			args := []string{
				"test",
				testDir,
				"--format=json",
				"--coverage",
				"--ignore=.*",
			}
			if cfg.InputFile != "" {
				args = append(args, "--data", cfg.InputFile)
			}
			args = append(args, cfg.TestFlags...)

			cmd := exec.CommandContext(localCtx, "opa", args...)
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			err := cmd.Run()
			output := out.String()

			res := TestResult{
				Dir:    testDir,
				Output: output,
				Passed: err == nil,
			}

			if summary, parseErr := summarizeCoverage([]byte(output)); parseErr == nil {
				res.Summary = summary
			}

			mu.Lock()
			results = append(results, res)
			if err != nil {
				errors = append(errors, fmt.Errorf("test failed for %s: %w\nOutput:\n%s",
					testDir, err, output))
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

func findTestDirs(rootDirs []string) ([]string, error) {
	dirSet := make(map[string]struct{})
	for _, root := range rootDirs {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if (!info.IsDir() && strings.HasSuffix(path, "_test.rego")) ||
				(info.IsDir() && filepath.Base(path) == "test") {
				dirSet[filepath.Dir(path)] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	dirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

func summarizeCoverage(raw []byte) (CoverageSummary, error) {
	var coverage struct {
		Coverage float64 `json:"coverage"`
		Files    map[string]struct {
			Covered []struct {
				Start struct{ Row int } `json:"start"`
				End   struct{ Row int } `json:"end"`
			} `json:"covered"`
			NotCovered []struct {
				Start struct{ Row int } `json:"start"`
				End   struct{ Row int } `json:"end"`
			} `json:"not_covered"`
		} `json:"files"`
	}

	if err := json.Unmarshal(raw, &coverage); err != nil {
		return CoverageSummary{}, err
	}

	total := 0
	covered := 0
	for _, file := range coverage.Files {
		total += len(file.NotCovered) + len(file.Covered)
		covered += len(file.Covered)
	}

	return CoverageSummary{
		FileCount:    len(coverage.Files),
		TotalRules:   total,
		CoveredRules: covered,
		Percent:      coverage.Coverage,
	}, nil
}
