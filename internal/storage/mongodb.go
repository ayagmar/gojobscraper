package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ayagmar/gojobscraper/internal/scraper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBStorage struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func NewMongoDBStorage(uri, dbName string) (*MongoDBStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB")

	database := client.Database(dbName)
	collection := database.Collection("jobs")

	// Create a unique index on platform_job_id
	_, err = collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "platform_job_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &MongoDBStorage{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

func (m *MongoDBStorage) SaveJobs(jobs []scraper.JobPosting) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var operations []mongo.WriteModel
	for _, job := range jobs {
		operation := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"platform_job_id": job.PlatformJobId}).
			SetUpdate(bson.M{"$set": job}).
			SetUpsert(true)
		operations = append(operations, operation)
	}

	opts := options.BulkWrite().SetOrdered(false)
	result, err := m.collection.BulkWrite(ctx, operations, opts)
	if err != nil {
		return fmt.Errorf("failed to save jobs: %w", err)
	}

	log.Printf("Upserted %d jobs, matched %d jobs", result.UpsertedCount, result.MatchedCount)
	return nil
}

func (m *MongoDBStorage) GetJobs() ([]scraper.JobPosting, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := m.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer cursor.Close(ctx)

	var jobs []scraper.JobPosting
	if err = cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}

	log.Printf("Retrieved %d jobs from MongoDB storage", len(jobs))
	return jobs, nil
}

func (m *MongoDBStorage) ClearJobs() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := m.collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to clear jobs: %w", err)
	}

	log.Printf("Cleared %d jobs from MongoDB storage", result.DeletedCount)
	return nil
}

func (m *MongoDBStorage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}
