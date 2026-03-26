package apperror

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrAlreadyExists         = errors.New("already exists")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrSelfFollow            = errors.New("cannot follow yourself")
	ErrAlreadyFollowing      = errors.New("already following")
	ErrNotFollowing          = errors.New("not following")
	ErrAddressLimitReached   = errors.New("address limit reached (max 10)")
	ErrAddressNotFound       = errors.New("address not found")
	ErrReviewAlreadyExists   = errors.New("review already exists for this order")
	ErrInvalidKYCState       = errors.New("invalid KYC state transition")
	ErrAlreadySeller         = errors.New("user is already a seller")
	ErrSellerNotFound        = errors.New("seller profile not found")
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}
