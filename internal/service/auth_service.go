package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	CreateInitialAdmin(ctx context.Context, email, password string) error
}

type authService struct {
	collection *mongo.Collection
	jwtSecret  string
}

func NewAuthService(db *mongo.Database, jwtSecret string) AuthService {
	return &authService{
		collection: db.Collection("users"),
		jwtSecret:  jwtSecret,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	var user model.User
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authService) CreateInitialAdmin(ctx context.Context, email, password string) error {
	count, err := s.collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if count > 0 {
		// Admin already exists, force update the password hash
		_, err := s.collection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"password_hash": string(hashed)}})
		return err
	}

	admin := model.User{
		Email:        email,
		PasswordHash: string(hashed),
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = s.collection.InsertOne(ctx, admin)
	return err
}
