package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/service"
	"github.com/modami/user-service/pkg/validator"
)

type AdminHandler struct {
	userService *service.UserService
	kycService  *service.KYCService
}

func NewAdminHandler(userService *service.UserService, kycService *service.KYCService) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		kycService:  kycService,
	}
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateStatus(c.Request.Context(), userID, domain.UserStatus(req.Status), req.Reason); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *AdminHandler) ApproveKYC(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.kycService.ApproveKYC(c.Request.Context(), userID, adminID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "KYC approved"})
}

func (h *AdminHandler) RejectKYC(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.RejectKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.kycService.RejectKYC(c.Request.Context(), userID, adminID, req.Reason); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "KYC rejected"})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	query := c.Query("q")
	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursor := c.Query("cursor")

	users, nextCursor, err := h.userService.SearchUsers(c.Request.Context(), query, limit, cursor)
	if err != nil {
		handleError(c, err)
		return
	}

	results := make([]dto.UserProfileResponse, 0, len(users))
	for _, u := range users {
		results = append(results, toUserProfileResponse(u))
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  results,
		"cursor": nextCursor,
	})
}
