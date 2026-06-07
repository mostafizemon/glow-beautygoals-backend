package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glow-and-beauty-goals/backend/internal/service"
)

type TrackingHandler struct {
	service service.TrackingService
}

func NewTrackingHandler(s service.TrackingService) *TrackingHandler {
	return &TrackingHandler{
		service: s,
	}
}

// TrackEvent receives events from the frontend and pushes them to Server-Side APIs
func (h *TrackingHandler) TrackEvent(c *gin.Context) {
	var payload service.TrackingPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Inject Client IP from the request for high Event Match Quality
	if payload.UserData == nil {
		payload.UserData = make(map[string]interface{})
	}
	payload.UserData["client_ip_address"] = getClientIP(c)

	// Run tracking asynchronously so it doesn't block the user's request
	go func() {
		_ = h.service.TrackEvent(context.Background(), payload)
	}()

	// Always return 200 immediately to the frontend
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}
