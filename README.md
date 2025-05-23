# OPA Toolkit

A unified Go library that simplifies policy development with the Open Policy Agent (OPA) by combining linting, formatting, testing, and benchmarking into one interface.

---

## Features

* **Linting** with [Regal](https://github.com/StyraInc/regal)
* **Formatting** using [OPA Formatter](https://www.openpolicyagent.org/docs/latest/tools/#format)
* **Testing** via `opa test` with JSON output and coverage support
* **Benchmarking** via `opa bench`
* **Unified interface** suitable for CI/CD integration or local development

---

## Installation

```bash
go get github.com/codejayish/opa-toolkit
```

---

## Usage

### 1. Import the Toolkit

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/codejayish/opa-toolkit/toolkit"
)
```

### 2. Initialize the Toolkit

```go
ctx := context.Background()
tk  := toolkit.New()
```

### 3. Run Linter on Policies

Leverage Regal to detect issues in your `.rego` files and receive structured findings.

```go
findings, err := tk.Lint(ctx, []string{"examples/policies"})
if err != nil {
    panic(err)
}
fmt.Printf("Lint findings:\n%+v\n", findings)
```

### 4. Format Rego Files

Format in-memory or on-disk Rego code (ideal for pre-commit hooks).

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
```

### 5. Run Tests with Coverage

Execute `opa test` with automatic JSON parsing and coverage reporting.

```go
testOutput, err := tk.Test(ctx, []string{"examples/policies", "examples/data"})
if err != nil {
    fmt.Println("Test failed:", err)
}
fmt.Println("Test output:\n", testOutput)
```

### 6. Run Benchmarks on Policies

Invoke `opa bench` to measure performance of specific rules.

```go
benchOutput, err := tk.Bench(ctx, []string{"examples/policies"}, "examples/data/input.json")
if err != nil {
    fmt.Println("Bench failed:", err)
}
fmt.Println("Bench output:\n", benchOutput)
```

---

## Example Project Structure

```
opa-demo/
├── go.mod
├── go.sum
├── main.go               # Your Go application using the OPA Toolkit
└── examples/             # OPA policies and data
    ├── policies/
    │   ├── example.rego      # Policy definition
    │   └── example_test.rego # Policy tests
    └── data/
        └── input.json        # Sample input for testing/benchmarking
```

---

