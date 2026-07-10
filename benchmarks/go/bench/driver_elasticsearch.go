package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchDriver struct {
	client *elasticsearch.Client
	cfg    *Config
}

func (d *ElasticsearchDriver) Name() string { return "elasticsearch" }

func (d *ElasticsearchDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.ElasticsearchURL},
	})
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	// Check connection
	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	res.Body.Close()

	d.client = client

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

	res, err = client.Indices.Create("bench_kv",
		client.Indices.Create.WithBody(strings.NewReader(body)),
		client.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		return nil // Index might already exist
	}
	res.Body.Close()
	return nil
}

func (d *ElasticsearchDriver) Cleanup() error {
	res, err := d.client.Indices.Delete([]string{"bench_kv"},
		d.client.Indices.Delete.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

func (d *ElasticsearchDriver) Close() error {
	return nil // elasticsearch-go client doesn't need explicit close
}

func (d *ElasticsearchDriver) Write(key string, value []byte) error {
	doc := map[string]interface{}{"v": value}
	body, _ := json.Marshal(doc)

	res, err := d.client.Index("bench_kv",
		bytes.NewReader(body),
		d.client.Index.WithDocumentID(key),
		d.client.Index.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	return nil
}

func (d *ElasticsearchDriver) Read(key string) ([]byte, error) {
	res, err := d.client.Get("bench_kv", key,
		d.client.Get.WithContext(context.Background()),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("get: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	source, ok := result["_source"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no _source")
	}
	v, _ := source["v"].(string)
	return []byte(v), nil
}
