package bench

import (
	"context"
	"os/exec"
	"strings"
)

// Run executes `opa bench "<query>" -i <input> -d <paths>` and returns the output.
func Run(ctx context.Context, query string, paths []string, inputFile string) (string, error) {
	args := []string{"bench", query}

	if inputFile != "" {
		args = append(args, "-i", inputFile)
	}

	for _, path := range paths {
		args = append(args, "-d", path)
	}

	cmd := exec.CommandContext(ctx, "opa", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
