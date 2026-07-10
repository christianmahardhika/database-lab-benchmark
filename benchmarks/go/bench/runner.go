package bench

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// DBDriver is the interface each database adapter must implement.
type DBDriver interface {
	Name() string
	Setup(cfg *Config) error
	Cleanup() error
	Close() error

	// Write inserts a key-value pair. Key format: "key_%08d"
	Write(key string, value []byte) error

	// Read fetches a value by key. Returns nil if not found.
	Read(key string) ([]byte, error)
}

// RunWrite benchmarks write throughput for a given database.
func RunWrite(dbName string, cfg *Config) (Result, error) {
	driver, err := GetDriver(dbName)
	if err != nil {
		return Result{}, err
	}

	if err := driver.Setup(cfg); err != nil {
		return Result{}, fmt.Errorf("setup: %w", err)
	}
	defer driver.Close()

	value := make([]byte, cfg.ValueSize)
	rand.Read(value)

	// Warmup
	for i := 0; i < cfg.WarmupRows; i++ {
		key := fmt.Sprintf("warmup_%08d", i)
		_ = driver.Write(key, value)
	}
	_ = driver.Cleanup()

	// Run multiple times, pick median
	var runs []Result
	for run := 0; run < cfg.Runs; run++ {
		r := execWrite(driver, cfg, value)
		runs = append(runs, r)
		_ = driver.Cleanup()
	}

	result := MedianResult(runs)
	result.Timestamp = time.Now().Format(time.RFC3339)
	return result, nil
}

func execWrite(driver DBDriver, cfg *Config, value []byte) Result {
	numOps := cfg.NumRows
	concurrency := cfg.Concurrency
	opsPerWorker := numOps / concurrency

	latencies := make([]float64, numOps)
	var idx atomic.Int64

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			base := workerID * opsPerWorker
			for i := 0; i < opsPerWorker; i++ {
				key := fmt.Sprintf("key_%08d", base+i)
				t0 := time.Now()
				_ = driver.Write(key, value)
				lat := float64(time.Since(t0).Microseconds()) / 1000.0 // ms
				pos := idx.Add(1) - 1
				if int(pos) < numOps {
					latencies[pos] = lat
				}
			}
		}(w)
	}
	wg.Wait()
	elapsed := time.Since(start)

	actualOps := int(idx.Load())
	p50, p95, p99, avg := ComputeLatencies(latencies[:actualOps])

	return Result{
		Database:    driver.Name(),
		Workload:    "write",
		TotalOps:    actualOps,
		ElapsedMs:   float64(elapsed.Milliseconds()),
		Throughput:  float64(actualOps) / elapsed.Seconds(),
		LatencyP50:  p50,
		LatencyP95:  p95,
		LatencyP99:  p99,
		LatencyAvg:  avg,
		Concurrency: cfg.Concurrency,
	}
}

// RunRead benchmarks read latency for a given database.
func RunRead(dbName string, cfg *Config) (Result, error) {
	driver, err := GetDriver(dbName)
	if err != nil {
		return Result{}, err
	}

	if err := driver.Setup(cfg); err != nil {
		return Result{}, fmt.Errorf("setup: %w", err)
	}
	defer driver.Close()

	value := make([]byte, cfg.ValueSize)
	rand.Read(value)

	// Pre-load data
	for i := 0; i < cfg.NumRows; i++ {
		key := fmt.Sprintf("key_%08d", i)
		if err := driver.Write(key, value); err != nil {
			return Result{}, fmt.Errorf("preload at %d: %w", i, err)
		}
	}

	// Run multiple times, pick median
	var runs []Result
	for run := 0; run < cfg.Runs; run++ {
		r := execRead(driver, cfg)
		runs = append(runs, r)
	}

	result := MedianResult(runs)
	result.Timestamp = time.Now().Format(time.RFC3339)
	_ = driver.Cleanup()
	return result, nil
}

