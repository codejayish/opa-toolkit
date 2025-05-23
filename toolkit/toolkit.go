package toolkit

import (
	"context"

	"github.com/codejayish/opa-toolkit/internal/bench"
	"github.com/codejayish/opa-toolkit/internal/format"
	"github.com/codejayish/opa-toolkit/internal/lint"
	"github.com/codejayish/opa-toolkit/internal/test"
)

// Toolkit is the unified interface for OPA tooling.
type Toolkit struct{}

// New returns a fresh Toolkit instance.
func New() *Toolkit {
	return &Toolkit{}
}

// Lint runs Regal on the given paths and returns any findings.
func (t *Toolkit) Lint(ctx context.Context, paths []string) ([]lint.Finding, error) {
	return lint.Run(ctx, paths)
}

// Format takes a map[filePath]rawBytes and returns map[filePath]formattedBytes.
func (t *Toolkit) Format(ctx context.Context, inputs map[string][]byte) (map[string][]byte, error) { //Accepts a map: {file path -> raw contents}
	out := make(map[string][]byte, len(inputs)) //Returns another map: {file path -> formatted contents}
	for path, content := range inputs {
		formatted, err := format.Format(content) //Formats each .rego file using OPA's AST + formatter
		if err != nil {
			return nil, err
		}
		out[path] = formatted
	}
	return out, nil
}

// FormatAll scans all .rego files under the given paths,
func (t *Toolkit) FormatAll(ctx context.Context, paths []string) (map[string][]byte, error) {
	return format.FormatAll(ctx, paths)
}

// Test runs `opa test --format json --coverage` under the hood.
func (t *Toolkit) Test(ctx context.Context, rootDirs []string) (string, error) {
	return test.Run(ctx, rootDirs)
}

// Bench runs `opa bench` and returns the raw output.
func (t *Toolkit) Bench(ctx context.Context, query string, paths []string, inputFile string) (string, error) {
	return bench.Run(ctx, query, paths, inputFile)
}

// BenchMany runs multiple opa bench queries concurrently and returns a map[query]output.
func (t *Toolkit) BenchMany(ctx context.Context, queries []string, paths []string, inputFile string) (map[string]string, error) {
	return bench.RunMany(ctx, queries, paths, inputFile)
}

// BenchSummary runs multiple benchmark queries and returns a summarized report string.
func (t *Toolkit) BenchSummary(ctx context.Context, queries []string, paths []string, inputFile string) (string, error) {
	// Run all benchmarks
	results, err := bench.RunMany(ctx, queries, paths, inputFile)
	if err != nil {
		return "", err
	}

	// Generate and return summary
	summary := bench.GenerateSummary(results)
	return summary, nil
}
