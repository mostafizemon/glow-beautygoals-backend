package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glow-and-beauty-goals/backend/internal/config"
	"github.com/glow-and-beauty-goals/backend/internal/handler"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
	"github.com/glow-and-beauty-goals/backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	db, err := repository.NewMongoDB(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Initialize Repositories
	productRepo := repository.NewProductRepository(db)

	// Initialize Services
	productService := service.NewProductService(productRepo)

	// Initialize Handlers
	productHandler := handler.NewProductHandler(productService)

	r := gin.Default()
	
	// Setup CORS
	config.SetupCORS(r)

	// Routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/products", productHandler.GetAllProducts)
		v1.POST("/products", productHandler.CreateProduct)
	}

	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
