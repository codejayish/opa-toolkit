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
	NsPerOp     int64
	BytesPerOp  int64
	AllocsPerOp int64
	RawOutput   string
}

var (
	nsOpRegex     = regexp.MustCompile(`ns/op\s+\|\s+(\d+)`)
	bytesOpRegex  = regexp.MustCompile(`B/op\s+\|\s+(\d+)`)
	allocsOpRegex = regexp.MustCompile(`allocs/op\s+\|\s+(\d+)`)
)

// ParseBenchOutput extracts key metrics from a single benchmark output.
func ParseBenchOutput(query, output string) BenchSummary {
	ns := extractInt(nsOpRegex, output)
	b := extractInt(bytesOpRegex, output)
	a := extractInt(allocsOpRegex, output)

	return BenchSummary{
		Query:       query,
		NsPerOp:     ns,
		BytesPerOp:  b,
		AllocsPerOp: a,
		RawOutput:   output,
	}
}

func extractInt(re *regexp.Regexp, text string) int64 {
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return -1
	}
	val, _ := strconv.ParseInt(matches[1], 10, 64)
	return val
}

// GenerateSummary parses and compares all benchmark outputs.
func GenerateSummary(results map[string]string) string {
	var summaries []BenchSummary
	for query, output := range results {
		summaries = append(summaries, ParseBenchOutput(query, output))
	}

	// Sort by NsPerOp ascending (fastest to slowest)
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].NsPerOp < summaries[j].NsPerOp
	})

	sb := strings.Builder{}
	sb.WriteString("📊 OPA Benchmark Summary\n")
	sb.WriteString("---------------------------------------------------\n")
	for _, s := range summaries {
		sb.WriteString(fmt.Sprintf(
			"🔹 %s\n   ns/op: %d | B/op: %d | allocs/op: %d\n\n",
			s.Query, s.NsPerOp, s.BytesPerOp, s.AllocsPerOp,
		))
	}

	if len(summaries) > 0 {
		sb.WriteString(fmt.Sprintf("✅ Fastest: %s (%d ns/op)\n", summaries[0].Query, summaries[0].NsPerOp))
		sb.WriteString(fmt.Sprintf("🐢 Slowest: %s (%d ns/op)\n", summaries[len(summaries)-1].Query, summaries[len(summaries)-1].NsPerOp))
	}

	return sb.String()
}
