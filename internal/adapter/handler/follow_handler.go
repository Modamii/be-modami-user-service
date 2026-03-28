package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/service"
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
)

type FollowHandler struct {
	followService *service.FollowService
}

func NewFollowHandler(followService *service.FollowService) *FollowHandler {
	return &FollowHandler{followService: followService}
}

// Follow godoc
// @Summary      Follow a user
// @Description  Follow another user by their ID
// @Tags         Follows
// @Produce      json
// @Param        id   path      string  true  "User ID to follow (UUID)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Security     BearerAuth
// @Router       /users/{id}/follow [post]
func (h *FollowHandler) Follow(c *gin.Context) {
	followerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "unauthorized")
		return
	}

	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		response.BadRequest(c.Writer, "invalid user id")
		return
	}

	if err := h.followService.Follow(c.Request.Context(), followerID, followingID); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, gin.H{"message": "followed successfully"})
}

// Unfollow godoc
// @Summary      Unfollow a user
// @Description  Unfollow a user by their ID
// @Tags         Follows
// @Produce      json
// @Param        id   path      string  true  "User ID to unfollow (UUID)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Security     BearerAuth
// @Router       /users/{id}/follow [delete]
func (h *FollowHandler) Unfollow(c *gin.Context) {
	followerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "unauthorized")
		return
	}

	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		response.BadRequest(c.Writer, "invalid user id")
		return
	}

	if err := h.followService.Unfollow(c.Request.Context(), followerID, followingID); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, gin.H{"message": "unfollowed successfully"})
}

// GetFollowers godoc
// @Summary      Get followers
// @Description  Returns a paginated list of a user's followers
// @Tags         Follows
// @Produce      json
// @Param        id      path      string  true   "User ID (UUID)"
// @Param        limit   query     int     false  "Results per page (max 100)"  default(20)
// @Param        cursor  query     string  false  "Pagination cursor"
// @Success      200     {object}  dto.FollowListResponse
// @Failure      400     {object}  response.Response
// @Router       /users/{id}/followers [get]
func (h *FollowHandler) GetFollowers(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c.Writer, "invalid user id")
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

	response.OK(c.Writer, dto.FollowListResponse{
		Users:  items,
		Cursor: nextCursor,
	})
}

// GetFollowing godoc
// @Summary      Get following
// @Description  Returns a paginated list of users that a user follows
// @Tags         Follows
// @Produce      json
// @Param        id      path      string  true   "User ID (UUID)"
// @Param        limit   query     int     false  "Results per page (max 100)"  default(20)
// @Param        cursor  query     string  false  "Pagination cursor"
// @Success      200     {object}  dto.FollowListResponse
// @Failure      400     {object}  response.Response
// @Router       /users/{id}/following [get]
func (h *FollowHandler) GetFollowing(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c.Writer, "invalid user id")
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

	response.OK(c.Writer, dto.FollowListResponse{
		Users:  items,
		Cursor: nextCursor,
	})
}

// CheckFollowStatus godoc
// @Summary      Check follow status
// @Description  Check if the authenticated user follows another user
// @Tags         Follows
// @Produce      json
// @Param        id   path      string  true  "User ID to check (UUID)"
// @Success      200  {object}  dto.FollowStatusResponse
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Security     BearerAuth
// @Router       /users/{id}/follow/status [get]
func (h *FollowHandler) CheckFollowStatus(c *gin.Context) {
	followerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "unauthorized")
		return
	}

	followingIDStr := c.Param("id")
	followingID, err := uuid.Parse(followingIDStr)
	if err != nil {
		response.BadRequest(c.Writer, "invalid user id")
		return
	}

	isFollowing, err := h.followService.CheckFollowStatus(c.Request.Context(), followerID, followingID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, dto.FollowStatusResponse{IsFollowing: isFollowing})
}
