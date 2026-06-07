package service

import (
	"context"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
)

type ConfigService interface {
	GetTrackingConfig(ctx context.Context) (*model.SiteConfig, error)
	UpdateTrackingConfig(ctx context.Context, config *model.SiteConfig) error
	GetContactConfig(ctx context.Context) (*model.ContactConfig, error)
	UpdateContactConfig(ctx context.Context, config *model.ContactConfig) error
}

type configService struct {
	repo repository.ConfigRepository
}

func NewConfigService(repo repository.ConfigRepository) ConfigService {
	return &configService{
		repo: repo,
	}
}

func (s *configService) GetTrackingConfig(ctx context.Context) (*model.SiteConfig, error) {
	return s.repo.GetSiteConfig(ctx, "tracking_pixels")
}

func (s *configService) UpdateTrackingConfig(ctx context.Context, config *model.SiteConfig) error {
	config.ConfigKey = "tracking_pixels"
	return s.repo.UpsertSiteConfig(ctx, config)
}

func (s *configService) GetContactConfig(ctx context.Context) (*model.ContactConfig, error) {
	return s.repo.GetContactConfig(ctx)
}

func (s *configService) UpdateContactConfig(ctx context.Context, config *model.ContactConfig) error {
	return s.repo.UpsertContactConfig(ctx, config)
}
