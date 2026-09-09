package invitations

import (
	"gobrewflow/shared"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InvitationHandler interface {
	SendInvitation(c *gin.Context)
	AcceptInvitation(c *gin.Context)
	SetPassword(c *gin.Context)
	CancelInvitation(c *gin.Context)
	GetInvitation(c *gin.Context)
	ListInvitations(c *gin.Context)
}

type invitationHandler struct {
	service InvitationService
}

func NewInvitationHandler(service InvitationService) InvitationHandler {
	return &invitationHandler{
		service: service,
	}
}

type SendInvitationResponse struct {
	Email           string    `json:"email"`
	InvitationToken string    `json:"invitation_token"`
	ExpiresAt       time.Time `json:"expires_at"`
	Message         string    `json:"message"`
}

type SendInvitationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *invitationHandler) SendInvitation(c *gin.Context) {
	var input SendInvitationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	requesterID, exists := c.Get(shared.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	requesterUUID, ok := requesterID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}

	invitation, err := h.service.SendInvitation(c.Request.Context(), SendInvitationInput{
		Email:     input.Email,
		InviterID: requesterUUID,
	})
	if err != nil {
		c.Error(err)
		return
	}

	resp := SendInvitationResponse{
		Email:           invitation.Email,
		InvitationToken: invitation.InvitationToken,
		ExpiresAt:       invitation.ExpiresAt,
		Message:         "Invitation sent successfully",
	}

	c.JSON(http.StatusCreated, resp)
}

type AcceptInvitationRequest struct {
	InvitationToken string `json:"invitation_token" binding:"required"`
}

type AcceptInvitationResponse struct {
	Email               string    `json:"email"`
	SetupToken          string    `json:"setup_token"`
	SetupTokenExpiresAt time.Time `json:"setup_token_expires_at"`
}

func (h *invitationHandler) AcceptInvitation(c *gin.Context) {
	var input AcceptInvitationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		return
	}

	invitation, err := h.service.AcceptInvitation(c.Request.Context(), input.InvitationToken)
	if err != nil {
		c.Error(err)
		return
	}

	resp := AcceptInvitationResponse{
		Email:               invitation.Email,
		SetupToken:          invitation.SetupToken,
		SetupTokenExpiresAt: *invitation.SetupTokenExpiresAt,
	}

	c.JSON(http.StatusOK, resp)
}

type SetPasswordRequest struct {
	SetupToken      string  `json:"setup_token" binding:"required"`
	FirstName       string  `json:"first_name" binding:"required"`
	MiddleName      *string `json:"middle_name"`
	LastName        string  `json:"last_name" binding:"required"`
	Password        string  `json:"password" binding:"required,min=6"`
	ConfirmPassword string  `json:"confirm_password" binding:"required"`
}

type SetPasswordResponse struct {
	ID     uuid.UUID `json:"id"`
	Email  string    `json:"email"`
	Status string    `json:"status"`
	UserID uuid.UUID `json:"user_id"`
}

func (h *invitationHandler) SetPassword(c *gin.Context) {
	var req SetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if req.Password != req.ConfirmPassword {
		c.Error(ErrPasswordMismatch)
		return
	}

	input := SetPasswordInput{
		SetupToken: req.SetupToken,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Password:   req.Password,
	}

	invitation, err := h.service.SetPassword(
		c.Request.Context(),
		input,
	)
	if err != nil {
		c.Error(err)
		return
	}

	resp := SetPasswordResponse{
		ID:     invitation.ID,
		Email:  invitation.Email,
		Status: string(invitation.Status),
		UserID: *invitation.UserID,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *invitationHandler) CancelInvitation(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
		return
	}

	requesterID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	requesterUUID := requesterID.(uuid.UUID)
	if err := h.service.CancelInvitation(c.Request.Context(), id, requesterUUID); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invitation cancelled"})
}

func (h *invitationHandler) GetInvitation(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
		return
	}

	requesterID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	requesterUUID := requesterID.(uuid.UUID)
	invitation, err := h.service.GetInvitation(c.Request.Context(), id, requesterUUID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          invitation.ID,
		"email":       invitation.Email,
		"status":      invitation.Status,
		"expires_at":  invitation.ExpiresAt,
		"accepted_at": invitation.AcceptedAt,
		"created_at":  invitation.CreatedAt,
	})
}

func (h *invitationHandler) ListInvitations(c *gin.Context) {
	requesterID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	inviterID := requesterID.(uuid.UUID)
	invitationsData, err := h.service.ListInvitations(c.Request.Context(), inviterID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitationsData,
		"count":       len(invitationsData),
	})
}
