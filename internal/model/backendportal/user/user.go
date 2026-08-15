package user

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	jwt "github.com/golang-jwt/jwt/v5"
)

type User struct {
	UUID             string                     `json:"uuid" db:"uuid"`
	Email            string                     `json:"email" db:"email"`
	Status           string                     `json:"status" db:"status"`
	Name             string                     `json:"name" db:"name"`
	Blocked          sql.NullTime               `json:"blockedAt" db:"blocked_at"`
	MerchantId       string                     `json:"merchantId" db:"merchant_id"`
	MerchantName     string                     `json:"merchantName,omitempty" db:"merchant_name"`
	RefreshToken     sql.NullString             `json:"refreshToken" db:"refresh_token"`
	IsChangePassword int                        `json:"isChangePassword" db:"is_change_password"`
	Role             sql.NullString             `json:"role" db:"role"`
	LastLoginAt      commonModel.CustomNullTime `json:"lastLoginAt" db:"last_login_at"`
	DeactivatedAt    sql.NullTime               `json:"deactivatedAt" db:"deactivate_at"`
	CreatedAt        time.Time                  `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time                  `json:"updatedAt" db:"updated_at"`
	DeletedAt        sql.NullTime               `json:"deletedAt" db:"deleted_at"`

	Password           string         `json:"-" db:"password"`
	PinHash            sql.NullString `json:"-" db:"pin_hash"`
	TOTPWrappedSecret  sql.NullString `json:"-" db:"totp_wrapped_secret"`
	TOTPEncryptVersion sql.NullInt16  `json:"-" db:"totp_encrypt_version"`
	TOTPStatus         string         `json:"totpStatus" db:"totp_status"`
	Preferred2FAMethod string         `json:"preferred2FAMethod" db:"preferred_2fa_method"`
	RoleId             sql.NullString `json:"-" db:"role_id"`
	AsMerchantPIC      bool           `json:"-" db:"as_merchant_pic"`

	DeviceIdentifier string `json:"-"`

	MerchantStatus sql.NullString `json:"merchantStatus" db:"merchant_status"`
	ReasonStatus   sql.NullString `json:"reasonStatus" db:"reason_status"`

	jwt.RegisteredClaims
}

type UserRegisterRequest struct {
	Email                string    `json:"email" validate:"required,email"`
	Name                 string    `json:"name" validate:"required"`
	Password             string    `json:"password" validate:"required"`
	PasswordConfirmation string    `json:"passwordConfirmation" validate:"required,eqfield=Password"`
	CreatedAt            time.Time `json:"-"`
}

type UserCreateRequest struct {
	Email     string    `json:"email" validate:"required,email"`
	Name      string    `json:"name" validate:"required"`
	RoleSlug  string    `json:"roleSlug" validate:"required"`
	CreatedAt time.Time `json:"-"`
}

type MerchantUserRequest struct {
	Email        string
	Name         string
	Role         string
	MerchantId   string
	MerchantName string
	Invitation   bool
}

func NewMerchantUser(req *MerchantUserRequest) *User {
	user := &User{
		UUID:             uuid.New().String(),
		MerchantId:       req.MerchantId,
		Email:            strings.ToLower(req.Email),
		Status:           constant.UserStatusInvited,
		Name:             req.Name,
		IsChangePassword: 1,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	randomPassword, _ := util.GenerateRandomString(10)
	user.Password = user.EncryptPassword(randomPassword)

	if !req.Invitation {
		user.Password = ""
	}
	return user
}

type CRMUserCreateRequest struct {
	Email      string    `json:"email" validate:"required,email"`
	Name       string    `json:"name" validate:"required"`
	RoleSlug   string    `json:"roleSlug" validate:"required"`
	MerchantId string    `json:"merchantId"`
	CreatedAt  time.Time `json:"-"`
}

type UserUpdateRequest struct {
	Email     string    `json:"email" validate:"required,email"`
	Name      string    `json:"name" validate:"required"`
	RoleSlug  string    `json:"roleSlug" validate:"required"`
	CreatedAt time.Time `json:"-"`
}

type CRMUserUpdateRequest struct {
	Email     string    `json:"email" validate:"required,email"`
	Name      string    `json:"name" validate:"required"`
	RoleSlug  string    `json:"roleSlug" validate:"required"`
	Status    string    `json:"status" validate:"required"`
	CreatedAt time.Time `json:"-"`
}

type UserForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UserResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

type User2FAResponse struct {
	Token               string                   `json:"token"`
	ResendDelaySeconds  int                      `json:"resendDelaySeconds,omitempty"`
	TwoFactorAuthMethod constant.TwoFactorMethod `json:"twoFactorAuthMethod"`
	IsTOTPActive        bool                     `json:"isTOTPActive"`
}

type UserLoginRequest struct {
	Email      string `json:"email" validate:"required,email" example:"test@gmail.com"`
	Password   string `json:"password" validate:"required" example:"GuaPukulLo10x"`
	IsRemember bool   `json:"isRemember" example:"true"`
}

type UserLogoutRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UserUnblockRequest struct {
	Email []string `json:"emails" validate:"min=1,dive"`
}

type UserAddRoleRequest struct {
	Email    string `json:"email" validate:"required,email"`
	RoleSlug string `json:"roleSlug" validate:"required"`
}

type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" validate:"required"`
	OldPassword string `json:"password" validate:"required"`
}

