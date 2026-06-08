package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
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
	EventName     string                 `json:"event_name"`
	EventID       string                 `json:"event_id"`
	EventTime     int64                  `json:"event_time"`
	EventURL      string                 `json:"event_url"`
	EventReferrer string                 `json:"event_referrer"`
	UserData      map[string]interface{} `json:"user_data"`
	CustomData    map[string]interface{} `json:"custom_data"`
}

var sha256Regex = regexp.MustCompile(`^[a-f0-9]{64}$`)

func hashSHA256(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || sha256Regex.MatchString(value) {
		return value
	}

	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func hashPhone(value string) string {
	digits := regexp.MustCompile(`\D+`).ReplaceAllString(value, "")
	return hashSHA256(digits)
}

func stringValue(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}

	if value, ok := data[key].(string); ok {
		return strings.TrimSpace(value)
	}

	return ""
}

func normalizeUserData(userData map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})
	for key, value := range userData {
		if value == nil {
			continue
		}

		if text, ok := value.(string); ok {
			if strings.TrimSpace(text) != "" {
				normalized[key] = text
			}
			continue
		}

		normalized[key] = value
	}

	if value := stringValue(normalized, "em"); value != "" {
		normalized["em"] = hashSHA256(value)
	}
	if value := stringValue(normalized, "ph"); value != "" {
		normalized["ph"] = hashPhone(value)
	}
	if value := stringValue(normalized, "fn"); value != "" {
		normalized["fn"] = hashSHA256(value)
	}
	if value := stringValue(normalized, "ln"); value != "" {
		normalized["ln"] = hashSHA256(value)
	}

	return normalized
}

func (s *trackingService) TrackEvent(ctx context.Context, payload TrackingPayload) error {
	payload.UserData = normalizeUserData(payload.UserData)

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

	var errs []string
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			fmt.Printf("Tracking Error: %v\n", err)
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("tracking errors: %s", strings.Join(errs, "; "))
	}
	return nil
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

	requestBody := map[string]interface{}{
		"data": []interface{}{eventData},
	}

	if cfg.TestEventCode != "" {
		requestBody["test_event_code"] = cfg.TestEventCode
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal Meta request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create Meta request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(cfg.AccessToken))

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
		fmt.Printf("[TikTok] SKIPPED event=%s reason: IsActive=%v PixelID=%q HasAccessToken=%v\n",
			payload.EventName, cfg.IsActive, cfg.PixelID, cfg.AccessToken != "")
		return nil
	}

	url := "https://business-api.tiktok.com/open_api/v1.3/event/track/"

	eventName := payload.EventName
	if eventName == "Purchase" {
		eventName = "CompletePayment"
	}

	pageData := map[string]interface{}{
		"url": payload.EventURL,
	}
	if payload.EventReferrer != "" {
		pageData["referrer"] = payload.EventReferrer
	}

	tikTokUser := make(map[string]interface{})
	if ip, ok := payload.UserData["client_ip_address"]; ok {
		tikTokUser["ip"] = ip
	}
	if ua, ok := payload.UserData["client_user_agent"]; ok {
		tikTokUser["user_agent"] = ua
	}
	if ttp, ok := payload.UserData["ttp"]; ok {
		tikTokUser["ttp"] = ttp
	}
	if ttclid, ok := payload.UserData["ttclid"]; ok {
		tikTokUser["ttclid"] = ttclid
	}
	if em, ok := payload.UserData["em"]; ok {
		tikTokUser["email"] = em
	}
	if ph, ok := payload.UserData["ph"]; ok {
		tikTokUser["phone_number"] = ph
	}
	if fn, ok := payload.UserData["fn"]; ok {
		tikTokUser["first_name"] = fn
	}
	if ln, ok := payload.UserData["ln"]; ok {
		tikTokUser["last_name"] = ln
	}

	eventObj := map[string]interface{}{
		"event": eventName,
		"event_id": payload.EventID,
		"event_time": payload.EventTime,
		"page": pageData,
		"properties": payload.CustomData,
	}

	if len(tikTokUser) > 0 {
		eventObj["user"] = tikTokUser
	}

	if callback, ok := payload.UserData["ttclid"]; ok {
		eventObj["ad"] = map[string]interface{}{
			"callback": callback,
		}
	}

	requestBody := map[string]interface{}{
		"event_source": "web",
		"event_source_id": cfg.PixelID,
		"data": []interface{}{eventObj},
	}

	if cfg.TestEventCode != "" {
		requestBody["test_event_code"] = cfg.TestEventCode
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal TikTok request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create TikTok request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", strings.TrimSpace(cfg.AccessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// TikTok ALWAYS returns HTTP 200 even on errors.
	// The real result is inside the JSON body: { "code": 0, "message": "OK" }
	var tikTokResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &tikTokResp); err != nil {
		fmt.Printf("[TikTok] Failed to parse response: %s\n", string(body))
		return fmt.Errorf("tiktok: failed to parse response: %w", err)
	}

	if tikTokResp.Code != 0 {
		fmt.Printf("[TikTok] API Error code=%d message=%s\n", tikTokResp.Code, tikTokResp.Message)
		return fmt.Errorf("tiktok API error %d: %s", tikTokResp.Code, tikTokResp.Message)
	}

	fmt.Printf("[TikTok] Event sent OK: %s (test_code=%s)\n", eventName, cfg.TestEventCode)
	return nil
}
