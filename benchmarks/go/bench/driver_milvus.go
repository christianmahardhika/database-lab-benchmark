package bench

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	mrand "math/rand"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const milvusDim = 128 // vector dimension

type MilvusDriver struct {
	client client.Client
	cfg    *Config
	seq    int64
}

func (d *MilvusDriver) Name() string { return "milvus" }

func (d *MilvusDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	c, err := client.NewClient(context.Background(), client.Config{
		Address: cfg.MilvusAddr,
	})
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	d.client = c

	// Drop if exists for clean state
	has, _ := c.HasCollection(context.Background(), "bench_kv")
	if has {
		_ = c.DropCollection(context.Background(), "bench_kv")
	}

	// Create collection with vector field
	schema := &entity.Schema{
		CollectionName: "bench_kv",
		Fields: []*entity.Field{
			{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true, AutoID: true},
			{Name: "k", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "64"}},
			{Name: "vec", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": fmt.Sprintf("%d", milvusDim)}},
		},
	}
	err = c.CreateCollection(context.Background(), schema, entity.DefaultShardNumber)
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}

	// Create index for search
	idx, _ := entity.NewIndexIvfFlat(entity.L2, 128)
	err = c.CreateIndex(context.Background(), "bench_kv", "vec", idx, false)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	// Load collection
	return c.LoadCollection(context.Background(), "bench_kv", false)
}

func (d *MilvusDriver) Cleanup() error {
	has, _ := d.client.HasCollection(context.Background(), "bench_kv")
	if has {
		return d.client.DropCollection(context.Background(), "bench_kv")
	}
	return nil
}

func (d *MilvusDriver) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

func (d *MilvusDriver) Write(key string, value []byte) error {
	// Generate a deterministic vector from key
	vec := keyToVector(key)

	keys := []string{key}
	vecs := [][]float32{vec}

	keyCol := entity.NewColumnVarChar("k", keys)
	vecCol := entity.NewColumnFloatVector("vec", milvusDim, vecs)

	_, err := d.client.Insert(context.Background(), "bench_kv", "", keyCol, vecCol)
	return err
}

func (d *MilvusDriver) Read(key string) ([]byte, error) {
	// Vector search (Milvus is designed for similarity search, not KV lookup)
	vec := keyToVector(key)
	sp, _ := entity.NewIndexIvfFlatSearchParam(16)

	results, err := d.client.Search(context.Background(), "bench_kv", nil,
		"", []string{"k"}, []entity.Vector{entity.FloatVector(vec)},
		"vec", entity.L2, 1, sp)
	if err != nil {
		return nil, err
	}
	if len(results) > 0 && results[0].ResultCount > 0 {
		col := results[0].Fields.GetColumn("k")
		if col != nil {
			v, _ := col.GetAsString(0)
			return []byte(v), nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func keyToVector(key string) []float32 {
	// Deterministic vector from key string
	var seed int64
	b := []byte(key)
	if len(b) >= 8 {
		seed = int64(binary.LittleEndian.Uint64(b[:8]))
	} else {
		buf := make([]byte, 8)
		copy(buf, b)
		rand.Read(buf[len(b):])
		seed = int64(binary.LittleEndian.Uint64(buf))
	}

	rng := mrand.New(mrand.NewSource(seed))
	vec := make([]float32, milvusDim)
	for i := range vec {
		vec[i] = float32(rng.NormFloat64())
	}
	// Normalize
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}
	norm = math.Sqrt(norm)
	for i := range vec {
		vec[i] /= float32(norm)
	}
	return vec
}
