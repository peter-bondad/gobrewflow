package invitation

import (
	"gobrewflow/internal/middleware"
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
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	InvitationToken string    `json:"invitation_token"`
	ExpiresAt       time.Time `json:"expires_at"`
	Status          string    `json:"status"`
}

type AcceptInvitationResponse struct {
	ID                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	Status              string     `json:"status"`
	AcceptedAt          *time.Time `json:"accepted_at"`
	SetupToken          string     `json:"setup_token"`
	SetupTokenExpiresAt time.Time  `json:"setup_token_expires_at"`
}

type SendInvitationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *invitationHandler) SendInvitation(c *gin.Context) {
	var input SendInvitationRequest
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	requesterID, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	requesterUUID, ok := requesterID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}

	invitation, err := h.service.SendInvitation(c.Request.Context(), input.Email, requesterUUID)
	if err != nil {
		switch err {
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "you do not have permission to send invitations"})
		case ErrInvitationAlreadySent:
			c.JSON(http.StatusConflict, gin.H{"error": "invitation already sent to this email"})
		case ErrInvitationLimitReached:
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "invitation limit reached"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send invitation"})
		}
		return
	}

	resp := SendInvitationResponse{
		ID:              invitation.ID,
		Email:           invitation.Email,
		InvitationToken: invitation.InvitationToken,
		ExpiresAt:       invitation.ExpiresAt,
		Status:          string(invitation.Status),
	}

	c.JSON(http.StatusCreated, resp)
}

type AcceptInvitationRequest struct {
	InvitationToken string `json:"invitation_token" binding:"required"`
}

func (h *invitationHandler) AcceptInvitation(c *gin.Context) {
	var input AcceptInvitationRequest
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	invitation, err := h.service.AcceptInvitation(c.Request.Context(), input.InvitationToken)
	if err != nil {
		switch err {
		case ErrInvitationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found"})
		case ErrInvitationNotPending:
			c.JSON(http.StatusConflict, gin.H{"error": "invitation is not pending"})
		case ErrInvitationExpired:
			c.JSON(http.StatusGone, gin.H{"error": "invitation has expired"})
		case ErrInvitationAlreadyAccepted:
			c.JSON(http.StatusConflict, gin.H{"error": "invitation already accepted"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to accept invitation"})
		}
		return
	}

	resp := AcceptInvitationResponse{
		ID:                  invitation.ID,
		Email:               invitation.Email,
		Status:              string(invitation.Status),
		AcceptedAt:          invitation.AcceptedAt,
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "passwords do not match",
		})
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
		switch err {
		case ErrSetupTokenInvalid:
			c.JSON(http.StatusNotFound, gin.H{"error": "setup token is invalid"})
		case ErrInvitationNotAccepted:
			c.JSON(http.StatusConflict, gin.H{"error": "invitation is not accepted"})
		case ErrSetupTokenExpired:
			c.JSON(http.StatusGone, gin.H{"error": "setup token has expired"})
		case ErrPasswordMismatch:
			c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
		case ErrEmailAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set password"})
		}
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
		switch err {
		case ErrInvitationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found"})
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case ErrInvitationNotPending:
			c.JSON(http.StatusConflict, gin.H{"error": "invitation is not pending"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel invitation"})
		}
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
		switch err {
		case ErrInvitationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found"})
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get invitation"})
		}
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
	invitations, err := h.service.ListInvitations(c.Request.Context(), inviterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list invitations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"count":       len(invitations),
	})
}
