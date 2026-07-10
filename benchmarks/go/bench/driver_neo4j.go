package bench

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jDriver struct {
	driver neo4j.DriverWithContext
	cfg    *Config
}

func (d *Neo4jDriver) Name() string { return "neo4j" }

func (d *Neo4jDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4jURI,
		neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""))
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	d.driver = driver

	// Verify connectivity
	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return fmt.Errorf("verify: %w", err)
	}

	// Create constraint for unique key
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err = session.Run(ctx,
		"CREATE CONSTRAINT IF NOT EXISTS FOR (n:KV) REQUIRE n.k IS UNIQUE", nil)
	if err != nil {
		// Ignore if constraint already exists
	}
	return nil
}

func (d *Neo4jDriver) Cleanup() error {
	ctx := context.Background()
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.Run(ctx, "MATCH (n:KV) DELETE n", nil)
	return err
}

func (d *Neo4jDriver) Close() error {
	if d.driver != nil {
		return d.driver.Close(context.Background())
	}
	return nil
}

func (d *Neo4jDriver) Write(key string, value []byte) error {
	ctx := context.Background()
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		"MERGE (n:KV {k: $key}) SET n.v = $value",
		map[string]interface{}{"key": key, "value": value})
	return err
}

func (d *Neo4jDriver) Read(key string) ([]byte, error) {
	ctx := context.Background()
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		"MATCH (n:KV {k: $key}) RETURN n.v AS v",
		map[string]interface{}{"key": key})
	if err != nil {
		return nil, err
	}

	if result.Next(ctx) {
		v, _ := result.Record().Get("v")
		if b, ok := v.([]byte); ok {
			return b, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
