package settlementHold

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type SettlementHold struct {
	UUID       string       `db:"uuid"`
	MerchantID string       `db:"merchant_id"`
	PaymentID  string       `db:"payment_id"`
	Status     string       `db:"status"`
	CreatedBy  string       `db:"created_by"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at"`
	DeletedAt  sql.NullTime `db:"deleted_at"`
}

type SettlementHoldHistory struct {
	UUID             string       `db:"uuid"`
	SettlementHoldID string       `db:"settlement_hold_id"`
	Status           string       `db:"status"`
	Reason           string       `db:"reason"`
	CreatedBy        string       `db:"created_by"`
	CreatedAt        time.Time    `db:"created_at"`
	DeletedAt        sql.NullTime `db:"deleted_at"`
}

type CreateUpdateSettlementHoldRequest struct {
	MerchantID string `json:"merchantId"`
	PaymentID  string `json:"paymentId" validate:"required"`
	Action     string `json:"action" validate:"required,oneof=HOLD RELEASE"`
	Reason     string `json:"reason" validate:"required"`
	CreatedBy  string `json:"createdBy" validate:"required"`
}

type CreateUpdateSettlementHoldResponse struct {
	UUID       string    `json:"uuid"`
	MerchantID string    `json:"merchantId"`
	PaymentID  string    `json:"paymentId"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
	CreatedBy  string    `json:"createdBy"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedBy  string    `json:"updatedBy"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func New(req *CreateUpdateSettlementHoldRequest) (*SettlementHold, *SettlementHoldHistory) {
	now := time.Now().UTC()
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}
	data := &SettlementHold{
		UUID:       id.String(),
		MerchantID: req.MerchantID,
		PaymentID:  req.PaymentID,
		Status:     req.Action,
		CreatedBy:  req.CreatedBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	historyId, err := uuid.NewV7()
	if err != nil {
		historyId = uuid.New()
	}
	dataHistory := &SettlementHoldHistory{
		UUID:             historyId.String(),
		SettlementHoldID: data.UUID,
		Status:           data.Status,
		Reason:           req.Reason,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        now,
	}

	return data, dataHistory
}

func (s *SettlementHold) Update(req *CreateUpdateSettlementHoldRequest) *SettlementHoldHistory {
	now := time.Now().UTC()
	historyId, err := uuid.NewV7()
	if err != nil {
		historyId = uuid.New()
	}
	s.Status = req.Action
	s.UpdatedAt = now

	dataHistory := &SettlementHoldHistory{
		UUID:             historyId.String(),
		SettlementHoldID: s.UUID,
		Status:           req.Action,
		Reason:           req.Reason,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        now,
	}

	return dataHistory
}
