package apperror

import "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"

// Domain-specific sentinel errors.
var (
	ErrNotFound            = apperror.New(apperror.CodeNotFound, "không tìm thấy")
	ErrAlreadyExists       = apperror.New(apperror.CodeConflict, "đã tồn tại")
	ErrUnauthorized        = apperror.New(apperror.CodeUnauthorized, "chưa xác thực")
	ErrForbidden           = apperror.New(apperror.CodeForbidden, "không có quyền truy cập")
	ErrSelfFollow          = apperror.New(apperror.CodeBadRequest, "không thể theo dõi chính mình")
	ErrAlreadyFollowing    = apperror.New(apperror.CodeConflict, "đã theo dõi người dùng này")
	ErrNotFollowing        = apperror.New(apperror.CodeBadRequest, "chưa theo dõi người dùng này")
	ErrAddressLimitReached = apperror.New(apperror.CodeBadRequest, "đã đạt giới hạn địa chỉ (tối đa 10)")
	ErrAddressNotFound     = apperror.New(apperror.CodeNotFound, "không tìm thấy địa chỉ")
	ErrReviewAlreadyExists = apperror.New(apperror.CodeConflict, "đơn hàng này đã được đánh giá")
	ErrInvalidKYCState     = apperror.New(apperror.CodeBadRequest, "trạng thái KYC không hợp lệ")
	ErrAlreadySeller       = apperror.New(apperror.CodeConflict, "người dùng đã là người bán")
	ErrSellerNotFound      = apperror.New(apperror.CodeNotFound, "không tìm thấy hồ sơ người bán")
)
