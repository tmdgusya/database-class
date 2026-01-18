package performance

import (
	"fmt"
	"runtime"
	"time"

	"github.com/tmdgusya/database-class/pkg/vector"
)

// LatencyResult holds latency measurement results
type LatencyResult struct {
	Mean   time.Duration // Average latency
	Median time.Duration // Median latency
	P95    time.Duration // 95th percentile
	P99    time.Duration // 99th percentile
	Min    time.Duration // Minimum latency
	Max    time.Duration // Maximum latency
}

// Index interface (same as in recall.go but defined here for independence)
type Index interface {
	Add(v vector.Vector) error
	Search(query vector.Vector, k int) ([]SearchResult, error)
	Size() int
}

type SearchResult struct {
	Vector   vector.Vector
	Distance float64
	Index    int
}

// MeasureSearchLatency measures search latency for a set of queries
func MeasureSearchLatency(
	index Index,
	queries []vector.Vector,
	k int,
) (*LatencyResult, error) {
	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries provided")
	}

	latencies := make([]time.Duration, len(queries))

	for i, query := range queries {
		start := time.Now()
		_, err := index.Search(query, k)
		latencies[i] = time.Since(start)

		if err != nil {
			return nil, fmt.Errorf("search failed at query %d: %w", i, err)
		}
	}

	return computeLatencyStats(latencies), nil
}

// MeasureBuildLatency measures time to build an index
func MeasureBuildLatency(
	buildFunc func() error,
) (time.Duration, error) {
	start := time.Now()
	err := buildFunc()
	duration := time.Since(start)

	if err != nil {
		return 0, fmt.Errorf("build failed: %w", err)
	}

	return duration, nil
}

// ThroughputResult holds throughput measurement results
type ThroughputResult struct {
	QueriesPerSecond float64       // Queries processed per second
	TotalQueries     int           // Total queries executed
	Duration         time.Duration // Total measurement duration
	AvgLatency       time.Duration // Average query latency
}

// MeasureThroughput measures query throughput over a duration
func MeasureThroughput(
	index Index,
	queries []vector.Vector,
	k int,
	duration time.Duration,
) (*ThroughputResult, error) {
	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries provided")
	}

	start := time.Now()
	queryCount := 0
	queryIdx := 0

	for time.Since(start) < duration {
		query := queries[queryIdx%len(queries)]
		_, err := index.Search(query, k)
		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}

		queryCount++
		queryIdx++
	}

	elapsed := time.Since(start)
	qps := float64(queryCount) / elapsed.Seconds()
	avgLatency := elapsed / time.Duration(queryCount)

	return &ThroughputResult{
		QueriesPerSecond: qps,
		TotalQueries:     queryCount,
		Duration:         elapsed,
		AvgLatency:       avgLatency,
	}, nil
}

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	AllocBytes      uint64 // Bytes allocated and still in use
	TotalAllocBytes uint64 // Cumulative bytes allocated
	SysBytes        uint64 // Bytes obtained from system
	NumGC           uint32 // Number of completed GC cycles
}

// MeasureMemoryUsage captures current memory statistics
func MeasureMemoryUsage() *MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryStats{
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		SysBytes:        m.Sys,
		NumGC:           m.NumGC,
	}
}

// MeasureIndexMemory estimates memory used by building an index
// Takes a snapshot before and after building
func MeasureIndexMemory(buildFunc func() error) (*MemoryStats, error) {
	// Force GC before measurement
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Let GC complete

	before := MeasureMemoryUsage()

	// Build index
	err := buildFunc()
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	// Force GC after build
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	after := MeasureMemoryUsage()

	// Calculate delta
	delta := &MemoryStats{
		AllocBytes:      after.AllocBytes - before.AllocBytes,
		TotalAllocBytes: after.TotalAllocBytes - before.TotalAllocBytes,
		SysBytes:        after.SysBytes - before.SysBytes,
		NumGC:           after.NumGC - before.NumGC,
	}

	return delta, nil
}

// computeLatencyStats computes statistics from latency samples
func computeLatencyStats(latencies []time.Duration) *LatencyResult {
	if len(latencies) == 0 {
		return &LatencyResult{}
	}

	// Sort for percentiles
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sortDurations(sorted)

	// Calculate statistics
	result := &LatencyResult{
		Min: sorted[0],
		Max: sorted[len(sorted)-1],
	}

	// Mean
	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}
	result.Mean = sum / time.Duration(len(latencies))

	// Median
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		result.Median = (sorted[mid-1] + sorted[mid]) / 2
	} else {
		result.Median = sorted[mid]
	}

	// Percentiles
	p95Idx := int(float64(len(sorted)) * 0.95)
	p99Idx := int(float64(len(sorted)) * 0.99)

	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	result.P95 = sorted[p95Idx]
	result.P99 = sorted[p99Idx]

	return result
}

// sortDurations sorts durations in place (simple insertion sort for small arrays)
func sortDurations(durations []time.Duration) {
	for i := 1; i < len(durations); i++ {
		key := durations[i]
		j := i - 1
		for j >= 0 && durations[j] > key {
			durations[j+1] = durations[j]
			j--
		}
		durations[j+1] = key
	}
}

// FormatBytes formats bytes in human-readable form
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration in human-readable form
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs", float64(d.Nanoseconds())/1000.0)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1000000.0)
	}
	return fmt.Sprintf("%.2f s", d.Seconds())
}
