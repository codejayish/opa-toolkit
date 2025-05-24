# 🔧 OPA Toolkit (Enhanced)

A unified Go library that simplifies policy development with the [Open Policy Agent (OPA)](https://www.openpolicyagent.org/), integrating:

-  **Linting** with Regal
-  **Formatting** using OPA Formatter
-  **Testing** with JSON output and rule-level coverage
-  **Benchmarking** with structured metrics (mean, p99, iterations, memory)
-  **Summaries** (text and markdown) for CI pipelines or local visibility

---

##  Features

| Feature       | Description                                      |
|---------------|--------------------------------------------------|
|  Linting     | Run Regal linter on all Rego files               |
|  Formatting  | Format `.rego` files using `opa fmt`             |
|  Testing     | Execute tests with rule-level coverage parsing   |
|  Benchmarking| Use `opa bench` and parse output into Go structs |
|  Summarize   | Generate readable performance summaries          |
|  Extensible  | Clean Go API with hook support                   |

---

##  Installation and Usage

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
findings, err := tk.Lint(ctx, []string{"examples/policies"}, toolkit.LintConfig{
    MaxWorkers:   8,
    OutputFormat: "text",
    PrintOutput:  true,
})
```

### 4. Format Rego Files

Format in-memory or on-disk Rego code (ideal for pre-commit hooks).

```go
formatted, err := tk.Format(ctx, []string{"examples/policies"}, toolkit.FormatConfig{
    MaxWorkers: 8,
    Write:      true,
    OnFileFormatted: func(path string) {
        fmt.Println(" Formatted:", path)
    },
})
```

### 5. Run Tests with Coverage

Execute `opa test` with automatic JSON parsing and coverage reporting.

```go
results, err := tk.Test(ctx, []string{"examples/policies"}, toolkit.TestConfig{
    InputFile:  "examples/data/input.json",
    Timeout:    10 * time.Second,
    MaxWorkers: 4,
    TestFlags:  []string{"--verbose"},
    OnTestComplete: func(res toolkit.TestResult) {
        fmt.Printf(" %s — Coverage: %.2f%% (%d/%d rules)\n",
            res.Dir,
            res.Summary.Percent,
            res.Summary.CoveredRules,
            res.Summary.TotalRules,
        )
    },
})
```

### 6. Benchmark a Single Query

Invoke `opa bench` to measure performance of specific rules.

```go
results, err := tk.Bench(ctx, toolkit.BenchConfig{
    Queries:         []string{"data.policies.allow"},
    Paths:           []string{"examples/policies"},
    InputFile:       "examples/data/input.json",
    MaxWorkers:      4,
    TimeoutPerQuery: 30 * time.Second,
    WarmupRuns:      5,
    OnQueryComplete: func(q string, r toolkit.BenchmarkResult) {
        fmt.Printf(" Benchmark: %s | Mean: %.2f µs | Iter: %d\n", q, r.Stats.MeanNs/1000, r.Stats.Iterations)
    },
})

```

### 7.  Benchmark Multiple Queries and Print Summary

Invoke `opa bench` to measure performance of specific rules.

```go
multiResults, _ := tk.Bench(ctx, toolkit.BenchConfig{
    Queries:         []string{"data.policies.allow == true", "data.policies.deny == false"},
    Paths:           []string{"examples/policies"},
    InputFile:       "examples/data/input.json",
    MaxWorkers:      4,
    TimeoutPerQuery: 30 * time.Second,
    WarmupRuns:      5,
})

fmt.Println(tk.BenchSummary(multiResults, "text"))

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

