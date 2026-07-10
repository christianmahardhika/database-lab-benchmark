package bench

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDriver struct {
	client *mongo.Client
	coll   *mongo.Collection
	cfg    *Config
}

func (d *MongoDriver) Name() string { return "mongodb" }

func (d *MongoDriver) Setup(cfg *Config) error {
	d.cfg = cfg
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	d.client = client
	d.coll = client.Database("benchmark").Collection("bench_kv")
	return nil
}

func (d *MongoDriver) Cleanup() error {
	return d.coll.Drop(context.Background())
}

func (d *MongoDriver) Close() error {
	if d.client != nil {
		return d.client.Disconnect(context.Background())
	}
	return nil
}

func (d *MongoDriver) Write(key string, value []byte) error {
	opts := options.Update().SetUpsert(true)
	_, err := d.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": key},
		bson.M{"$set": bson.M{"v": value}},
		opts,
	)
	return err
}

func (d *MongoDriver) Read(key string) ([]byte, error) {
	var doc struct {
		V []byte `bson:"v"`
	}
	err := d.coll.FindOne(context.Background(), bson.M{"_id": key}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return doc.V, nil
}
