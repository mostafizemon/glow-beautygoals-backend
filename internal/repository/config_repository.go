package repository

import (
	"context"
	"time"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConfigRepository interface {
	GetSiteConfig(ctx context.Context, key string) (*model.SiteConfig, error)
	UpsertSiteConfig(ctx context.Context, config *model.SiteConfig) error
}

type configRepository struct {
	collection *mongo.Collection
}

func NewConfigRepository(db *mongo.Database) ConfigRepository {
	return &configRepository{
		collection: db.Collection("site_configs"),
	}
}

func (r *configRepository) GetSiteConfig(ctx context.Context, key string) (*model.SiteConfig, error) {
	var config model.SiteConfig
	err := r.collection.FindOne(ctx, bson.M{"config_key": key}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return an empty config if none exists yet
			return &model.SiteConfig{ConfigKey: key}, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *configRepository) UpsertSiteConfig(ctx context.Context, config *model.SiteConfig) error {
	config.UpdatedAt = time.Now()
	
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"config_key": config.ConfigKey}
	update := bson.M{"$set": config}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}
