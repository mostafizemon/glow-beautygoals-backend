package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glow-and-beauty-goals/backend/internal/model"
	"github.com/glow-and-beauty-goals/backend/internal/service"
	"github.com/glow-and-beauty-goals/backend/pkg/cloud"
)

type ProductHandler struct {
	service           service.ProductService
	cloudinaryService cloud.CloudinaryService
}

func NewProductHandler(s service.ProductService, c cloud.CloudinaryService) *ProductHandler {
	return &ProductHandler{
		service:           s,
		cloudinaryService: c,
	}
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	filter := make(map[string]interface{})

	if isFeatured := c.Query("is_featured"); isFeatured == "true" {
		filter["is_featured"] = true
	}
	if category := c.Query("category"); category != "" {
		filter["categories"] = category
	}

	products, err := h.service.GetAllProducts(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product slug is required"})
		return
	}

	product, err := h.service.GetProductBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateProduct(c.Request.Context(), &product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	if h.cloudinaryService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cloudinary not configured"})
		return
	}

	url, err := h.cloudinaryService.UploadImage(c.Request.Context(), file, "glow_store_uploads")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	// Get product to retrieve image URLs
	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	// Delete from Cloudinary
	if h.cloudinaryService != nil {
		for _, imgURL := range product.Images {
			// We try to delete all images. If one fails, we log it but continue
			_ = h.cloudinaryService.DeleteImage(c.Request.Context(), imgURL)
		}
	}

	// Delete from Database
	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product ID is required"})
		return
	}

	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateProduct(c.Request.Context(), id, &product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product updated successfully"})
}
