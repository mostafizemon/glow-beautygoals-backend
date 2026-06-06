package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Slug      string             `bson:"slug" json:"slug"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	Role         string             `bson:"role" json:"role"` // 'admin' or 'customer'
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type SEO struct {
	MetaTitle       string `bson:"meta_title" json:"meta_title"`
	MetaDescription string `bson:"meta_description" json:"meta_description"`
	OpenGraphImage  string `bson:"open_graph_image" json:"open_graph_image"`
}

type Product struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Slug           string             `bson:"slug" json:"slug"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description" json:"description"`
	Category       string             `bson:"category" json:"category"`
	Price          float64            `bson:"price" json:"price"`
	OfferPrice     float64            `bson:"offer_price" json:"offer_price"`
	Stock          int                `bson:"stock" json:"stock"`
	Images         []string           `bson:"images" json:"images"`
	SEO            SEO                `bson:"seo" json:"seo"`
	IsFeatured     bool               `bson:"is_featured" json:"is_featured"`
	IsActive       bool               `bson:"is_active" json:"is_active"`
	IsFreeDelivery bool               `bson:"is_free_delivery" json:"is_free_delivery"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type Customer struct {
	Name      string `bson:"name" json:"name"`
	Phone     string `bson:"phone" json:"phone"`
	Address   string `bson:"address" json:"address"`
	IPAddress string `bson:"ip_address" json:"ip_address"`
	UserAgent string `bson:"user_agent" json:"user_agent"`
}

type OrderItem struct {
	ProductID       primitive.ObjectID `bson:"product_id" json:"product_id"`
	Quantity        int                `bson:"quantity" json:"quantity"`
	PriceAtPurchase float64            `bson:"price_at_purchase" json:"price_at_purchase"`
}

type TrackingEvents struct {
	Fbp     string `bson:"fbp" json:"fbp"`
	Fbc     string `bson:"fbc" json:"fbc"`
	Ttp     string `bson:"ttp" json:"ttp"`
	EventID string `bson:"event_id" json:"event_id"`
}

type Order struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderNumber    string             `bson:"order_number" json:"order_number"`
	Customer       Customer           `bson:"customer" json:"customer"`
	Items          []OrderItem        `bson:"items" json:"items"`
	TotalAmount    float64            `bson:"total_amount" json:"total_amount"`
	Status         string             `bson:"status" json:"status"` // pending, confirmed, shipped, delivered, cancelled
	TrackingEvents TrackingEvents     `bson:"tracking_events" json:"tracking_events"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type PixelConfig struct {
	PixelID       string `bson:"pixel_id" json:"pixel_id"`
	AccessToken   string `bson:"access_token" json:"access_token"`
	TestEventCode string `bson:"test_event_code" json:"test_event_code"`
	IsActive      bool   `bson:"is_active" json:"is_active"`
}

type SiteConfig struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConfigKey string             `bson:"config_key" json:"config_key"` // e.g., 'tracking_pixels'
	Meta      PixelConfig        `bson:"meta" json:"meta"`
	TikTok    PixelConfig        `bson:"tiktok" json:"tiktok"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
