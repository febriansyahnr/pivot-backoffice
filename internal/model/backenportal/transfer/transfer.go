package transfer

import (
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
)

var (
	SortColCreatedAt = "createdAt"
	SortColAmount    = "amount"
	SortColRecipient = "recipientName"
	SortColSender    = "senderName"
)

type Transfer struct {
	UUID              uuid.UUID  `json:"uuid" db:"uuid"`
	MerchantID        uuid.UUID  `json:"merchantId" db:"merchant_id"`
	RecipientID       uuid.UUID  `json:"recipientId" db:"recipient_id"`
	ReferenceID       string     `json:"referenceId" db:"reference_id"`
	TransferType      string     `json:"transferType" db:"transfer_type"`
	Direction         string     `json:"-" db:"direction"` // income outcome, income when the recipient is the merchant
	Currency          string     `json:"currency" db:"currency"`
	Amount            float64    `json:"amount" db:"amount"`
	Status            string     `json:"status" db:"status"`
	Remarks           string     `json:"remarks" db:"remarks"`
	ReasonDescription string     `json:"reasonDescription" db:"reason_description"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt         *time.Time `json:"deletedAt" db:"deleted_at"`
	Beneficiary       string     `json:"beneficiary" db:"beneficiary"`
}

type TransferRequest struct {
	SourceMerchantID uuid.UUID `validate:"required"`
	RecipientID      string    `json:"recipientId" validate:"required,uuid"`
	ReferenceID      string    `json:"referenceId" validate:"required,maxChar=100"`
	TransferType     string    `json:"transferType" validate:"required,oneof=DIRECT"`
	Amount           float64   `json:"amount" validate:"required,min=1"`
	Remarks          string    `json:"remarks" validate:"maxChar=100"`
	ParentMerchantID uuid.UUID `json:"-"`
	Usecase          string    `json:"usecase"`
}

type GetTransferTransactionRequest struct {
	ParentID      string
	MerchantID    string
	TransactionID string
}

type TransferResponse struct {
	UUID         string    `json:"uuid"`
	RecipientID  string    `json:"recipientId"`
	ReferenceID  string    `json:"referenceId"`
	TransferType string    `json:"transferType"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"`
	Remarks      string    `json:"remarks"`
	UpdatedAt    time.Time `json:"updatedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

type TransferTransactionDetail struct {
	UUID               string    `json:"uuid" db:"uuid"`
	ReferenceID        string    `json:"referenceId" db:"reference_id"` // it will contain payment ID
	SenderID           string    `json:"senderId" db:"sender_id"`
	SenderName         string    `json:"senderName" db:"sender_name"`
	RecipientID        string    `json:"recipientId" db:"recipient_id"`
	RecipientName      string    `json:"recipientName" db:"recipient_name"`
	Type               string    `json:"type" db:"type"`
	Currency           string    `json:"currency" db:"currency"`
	Amount             float64   `json:"amount" db:"amount"`
	Status             string    `json:"status" db:"status"`
	Remarks            string    `json:"remarks" db:"remarks"`
	PaymentID          *string   `json:"paymentId" db:"payment_id"`
	FeeAmount          float64   `json:"feeAmount" db:"fee_amount"`
	FeeCurrency        *string   `json:"feeCurrency" db:"fee_currency"`
	PaymentReferenceID *string   `json:"paymentReferenceId" db:"payment_reference_id"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
}

