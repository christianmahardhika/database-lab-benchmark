package bench

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Result holds benchmark measurements for a single database + workload.
type Result struct {
	Database    string  `json:"database"`
	Workload    string  `json:"workload"`
	TotalOps    int     `json:"total_ops"`
	ElapsedMs   float64 `json:"elapsed_ms"`
	Throughput  float64 `json:"throughput_ops_s"`
	LatencyP50  float64 `json:"latency_p50_ms"`
	LatencyP95  float64 `json:"latency_p95_ms"`
	LatencyP99  float64 `json:"latency_p99_ms"`
	LatencyAvg  float64 `json:"latency_avg_ms"`
	Concurrency int     `json:"concurrency"`
	Timestamp   string  `json:"timestamp"`
	Error       string  `json:"error,omitempty"`
}

// Summary returns a one-line human-readable summary.
func (r Result) Summary() string {
	return fmt.Sprintf("%.0f ops/s | p50=%.2fms p95=%.2fms p99=%.2fms",
		r.Throughput, r.LatencyP50, r.LatencyP95, r.LatencyP99)
}

// ComputeLatencies calculates percentiles from a slice of durations (in ms).
func ComputeLatencies(latencies []float64) (p50, p95, p99, avg float64) {
	n := len(latencies)
	if n == 0 {
		return
	}
	sort.Float64s(latencies)
	p50 = latencies[n*50/100]
	p95 = latencies[n*95/100]
	idx99 := n * 99 / 100
	if idx99 >= n {
		idx99 = n - 1
	}
	p99 = latencies[idx99]

	var total float64
	for _, l := range latencies {
		total += l
	}
	avg = total / float64(n)
	return
}

// MedianResult picks the median throughput from multiple runs.
func MedianResult(results []Result) Result {
	if len(results) == 0 {
		return Result{}
	}
	if len(results) == 1 {
		return results[0]
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Throughput < results[j].Throughput
	})
	return results[len(results)/2]
}

// PrintTable prints results in a formatted table.
func PrintTable(results []Result) {
	if len(results) == 0 {
		fmt.Println("  (no results)")
		return
	}
	fmt.Printf("\n  %-14s %12s %9s %9s %9s %9s\n",
		"Database", "Throughput", "p50(ms)", "p95(ms)", "p99(ms)", "Avg(ms)")
	fmt.Println("  ────────────── ──────────── ───────── ───────── ───────── ─────────")
	for _, r := range results {
		fmt.Printf("  %-14s %10.0f/s %9.2f %9.2f %9.2f %9.2f\n",
			r.Database, r.Throughput, r.LatencyP50, r.LatencyP95, r.LatencyP99, r.LatencyAvg)
	}
	fmt.Println()
}

// SaveJSON writes results to a timestamped JSON file.
func SaveJSON(results []Result, outputDir, prefix string) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "  Failed to create output dir: %v\n", err)
		return
	}

	filename := fmt.Sprintf("%s_%s.json", prefix, time.Now().Format("20060102_150405"))
	path := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Failed to marshal results: %v\n", err)
		return
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "  Failed to write results: %v\n", err)
		return
	}

	fmt.Printf("  📁 Results saved: %s\n", path)
}
