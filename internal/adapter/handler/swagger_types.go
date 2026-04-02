package handler

import (
	"be-modami-user-service/internal/dto"

	"gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
)

// SwaggerResponse is the standard API envelope used in Swagger docs.
type SwaggerResponse struct {
	Success bool          `json:"success" example:"true"`
	Data    any           `json:"data,omitempty"`
	Error   *SwaggerError `json:"error,omitempty"`
	Meta    *SwaggerMeta  `json:"meta,omitempty"`
}

// SwaggerError represents the error portion of a response.
type SwaggerError struct {
	Code    apperror.Code `json:"code" example:"BAD_REQUEST"`
	Message string        `json:"message" example:"error message"`
	Detail  string        `json:"detail,omitempty"`
}

// SwaggerMeta contains response metadata.
type SwaggerMeta struct {
	Timestamp int64 `json:"timestamp" example:"1711584000"`
}

// SearchUsersResponse represents a paginated user search result.
type SearchUsersResponse struct {
	Users  []dto.UserProfileResponse `json:"users"`
	Cursor string                    `json:"cursor,omitempty"`
}

// AddressListResponse represents a list of addresses.
type AddressListResponse struct {
	Addresses []dto.AddressResponse `json:"addresses"`
}
