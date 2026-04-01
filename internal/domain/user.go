package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string
type UserStatus string
type GenderType string
type KYCStatus string
type KYCDocType string
type KYCDocStatus string
type BusinessType string
type ReviewRole string

const (
	UserRoleBuyer  UserRole = "buyer"
	UserRoleSeller UserRole = "seller"
	UserRoleAdmin  UserRole = "admin"
)

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
)

const (
	GenderMale        GenderType = "male"
	GenderFemale      GenderType = "female"
	GenderOther       GenderType = "other"
	GenderUndisclosed GenderType = "undisclosed"
)

const (
	KYCStatusNone     KYCStatus = "none"
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusApproved KYCStatus = "approved"
	KYCStatusRejected KYCStatus = "rejected"
)

const (
	KYCDocTypeIDCardFront     KYCDocType = "id_card_front"
	KYCDocTypeIDCardBack      KYCDocType = "id_card_back"
	KYCDocTypeBusinessLicense KYCDocType = "business_license"
	KYCDocTypeSelfieWithID    KYCDocType = "selfie_with_id"
)

const (
	KYCDocStatusPending  KYCDocStatus = "pending"
	KYCDocStatusApproved KYCDocStatus = "approved"
	KYCDocStatusRejected KYCDocStatus = "rejected"
)

const (
	BusinessTypeIndividual BusinessType = "individual"
	BusinessTypeBusiness   BusinessType = "business"
)

const (
	ReviewRoleBuyer  ReviewRole = "buyer"
	ReviewRoleSeller ReviewRole = "seller"
)

type User struct {
	ID             uuid.UUID  `json:"id"`
	KeycloakID     string     `json:"keycloak_id"`
	Email          string     `json:"email"`
	UserName       string     `json:"username"`
	FullName       string     `json:"full_name"`
	Phone          string     `json:"phone"`
	AvatarURL      string     `json:"avatar_url"`
	CoverURL       string     `json:"cover_url"`
	Bio            string     `json:"bio"`
	Gender         GenderType `json:"gender"`
	DateOfBirth    *time.Time `json:"date_of_birth"`
	Role           UserRole   `json:"role"`
	Status         UserStatus `json:"status"`
	EmailVerified  bool       `json:"email_verified"`
	TrustScore     float64    `json:"trust_score"`
	FollowerCount  int        `json:"follower_count"`
	FollowingCount int        `json:"following_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}
