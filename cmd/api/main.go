package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glow-and-beauty-goals/backend/internal/config"
	"github.com/glow-and-beauty-goals/backend/internal/handler"
	"github.com/glow-and-beauty-goals/backend/internal/middleware"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
	"github.com/glow-and-beauty-goals/backend/internal/service"
	"github.com/glow-and-beauty-goals/backend/pkg/cloud"
)

func main() {
	cfg := config.LoadConfig()

	db, err := repository.NewMongoDB(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Initialize Cloudinary
	var cloudinaryService cloud.CloudinaryService
	if cfg.CloudinaryURL != "" {
		var err error
		cloudinaryService, err = cloud.NewCloudinaryService(cfg.CloudinaryURL)
		if err != nil {
			log.Printf("Failed to initialize Cloudinary: %v", err)
		}
	}

	// Initialize Repositories
	productRepo := repository.NewProductRepository(db)
	configRepo := repository.NewConfigRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Initialize Services
	productService := service.NewProductService(productRepo)
	authService := service.NewAuthService(db, cfg.JWTSecret)
	configService := service.NewConfigService(configRepo)
	trackingService := service.NewTrackingService(configRepo)
	userService := service.NewUserService(db)
	orderService := service.NewOrderService(orderRepo)

	// Create initial admin if doesn't exist
	if err := authService.CreateInitialAdmin(context.Background(), "mostafizemon09@gmail.com", "Emon@548"); err != nil {
		log.Fatalf("Failed to seed initial admin: %v", err)
	}

	// Initialize Handlers
	productHandler := handler.NewProductHandler(productService, cloudinaryService)
	authHandler := handler.NewAuthHandler(authService)
	configHandler := handler.NewConfigHandler(configService)
	trackingHandler := handler.NewTrackingHandler(trackingService)
	userHandler := handler.NewUserHandler(userService)
	orderHandler := handler.NewOrderHandler(orderService)

	r := gin.Default()
	
	// Setup CORS
	config.SetupCORS(r)

	// Routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/admin/login", authHandler.Login)
		v1.GET("/products", productHandler.GetAllProducts)
		v1.GET("/products/:id", productHandler.GetProduct)
		v1.GET("/products/slug/:slug", productHandler.GetProductBySlug)
		v1.GET("/config/pixels", configHandler.GetPublicPixels)
		v1.POST("/events/track", trackingHandler.TrackEvent)
		v1.POST("/orders", orderHandler.CreateOrder) // Publicly accessible to place orders

		// Protected Admin routes (we will add JWT middleware later)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			// All logged-in staff can change their password
			admin.PUT("/users/password", userHandler.ChangePassword)

			// Admin-only routes
			superAdmin := admin.Group("/")
			superAdmin.Use(middleware.RoleRequired("admin"))
			{
				superAdmin.POST("/products", productHandler.CreateProduct)
				superAdmin.PUT("/products/:id", productHandler.UpdateProduct)
				superAdmin.DELETE("/products/:id", productHandler.DeleteProduct)
				superAdmin.POST("/products/upload", productHandler.UploadImage)
				
				superAdmin.GET("/config/tracking", configHandler.GetTrackingConfig)
				superAdmin.POST("/config/tracking", configHandler.UpdateTrackingConfig)

				superAdmin.GET("/users", userHandler.GetAllUsers)
				superAdmin.POST("/users", userHandler.CreateUser)
				
				superAdmin.GET("/orders", orderHandler.GetAllOrders)
				superAdmin.GET("/orders/:id", orderHandler.GetOrder)
				superAdmin.PUT("/orders/:id/status", orderHandler.UpdateOrderStatus)
			}
		}
	}

	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
