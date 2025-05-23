package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/styrainc/regal/pkg/linter"
)

// Finding represents a single lint violation.
type Finding struct {
	File    string `json:"file"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

// Config allows configurable linting behavior.
type Config struct {
	MaxWorkers   int
	OutputFormat string // Options: "text", "json", "github"
	PrintOutput  bool
}

// Run performs concurrent linting of all .rego files and prints formatted output if enabled.
func Run(ctx context.Context, paths []string, cfg Config) ([]Finding, error) {
	allFiles, err := findRegoFiles(paths)
	if err != nil {
		return nil, fmt.Errorf("file discovery failed: %w", err)
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		allFindings []Finding
		allErrors   []error
	)

	semaphore := make(chan struct{}, cfg.MaxWorkers)

	for _, file := range allFiles {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			l := linter.NewLinter()

			rep, err := l.WithInputPaths([]string{filePath}).Lint(ctx)
			if err != nil {
				mu.Lock()
				allErrors = append(allErrors, fmt.Errorf("%s: %w", filePath, err))
				mu.Unlock()
				return
			}

			var localFindings []Finding
			for _, v := range rep.Violations {
				localFindings = append(localFindings, Finding{
					File:    v.Location.File,
					Rule:    v.Title,
					Message: v.Description,
					Line:    v.Location.Row,
				})
			}

			mu.Lock()
			allFindings = append(allFindings, localFindings...)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	if cfg.PrintOutput {
		formatOutput(cfg.OutputFormat, allFindings)
	}

	if len(allErrors) > 0 {
		return allFindings, fmt.Errorf("encountered %d errors (first: %v)", len(allErrors), allErrors[0])
	}

	return allFindings, nil
}

// findRegoFiles recursively finds all .rego files in given paths.
func findRegoFiles(paths []string) ([]string, error) {
	var files []string
	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".rego") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// formatOutput prints findings in the specified output format.
func formatOutput(format string, findings []Finding) {
	switch strings.ToLower(format) {
	case "json":
		jsonOut, _ := json.MarshalIndent(findings, "", "  ")
		fmt.Println(string(jsonOut))

	case "github":
		for _, f := range findings {
			fmt.Printf("::error file=%s,line=%d::[%s] %s\n", f.File, f.Line, f.Rule, f.Message)
		}

	default: // text
		for _, f := range findings {
			fmt.Printf("%s:%d [%s] %s\n", f.File, f.Line, f.Rule, f.Message)
		}
	}
}
