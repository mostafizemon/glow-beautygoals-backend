package service

import (
	"context"
	"errors"
	"time"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetAllUsers(ctx context.Context) ([]model.User, error)
	CreateUser(ctx context.Context, email, password, role string) error
	ChangePassword(ctx context.Context, userID, newPassword string) error
}

type userService struct {
	collection *mongo.Collection
}

func NewUserService(db *mongo.Database) UserService {
	return &userService{
		collection: db.Collection("users"),
	}
}

func (s *userService) GetAllUsers(ctx context.Context) ([]model.User, error) {
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *userService) CreateUser(ctx context.Context, email, password, role string) error {
	count, err := s.collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("a user with this email already exists")
	}

	if role != "admin" && role != "salesperson" {
		return errors.New("invalid role. Must be 'admin' or 'salesperson'")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := model.User{
		Email:        email,
		PasswordHash: string(hashed),
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = s.collection.InsertOne(ctx, user)
	return err
}

func (s *userService) ChangePassword(ctx context.Context, userID, newPassword string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{
			"password_hash": string(hashed),
			"updated_at":    time.Now(),
		}},
	)

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}