type CheckPasswordRequest struct {
	Password string `json:"password" validate:"required"`
}

type ChangePasswordResponse struct {
	Updated bool `json:"updated"`
}

type GenerateRandomPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}
type GenerateRandomPasswordResponse struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	UUID               string     `json:"uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email              string     `json:"email" example:"test@gmail.com"`
	Status             string     `json:"status" example:"ACTIVE"`
	Name               string     `json:"name" example:"John Doe"`
	Blocked            time.Time  `json:"blockedAt" example:"2021-01-01T00:00:00Z"`
	MerchantId         string     `json:"merchantId" example:"123e4567-e89b-12d3-a456-426614174000"`
	MerchantName       string     `json:"merchantName,omitempty" example:"Merchant Example"`
	IsChangePassword   int        `json:"isChangePassword" example:"0"`
	IsEmptyPin         int        `json:"isEmptyPin" example:"0"`
	Role               string     `json:"role" example:"admin"`
	TOTPStatus         string     `json:"totpStatus"`
	Preferred2FAMethod string     `json:"preferred2FAMethod"`
	DeactivatedAt      string     `json:"deactivatedAt" example:"2021-01-01T00:00:00Z"`
	LastChangePassword *time.Time `json:"lastChangePassword,omitempty" example:"2021-01-01T00:00:00Z"`
	CreatedAt          time.Time  `json:"createdAt" example:"2021-01-01T00:00:00Z"`
	UpdatedAt          time.Time  `json:"updatedAt" example:"2021-01-01T00:00:00Z"`
}

type UserLoggedInResponse struct {
	UserInfo     *UserResponse `json:"userInfo"`
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
}

type UserRefreshTokenRequest struct {
	Email        string `json:"email" validate:"required,email"`
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// ActivateDeactivateRequest holds the validated data needed to activate/deactivate a user.
type ActivateDeactivateRequest struct {
	UserID         string
	User           *User
	UserTokenClaim *UserTokenClaims // actor token claim that do request
}

type UserTokenClaims struct {
	jwt.RegisteredClaims

	UUID         string    `json:"uuid"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	RoleID       string    `json:"roleId"`
	Blocked      time.Time `json:"blockedAt"`
	MerchantId   string    `json:"merchantId"`
	RefreshToken string    `json:"refreshToken"`

	DeviceIdentifier string `json:"deviceIdentifier"`
}

// ToResponse convert User to UserResponse
func (u *User) ToResponse() *UserResponse {
	var (
		role, deactivatedAt string
		blockedAt           time.Time
		isEmptyPin          = 1
	)

	if u.Role.Valid {
		role = u.Role.String
	}

	if u.Blocked.Valid {
		blockedAt = u.Blocked.Time
	}

	if u.PinHash.Valid {
		isEmptyPin = 0
	}

	if !u.DeactivatedAt.Time.IsZero() {
		deactivatedAt = u.DeactivatedAt.Time.String()
	}

	return &UserResponse{
		UUID:               u.UUID,
		Email:              u.Email,
		Status:             u.Status,
		Name:               u.Name,
		Blocked:            blockedAt,
		MerchantId:         u.MerchantId,
		MerchantName:       u.MerchantName,
		IsChangePassword:   u.IsChangePassword,
		IsEmptyPin:         isEmptyPin,
		Role:               role,
		DeactivatedAt:      deactivatedAt,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
		TOTPStatus:         u.TOTPStatus,
		Preferred2FAMethod: u.Preferred2FAMethod,
	}
}

// EncryptPassword encrypt password using built-in sha256
func (u *User) EncryptPassword(pass string) (hashed string) {
	hasher := sha256.New()
	hasher.Write([]byte(pass))
	hash := hasher.Sum(nil)

	return hex.EncodeToString(hash)
}

// ComparePassword compare password with encrypted password
func (u *User) ComparePassword(pass string) bool {
	return u.EncryptPassword(pass) == u.Password
}
