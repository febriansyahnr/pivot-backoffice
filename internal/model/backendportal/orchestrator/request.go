package orchestrator_model

import (
	"database/sql"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

const (
	TypePayment           = "PAYMENT"
	TypeDisbursement      = "DISBURSEMENT"
	TypeDisbursementTopUp = "DISBURSEMENT_TOP_UP"
	TypeMerchantTopUp     = "MERCHANT_TOP_UP"
	TypeFee               = "FEE"
)

type CreateAccountTransactionRequest struct {
	UUID                 uuid.UUID          `json:"uuid"`
	ReferenceID          string             `json:"reference_id"`
	MerchantID           uuid.UUID          `json:"merchant_id"`
	Currency             string             `json:"currency"`
	Credit               float64            `json:"credit"`
	Debit                float64            `json:"debit"`
	Type                 string             `json:"type"` // Type merupakan use case dari setiap transaksi yang dicatat
	Channel              string             `json:"channel"`
	Status               string             `json:"status"`
	Remarks              string             `json:"remarks"`
	TransactionTimestamp time.Time          `json:"transaction_timestamp"`
	AdditionalInfo       types.NullJSONText `json:"additionalInfo"`
	ReasonType           *string            `json:"reason_type,omitempty"`
	ReasonDescription    *string            `json:"reason_description,omitempty"`
	SettlementStatus     *string            `json:"settlementStatus"`
	SettlementAt         *time.Time         `json:"settlementAt"`
	SettlementModel      *string            `json:"settlementModel"`

	Usecase                string `json:"-"` // Usecase merupakan account balance (aka AccountName) yang digunakan setiap transaksi
	Processor              string `json:"-"` // Processor merupakan nama processor yang digunakan setiap transaksi
	ProcessorID            string `json:"-"` // ProcessorID merupakan ID dari master bank transfer, qris dan va
	ProcessorTransactionID string `json:"-"` // ProcessorTransactionID merupakan ID transaksi pada processor
	Reference              string `json:"-"` // Set the reference directly instead of using usecase name
}

func (req CreateAccountTransactionRequest) RoutingChannel(channel string) CreateAccountTransactionRequest {
	req.Channel = channel

	switch {
	case channel == constant.ChannelBalance:
		req.Type = constant.TypeDisbursement
	case channel == constant.ChannelVirtualAccount ||
		channel == constant.ChannelCreditCard ||
		channel == constant.ChannelBankTransfer:
		req.Type = constant.TypePayment
	default:
		// Need to defined later
	}

	return CreateAccountTransactionRequest(req)
}

func (req *CreateAccountTransactionRequest) ToAccountTransactionDTO(account *account_model.Account) *AccountTransaction {
	id := uuid.New()
	if req.UUID != uuid.Nil {
		id = req.UUID
	}

	accTrx := &AccountTransaction{
		UUID:                   id,
		ReferenceID:            req.ReferenceID,
		MerchantID:             req.MerchantID,
		AccountID:              account.UUID,
		Currency:               req.Currency,
		Credit:                 req.Credit,
		Debit:                  req.Debit,
		Type:                   req.Type,
		Reference:              req.Usecase,
		Channel:                req.Channel,
		Status:                 req.Status,
		Remarks:                req.Remarks,
		TransactionTimestamp:   req.TransactionTimestamp,
		AdditionalInfo:         req.AdditionalInfo,
		Processor:              req.Processor,
		ProcessorID:            req.ProcessorID,
		ProcessorTransactionID: req.ProcessorTransactionID,
	}
	if req.Usecase == "" {
		accTrx.Reference = account.Name
	}

	if req.Reference != "" {
		accTrx.Reference = req.Reference
	}

	if req.ReasonType != nil {
		accTrx.ReasonType = sql.NullString{
			Valid: true, String: *req.ReasonType,
		}
	}
	if req.ReasonDescription != nil {
		accTrx.ReasonDescription = sql.NullString{
			Valid: true, String: *req.ReasonDescription,
		}
	}
	if req.SettlementStatus != nil {
		accTrx.SettlementStatus = sql.NullString{
			Valid: true, String: *req.SettlementStatus,
		}
	}
	if req.SettlementAt != nil && !req.SettlementAt.IsZero() {
		accTrx.SettlementAt = sql.NullTime{
			Valid: true, Time: *req.SettlementAt,
		}
	}
	if req.SettlementModel != nil {
		accTrx.SettlementModel = sql.NullString{
			Valid: true, String: *req.SettlementModel,
		}
	}
	return accTrx
}

type TransactionHistoryFilterRequest struct {
	MerchantID             string    `json:"merchantId"`
	CreatedAt              string    `json:"createdAt"`
	ApprovedAt             string    `json:"approvedAt"`
	UpdatedAt              string    `json:"updatedAt"`
	TrxTypes               []string  `json:"trxTypes"`
	Status                 string    `json:"status"`
	Amount                 string    `json:"amount"`
	BeneficiaryAccountNo   string    `json:"beneficiaryAccountNo"`
	BeneficiaryAccountName string    `json:"beneficiaryAccountName"`
	BeneficiaryBankName    string    `json:"beneficiaryBankName"`
	TrxID                  string    `json:"trxId"`
	ProcessorReferenceID   string    `json:"trxRef"`
	BalanceTypes           []string  `json:"balanceTypes"`
	MerchantReferenceID    string    `json:"merchantReferenceId"`
	StartDate              time.Time `json:"-"`
	EndDate                time.Time `json:"-"`
	StartSettlementDate    time.Time `json:"-"`
	EndSettlementDate      time.Time `json:"-"`
	CreatedAtStartDate     time.Time `json:"-"`
	CreatedAtEndDate       time.Time `json:"-"`
	FilteredSortQuery      string    `json:"-"`
	TransactionId          string    `json:"-"`
	SettlementModel        string    `json:"settlementModel"`
}

type VoidTransactionRequest struct {
	TrxID             string
	ReasonType        string
	ReasonDescription string
	Status            string
	SettlementStatus  string
}

type UpdateTransactionRequest struct {
	TransactionID string
	Channel       string
}

type GetMerchantBalanceRequest struct {
	Date        time.Time
	MerchantID  string
	BalanceName string
}
