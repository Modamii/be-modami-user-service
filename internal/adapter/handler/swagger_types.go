package handler

import "github.com/modami/user-service/internal/dto"

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// MessageResponse represents a success message response.
type MessageResponse struct {
	Message string `json:"message" example:"operation successful"`
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
