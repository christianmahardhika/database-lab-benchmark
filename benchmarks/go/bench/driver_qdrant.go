package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// QdrantDriver benchmarks Qdrant via REST API.
type QdrantDriver struct {
	url    string
	client *http.Client
	cfg    *Config
	seq    int64
}

func (d *QdrantDriver) Name() string { return "qdrant" }

func (d *QdrantDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	d.url = cfg.QdrantURL
	d.client = &http.Client{}

	// Delete collection if exists
	req, _ := http.NewRequest("DELETE", d.url+"/collections/bench_kv", nil)
	resp, _ := d.client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	// Create collection with vector config
	body := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     benchVectorDim,
			"distance": "Cosine",
		},
		"optimizers_config": map[string]interface{}{
			"indexing_threshold": 20000,
		},
		"hnsw_config": map[string]interface{}{
			"m":            16,
			"ef_construct": 128,
		},
	}
	return d.putJSON("/collections/bench_kv", body)
}

func (d *QdrantDriver) Cleanup() error {
	req, _ := http.NewRequest("DELETE", d.url+"/collections/bench_kv", nil)
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (d *QdrantDriver) Close() error {
	return nil
}

func (d *QdrantDriver) Write(key string, value []byte) error {
	d.seq++
	vec := deterministicVector(key, benchVectorDim)

	body := map[string]interface{}{
		"points": []map[string]interface{}{
			{
				"id":      d.seq,
				"vector":  vec,
				"payload": map[string]interface{}{"k": key, "v": string(value)},
			},
		},
	}
	return d.putJSON("/collections/bench_kv/points", body)
}

func (d *QdrantDriver) Read(key string) ([]byte, error) {
	vec := deterministicVector(key, benchVectorDim)

	body := map[string]interface{}{
		"vector": vec,
		"limit":  1,
		"params": map[string]interface{}{
			"hnsw_ef": 64,
		},
		"with_payload": true,
	}

	data, _ := json.Marshal(body)
	resp, err := d.client.Post(d.url+"/collections/bench_kv/points/search", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Result) > 0 {
		if v, ok := result.Result[0].Payload["v"].(string); ok {
			return []byte(v), nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (d *QdrantDriver) putJSON(path string, body interface{}) error {
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("PUT", d.url+path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("qdrant error: %s", resp.Status)
	}
	return nil
}
