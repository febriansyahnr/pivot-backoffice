package installmentPlanModel

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/creditcardCoreProcessor"
)

type CardInstallmentMetadata struct {
	MidID         string   `json:"midId"`
	Mid           string   `json:"mid"`
	AllowedBins   []string `json:"allowedBins"`
	Interest      float64  `json:"interest"`
	MinimumAmount float64  `json:"minimumAmount"`
	MaximumAmount float64  `json:"maximumAmount"`
}

type InstallmentPlanMetadata struct {
	Card *CardInstallmentMetadata `json:"card,omitempty"`
}

type InstallmentPlan struct {
	UUID            string         `json:"uuid" db:"uuid"`
	MerchantID      string         `json:"merchantId" db:"merchant_id"`
	Acquirer        string         `json:"acquirer" db:"acquirer"`
	SettlementType  string         `json:"settlementType" db:"settlement_type"`
	InstallmentType string         `json:"installmentType" db:"installment_type"`
	PaymentMethod   string         `json:"paymentMethod" db:"payment_method"`
	Title           string         `json:"title" db:"title"`
	Description     string         `json:"description" db:"description"`
	Tenor           int            `json:"tenor" db:"tenor"`
	Status          string         `json:"status" db:"status"`
	Metadata        types.JSONText `json:"metadata" db:"metadata"`
	CreatedAt       time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time      `json:"updatedAt" db:"updated_at"`
	DeletedAt       sql.NullTime   `json:"deletedAt" db:"deleted_at"`

	PlanMetadata *InstallmentPlanMetadata `json:"planMetadata"`
}

type CreateInstallmentPlanRequest struct {
	MerchantID     string                      `json:"merchantId"`
	Acquirer       string                      `json:"acquirer" validate:"required"`
	SettlementType string                      `json:"settlementType" validate:"required,oneof=AGGREGATOR DIRECT"`
	PaymentMethod  string                      `json:"paymentMethod" validate:"required,oneof=CARD"`
	Title          string                      `json:"title" validate:"required"`
	Description    string                      `json:"description" validate:"required"`
	Tenor          int                         `json:"tenor" validate:"required,gte=1"`
	CardDetail     *CardInstallmentPlanRequest `json:"cardDetail" validate:"required_if=PaymentMethod CARD"`
}

type CardInstallmentPlanRequest struct {
	MidID              string `json:"midId" validate:"required,uuid"`
	Mid                string
	MidInstallmentType string
	AllowedBins        []string `json:"allowedBins" validate:"required,min=1"`
	Interest           float64  `json:"interest"`
	MinimumAmount      float64  `json:"minimumAmount" validate:"required,gte=1"`
	MaximumAmount      float64  `json:"maximumAmount" validate:"required,gte=1"`
}

func New(req *CreateInstallmentPlanRequest) *InstallmentPlan {
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}

	installmentPlan := &InstallmentPlan{
		UUID:           id.String(),
		MerchantID:     req.MerchantID,
		Acquirer:       req.Acquirer,
		PaymentMethod:  req.PaymentMethod,
		SettlementType: req.SettlementType,
		Title:          req.Title,
		Description:    req.Description,
		Tenor:          req.Tenor,
		Status:         constant.InstallmentPlanStatusActive,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	metadata := InstallmentPlanMetadata{}
	if req.CardDetail != nil {
		metadata.Card = &CardInstallmentMetadata{
			MidID:         req.CardDetail.MidID,
			Mid:           req.CardDetail.Mid,
			AllowedBins:   req.CardDetail.AllowedBins,
			Interest:      req.CardDetail.Interest,
			MinimumAmount: req.CardDetail.MinimumAmount,
			MaximumAmount: req.CardDetail.MaximumAmount,
		}

		if req.CardDetail.MidInstallmentType != "" {
			installmentPlan.InstallmentType = req.CardDetail.MidInstallmentType
		}
	}
	installmentPlan.PlanMetadata = &metadata

	metadataJSON, _ := json.Marshal(metadata)
	installmentPlan.Metadata = metadataJSON

	return installmentPlan
}

type UpdateInstallmentPlanRequest struct {
	UUID           string                            `json:"uuid"`
	MerchantID     string                            `json:"merchantId"`
	Acquirer       string                            `json:"acquirer"`
	SettlementType string                            `json:"settlementType" validate:"omitempty,oneof=AGGREGATOR DIRECT"`
	PaymentMethod  string                            `json:"paymentMethod" validate:"omitempty,oneof=CARD"`
	Title          string                            `json:"title"`
	Description    string                            `json:"description"`
	Tenor          *int                              `json:"tenor" validate:"omitempty,gte=1"`
	Status         string                            `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE"`
	CardDetail     *UpdateCardInstallmentPlanRequest `json:"cardDetail"`
}

