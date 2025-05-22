// internal/lint/lint.go
package lint

import (
	"context"

	"github.com/styrainc/regal/pkg/linter"
)

// Finding represents a single lint violation.
type Finding struct {
	File    string
	Rule    string
	Message string
}

// Run lints all .rego files under the given paths.
func Run(ctx context.Context, paths []string) ([]Finding, error) {
	// Create a new Regal linter configured to scan the given paths.
	l := linter.NewLinter().
		WithInputPaths(paths)

	// Execute linting
	rep, err := l.Lint(ctx)
	if err != nil {
		return nil, err
	}

	// Flatten the Report.Violations into our Finding struct
	var results []Finding
	for _, v := range rep.Violations {
		results = append(results, Finding{
			File:    v.Location.File,
			Rule:    v.Title,
			Message: v.Description,
		})
	}
	return results, nil
}
