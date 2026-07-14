package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenSearchDriver uses raw HTTP client for OpenSearch compatibility.
// OpenSearch is API-compatible with Elasticsearch 7.x but diverged after AWS fork.
type OpenSearchDriver struct {
	client  *http.Client
	baseURL string
	cfg     *Config
}

func (d *OpenSearchDriver) Name() string { return "opensearch" }

func (d *OpenSearchDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	d.baseURL = strings.TrimSuffix(cfg.OpenSearchURL, "/")
	d.client = &http.Client{Timeout: 30 * time.Second}

	// Check connection
	resp, err := d.client.Get(d.baseURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	resp.Body.Close()

	// Create index with minimal settings for benchmark
	body := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"refresh_interval": "30s"
		},
		"mappings": {
			"properties": {
				"v": { "type": "binary" }
			}
		}
	}`

	req, _ := http.NewRequestWithContext(context.Background(), "PUT",
		d.baseURL+"/bench_kv", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = d.client.Do(req)
	if err != nil {
		return nil // Index might already exist
	}
	resp.Body.Close()
	return nil
}

func (d *OpenSearchDriver) Cleanup() error {
	req, _ := http.NewRequestWithContext(context.Background(), "DELETE",
		d.baseURL+"/bench_kv", nil)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (d *OpenSearchDriver) Close() error {
	return nil // HTTP client doesn't need explicit close
}

func (d *OpenSearchDriver) Write(key string, value []byte) error {
	doc := map[string]interface{}{"v": value}
	body, _ := json.Marshal(doc)

	req, _ := http.NewRequestWithContext(context.Background(), "PUT",
		fmt.Sprintf("%s/bench_kv/_doc/%s", d.baseURL, key),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return nil
}

func (d *OpenSearchDriver) Read(key string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET",
		fmt.Sprintf("%s/bench_kv/_doc/%s", d.baseURL, key), nil)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	source, ok := result["_source"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no _source")
	}
	v, _ := source["v"].(string)
	return []byte(v), nil
}
