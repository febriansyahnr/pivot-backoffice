package platform

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
)

type GetSubMerchantUsersRequest struct {
	MerchantId       string
	ParentMerchantId string
	Keyword          string
	RoleId           string
	Page             int64
	PerPage          int64
	SortBy           string
	SortOrder        string
}

type SubMerchantUserResponse struct {
	UUID          string    `json:"uuid"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	AsMerchantPIC bool      `json:"asMerchantPIC"`
	Status        string    `json:"status"`
	LastLoginAt   time.Time `json:"lastLoginAt"`
}

func ToSubMerchantUserResponse(u *user.User) *SubMerchantUserResponse {
	return &SubMerchantUserResponse{
		UUID:          u.UUID,
		Name:          u.Name,
		Email:         u.Email,
		Role:          u.Role.String,
		AsMerchantPIC: u.AsMerchantPIC,
		Status:        u.Status,
		LastLoginAt:   u.LastLoginAt.Time,
	}
}
