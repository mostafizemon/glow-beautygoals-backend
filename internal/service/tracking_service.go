package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/repository"
)

type TrackingService interface {
	TrackEvent(ctx context.Context, payload TrackingPayload) error
}

type trackingService struct {
	configRepo repository.ConfigRepository
	httpClient *http.Client
}

func NewTrackingService(configRepo repository.ConfigRepository) TrackingService {
	return &trackingService{
		configRepo: configRepo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Payload coming from the Next.js frontend
type TrackingPayload struct {
	EventName  string                 `json:"event_name"`
	EventID    string                 `json:"event_id"`
	EventTime  int64                  `json:"event_time"`
	EventURL   string                 `json:"event_url"`
	UserData   map[string]interface{} `json:"user_data"`
	CustomData map[string]interface{} `json:"custom_data"`
}

func (s *trackingService) TrackEvent(ctx context.Context, payload TrackingPayload) error {
	// Get tracking config
	config, err := s.configRepo.GetSiteConfig(ctx, "tracking_pixels")
	if err != nil {
		return err // Should not happen ideally, but if DB fails, we fail
	}

	// Fire concurrently
	errChan := make(chan error, 2)
	go func() {
		errChan <- s.sendToMetaCAPI(ctx, config.Meta, payload)
	}()
	go func() {
		errChan <- s.sendToTikTokAPI(ctx, config.TikTok, payload)
	}()

	var finalErr error
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			fmt.Printf("Tracking Error: %v\n", err)
			finalErr = err
		}
	}

	return finalErr // Returning the last error if any, but tracking should usually not block the user flow
}

func (s *trackingService) sendToMetaCAPI(ctx context.Context, cfg model.PixelConfig, payload TrackingPayload) error {
	if !cfg.IsActive || cfg.PixelID == "" || cfg.AccessToken == "" {
		return nil
	}

	url := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events", cfg.PixelID)

	eventData := map[string]interface{}{
		"event_name": payload.EventName,
		"event_time": payload.EventTime,
		"action_source": "website",
		"event_id": payload.EventID,
		"event_source_url": payload.EventURL,
		"user_data": payload.UserData,
		"custom_data": payload.CustomData,
	}

	// Make sure client_ip_address is passed if available
	// Meta requires client_ip_address for high match quality
	if ip, ok := payload.UserData["client_ip_address"]; !ok || ip == "" {
		// Just in case it wasn't passed, though it should be populated by the handler
		// We'll leave it as is, but ensure the struct accepts it
	}

	requestBody := map[string]interface{}{
		"data": []interface{}{eventData},
	}

	if cfg.TestEventCode != "" {
		requestBody["test_event_code"] = cfg.TestEventCode
	}

	urlWithToken := fmt.Sprintf("%s?access_token=%s", url, cfg.AccessToken)

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", urlWithToken, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Meta CAPI failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *trackingService) sendToTikTokAPI(ctx context.Context, cfg model.PixelConfig, payload TrackingPayload) error {
	if !cfg.IsActive || cfg.PixelID == "" || cfg.AccessToken == "" {
		return nil
	}

	url := "https://business-api.tiktok.com/open_api/v1.3/pixel/track/"

	eventName := payload.EventName
	if eventName == "Purchase" {
		eventName = "CompletePayment"
	}

	// TikTok expects ip and user_agent in context, while hashed identifiers live in context.user.
	contextData := map[string]interface{}{
		"page": map[string]interface{}{
			"url": payload.EventURL,
		},
	}

	tikTokUser := make(map[string]interface{})
	if ip, ok := payload.UserData["client_ip_address"]; ok {
		contextData["ip"] = ip
	}
	if ua, ok := payload.UserData["client_user_agent"]; ok {
		contextData["user_agent"] = ua
	}
	if ttp, ok := payload.UserData["ttp"]; ok {
		tikTokUser["ttp"] = ttp
	}
	if em, ok := payload.UserData["em"]; ok {
		tikTokUser["email"] = em
	}
	if ph, ok := payload.UserData["ph"]; ok {
		tikTokUser["phone_number"] = ph
	}
	if len(tikTokUser) > 0 {
		contextData["user"] = tikTokUser
	}

	requestBody := map[string]interface{}{
		"pixel_code": cfg.PixelID,
		"event": eventName,
		"event_id": payload.EventID,
		"timestamp": time.Unix(payload.EventTime, 0).UTC().Format(time.RFC3339),
		"context": contextData,
		"properties": payload.CustomData,
	}

	if cfg.TestEventCode != "" {
		requestBody["test_event_code"] = cfg.TestEventCode
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", cfg.AccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("TikTok API failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
