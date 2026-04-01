package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/service"
	"github.com/modami/user-service/pkg/validator"
	"gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMyProfile godoc
// @Summary      Get my profile
// @Description  Returns the authenticated user's profile
// @Tags         Users
// @Produce      json
// @Success      200  {object}  dto.UserProfileResponse
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	keycloakID, ok := middleware.GetKeycloakID(c)
	if !ok {
		response.Unauthorized(c.Writer, "chưa xác thực")
		return
	}

	user, err := h.userService.GetMyProfile(c.Request.Context(), keycloakID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, toUserProfileResponse(user))
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Returns a user's public profile by ID
// @Tags         Users
// @Produce      json
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  dto.UserProfileResponse
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /users/{id} [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c.Writer, "ID người dùng không hợp lệ")
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, toUserProfileResponse(user))
}

// UpdateProfile godoc
// @Summary      Update my profile
// @Description  Updates the authenticated user's profile fields
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body  body      dto.UpdateProfileRequest  true  "Profile fields to update"
// @Success      200   {object}  dto.UserProfileResponse
// @Failure      400   {object}  response.Response
// @Failure      401   {object}  response.Response
// @Security     BearerAuth
// @Router       /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "chưa xác thực")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}
	if err := validator.Validate(req); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, toUserProfileResponse(user))
}

// UpdateAvatar godoc
// @Summary      Update avatar
// @Description  Updates the authenticated user's avatar URL
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body  body      dto.UpdateAvatarRequest  true  "Avatar URL"
// @Success      200   {object}  response.Response
// @Failure      400   {object}  response.Response
// @Failure      401   {object}  response.Response
// @Security     BearerAuth
// @Router       /users/me/avatar [put]
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "chưa xác thực")
		return
	}

	var req dto.UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	if err := h.userService.UpdateAvatar(c.Request.Context(), userID, req.AvatarURL); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, gin.H{"message": "cập nhật ảnh đại diện thành công"})
}

// UpdateCover godoc
// @Summary      Update cover image
// @Description  Updates the authenticated user's cover image URL
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body  body      dto.UpdateCoverRequest  true  "Cover URL"
// @Success      200   {object}  response.Response
// @Failure      400   {object}  response.Response
// @Failure      401   {object}  response.Response
// @Security     BearerAuth
// @Router       /users/me/cover [put]
func (h *UserHandler) UpdateCover(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "chưa xác thực")
		return
	}

	var req dto.UpdateCoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	if err := h.userService.UpdateCover(c.Request.Context(), userID, req.CoverURL); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, gin.H{"message": "cập nhật ảnh bìa thành công"})
}

// DeactivateAccount godoc
// @Summary      Deactivate account
// @Description  Deactivates the authenticated user's account
// @Tags         Users
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Security     BearerAuth
// @Router       /users/me [delete]
func (h *UserHandler) DeactivateAccount(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "chưa xác thực")
		return
	}

	if err := h.userService.DeactivateAccount(c.Request.Context(), userID); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, gin.H{"message": "tài khoản đã được vô hiệu hóa"})
}

// SearchUsers godoc
// @Summary      Search users
// @Description  Search users by name or email
// @Tags         Users
// @Produce      json
// @Param        q       query     string  true   "Search query"
// @Param        limit   query     int     false  "Results per page (max 100)"  default(20)
// @Param        cursor  query     string  false  "Pagination cursor"
// @Success      200     {object}  response.Response
// @Failure      400     {object}  response.Response
// @Router       /users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.BadRequest(c.Writer, "tham số tìm kiếm 'q' là bắt buộc")
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

	response.OK(c.Writer, gin.H{
		"users":  results,
		"cursor": nextCursor,
	})
}

func toUserProfileResponse(u *domain.User) dto.UserProfileResponse {
	resp := dto.UserProfileResponse{
		ID:             u.ID.String(),
		Email:          u.Email,
		UserName:       u.UserName,
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
	if u.DateOfBirth != nil {
		s := u.DateOfBirth.Format("2006-01-02")
		resp.DateOfBirth = &s
	}
	return resp
}

func handleError(c *gin.Context, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Err(c.Writer, ae)
		return
	}
	response.InternalError(c.Writer, "lỗi máy chủ nội bộ")
}
