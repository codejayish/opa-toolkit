// examples/main.go
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/codejayish/opa-toolkit/pkg/toolkit"
)

func main() {
	tk := toolkit.New()
	ctx := context.Background()

	// 1) Lint
	findings, err := tk.Lint(ctx, []string{"examples/policies"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Lint findings: %+v\n\n", findings)

	// 2) Format
	inPath := "examples/policies/example.rego"
	raw, err := ioutil.ReadFile(inPath)
	if err != nil {
		panic(err)
	}
	formattedMap, err := tk.Format(ctx, map[string][]byte{filepath.Base(inPath): raw})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Formatted %s:\n%s\n\n", inPath, string(formattedMap[filepath.Base(inPath)]))

	// 3) Test (with coverage)
	testOut, err := tk.Test(ctx, []string{"examples/policies", "examples/data"})
	if err != nil {
		panic(fmt.Errorf("test failed: %v\n%s", err, testOut))
	}
	fmt.Printf("Test output (JSON+coverage):\n%s\n\n", testOut)

	// 4) Bench
	benchOut, err := tk.Bench(ctx, []string{"examples/policies"}, "examples/data/input.json")
	if err != nil {
		panic(fmt.Errorf("bench failed: %v\n%s", err, benchOut))
	}
	fmt.Printf("Bench output:\n%s\n", benchOut)
}
