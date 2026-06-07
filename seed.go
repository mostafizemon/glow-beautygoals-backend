package main

import (
	"context"
	"fmt"
	"log"

	"github.com/glow-and-beauty-goals/backend/internal/config"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
	"github.com/glow-and-beauty-goals/backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()
	db, err := repository.NewMongoDB(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	authService := service.NewAuthService(db, cfg.JWTSecret)
	
	email := "mostafizemon09@gmail.com"
	password := "Emon@548"
	
	err = authService.CreateInitialAdmin(context.Background(), email, password)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}
	
	fmt.Printf("Successfully created/verified admin: %s\n", email)
}
