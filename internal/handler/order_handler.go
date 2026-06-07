package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/service"
)

type OrderHandler struct {
	service         service.OrderService
	trackingService service.TrackingService
}

func NewOrderHandler(s service.OrderService, trackingService service.TrackingService) *OrderHandler {
	return &OrderHandler{
		service:         s,
		trackingService: trackingService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var order model.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Capture IP Address and User Agent for fraud detection/tracking if available
	order.Customer.IPAddress = getClientIP(c)
	order.Customer.UserAgent = c.Request.UserAgent()

	if err := h.service.CreateOrder(c.Request.Context(), &order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.trackPurchase(c, &order)

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) trackPurchase(c *gin.Context, order *model.Order) {
	if h.trackingService == nil {
		return
	}

	eventID := order.TrackingEvents.EventID
	if eventID == "" {
		eventID = order.OrderNumber
	}

	eventURL := order.TrackingEvents.EventURL
	if eventURL == "" {
		eventURL = c.GetHeader("Referer")
	}

	contentIDs := make([]string, 0, len(order.Items))
	contents := make([]map[string]interface{}, 0, len(order.Items))
	for _, item := range order.Items {
		productID := item.ProductID.Hex()
		contentIDs = append(contentIDs, productID)
		contents = append(contents, map[string]interface{}{
			"content_id": productID,
			"id":         productID,
			"quantity":   item.Quantity,
			"price":      item.PriceAtPurchase,
			"item_price": item.PriceAtPurchase,
		})
	}

	payload := service.TrackingPayload{
		EventName:     "Purchase",
		EventID:       eventID,
		EventTime:     time.Now().Unix(),
		EventURL:      eventURL,
		EventReferrer: c.GetHeader("Referer"),
		CustomData: map[string]interface{}{
			"value":        order.TotalAmount,
			"currency":     "BDT",
			"content_type": "product",
			"content_ids":  contentIDs,
			"contents":     contents,
			"order_id":     order.OrderNumber,
		},
		UserData: map[string]interface{}{
			"client_ip_address": order.Customer.IPAddress,
			"client_user_agent": order.Customer.UserAgent,
			"fbp":               order.TrackingEvents.Fbp,
			"fbc":               order.TrackingEvents.Fbc,
			"ttp":               order.TrackingEvents.Ttp,
			"ttclid":            order.TrackingEvents.Ttclid,
			"fn":                order.Customer.Name,
			"ph":                order.Customer.Phone,
		},
	}

	go func() {
		_ = h.trackingService.TrackEvent(context.Background(), payload)
	}()
}

func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	orders, err := h.service.GetAllOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated successfully"})
}
