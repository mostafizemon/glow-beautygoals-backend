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
	GetContactConfig(ctx context.Context) (*model.ContactConfig, error)
	UpsertContactConfig(ctx context.Context, config *model.ContactConfig) error
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

func (r *configRepository) GetContactConfig(ctx context.Context) (*model.ContactConfig, error) {
	var config model.ContactConfig
	err := r.collection.FindOne(ctx, bson.M{"config_key": "contact_info"}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.ContactConfig{ConfigKey: "contact_info", IsActive: true}, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *configRepository) UpsertContactConfig(ctx context.Context, config *model.ContactConfig) error {
	config.ConfigKey = "contact_info"
	config.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"config_key": "contact_info"}
	update := bson.M{"$set": config}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}
