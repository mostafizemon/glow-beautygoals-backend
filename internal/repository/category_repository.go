package repository

import (
	"context"
	"strings"
	"time"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryRepository interface {
	GetAll(ctx context.Context) ([]model.Category, error)
	GetByID(ctx context.Context, id string) (*model.Category, error)
	Create(ctx context.Context, category *model.Category) error
	Update(ctx context.Context, id string, category *model.Category) error
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, categoryIDs []string) error
}

type categoryRepository struct {
	collection *mongo.Collection
}

func NewCategoryRepository(db *mongo.Database) CategoryRepository {
	return &categoryRepository{
		collection: db.Collection("categories"),
	}
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]model.Category, error) {
	opts := options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []model.Category
	if err = cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*model.Category, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var category model.Category
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&category)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Create(ctx context.Context, category *model.Category) error {
	category.ID = primitive.NewObjectID()
	if category.Slug == "" {
		category.Slug = strings.ToLower(strings.ReplaceAll(category.Name, " ", "-"))
	}
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, category)
	return err
}

func (r *categoryRepository) Update(ctx context.Context, id string, category *model.Category) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	category.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":       category.Name,
			"slug":       category.Slug,
			"sort_order": category.SortOrder,
			"updated_at": category.UpdatedAt,
		},
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *categoryRepository) Reorder(ctx context.Context, categoryIDs []string) error {
	var models []mongo.WriteModel
	for i, id := range categoryIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue // skip invalid ids
		}
		update := bson.M{"$set": bson.M{"sort_order": i}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": objID}).SetUpdate(update))
	}
	if len(models) == 0 {
		return nil
	}
	_, err := r.collection.BulkWrite(ctx, models)
	return err
}
