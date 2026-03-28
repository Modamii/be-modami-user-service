package apperror

import "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"

// Domain-specific sentinel errors.
var (
	ErrNotFound            = apperror.New(apperror.CodeNotFound, "not found")
	ErrAlreadyExists       = apperror.New(apperror.CodeConflict, "already exists")
	ErrUnauthorized        = apperror.New(apperror.CodeUnauthorized, "unauthorized")
	ErrForbidden           = apperror.New(apperror.CodeForbidden, "forbidden")
	ErrSelfFollow          = apperror.New(apperror.CodeBadRequest, "cannot follow yourself")
	ErrAlreadyFollowing    = apperror.New(apperror.CodeConflict, "already following")
	ErrNotFollowing        = apperror.New(apperror.CodeBadRequest, "not following")
	ErrAddressLimitReached = apperror.New(apperror.CodeBadRequest, "address limit reached (max 10)")
	ErrAddressNotFound     = apperror.New(apperror.CodeNotFound, "address not found")
	ErrReviewAlreadyExists = apperror.New(apperror.CodeConflict, "review already exists for this order")
	ErrInvalidKYCState     = apperror.New(apperror.CodeBadRequest, "invalid KYC state transition")
	ErrAlreadySeller       = apperror.New(apperror.CodeConflict, "user is already a seller")
	ErrSellerNotFound      = apperror.New(apperror.CodeNotFound, "seller profile not found")
)
