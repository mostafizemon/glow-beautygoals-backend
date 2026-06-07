package config

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	MongoURI     string
	DBName       string
	JWTSecret    string
	CloudinaryURL string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "beauty_db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super_secret_key_for_admin"
	}

	cloudinaryURL := os.Getenv("CLOUDINARY_URL")

	return &Config{
		Port:          port,
		MongoURI:      mongoURI,
		DBName:        dbName,
		JWTSecret:     jwtSecret,
		CloudinaryURL: cloudinaryURL,
	}
}

func SetupCORS(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // As specified: "allow all origins (no restriction needed)"
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))
}