func (t *Transfer) ToTransferResponse() *TransferResponse {
	return &TransferResponse{
		UUID:         t.UUID.String(),
		RecipientID:  t.RecipientID.String(),
		ReferenceID:  t.ReferenceID,
		TransferType: t.TransferType,
		Amount:       t.Amount,
		Status:       t.Status,
		Remarks:      t.Remarks,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

type ListTransferResponse struct {
	UUID          string    `json:"uuid"`
	ReferenceID   string    `json:"referenceId"`
	Type          string    `json:"type"`
	SenderID      string    `json:"senderId"`
	SenderName    string    `json:"senderName"`
	RecipientID   string    `json:"recipientId"`
	RecipientName string    `json:"recipientName"`
	TransferType  string    `json:"transferType"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	Remarks       string    `json:"remarks"`
	UpdatedAt     time.Time `json:"updatedAt"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (t *Transfer) ToListTransferResponse(merchantId string, participantNameMap map[string]string) *ListTransferResponse {
	data := &ListTransferResponse{
		UUID:          t.UUID.String(),
		ReferenceID:   t.ReferenceID,
		Type:          t.Direction,
		SenderID:      t.MerchantID.String(),
		SenderName:    participantNameMap[t.MerchantID.String()],
		RecipientID:   t.RecipientID.String(),
		RecipientName: participantNameMap[t.RecipientID.String()],
		TransferType:  t.TransferType,
		Amount:        t.Amount,
		Status:        t.Status,
		Remarks:       t.Remarks,
		UpdatedAt:     t.UpdatedAt,
		CreatedAt:     t.CreatedAt,
	}

	if data.Type == "" {
		data.Type = constant.TransferTypeIN
	}

	return data
}

func (r *TransferRequest) Validate() error {
	if r.SourceMerchantID == uuid.Nil || r.ParentMerchantID == uuid.Nil || r.RecipientID == "" {
		return constant.ErrInvalidMerchantId
	}

	if r.SourceMerchantID == util.ParseUUID(r.RecipientID) {
		return constant.ErrSameMerchant
	}

	if r.TransferType != constant.MoneyFlowDirect && r.TransferType != constant.MoneyFlowIndirect {
		return constant.ErrInvalidTransferType
	}

	if r.Amount <= 0 {
		return constant.ErrInvalidAmount
	}

	return nil
}

func NewTransfer(request *TransferRequest) (*Transfer, error) {
	err := request.Validate()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Transfer{
		UUID:         uuid.New(),
		MerchantID:   request.SourceMerchantID,
		RecipientID:  util.ParseUUID(request.RecipientID),
		ReferenceID:  request.ReferenceID,
		TransferType: request.TransferType,
		Currency:     constant.CurrencyIDR,
		Amount:       request.Amount,
		Status:       constant.TransferStatusPending,
		Remarks:      request.Remarks,
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
	}, nil
}

func (t *Transfer) Update(status, reason string) {
	t.Status = status
	t.ReasonDescription = reason
	t.UpdatedAt = time.Now().UTC()
}

type GetTransferListRequest struct {
	UUID               string    `validate:"-"`
	MerchantID         string    `validate:"required,uuid"`
	ParentID           string    `validate:"-"`
	PaymentReferenceID string    `validate:"-"`
	PaymentID          string    `validate:"-"`
	ReferenceID        string    `validate:"-"`
	Type               string    `validate:"-"`
	StartDate          time.Time `validate:"-"`
	EndDate            time.Time `validate:"-"`
	Status             string    `validate:"-"`
	SortBy             string    `validate:"-"`
	SortOrder          string    `validate:"-"`
	Page               int64     `validate:"-"`
	PerPage            int64     `validate:"-"`
	StrStartDate       string    `validate:"required_with=StrEndDate,omitempty,iso_8601_datetime" name:"StartDate"`                         // Optional time range
	StrEndDate         string    `validate:"required_with=StrStartDate,omitempty,iso_8601_datetime,gtecsfield=StrStartDate" name:"EndDate"` // Optional time range
}

func (r *GetTransferListRequest) ValidateAndAdjust() error {
	if r.MerchantID == "" {
		return constant.ErrInvalidMerchantId
	}

	if r.Type != "" {
		r.Type = strings.ToUpper(r.Type)
		if r.Type != constant.TransferTypeIN && r.Type != constant.TransferTypeOUT {
			return constant.ErrInvalidTransferType
		}
	}

	if r.Status != "" {
		r.Status = strings.ToUpper(r.Status)
		if r.Status != constant.StatusPending && r.Status != constant.StatusSuccess && r.Status != constant.StatusFailed {
			return constant.ErrInvalidTransferStatus
		}
	}

	if r.SortBy != "" {
		if r.SortBy != SortColCreatedAt && r.SortBy != SortColAmount && r.SortBy != SortColRecipient && r.SortBy != SortColSender {
			return constant.ErrInvalidTransferSortColumn
		}
	}

	if r.StartDate.IsZero() && r.EndDate.IsZero() {
		r.EndDate = time.Now().UTC()
		r.StartDate = r.EndDate.AddDate(0, 0, -constant.TransferDefaultRangeDaysDuration)
	}
	if r.StartDate.IsZero() {
		r.StartDate = r.EndDate.AddDate(0, 0, -constant.TransferDefaultRangeDaysDuration)
	}
	if r.EndDate.IsZero() || r.StartDate.After(r.EndDate) {
		r.EndDate = r.StartDate.AddDate(0, 0, constant.TransferMaxRangeDaysDuration)
	}

	return nil
}
