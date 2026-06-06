package service

import (
	"context"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
)

type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]model.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*model.Category, error)
	CreateCategory(ctx context.Context, category *model.Category) error
	UpdateCategory(ctx context.Context, id string, category *model.Category) error
	DeleteCategory(ctx context.Context, id string) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) GetAllCategories(ctx context.Context) ([]model.Category, error) {
	return s.repo.GetAll(ctx)
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id string) (*model.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *categoryService) CreateCategory(ctx context.Context, category *model.Category) error {
	return s.repo.Create(ctx, category)
}

func (s *categoryService) UpdateCategory(ctx context.Context, id string, category *model.Category) error {
	return s.repo.Update(ctx, id, category)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
