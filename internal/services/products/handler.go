package products

import (
	"gobrewflow/shared"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductHandler interface {
	CreateProducts(c *gin.Context)
	FindProductByID(c *gin.Context)
	FindProductBySKU(c *gin.Context)
	ListProducts(c *gin.Context)
}

type productHandler struct {
	service ProductService
}

func NewProductHandler(service ProductService) ProductHandler {
	return &productHandler{
		service: service,
	}
}

type CreateProductRequest struct {
	Name       string `json:"name" binding:"required"`
	Slug       string `json:"slug"`
	Price      int64  `json:"price" binding:"required"`
	CategoryID string `json:"category_id" binding:"required"`
}

func (h *productHandler) CreateProducts(c *gin.Context) {
	var req []CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if len(req) == 0 {
		c.Error(ErrNoProducts)
		return
	}

	inputs := make([]CreateProductInput, len(req))

	for i, product := range req {
		inputs[i] = CreateProductInput{
			Name:       product.Name,
			Slug:       product.Slug,
			Price:      product.Price,
			CategoryID: product.CategoryID,
		}
	}

	if err := h.service.CreateProducts(
		c.Request.Context(),
		inputs,
	); err != nil {
		c.Error(err)
		return
	}

	if len(req) > 1 {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Products created successfully",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
	})
}
func (h *productHandler) FindProductByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(ProductIdIsRequired)
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
		c.Error(ProductSKUIsRequired)
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
		c.Error(err)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Limit < 1 || req.Limit > 100 {
		c.Error(InvalidLimitParameter)
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.Page < 1 {
		c.Error(InvalidPageParameter)
		return
	}

	input := ProductListInput{
		Name:     req.Name,
		Category: req.Category,
		IsActive: req.IsActive,
		Page:     req.Page,
	}

	productsData, err := h.service.ListProducts(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	resp := ListProductsResponse{
		Data: productsData.Data,
		Pagination: shared.Pagination{
			Page:       productsData.Page,
			Limit:      productsData.Limit,
			Total:      productsData.Total,
			TotalPages: productsData.TotalPages,
		},
	}

	c.JSON(http.StatusOK, resp)
}
