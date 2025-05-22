package bench

import (
	"context"
	"os/exec"
	"strings"
)

// Run executes `opa bench` on the given paths, optionally supplying an input file.
func Run(ctx context.Context, paths []string, inputFile string) (string, error) {
	args := []string{"bench"}
	args = append(args, paths...)
	if inputFile != "" {
		args = append(args, "--input", inputFile)
	}
	cmd := exec.CommandContext(ctx, "opa", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
