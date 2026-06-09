package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetAllOrders(ctx context.Context) ([]model.Order, error)
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status string) error
	UpdateOrderDetails(ctx context.Context, id string, customer model.Customer, totalAmount float64) error
	DeleteOrder(ctx context.Context, id string) error
}

type orderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	return &orderService{
		repo: repo,
	}
}

func generateOrderNumber() string {
	// Simple order number generator: ORD-YYYYMMDD-XXXX
	now := time.Now()
	dateStr := now.Format("20060102")
	randomNum := rand.Intn(9000) + 1000 // 4 digit random number
	return fmt.Sprintf("ORD-%s-%d", dateStr, randomNum)
}

func (s *orderService) CreateOrder(ctx context.Context, order *model.Order) error {
	order.OrderNumber = generateOrderNumber()
	order.Status = "pending" // Initial status
	
	// Future: here we could also calculate TotalAmount server-side by verifying against ProductRepository
	// For now, we trust the client's TotalAmount or we can implement product lookup if needed.
	
	return s.repo.Create(ctx, order)
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	return s.repo.GetAll(ctx, nil)
}

func (s *orderService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id string, status string) error {
	// Basic validation of status
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"shipped":   true,
		"delivered": true,
		"cancelled": true,
	}
	
	if !validStatuses[status] {
		return fmt.Errorf("invalid order status: %s", status)
	}

	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *orderService) UpdateOrderDetails(ctx context.Context, id string, customer model.Customer, totalAmount float64) error {
	// Simple validation
	if customer.Name == "" || customer.Phone == "" || customer.Address == "" {
		return fmt.Errorf("customer name, phone, and address are required")
	}
	if totalAmount < 0 {
		return fmt.Errorf("total amount cannot be negative")
	}
	return s.repo.UpdateDetails(ctx, id, customer, totalAmount)
}

func (s *orderService) DeleteOrder(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
