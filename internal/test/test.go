package test

import (
	"context"
	"os/exec"
	"strings"
)

// Run executes `opa test --format json --coverage` on the provided directories.
func Run(ctx context.Context, dirs []string) (string, error) {
	args := []string{"test", "--format", "json", "--coverage"}
	args = append(args, dirs...)
	cmd := exec.CommandContext(ctx, "opa", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
