package products

import (
	"gobrewflow/shared"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductHandler interface {
	CreateProduct(c *gin.Context)
	FindProductByID(c *gin.Context)
	FindProductBySKU(c *gin.Context)
	ListProducts(c *gin.Context)
}

type productHandler struct {
	service ProductServiceInterface
}

func NewProductHandler(service ProductServiceInterface) ProductHandler {
	return &productHandler{
		service: service,
	}
}

type CreateProductRequest struct {
	Name       string `json:"name" binding:"required"`
	Slug       string `json:"slug"`
	CategoryID string `json:"category_id"`
}

func (h *productHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Product name is required",
		})
		return
	}

	if req.CategoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Category ID is required",
		})
		return
	}

	err := h.service.InsertProduct(c.Request.Context(), &ProductInput{
		Name:       req.Name,
		Slug:       req.Slug,
		CategoryID: req.CategoryID,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
	})
}

func (h *productHandler) FindProductByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "Product ID is required"})
		return
	}

	product, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, product)
}

func (h *productHandler) FindProductBySKU(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		c.JSON(400, gin.H{"error": "Product SKU is required"})
		return
	}

	product, err := h.service.FindBySKU(c.Request.Context(), sku)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, product)
}

type ListProductsRequest struct {
	Name     string `form:"name"`
	Category string `form:"category"`
	IsActive *bool  `form:"is_active"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
}

type ListProductsResponse struct {
	Data       []ProductListItem `json:"data"`
	Pagination shared.Pagination `json:"pagination"`
}

func (h *productHandler) ListProducts(c *gin.Context) {
	var req ListProductsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid query parameters",
		})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Limit < 1 || req.Limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid limit parameter",
		})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.Page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid page parameter",
		})
		return
	}

	input := ProductListInput{
		Name:     req.Name,
		Category: req.Category,
		IsActive: req.IsActive,
		Page:     req.Page,
	}

	products, err := h.service.ListProducts(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}

	resp := ListProductsResponse{
		Data: products.Data,
		Pagination: shared.Pagination{
			Page:       products.Page,
			Limit:      products.Limit,
			Total:      products.Total,
			TotalPages: products.TotalPages,
		},
	}

	c.JSON(http.StatusOK, resp)
}
