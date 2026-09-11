package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryHandler interface {
	FindByProductID(ctx *gin.Context)
}

type inventoryHandler struct {
	service InventoryServiceInterface
}

func NewInventoryHandler(inventoryService inventoryService) InventoryHandler {
	return &inventoryHandler{
		service: inventoryService,
	}
}

type FindByProductIDRequestParam struct {
	ProductInventoryID string `uri:"productInventoryId" binding:"required,uuid"`
}

type ProductInventoryResponse struct {
	ProductID   uuid.UUID
	Quantity    int
	ProductName string
	SKU         string
	Price       int64
}

func (h *inventoryHandler) FindByProductID(c *gin.Context) {
	var param FindByProductIDRequestParam

	if err := c.ShouldBindUri(&param); err != nil {
		c.Error(err)
		return
	}
	productInventoryID, err := uuid.Parse(param.ProductInventoryID)
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
