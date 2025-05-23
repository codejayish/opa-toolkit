package bench

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type BenchSummary struct {
	Query       string
	NsPerOp     int64   // Nanoseconds per operation
	BytesPerOp  int64   // Bytes allocated per operation
	AllocsPerOp int64   // Allocations per operation
	MemMB       float64 // Peak memory usage in MB
	OpsPerSec   int64   // Throughput
	PercentDiff float64 // Percent of total runtime
	RawOutput   string
}

var (
	nsOpRegex     = regexp.MustCompile(`ns/op\s+\|\s+(\d+)`)
	bytesOpRegex  = regexp.MustCompile(`B/op\s+\|\s+(\d+)`)
	allocsOpRegex = regexp.MustCompile(`allocs/op\s+\|\s+(\d+)`)
	memMBRegex    = regexp.MustCompile(`(\d+\.\d+)\s+MB`)
	opsSecRegex   = regexp.MustCompile(`(\d+)\s+op/s`)
)

// ParseBenchOutput extracts metrics from a single benchmark output string.
func ParseBenchOutput(query, output string) BenchSummary {
	return BenchSummary{
		Query:       query,
		NsPerOp:     extractInt(nsOpRegex, output),
		BytesPerOp:  extractInt(bytesOpRegex, output),
		AllocsPerOp: extractInt(allocsOpRegex, output),
		MemMB:       extractFloat(memMBRegex, output),
		OpsPerSec:   extractInt(opsSecRegex, output),
		RawOutput:   output,
		PercentDiff: 0, // will be calculated later
	}
}

// GenerateSummary creates a readable summary report of benchmark results.
func GenerateSummary(results map[string]string, format string) string {
	var summaries []BenchSummary
	for query, output := range results {
		s := ParseBenchOutput(query, output)
		if s.NsPerOp > 0 {
			summaries = append(summaries, s)
		}
	}

	if len(summaries) == 0 {
		return "⚠️ No valid benchmark data to summarize.\n"
	}

	// Calculate total and relative percentages
	var totalNs int64
	for _, s := range summaries {
		totalNs += s.NsPerOp
	}
	for i := range summaries {
		summaries[i].PercentDiff = (float64(summaries[i].NsPerOp) / float64(totalNs)) * 100
	}

	// Sort: fastest (lowest NsPerOp) first
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

// textSummary returns a clean CLI-friendly summary.
func textSummary(summaries []BenchSummary) string {
	sb := strings.Builder{}
	sb.WriteString("📊 OPA Benchmark Summary\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	for _, s := range summaries {
		sb.WriteString(fmt.Sprintf(
			"🔹 %s\n   %.2f µs/op (%.1f%%) | %d B/op | %d allocs/op | %.1f MB\n\n",
			s.Query,
			float64(s.NsPerOp)/1000,
			s.PercentDiff,
			s.BytesPerOp,
			s.AllocsPerOp,
			s.MemMB,
		))
	}

	if len(summaries) > 1 {
		fastest, slowest := summaries[0], summaries[len(summaries)-1]
		sb.WriteString(fmt.Sprintf("✅ Fastest: %s (%.2f µs/op)\n", fastest.Query, float64(fastest.NsPerOp)/1000))
		sb.WriteString(fmt.Sprintf("🐢 Slowest: %s (%.2f µs/op)\n", slowest.Query, float64(slowest.NsPerOp)/1000))
		sb.WriteString(fmt.Sprintf("📈 Performance spread: %.1fx\n",
			float64(slowest.NsPerOp)/float64(fastest.NsPerOp)))
	}

	return sb.String()
}

// markdownSummary returns the summary in GitHub-flavored markdown table.
func markdownSummary(summaries []BenchSummary) string {
	sb := strings.Builder{}
	sb.WriteString("## OPA Benchmark Summary\n\n")
	sb.WriteString("| Query | µs/op | % Total | B/op | allocs/op | Mem (MB) |\n")
	sb.WriteString("|-------|--------|----------|------|------------|-----------|\n")

	for _, s := range summaries {
		sb.WriteString(fmt.Sprintf(
			"| %s | %.2f | %.1f%% | %d | %d | %.2f |\n",
			s.Query,
			float64(s.NsPerOp)/1000,
			s.PercentDiff,
			s.BytesPerOp,
			s.AllocsPerOp,
			s.MemMB,
		))
	}

	return sb.String()
}

// extractInt extracts a single int64 from regex match in string.
func extractInt(r *regexp.Regexp, s string) int64 {
	matches := r.FindStringSubmatch(s)
	if len(matches) >= 2 {
		if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			return val
		}
	}
	return 0
}

// extractFloat extracts a single float64 from regex match in string.
func extractFloat(r *regexp.Regexp, s string) float64 {
	matches := r.FindStringSubmatch(s)
	if len(matches) >= 2 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return val
		}
	}
	return 0
}
