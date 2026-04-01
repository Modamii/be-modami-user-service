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
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
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

// UpdateUserStatus godoc
// @Summary      Update user status
// @Description  Update a user's account status (admin only)
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "User ID (UUID)"
// @Param        body  body      dto.UpdateStatusRequest  true  "New status"
// @Success      200   {object}  response.Response
// @Failure      400   {object}  response.Response
// @Failure      401   {object}  response.Response
// @Failure      403   {object}  response.Response
// @Failure      404   {object}  response.Response
// @Security     BearerAuth
// @Router       /admin/users/{id}/status [put]
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c.Writer, "ID người dùng không hợp lệ")
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}
	if err := validator.Validate(req); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	if err := h.userService.UpdateStatus(c.Request.Context(), userID, domain.UserStatus(req.Status), req.Reason); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, gin.H{"message": "cập nhật trạng thái thành công"})
}

// ApproveKYC godoc
// @Summary      Approve KYC
// @Description  Approve a user's KYC verification (admin only)
// @Tags         Admin
// @Produce      json
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Security     BearerAuth
// @Router       /admin/users/{id}/kyc/approve [put]
func (h *AdminHandler) ApproveKYC(c *gin.Context) {
	adminID, ok := middleware.UserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "chưa xác thực")
		return
	}

	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c.Writer, "ID người dùng không hợp lệ")
		return
	}

	if err := h.kycService.ApproveKYC(c.Request.Context(), userID, adminID); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, gin.H{"message": "duyệt KYC thành công"})
}

// RejectKYC godoc
// @Summary      Reject KYC
// @Description  Reject a user's KYC verification with a reason (admin only)
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        id    path      string               true  "User ID (UUID)"
// @Param        body  body      dto.RejectKYCRequest  true  "Rejection reason"
// @Success      200   {object}  response.Response
// @Failure      400   {object}  response.Response
// @Failure      401   {object}  response.Response
// @Failure      403   {object}  response.Response
// @Security     BearerAuth
// @Router       /admin/users/{id}/kyc/reject [put]
func (h *AdminHandler) RejectKYC(c *gin.Context) {
	adminID, ok := middleware.UserID(c)
	if !ok {
		response.Unauthorized(c.Writer, "chưa xác thực")
		return
	}

	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c.Writer, "ID người dùng không hợp lệ")
		return
	}

	var req dto.RejectKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}
	if err := validator.Validate(req); err != nil {
		response.BadRequest(c.Writer, err.Error())
		return
	}

	if err := h.kycService.RejectKYC(c.Request.Context(), userID, adminID, req.Reason); err != nil {
		handleError(c, err)
		return
	}

	response.OK(c.Writer, gin.H{"message": "từ chối KYC thành công"})
}

// ListUsers godoc
// @Summary      List users
// @Description  List all users with optional search (admin only)
// @Tags         Admin
// @Produce      json
// @Param        q       query     string  false  "Search query"
// @Param        limit   query     int     false  "Results per page (max 100)"  default(20)
// @Param        cursor  query     string  false  "Pagination cursor"
// @Success      200     {object}  response.Response
// @Failure      401     {object}  response.Response
// @Failure      403     {object}  response.Response
// @Security     BearerAuth
// @Router       /admin/users [get]
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

	response.OK(c.Writer, gin.H{
		"users":  results,
		"cursor": nextCursor,
	})
}