type UpdateCardInstallmentPlanRequest struct {
	MidID         string   `json:"midId" validate:"omitempty,uuid"`
	AllowedBins   []string `json:"allowedBins"`
	Interest      *float64 `json:"interest"`
	MinimumAmount float64  `json:"minimumAmount" validate:"omitempty,gte=1"`
	MaximumAmount float64  `json:"maximumAmount" validate:"omitempty,gte=1"`
}

func (i *InstallmentPlan) LoadMetadata() error {
	return json.Unmarshal(i.Metadata, &i.PlanMetadata)
}

func (i *InstallmentPlan) Update(req *UpdateInstallmentPlanRequest) error {

	if req.Acquirer != "" {
		i.Acquirer = req.Acquirer
	}
	if req.SettlementType != "" {
		i.SettlementType = req.SettlementType
	}
	if req.PaymentMethod != "" {
		i.PaymentMethod = req.PaymentMethod
	}
	if req.Title != "" {
		i.Title = req.Title
	}
	if req.Description != "" {
		i.Description = req.Description
	}
	if req.Tenor != nil && *req.Tenor != i.Tenor {
		i.Tenor = *req.Tenor
	}
	if req.Status != "" {
		i.Status = req.Status
	}
	if req.CardDetail != nil {
		if err := i.LoadMetadata(); err != nil {
			return err
		}

		if req.CardDetail.MidID != "" {
			i.PlanMetadata.Card.MidID = req.CardDetail.MidID
		}

		if req.CardDetail.AllowedBins != nil {
			i.PlanMetadata.Card.AllowedBins = req.CardDetail.AllowedBins
		}

		if req.CardDetail.Interest != nil && *req.CardDetail.Interest != i.PlanMetadata.Card.Interest {
			i.PlanMetadata.Card.Interest = *req.CardDetail.Interest
		}

		if req.CardDetail.MinimumAmount != 0 {
			i.PlanMetadata.Card.MinimumAmount = req.CardDetail.MinimumAmount
		}

		if req.CardDetail.MaximumAmount != 0 {
			i.PlanMetadata.Card.MaximumAmount = req.CardDetail.MaximumAmount
		}

		metadataJSON, _ := json.Marshal(i.PlanMetadata)
		i.Metadata = metadataJSON
	}

	if i.Status == constant.InstallmentPlanStatusActive && len(i.PlanMetadata.Card.AllowedBins) == 0 {
		return constant.ErrActiveInstallmentEmptyBins
	}

	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *InstallmentPlan) UpdateMIDInfo(mid *creditcardCoreProcessorModel.MIDResponseData) {
	i.PlanMetadata.Card.Mid = mid.Mid
	i.InstallmentType = mid.InstallmentType
	metadataJSON, _ := json.Marshal(i.PlanMetadata)
	i.Metadata = metadataJSON
}

func (i *InstallmentPlan) GetStringTenor() string {
	return fmt.Sprintf("%dMO", i.Tenor)
}

type InstallmentPlanResponse struct {
	UUID            string       `json:"uuid"`
	MerchantID      string       `json:"merchantId"`
	Acquirer        string       `json:"acquirer"`
	SettlementType  string       `json:"settlementType"`
	InstallmentType string       `json:"installmentType"`
	PaymentMethod   string       `json:"paymentMethod"`
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	Tenor           int          `json:"tenor"`
	Status          string       `json:"status"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
	DeletedAt       sql.NullTime `json:"deletedAt"`

	PlanMetadata *InstallmentPlanMetadata `json:"planMetadata"`
}

func (i *InstallmentPlan) ToResponseModel() *InstallmentPlanResponse {
	return &InstallmentPlanResponse{
		UUID:            i.UUID,
		MerchantID:      i.MerchantID,
		Acquirer:        i.Acquirer,
		SettlementType:  i.SettlementType,
		InstallmentType: i.InstallmentType,
		PaymentMethod:   i.PaymentMethod,
		Title:           i.Title,
		Description:     i.Description,
		Tenor:           i.Tenor,
		Status:          i.Status,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		DeletedAt:       i.DeletedAt,
		PlanMetadata:    i.PlanMetadata,
	}
}

type ValidateCardInstallmentPlanRequest struct {
	MidId          string
	Tenor          int
	SettlementType string
	AllowedBins    []string
}
