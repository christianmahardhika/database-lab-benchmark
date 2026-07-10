package bench

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxDBDriver struct {
	client influxdb2.Client
	cfg    *Config
	seq    int64
}

func (d *InfluxDBDriver) Name() string { return "influxdb" }

func (d *InfluxDBDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	d.client = influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)

	// Verify health
	health, err := d.client.Health(context.Background())
	if err != nil {
		return fmt.Errorf("health check: %w", err)
	}
	if health.Status != "pass" {
		return fmt.Errorf("influxdb unhealthy: %s", health.Status)
	}
	return nil
}

func (d *InfluxDBDriver) Cleanup() error {
	deleteAPI := d.client.DeleteAPI()
	// Delete all data in the bucket
	start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	stop := time.Now().Add(24 * time.Hour)
	return deleteAPI.DeleteWithName(context.Background(), d.cfg.InfluxDBOrg, d.cfg.InfluxDBBucket, start, stop, "")
}

func (d *InfluxDBDriver) Close() error {
	if d.client != nil {
		d.client.Close()
	}
	return nil
}

func (d *InfluxDBDriver) Write(key string, value []byte) error {
	writeAPI := d.client.WriteAPIBlocking(d.cfg.InfluxDBOrg, d.cfg.InfluxDBBucket)

	d.seq++
	p := influxdb2.NewPoint("bench_kv",
		map[string]string{"k": key},
		map[string]interface{}{"v": string(value)},
		time.Now().Add(time.Duration(d.seq)*time.Microsecond),
	)
	return writeAPI.WritePoint(context.Background(), p)
}

func (d *InfluxDBDriver) Read(key string) ([]byte, error) {
	queryAPI := d.client.QueryAPI(d.cfg.InfluxDBOrg)

	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -1h)
		|> filter(fn: (r) => r._measurement == "bench_kv" and r.k == "%s")
		|> last()
	`, d.cfg.InfluxDBBucket, key)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	if result.Next() {
		v := result.Record().Value()
		if s, ok := v.(string); ok {
			return []byte(s), nil
		}
	}
	return nil, fmt.Errorf("not found")
}
