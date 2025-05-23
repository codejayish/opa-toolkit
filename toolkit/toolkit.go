package toolkit

import (
	"context"

	"github.com/codejayish/opa-toolkit/internal/bench"
	"github.com/codejayish/opa-toolkit/internal/format"
	"github.com/codejayish/opa-toolkit/internal/lint"
	"github.com/codejayish/opa-toolkit/internal/test"
)

// Re-export config and result types so external users (like opa-demo) can use them
type (
	LintConfig      = lint.Config
	FormatConfig    = format.Config
	TestConfig      = test.Config
	BenchConfig     = bench.Config
	BenchmarkResult = bench.BenchmarkResult
	TestResult      = test.TestResult
)

// Toolkit is the unified interface for OPA tooling.
type Toolkit struct{}

// New returns a fresh Toolkit instance.
func New() *Toolkit {
	return &Toolkit{}
}

// Lint runs Regal on the given paths and returns any findings.
func (t *Toolkit) Lint(ctx context.Context, paths []string, cfg lint.Config) ([]lint.Finding, error) {
	return lint.Run(ctx, paths, cfg)
}

// Format takes a map[filePath]rawBytes and returns map[filePath]formattedBytes.
func (t *Toolkit) Format(ctx context.Context, paths []string, cfg format.Config) (map[string][]byte, error) {
	return format.FormatAll(ctx, paths, cfg)
}

// Test runs `opa test --format json --coverage` under the hood.
func (t *Toolkit) Test(ctx context.Context, paths []string, cfg test.Config) ([]test.TestResult, error) {
	return test.Run(ctx, paths, cfg)
}

// Bench runs benchmarking for given queries and returns performance results.
func (t *Toolkit) Bench(ctx context.Context, cfg bench.Config) (map[string]bench.BenchmarkResult, error) {
	return bench.Run(ctx, cfg)
}

// BenchSummary generates a report based on raw benchmark outputs.
func (t *Toolkit) BenchSummary(results map[string]bench.BenchmarkResult, format string) string {
	raw := make(map[string]string)
	for q, res := range results {
		if res.Error == nil {
			raw[q] = res.Output
		}
	}
	return bench.GenerateSummary(raw, format)
}
