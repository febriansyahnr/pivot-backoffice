package ipwhitelistModel

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

type IPWhitelistConfiguration struct {
	ID          string       `json:"id" db:"id"`
	MerchantID  string       `json:"merchantId" db:"merchant_id"`
	IP          string       `json:"ip" db:"ip"`
	Subnet      string       `json:"subnet" db:"subnet"`
	Priority    int          `json:"priority" db:"priority"`
	Action      string       `json:"action" db:"action"`
	Status      string       `json:"status" db:"status"`
	Description string       `json:"description" db:"description"`
	CreatedAt   time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time    `json:"updatedAt" db:"updated_at"`
	DeletedAt   sql.NullTime `json:"deletedAt" db:"deleted_at"`
}

type CreateIPWhitelistConfiguration struct {
	MerchantID  string `json:"merchantId" validate:"required,uuid"`
	IP          string `json:"ip" validate:"ip"`
	Subnet      string `json:"subnet" `
	Priority    int    `json:"priority" validate:"gte=0"`
	Action      string `json:"action" validate:"required,oneof=ALLOW BLOCK"`
	Description string `json:"description"`
}

type UpdateIPWhitelistConfiguration struct {
	ID          string `json:"id" validate:"required,uuid"`
	MerchantID  string `json:"merchantId" validate:"required,uuid"`
	IP          string `json:"ip" validate:"ip"`
	Subnet      string `json:"subnet"`
	Priority    int    `json:"priority" validate:"gte=0"`
	Action      string `json:"action" validate:"required,oneof=ALLOW BLOCK"`
	Status      string `json:"status" validate:"required,oneof=ACTIVE INACTIVE"`
	Description string `json:"description"`
}

type IPWhitelistConfigurationResponse struct {
	ID          string    `json:"id"`
	MerchantID  string    `json:"merchantId"`
	IP          string    `json:"ip"`
	Subnet      string    `json:"subnet"`
	Priority    int       `json:"priority"`
	Action      string    `json:"action"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt" `
	UpdatedAt   time.Time `json:"updatedAt"`
}

// UUIDGenerator is a function type for generating UUIDs
type UUIDGenerator func() (uuid.UUID, error)

// defaultUUIDGenerator is the default implementation using uuid.NewV7
var defaultUUIDGenerator UUIDGenerator = uuid.NewV7

func New(req *CreateIPWhitelistConfiguration) (*IPWhitelistConfiguration, error) {
	err := util.ValidateIPAddress(req.IP, req.Subnet)
	if err != nil {
		return nil, err
	}

	id, err := defaultUUIDGenerator()
	if err != nil {
		return nil, err
	}
	return &IPWhitelistConfiguration{
		ID:          id.String(),
		MerchantID:  req.MerchantID,
		IP:          req.IP,
		Subnet:      req.Subnet,
		Priority:    req.Priority,
		Action:      req.Action,
		Status:      constant.StatusActive,
		Description: req.Description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

func (i *IPWhitelistConfiguration) Update(req *UpdateIPWhitelistConfiguration) error {
	err := util.ValidateIPAddress(req.IP, req.Subnet)
	if err != nil {
		return err
	}

	i.IP = req.IP
	i.Subnet = req.Subnet
	i.Priority = req.Priority
	i.Action = req.Action
	i.Status = req.Status
	i.Description = req.Description
	i.UpdatedAt = time.Now().UTC()

	return nil
}

type GetIPWhitelistConfiguration struct {
	MerchantID  string
	IP          string
	Subnet      string `validate:"omitempty,cidrv4"`
	Status      string `validate:"omitempty,oneof=ACTIVE INACTIVE"`
	ExcludedIDs []string
	Page        int64
	PageSize    int64
}

func (i *IPWhitelistConfiguration) ToResponseModel() *IPWhitelistConfigurationResponse {
	return &IPWhitelistConfigurationResponse{
		ID:          i.ID,
		MerchantID:  i.MerchantID,
		IP:          i.IP,
		Subnet:      i.Subnet,
		Priority:    i.Priority,
		Action:      i.Action,
		Status:      i.Status,
		Description: i.Description,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}
