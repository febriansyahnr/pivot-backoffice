package user

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type ListUsersByMerchantIDRequest struct {
	MerchantID     string     `json:"merchantId"`
	StartCreatedAt *time.Time `json:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt"`
	Name           string     `json:"name"`
	RoleID         string     `json:"roleId"`
	SortBy         string     `json:"sortBy"`
	SortOrder      string     `json:"sortOrder"`
}

func (req *ListUsersByMerchantIDRequest) Validate() error {
	if req.SortBy != "" {
		if req.SortBy != constant.UserSortColName {
			return constant.ErrInvalidUserListSortColumn
		}
	}

	if req.SortOrder != "" {
		if err := commonModel.ValidateSortOrder(req.SortOrder); err != nil {
			return err
		}
	}

	return nil
}

type CreatePinRequest struct {
	Pin string `json:"pin" validate:"required,numeric,len=6"`
}

type ResetPinRequest CreatePinRequest

type CheckPinRequest struct {
	Pin string `json:"pin" validate:"required,numeric,len=6"`
}

type ChangePinRequest struct {
	Pin    string `json:"pin" validate:"required,numeric,len=6"`
	NewPin string `json:"newPin" validate:"required,numeric,len=6"`
}

type ValidateInvitationRequest struct {
	Token string `json:"token" validate:"required"` // Activation token, generated from create new member.
}

type ActivateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	PIN      string `json:"pin" validate:"required,numeric,len=6"`
	Token    string `json:"-"` // Activation token, generated from create new member.
}

type SendGeneratedInvitationRequest struct {
	Inviter      string `json:"-"`
	Email        string `json:"-"`
	MerchantName string `json:"-"`
	MerchantID   string `json:"-"`
	UserID       string `json:"-"`
	UserName     string `json:"-"`
	IsResend     bool   `json:"-"`
}

type ResendEmailInvitationRequest struct {
	Email string `json:"email" validate:"required,email"`
}
