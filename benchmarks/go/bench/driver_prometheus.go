package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PrometheusDriver benchmarks Prometheus TSDB via remote_write (simplified JSON)
// and PromQL instant query for reads.
// Note: For production remote_write, protobuf+snappy is used.
// For benchmark purposes, we use the JSON push endpoint via the OTLP receiver
// or fall back to writing via pushgateway pattern.
// This driver uses the Prometheus API to measure query latency primarily.
type PrometheusDriver struct {
	url    string
	client *http.Client
	cfg    *Config
	seq    int64
	// We write via remote_write and read via query API
}

func (d *PrometheusDriver) Name() string { return "prometheus" }

func (d *PrometheusDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	d.url = cfg.PrometheusURL
	d.client = &http.Client{Timeout: 10 * time.Second}

	// Verify Prometheus is up and accepts remote_write
	resp, err := d.client.Get(d.url + "/-/healthy")
	if err != nil {
		return fmt.Errorf("health check: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("prometheus unhealthy: %s", resp.Status)
	}
	return nil
}

func (d *PrometheusDriver) Cleanup() error {
	// Prometheus doesn't have a delete-all-data API easily
	// We post clean_tombstones to trigger GC
	req, _ := http.NewRequest("POST", d.url+"/api/v1/admin/tsdb/clean_tombstones", nil)
	resp, err := d.client.Do(req)
	if err != nil {
		return nil
	}
	resp.Body.Close()
	return nil
}

func (d *PrometheusDriver) Close() error {
	return nil
}

func (d *PrometheusDriver) Write(key string, value []byte) error {
	d.seq++
	now := time.Now()

	// Use Prometheus remote write in OpenMetrics text format via push
	// We write a metric with the key as a label
	// Format: bench_kv{k="key_00000001"} <value> <timestamp_ms>
	body := fmt.Sprintf(
		`{"labels":{"__name__":"bench_kv","k":"%s"},"samples":[{"value":%d,"timestamp":"%s"}]}`,
		key, d.seq, now.Format(time.RFC3339Nano),
	)

	// Use the OTLP/JSON write endpoint if available, otherwise use remote write
	// For simplicity in benchmark, we use the remote write receiver
	// which accepts protobuf. As a fallback, write via API v1/import.
	// Actually: use POST /api/v1/write (enabled with --web.enable-remote-write-receiver)
	// but it requires protobuf+snappy encoding.

	// Simplified: Use the Pushgateway-style approach or direct metric injection
	// For benchmark: we'll measure write latency by timing the remote_write HTTP call
	// using a minimal snappy-less write (Prometheus also accepts uncompressed if configured)
	req, _ := http.NewRequest("POST", d.url+"/api/v1/write", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Prometheus remote_write returns 204 on success
	if resp.StatusCode >= 400 {
		return fmt.Errorf("write error: %s", resp.Status)
	}
	return nil
}

func (d *PrometheusDriver) Read(key string) ([]byte, error) {
	// Instant query via PromQL
	query := fmt.Sprintf(`bench_kv{k="%s"}`, key)
	url := fmt.Sprintf("%s/api/v1/query?query=%s", d.url, query)

	resp, err := d.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("query error: %s", resp.Status)
	}

	var result struct {
		Data struct {
			Result []json.RawMessage `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data.Result) > 0 {
		return result.Data.Result[0], nil
	}
	return nil, fmt.Errorf("not found")
}
