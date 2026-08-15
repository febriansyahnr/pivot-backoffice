package ratelimiter

import (
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type RateLimit struct {
	Attribute            string
	IsCheckResultCorrect bool
	FeatureName          string
	Timestamp            time.Time
}

type RateLimitConfiguration struct {
	UUID              string     `json:"uuid" db:"uuid"`
	MerchantID        string     `json:"merchantId" db:"merchant_id"`
	Limit             int        `json:"limit" db:"limit"`
	Burst             int        `json:"burst" db:"burst"` // Allowed number of requests after limit
	Order             int        `json:"order" db:"order"`
	Time              string     `json:"time" db:"time"`         // ENUM('SECOND', 'MINUTE', 'HOUR', 'DAILY')
	Variable          string     `json:"variable" db:"variable"` // ENUM('IP', 'PATH')
	VariableValue     string     `json:"variableValue" db:"variable_value"`
	VariableMatchType string     `json:"variableMatchType" db:"variable_match_type"` // ENUM('EXACT', 'CONTAINS', 'PREFIX')
	HTTPMethod        string     `json:"httpMethod" db:"http_method"`                // ENUM('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD', 'TRACE', 'CONNECT')
	Status            string     `json:"status" db:"status"`
	Description       string     `json:"description" db:"description"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty" db:"deleted_at"` // Nullable
}

type MerchantRateLimitHeaderMetadata struct {
	RateLimitLimit     int
	RateLimitRemaining int
	RateLimitReset     int64
}

type MerchantRateLimitRequest struct {
	MerchantID    string
	Path          string
	IPAddress     string
	Status        string
	Variable      string
	VariableValue string
	HTTPMethod    string
	Page          int64
	PageSize      int64
}

type MerchantRateLimitConfig struct {
	UUID              string `json:"uuid" db:"uuid"`
	Variable          string `json:"variable" db:"variable"`
	VariableValue     string `json:"variable_value" db:"variable_value"`
	Limit             int    `json:"limit" db:"limit"`
	Burst             int    `json:"burst" db:"burst"`
	VariableMatchType string `json:"variable_match_type" db:"variable_match_type"`
	HTTPMethod        string `json:"http_method" db:"http_method"`
	Time              string `json:"time" db:"time"`
}

func (r *MerchantRateLimitConfig) IsExactType() bool {
	return r.VariableMatchType == constant.RateLimitConfigVariableMatchTypeExact
}

func (r *MerchantRateLimitConfig) IsContainsType() bool {
	return r.VariableMatchType == constant.RateLimitConfigVariableMatchTypeContains
}

func (r *MerchantRateLimitConfig) IsPrefixType() bool {
	return r.VariableMatchType == constant.RateLimitConfigVariableMatchTypePrefix
}

func (r *MerchantRateLimitConfig) GetDuration() time.Duration {
	switch r.Time {
	case constant.RateLimitConfigTimeSecond:
		return constant.RateLimitConfigTimeSecondDuration
	case constant.RateLimitConfigTimeMinute:
		return constant.RateLimitConfigTimeMinuteDuration
	case constant.RateLimitConfigTimeHour:
		return constant.RateLimitConfigTimeHourDuration
	case constant.RateLimitConfigTimeDaily:
		return constant.RateLimitConfigTimeDailyDuration
	default:
		return constant.RateLimitConfigTimeSecondDuration
	}
}

type CreateRateLimitConfiguration struct {
	MerchantID        string `json:"merchantId" validate:"required,uuid"`
	Limit             int    `json:"limit" validate:"required,gte=1"`
	Order             int    `json:"order" validate:"required,gte=0"`
	Burst             int    `json:"burst" validate:"required,gte=0"`
	Time              string `json:"time" validate:"required,oneof=SECOND MINUTE HOUR DAILY"`
	Variable          string `json:"variable" validate:"required,oneof=IP PATH"`
	VariableValue     string `json:"variableValue" validate:"required"`
	VariableMatchType string `json:"variableMatchType" validate:"required,oneof=EXACT CONTAINS PREFIX"`
	HTTPMethod        string `json:"httpMethod" validate:"required,oneof=GET POST PUT DELETE PATCH OPTIONS HEAD TRACE CONNECT"`
	Description       string `json:"description"`
}

type UpdateRateLimitConfiguration struct {
	ID                string `json:"id" validate:"required,uuid"`
	MerchantID        string `json:"merchantId" validate:"required,uuid"`
	Limit             int    `json:"limit" validate:"required,gte=1"`
	Order             int    `json:"order" validate:"required,gte=0"`
	Burst             int    `json:"burst" validate:"required,gte=0"`
	Time              string `json:"time" validate:"required,oneof=SECOND MINUTE HOUR DAILY"`
	Variable          string `json:"variable" validate:"required,oneof=IP PATH"`
	VariableValue     string `json:"variableValue" validate:"required"`
	VariableMatchType string `json:"variableMatchType" validate:"required,oneof=EXACT CONTAINS PREFIX"`
	HTTPMethod        string `json:"httpMethod" validate:"required,oneof=GET POST PUT DELETE PATCH OPTIONS HEAD TRACE CONNECT"`
	Description       string `json:"description"`
	Status            string `json:"status" validate:"required,oneof=ACTIVE INACTIVE"`
}

// UUIDGenerator is a function type for generating UUIDs
type UUIDGenerator func() (uuid.UUID, error)

// defaultUUIDGenerator is the default implementation using uuid.NewV7
var defaultUUIDGenerator UUIDGenerator = uuid.NewV7

func New(req *CreateRateLimitConfiguration) (*RateLimitConfiguration, error) {
	id, err := defaultUUIDGenerator()
	if err != nil {
		return nil, err
	}

	return &RateLimitConfiguration{
		UUID:              id.String(),
		MerchantID:        req.MerchantID,
		Limit:             req.Limit,
		Burst:             req.Burst,
		Order:             req.Order,
		Time:              req.Time,
		Variable:          req.Variable,
		VariableValue:     req.VariableValue,
		VariableMatchType: req.VariableMatchType,
		HTTPMethod:        req.HTTPMethod,
		Status:            constant.StatusActive,
		Description:       req.Description,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}, nil
}

func (i *RateLimitConfiguration) Update(req *UpdateRateLimitConfiguration) error {

	i.MerchantID = req.MerchantID
	i.Limit = req.Limit
	i.Burst = req.Burst
	i.Order = req.Order
	i.Time = req.Time
	i.Variable = req.Variable
	i.VariableValue = req.VariableValue
	i.VariableMatchType = req.VariableMatchType
	i.HTTPMethod = req.HTTPMethod
	i.Status = req.Status
	i.Description = req.Description
	i.UpdatedAt = time.Now().UTC()

	return nil
}
