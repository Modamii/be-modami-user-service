package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/service"
)

type FollowHandler struct {
	followService *service.FollowService
}

func NewFollowHandler(followService *service.FollowService) *FollowHandler {
	return &FollowHandler{followService: followService}
}

func (h *FollowHandler) Follow(c *gin.Context) {
	followerID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.followService.Follow(c.Request.Context(), followerID, followingID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "followed successfully"})
}

func (h *FollowHandler) Unfollow(c *gin.Context) {
	followerID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.followService.Unfollow(c.Request.Context(), followerID, followingID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unfollowed successfully"})
}

func (h *FollowHandler) GetFollowers(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursor := c.Query("cursor")

	users, nextCursor, err := h.followService.GetFollowers(c.Request.Context(), userID, limit, cursor)
	if err != nil {
		handleError(c, err)
		return
	}

	items := make([]dto.FollowUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, dto.FollowUserItem{
			ID:         u.ID.String(),
			FullName:   u.FullName,
			AvatarURL:  u.AvatarURL,
			FollowedAt: u.FollowedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, dto.FollowListResponse{
		Users:  items,
		Cursor: nextCursor,
	})
}

func (h *FollowHandler) GetFollowing(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	cursor := c.Query("cursor")

	users, nextCursor, err := h.followService.GetFollowing(c.Request.Context(), userID, limit, cursor)
	if err != nil {
		handleError(c, err)
		return
	}

	items := make([]dto.FollowUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, dto.FollowUserItem{
			ID:         u.ID.String(),
			FullName:   u.FullName,
			AvatarURL:  u.AvatarURL,
			FollowedAt: u.FollowedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, dto.FollowListResponse{
		Users:  items,
		Cursor: nextCursor,
	})
}

func (h *FollowHandler) CheckFollowStatus(c *gin.Context) {
	followerID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	isFollowing, err := h.followService.CheckFollowStatus(c.Request.Context(), followerID, followingID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.FollowStatusResponse{IsFollowing: isFollowing})
}
