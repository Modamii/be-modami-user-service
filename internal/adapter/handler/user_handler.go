package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/service"
	"github.com/modami/user-service/pkg/apperror"
	"github.com/modami/user-service/pkg/validator"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMyProfile(c *gin.Context) {
	keycloakID, ok := middleware.GetKeycloakID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userService.GetMyProfile(c.Request.Context(), keycloakID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toUserProfileResponse(user))
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toUserProfileResponse(user))
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toUserProfileResponse(user))
}

func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateAvatar(c.Request.Context(), userID, req.AvatarURL); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "avatar updated"})
}

func (h *UserHandler) UpdateCover(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateCoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateCover(c.Request.Context(), userID, req.CoverURL); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cover updated"})
}

func (h *UserHandler) DeactivateAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.userService.DeactivateAccount(c.Request.Context(), userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deactivated"})
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

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

	var results []dto.UserProfileResponse
	for _, u := range users {
		results = append(results, toUserProfileResponse(u))
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  results,
		"cursor": nextCursor,
	})
}

func toUserProfileResponse(u *domain.User) dto.UserProfileResponse {
	return dto.UserProfileResponse{
		ID:             u.ID.String(),
		Email:          u.Email,
		FullName:       u.FullName,
		Phone:          u.Phone,
		AvatarURL:      u.AvatarURL,
		CoverURL:       u.CoverURL,
		Bio:            u.Bio,
		Gender:         string(u.Gender),
		Role:           string(u.Role),
		Status:         string(u.Status),
		TrustScore:     u.TrustScore,
		FollowerCount:  u.FollowerCount,
		FollowingCount: u.FollowingCount,
		EmailVerified:  u.EmailVerified,
		CreatedAt:      u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrSelfFollow):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrAlreadyFollowing):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrNotFollowing):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrAddressLimitReached):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrAddressNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrReviewAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrInvalidKYCState):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrAlreadySeller):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, apperror.ErrSellerNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
