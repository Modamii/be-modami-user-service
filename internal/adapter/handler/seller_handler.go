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
