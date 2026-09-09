package categories

import (
	"gobrewflow/shared"
	"net/http"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type CategoryHandler interface {
	CreateCategory(c *gin.Context)
	UpdateCategoryName(c *gin.Context)
	ListCategories(c *gin.Context)
	SetCategoryStatus(c *gin.Context)
}

type categoryHandler struct {
	service CategoryServiceInterface
}

func NewCategoryHandler(service CategoryServiceInterface) CategoryHandler {
	return &categoryHandler{
		service: service,
	}
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *categoryHandler) CreateCategory(c *gin.Context) {
	var input CreateCategoryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	err := h.service.CreateCategory(c.Request.Context(), input.Name)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Category created successfully",
	})
}

type UpdateCategoryNameRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCategoryNameResponse struct {
	Success bool `json:"success"`
}

func (h *categoryHandler) UpdateCategoryName(c *gin.Context) {
	var input UpdateCategoryNameRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	id := c.Param("id")
	categoryID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	success, err := h.service.UpdateCategoryName(c.Request.Context(), categoryID, input.Name)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, UpdateCategoryNameResponse{
		Success: success,
	})
}

type CategoryListRequest struct {
	Name  string `form:"name"`
	Limit int    `form:"limit"`
	Page  int    `form:"page"`
}

type CategoryListResponse struct {
	Data       []CategoryList    `json:"data"`
	Pagination shared.Pagination `json:"pagination"`
}

func (h *categoryHandler) ListCategories(c *gin.Context) {
	var req CategoryListRequest

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

	input := CategoryListInput{
		Name:  req.Name,
		Limit: req.Limit,
		Page:  req.Page,
	}

	categoriesData, err := h.service.ListCategories(
		c.Request.Context(),
		input,
	)
	if err != nil {
		c.Error(err)
		return
	}
	resp := CategoryListResponse{
		Data: categoriesData.Data,
		Pagination: shared.Pagination{
			Page:       categoriesData.Page,
			Limit:      categoriesData.Limit,
			Total:      categoriesData.Total,
			TotalPages: categoriesData.TotalPages,
		},
	}

	c.JSON(http.StatusOK, resp)
}

type SetCategoryStatusParam struct {
	ID string `uri:"id" binding:"required,uuid"`
}
type SetCategoryStatusRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *categoryHandler) SetCategoryStatus(c *gin.Context) {
	var param SetCategoryStatusParam
	var input SetCategoryStatusRequest
	if err := c.ShouldBindUri(&param); err != nil {
		c.Error(err)
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	categoryID, err := uuid.Parse(param.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	success, err := h.service.SetCategoryStatus(c.Request.Context(), categoryID, input.IsActive)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
	})

}
