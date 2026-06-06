package service

import (
	"context"
	"fmt"
	"time"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
)

type ProductService interface {
	GetAllProducts(ctx context.Context, filter map[string]interface{}) ([]model.Product, error)
	GetProductByID(ctx context.Context, id string) (*model.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*model.Product, error)
	CreateProduct(ctx context.Context, product *model.Product) error
	UpdateProduct(ctx context.Context, id string, product *model.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{
		repo: repo,
	}
}

func (s *productService) GetAllProducts(ctx context.Context, filter map[string]interface{}) ([]model.Product, error) {
	return s.repo.GetAll(ctx, filter)
}
func (s *productService) CreateProduct(ctx context.Context, product *model.Product) error {
	existing, err := s.repo.GetBySlug(ctx, product.Slug)
	if err == nil && existing != nil {
		product.Slug = fmt.Sprintf("%s-%d", product.Slug, time.Now().Unix())
	}
	return s.repo.Create(ctx, product)
}

func (s *productService) GetProductByID(ctx context.Context, id string) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) GetProductBySlug(ctx context.Context, slug string) (*model.Product, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *productService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *productService) UpdateProduct(ctx context.Context, id string, product *model.Product) error {
	existing, err := s.repo.GetBySlug(ctx, product.Slug)
	if err == nil && existing != nil && existing.ID.Hex() != id {
		product.Slug = fmt.Sprintf("%s-%d", product.Slug, time.Now().Unix())
	}
	return s.repo.Update(ctx, id, product)
}
