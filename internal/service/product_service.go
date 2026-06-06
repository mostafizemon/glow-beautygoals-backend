package service

import (
	"context"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
)

type ProductService interface {
	GetAllProducts(ctx context.Context) ([]model.Product, error)
	CreateProduct(ctx context.Context, product *model.Product) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{
		repo: repo,
	}
}

func (s *productService) GetAllProducts(ctx context.Context) ([]model.Product, error) {
	return s.repo.GetAll(ctx)
}

func (s *productService) CreateProduct(ctx context.Context, product *model.Product) error {
	return s.repo.Create(ctx, product)
}
