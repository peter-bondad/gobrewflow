package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryHandler interface {
	FindByProductID(ctx *gin.Context)
	GetInventoryByProductID(ctx *gin.Context)
	AdjustStock(c *gin.Context)
}

type inventoryHandler struct {
	service InventoryService
}

func NewInventoryHandler(inventoryService InventoryService) InventoryHandler {
	return &inventoryHandler{
		service: inventoryService,
	}
}

type ProductInventoryResponse struct {
	ProductID   uuid.UUID
	Quantity    int
	ProductName string
	SKU         string
	Price       int64
}

func (h *inventoryHandler) FindByProductID(c *gin.Context) {
	var param GetInventoryByProductIDRequestParam

	if err := c.ShouldBindUri(&param); err != nil {
		c.Error(err)
		return
	}
	productInventoryID, err := uuid.Parse(param.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
		return
	}

	productInventory, err := h.service.FindByProductID(c, productInventoryID)
	if err != nil {
		c.Error(err)
		return
	}

	resp := &ProductInventoryResponse{
		ProductID:   productInventoryID,
		ProductName: productInventory.ProductName,
		SKU:         productInventory.SKU,
		Price:       productInventory.Price,
		Quantity:    productInventory.Quantity,
	}

	c.JSON(200, resp)
}

type GetInventoryByProductIDRequestParam struct {
	ProductID string `uri:"productId" binding:"required,uuid"`
}

type ProductInventoryQuantityResponse struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
}

func (h *inventoryHandler) GetInventoryByProductID(c *gin.Context) {
	var param GetInventoryByProductIDRequestParam

	if err := c.ShouldBindUri(&param); err != nil {
		c.Error(err)
		return
	}

	productID, err := uuid.Parse(param.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	inventory, err := h.service.GetInventoryByProductID(c.Request.Context(), productID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, &ProductInventoryQuantityResponse{
		ProductID: inventory.ProductID,
		Quantity:  inventory.Quantity,
	})
}

type AdjustStockRequest struct {
	AdjustedStock int `json:"adjusted_stock" binding:"gte=0"`
}

type AdjustStockResponse struct {
	ProductID   uuid.UUID `json:"productId"`
	BeforeStock int       `json:"beforeStock"`
	Adjustment  int       `json:"adjustment"`
	AfterStock  int       `json:"afterStock"`
}

func (h *inventoryHandler) AdjustStock(c *gin.Context) {
	var param GetInventoryByProductIDRequestParam

	if err := c.ShouldBindUri(&param); err != nil {
		c.Error(err)
		return
	}

	productID, err := uuid.Parse(param.ProductID)
	if err != nil {
		c.Error(err)
		return
	}

	var req AdjustStockRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	output, err := h.service.AdjustStock(
		c.Request.Context(),
		productID,
		req.AdjustedStock,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, output)
}
