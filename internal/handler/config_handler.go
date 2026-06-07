package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/service"
)

type ConfigHandler struct {
	service service.ConfigService
}

func NewConfigHandler(s service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		service: s,
	}
}

// GetTrackingConfig returns the full config for admins
func (h *ConfigHandler) GetTrackingConfig(c *gin.Context) {
	config, err := h.service.GetTrackingConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

// UpdateTrackingConfig updates the full config from admins
func (h *ConfigHandler) UpdateTrackingConfig(c *gin.Context) {
	var config model.SiteConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateTrackingConfig(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetPublicPixels returns ONLY the pixel IDs for the frontend script injection (hides access tokens)
func (h *ConfigHandler) GetPublicPixels(c *gin.Context) {
	config, err := h.service.GetTrackingConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Safe response hiding secret tokens
	safeConfig := gin.H{
		"meta": gin.H{
			"pixel_id":  config.Meta.PixelID,
			"is_active": config.Meta.IsActive,
		},
		"tiktok": gin.H{
			"pixel_id":  config.TikTok.PixelID,
			"is_active": config.TikTok.IsActive,
		},
	}

	c.JSON(http.StatusOK, safeConfig)
}

// GetContactConfig returns the contact numbers (public endpoint)
func (h *ConfigHandler) GetContactConfig(c *gin.Context) {
	config, err := h.service.GetContactConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

// UpdateContactConfig saves admin-entered WhatsApp and Phone numbers
func (h *ConfigHandler) UpdateContactConfig(c *gin.Context) {
	var config model.ContactConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateContactConfig(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}
