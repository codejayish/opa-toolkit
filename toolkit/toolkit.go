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
func (t *Toolkit) Format(ctx context.Context, inputs map[string][]byte) (map[string][]byte, error) {
	out := make(map[string][]byte, len(inputs))
	for path, content := range inputs {
		formatted, err := format.Format(content)
		if err != nil {
			return nil, err
		}
		out[path] = formatted
	}
	return out, nil
}

// Test runs `opa test --format json --coverage` under the hood.
func (t *Toolkit) Test(ctx context.Context, dirs []string) (string, error) {
	return test.Run(ctx, dirs)
}

// Bench runs `opa bench` and returns the raw output.
func (t *Toolkit) Bench(ctx context.Context, query string, paths []string, inputFile string) (string, error) {
	return bench.Run(ctx, query, paths, inputFile)
}
