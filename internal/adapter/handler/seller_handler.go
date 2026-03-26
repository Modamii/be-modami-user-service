package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/user-service/internal/adapter/handler/middleware"
	"github.com/modami/user-service/internal/domain"
	"github.com/modami/user-service/internal/dto"
	"github.com/modami/user-service/internal/service"
	"github.com/modami/user-service/pkg/validator"
)

type SellerHandler struct {
	sellerService *service.SellerService
	kycService    *service.KYCService
}

func NewSellerHandler(sellerService *service.SellerService, kycService *service.KYCService) *SellerHandler {
	return &SellerHandler{
		sellerService: sellerService,
		kycService:    kycService,
	}
}

// Register godoc
// @Summary      Register as seller
// @Description  Register the authenticated user as a seller
// @Tags         Sellers
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterSellerRequest  true  "Seller registration details"
// @Success      201   {object}  dto.SellerProfileResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse  "Already a seller"
// @Security     BearerAuth
// @Router       /users/me/seller/register [post]
func (h *SellerHandler) Register(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.RegisterSellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.sellerService.Register(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toSellerProfileResponse(profile))
}

// UpdateProfile godoc
// @Summary      Update seller profile
// @Description  Update the authenticated seller's profile
// @Tags         Sellers
// @Accept       json
// @Produce      json
// @Param        body  body      dto.UpdateSellerProfileRequest  true  "Fields to update"
// @Success      200   {object}  dto.SellerProfileResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/seller/profile [put]
func (h *SellerHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateSellerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.sellerService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSellerProfileResponse(profile))
}

// GetShopProfile godoc
// @Summary      Get shop profile
// @Description  Returns a seller's public shop profile
// @Tags         Sellers
// @Produce      json
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  dto.SellerProfileResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/{id}/shop [get]
func (h *SellerHandler) GetShopProfile(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	profile, err := h.sellerService.GetShopProfile(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSellerProfileResponse(profile))
}

// SubmitKYC godoc
// @Summary      Submit KYC documents
// @Description  Submit KYC documents for verification
// @Tags         KYC
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SubmitKYCRequest  true  "KYC documents"
// @Success      200   {object}  MessageResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/seller/kyc [post]
func (h *SellerHandler) SubmitKYC(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.SubmitKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.kycService.SubmitKYC(c.Request.Context(), userID, req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "KYC documents submitted"})
}

// GetKYCStatus godoc
// @Summary      Get KYC status
// @Description  Returns the KYC verification status for the authenticated seller
// @Tags         KYC
// @Produce      json
// @Success      200  {object}  dto.KYCStatusResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/seller/kyc/status [get]
func (h *SellerHandler) GetKYCStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	status, err := h.kycService.GetKYCStatus(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.KYCStatusResponse{Status: string(status)})
}

func toSellerProfileResponse(p *domain.SellerProfile) dto.SellerProfileResponse {
	return dto.SellerProfileResponse{
		UserID:          p.UserID.String(),
		ShopName:        p.ShopName,
		ShopSlug:        p.ShopSlug,
		ShopDescription: p.ShopDescription,
		ShopLogoURL:     p.ShopLogoURL,
		ShopBannerURL:   p.ShopBannerURL,
		BusinessType:    string(p.BusinessType),
		KYCStatus:       string(p.KYCStatus),
		AvgRating:       p.AvgRating,
		TotalReviews:    p.TotalReviews,
		CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