func execRead(driver DBDriver, cfg *Config) Result {
	numOps := cfg.NumRows
	concurrency := cfg.Concurrency
	opsPerWorker := numOps / concurrency

	latencies := make([]float64, numOps)
	var idx atomic.Int64

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(workerID)))
			for i := 0; i < opsPerWorker; i++ {
				key := fmt.Sprintf("key_%08d", rng.Intn(numOps))
				t0 := time.Now()
				_, _ = driver.Read(key)
				lat := float64(time.Since(t0).Microseconds()) / 1000.0
				pos := idx.Add(1) - 1
				if int(pos) < numOps {
					latencies[pos] = lat
				}
			}
		}(w)
	}
	wg.Wait()
	elapsed := time.Since(start)

	actualOps := int(idx.Load())
	p50, p95, p99, avg := ComputeLatencies(latencies[:actualOps])

	return Result{
		Database:    driver.Name(),
		Workload:    "read",
		TotalOps:    actualOps,
		ElapsedMs:   float64(elapsed.Milliseconds()),
		Throughput:  float64(actualOps) / elapsed.Seconds(),
		LatencyP50:  p50,
		LatencyP95:  p95,
		LatencyP99:  p99,
		LatencyAvg:  avg,
		Concurrency: cfg.Concurrency,
	}
}

// RunMixed benchmarks a mixed workload (80% read, 20% write).
func RunMixed(dbName string, cfg *Config) (Result, error) {
	driver, err := GetDriver(dbName)
	if err != nil {
		return Result{}, err
	}

	if err := driver.Setup(cfg); err != nil {
		return Result{}, fmt.Errorf("setup: %w", err)
	}
	defer driver.Close()

	value := make([]byte, cfg.ValueSize)
	rand.Read(value)

	// Pre-load half the data
	preload := cfg.NumRows / 2
	for i := 0; i < preload; i++ {
		key := fmt.Sprintf("key_%08d", i)
		_ = driver.Write(key, value)
	}

	var runs []Result
	for run := 0; run < cfg.Runs; run++ {
		r := execMixed(driver, cfg, value, preload)
		runs = append(runs, r)
	}

	result := MedianResult(runs)
	result.Timestamp = time.Now().Format(time.RFC3339)
	_ = driver.Cleanup()
	return result, nil
}

func execMixed(driver DBDriver, cfg *Config, value []byte, preloaded int) Result {
	numOps := cfg.NumRows
	concurrency := cfg.Concurrency
	opsPerWorker := numOps / concurrency

	latencies := make([]float64, numOps)
	var idx atomic.Int64
	var writeCounter atomic.Int64

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(workerID * 1000)))
			for i := 0; i < opsPerWorker; i++ {
				t0 := time.Now()
				if rng.Float64() < 0.8 {
					// Read
					key := fmt.Sprintf("key_%08d", rng.Intn(preloaded))
					_, _ = driver.Read(key)
				} else {
					// Write
					wID := writeCounter.Add(1)
					key := fmt.Sprintf("mixed_%08d", wID)
					_ = driver.Write(key, value)
				}
				lat := float64(time.Since(t0).Microseconds()) / 1000.0
				pos := idx.Add(1) - 1
				if int(pos) < numOps {
					latencies[pos] = lat
				}
			}
		}(w)
	}
	wg.Wait()
	elapsed := time.Since(start)

	actualOps := int(idx.Load())
	p50, p95, p99, avg := ComputeLatencies(latencies[:actualOps])

	return Result{
		Database:    driver.Name(),
		Workload:    "mixed_80r_20w",
		TotalOps:    actualOps,
		ElapsedMs:   float64(elapsed.Milliseconds()),
		Throughput:  float64(actualOps) / elapsed.Seconds(),
		LatencyP50:  p50,
		LatencyP95:  p95,
		LatencyP99:  p99,
		LatencyAvg:  avg,
		Concurrency: cfg.Concurrency,
	}
}
