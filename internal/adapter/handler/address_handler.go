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

// AddAddress godoc
// @Summary      Add address
// @Description  Add a new address for the authenticated user
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateAddressRequest  true  "Address details"
// @Success      201   {object}  dto.AddressResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      422   {object}  ErrorResponse  "Address limit reached"
// @Security     BearerAuth
// @Router       /users/me/addresses [post]
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

// ListAddresses godoc
// @Summary      List addresses
// @Description  Returns all addresses for the authenticated user
// @Tags         Addresses
// @Produce      json
// @Success      200  {object}  AddressListResponse
// @Failure      401  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/addresses [get]
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

// UpdateAddress godoc
// @Summary      Update address
// @Description  Update an existing address
// @Tags         Addresses
// @Accept       json
// @Produce      json
// @Param        addr_id  path      string                   true  "Address ID (UUID)"
// @Param        body     body      dto.UpdateAddressRequest  true  "Fields to update"
// @Success      200      {object}  dto.AddressResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/addresses/{addr_id} [put]
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

// DeleteAddress godoc
// @Summary      Delete address
// @Description  Delete an address by ID
// @Tags         Addresses
// @Produce      json
// @Param        addr_id  path      string  true  "Address ID (UUID)"
// @Success      200      {object}  MessageResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/addresses/{addr_id} [delete]
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

// SetDefault godoc
// @Summary      Set default address
// @Description  Set an address as the default
// @Tags         Addresses
// @Produce      json
// @Param        addr_id  path      string  true  "Address ID (UUID)"
// @Success      200      {object}  MessageResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/me/addresses/{addr_id}/default [put]
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
