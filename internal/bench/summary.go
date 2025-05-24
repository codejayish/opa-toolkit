package bench

import (
	"fmt"
	"sort"
	"strings"
)

// BenchSummary represents a normalized benchmark result for reporting.
type BenchSummary struct {
	Query       string
	NsPerOp     float64
	MemMB       float64
	OpsPerSec   float64
	Iterations  int
	PercentDiff float64
}

// GenerateSummary produces a formatted summary string from parsed results.
func GenerateSummary(results map[string]BenchmarkResult, format string) string {
	var summaries []BenchSummary
	var totalNs float64

	// Collect summaries from parsed BenchmarkResult data
	for query, result := range results {
		if result.Stats.Iterations == 0 {
			continue // skip empty
		}
		nsPerOp := result.Stats.MeanNs
		opsPerSec := float64(result.Stats.Iterations) / result.Duration.Seconds()
		memMB := float64(result.Stats.MemoryKB) / 1024

		summary := BenchSummary{
			Query:      query,
			NsPerOp:    nsPerOp,
			MemMB:      memMB,
			OpsPerSec:  opsPerSec,
			Iterations: result.Stats.Iterations,
		}
		summaries = append(summaries, summary)
		totalNs += nsPerOp
	}

	if len(summaries) == 0 {
		return "⚠️ No valid benchmark data to summarize.\n"
	}

	// Compute relative performance
	for i := range summaries {
		summaries[i].PercentDiff = (summaries[i].NsPerOp / totalNs) * 100
	}

	// Sort by fastest (lowest NsPerOp)
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].NsPerOp < summaries[j].NsPerOp
	})

	switch format {
	case "markdown":
		return markdownSummary(summaries)
	default:
		return textSummary(summaries)
	}
}

// textSummary prints results for terminal output.
func textSummary(summaries []BenchSummary) string {
	var b strings.Builder
	b.WriteString("📊 OPA Benchmark Summary\n")
	b.WriteString(strings.Repeat("─", 60) + "\n")

	for _, s := range summaries {
		b.WriteString(fmt.Sprintf(
			"🔹 %s\n   %.2f µs/op (%.1f%%) | %.2f MB | %.0f ops/sec | %d iterations\n\n",
			s.Query, s.NsPerOp/1000, s.PercentDiff, s.MemMB, s.OpsPerSec, s.Iterations,
		))
	}

	if len(summaries) > 1 {
		fastest := summaries[0]
		slowest := summaries[len(summaries)-1]
		spread := slowest.NsPerOp / fastest.NsPerOp

		b.WriteString(fmt.Sprintf("✅ Fastest: %s (%.2f µs/op)\n", fastest.Query, fastest.NsPerOp/1000))
		b.WriteString(fmt.Sprintf("🐢 Slowest: %s (%.2f µs/op)\n", slowest.Query, slowest.NsPerOp/1000))
		b.WriteString(fmt.Sprintf("📈 Performance spread: %.1fx\n", spread))
	}

	return b.String()
}

// markdownSummary prints results in Markdown table format.
func markdownSummary(summaries []BenchSummary) string {
	var b strings.Builder
	b.WriteString("## OPA Benchmark Summary\n\n")
	b.WriteString("| Query | µs/op | % Total | Mem (MB) | Ops/sec | Iterations |\n")
	b.WriteString("|-------|--------|----------|-----------|----------|-------------|\n")

	for _, s := range summaries {
		b.WriteString(fmt.Sprintf(
			"| %s | %.2f | %.1f%% | %.2f | %.0f | %d |\n",
			s.Query, s.NsPerOp/1000, s.PercentDiff, s.MemMB, s.OpsPerSec, s.Iterations,
		))
	}

	return b.String()
}
