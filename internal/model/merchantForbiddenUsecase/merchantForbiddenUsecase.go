package merchantForbiddenUsecase

import (
	"database/sql"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
)

var (
	allowedUsecases = []string{
		constant.ReferenceDisbursement,
		constant.ReferenceWithdrawal,
	}
)

type MerchantForbiddenUseCase struct {
	UUID       string       `json:"uuid" db:"uuid"`
	MerchantID string       `json:"merchantId" db:"merchant_id"`
	UseCase    string       `json:"usecase" db:"use_case"`
	CreatedAt  time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time    `json:"updatedAt" db:"updated_at"`
	DeletedAt  sql.NullTime `json:"deletedAt" db:"deleted_at"`
}

type NewMerchantForbiddenUseCaseRequest struct {
	MerchantID string
	UseCase    string
}

type GetMerchantForbiddenUseCaseRequest struct {
	MerchantID string
	UseCase    string
}

type MerchantForbiddenUseCaseRequest struct {
	Requester  string
	MerchantID string `json:"merchantId" validate:"required,uuid"`
	UseCase    string `json:"usecase" validate:"required,oneof=DISBURSEMENT WITHDRAWAL"`
	SetStatus  bool   `json:"setStatus"`
}

func NewMerchantForbiddenUseCase(req *NewMerchantForbiddenUseCaseRequest) *MerchantForbiddenUseCase {
	return &MerchantForbiddenUseCase{
		UUID:       uuid.New().String(),
		MerchantID: req.MerchantID,
		UseCase:    req.UseCase,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
}

func IsUseCaseExists(useCase string) bool {
	return slices.Contains(allowedUsecases, strings.ToUpper(useCase))
}
