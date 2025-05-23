# OPA Toolkit

A unified Go library that simplifies policy development with the Open Policy Agent (OPA) by combining linting, formatting, testing, and benchmarking into one interface.

---

## Features

-  Linting with [Regal](https://github.com/StyraInc/regal)
-  Formatting using [OPA Formatter](https://www.openpolicyagent.org/docs/latest/tools/#format)
-  Testing using `opa test` with JSON and coverage support
-  Benchmarking using `opa bench`
-  Simple and unified interface for use in CI/CD or local development

---

## Installation

```bash
go get github.com/codejayish/opa-toolkit


---

## How to Use the OPA Toolkit

Here's a step-by-step guide to using the toolkit in your Go application:

### 1. Import the Toolkit

Import the package in your Go code:

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/codejayish/opa-toolkit/toolkit"
)

### 2. Initialize the Toolkit

Create a new instance and a context:

```go
ctx := context.Background()
tk := toolkit.New()

### 3. Run Linter on Policies

Use the linter to detect issues in your `.rego` files. It leverages Regal and returns structured findings.

```go
findings, err := tk.Lint(ctx, []string{"examples/policies"})
if err != nil {
    panic(err)
}
fmt.Printf("Lint findings:\n%+v\n", findings)

### 4. Format Rego Files

Format raw Rego source code. This is useful for in-memory formatting or pre-commit hooks.

```go
raw, err := os.ReadFile("examples/policies/example.rego")
if err != nil {
    panic(err)
}

formattedMap, err := tk.Format(ctx, map[string][]byte{
    "examples/policies/example.rego": raw,
})
if err != nil {
    panic(err)
}

fmt.Println("Formatted Code:\n", string(formattedMap["examples/policies/example.rego"]))

### 5. Run Tests with Coverage

Execute `opa test` with automatic JSON output and coverage reporting.

```go
testOutput, err := tk.Test(ctx, []string{"examples/policies", "examples/data"})
if err != nil {
    fmt.Println("Test failed:", err)
}
fmt.Println("Test output:\n", testOutput)

### 6. Run Benchmarks on Policies

Benchmark specific rules using `opa bench`, providing policy paths and an optional input file.

```go
benchOutput, err := tk.Bench(ctx, []string{"examples/policies"}, "examples/data/input.json")
if err != nil {
    fmt.Println("Bench failed:", err)
}
fmt.Println("Bench output:\n", benchOutput)

## 📁 Example Project Structure

A typical project using this toolkit might look like this:

```bash
opa-demo/
├── main.go               # Your Go application using the toolkit
├── go.mod
├── go.sum
└── examples/             # Your OPA policies and data
    ├── policies/
    │   ├── example.rego      # Your policy file
    │   └── example_test.rego # Your test file
    └── data/
        └── input.json        # Sample input data for testing/benchmarking
