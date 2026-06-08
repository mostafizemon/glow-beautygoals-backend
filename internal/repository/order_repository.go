package repository

import (
	"context"
	"time"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetAll(ctx context.Context, filter bson.M) ([]model.Order, error)
	GetByID(ctx context.Context, id string) (*model.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	Delete(ctx context.Context, id string) error
}

type orderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) OrderRepository {
	return &orderRepository{
		collection: db.Collection("orders"),
	}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, order)
	return err
}

func (r *orderRepository) GetAll(ctx context.Context, filter bson.M) ([]model.Order, error) {
	if filter == nil {
		filter = bson.M{}
	}

	// Sort by created_at descending
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []model.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var order model.Order
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *orderRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
