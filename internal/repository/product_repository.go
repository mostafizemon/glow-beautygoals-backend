package repository

import (
	"context"
	"time"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductRepository interface {
	GetAll(ctx context.Context, filter bson.M) ([]model.Product, error)
	GetByID(ctx context.Context, id string) (*model.Product, error)
	GetBySlug(ctx context.Context, slug string) (*model.Product, error)
	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, id string, product *model.Product) error
	Delete(ctx context.Context, id string) error
}

type productRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) ProductRepository {
	return &productRepository{
		collection: db.Collection("products"),
	}
}

func (r *productRepository) GetAll(ctx context.Context, filter bson.M) ([]model.Product, error) {
	if filter == nil {
		filter = bson.M{}
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []model.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	product.ID = primitive.NewObjectID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, product)
	return err
}

func (r *productRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var product model.Product
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetBySlug(ctx context.Context, slug string) (*model.Product, error) {
	var product model.Product
	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *productRepository) Update(ctx context.Context, id string, product *model.Product) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	product.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":        product.Name,
			"slug":        product.Slug,
			"description": product.Description,
			"price":       product.Price,
			"stock":       product.Stock,
			"category":    product.Category,
			"is_featured": product.IsFeatured,
			"is_active":   product.IsActive,
			"images":      product.Images,
			"updated_at":  product.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}
