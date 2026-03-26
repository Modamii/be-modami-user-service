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

type AddressHandler struct {
	addressService *service.AddressService
}

func NewAddressHandler(addressService *service.AddressService) *AddressHandler {
	return &AddressHandler{addressService: addressService}
}

func (h *AddressHandler) AddAddress(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addr, err := h.addressService.AddAddress(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toAddressResponse(addr))
}

func (h *AddressHandler) ListAddresses(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	addrs, err := h.addressService.ListAddresses(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	items := make([]dto.AddressResponse, 0, len(addrs))
	for _, a := range addrs {
		items = append(items, toAddressResponse(a))
	}

	c.JSON(http.StatusOK, gin.H{"addresses": items})
}

func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	addrIDStr := c.Param("addr_id")
	addrID, err := uuid.Parse(addrIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}

	var req dto.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.Validate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addr, err := h.addressService.UpdateAddress(c.Request.Context(), userID, addrID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toAddressResponse(addr))
}

func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	addrIDStr := c.Param("addr_id")
	addrID, err := uuid.Parse(addrIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}

	if err := h.addressService.DeleteAddress(c.Request.Context(), userID, addrID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "address deleted"})
}

func (h *AddressHandler) SetDefault(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	addrIDStr := c.Param("addr_id")
	addrID, err := uuid.Parse(addrIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address id"})
		return
	}

	if err := h.addressService.SetDefault(c.Request.Context(), userID, addrID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "default address updated"})
}

func toAddressResponse(a *domain.Address) dto.AddressResponse {
	return dto.AddressResponse{
		ID:            a.ID.String(),
		Label:         a.Label,
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		AddressLine1:  a.AddressLine1,
		AddressLine2:  a.AddressLine2,
		Ward:          a.Ward,
		District:      a.District,
		Province:      a.Province,
		PostalCode:    a.PostalCode,
		Country:       a.Country,
		IsDefault:     a.IsDefault,
		CreatedAt:     a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
