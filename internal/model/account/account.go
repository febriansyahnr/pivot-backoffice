package account_model

import (
	"database/sql"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/google/uuid"
)

type Account struct {
	UUID                       uuid.UUID    `db:"uuid" json:"uuid"`
	ReferenceID                uuid.UUID    `db:"reference_id" json:"referenceId"`
	Name                       string       `db:"name" json:"name"`
	EODBalance                 float64      `db:"eod_balance" json:"eodBalance"`
	HoldedBalance              float64      `db:"holded_balance" json:"holdedBalance"`
	Currency                   string       `db:"currency" json:"currency"`
	Type                       string       `db:"type" json:"type"`
	UserType                   string       `db:"user_type" json:"userType"`
	LastUpdateBalanceAt        time.Time    `db:"last_update_balance_at" json:"lastUpdateBalanceAt"`
	PendingTransactionCutoffAt *time.Time   `db:"pending_transaction_cutoff_at" json:"-"`
	CreatedAt                  time.Time    `db:"created_at" json:"createdAt"`
	UpdatedAt                  time.Time    `db:"updated_at" json:"updatedAt"`
	DeletedAt                  sql.NullTime `db:"deleted_at" json:"deletedAt,omitempty"`

	CurrentBalance          float64   `db:"-" json:"currentBalance,omitempty"`
	CurrentBalanceCheckTime time.Time `db:"-" json:"currentBalanceCheckTime,omitempty"`
}

func (a *Account) RequiresPendingBalanceCalculation() bool {
	return a.Name == constant.AccountNamePayment ||
		a.Name == constant.AccountNameWallet ||
		a.Name == constant.AccountNameVirtualTerminal
}

func (a *Account) GetPendingTransactionCutoffOrBackdate() time.Time {
	if a.PendingTransactionCutoffAt != nil {
		return *a.PendingTransactionCutoffAt
	}
	return time.Now().UTC().AddDate(0, 0, -constant.DefaultPendingTransactionBackdateDays)
}

type AccountResponse struct {
	UUID                uuid.UUID `db:"uuid" json:"uuid"`
	ReferenceID         uuid.UUID `db:"reference_id" json:"merchantId"`
	Name                string    `db:"name" json:"name"`
	EODBalance          float64   `db:"eod_balance" json:"eodBalance"`
	Currency            string    `db:"currency" json:"currency"`
	Type                string    `db:"type" json:"type"`
	UserType            string    `db:"user_type" json:"userType"`
	LastUpdateBalanceAt time.Time `db:"last_update_balance_at" json:"lastUpdateBalanceAt"`
	CreatedAt           time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt           time.Time `db:"updated_at" json:"updatedAt"`

	CurrentBalance          float64   `db:"-" json:"current_balance,omitempty"`
	CurrentBalanceCheckTime time.Time `db:"-" json:"current_balance_check_time,omitempty"`
}

type WalletAccountResponse struct {
	UUID        uuid.UUID `db:"uuid" json:"uuid"`
	ReferenceID uuid.UUID `db:"reference_id" json:"entityId"`
	Name        string    `db:"name" json:"name"`
	Currency    string    `db:"currency" json:"currency"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type NewAccountRequest struct {
	ReferenceID uuid.UUID `json:"referenceId" validate:"required,uuid"`
	UserType    string    `json:"userType" validate:"required"`
	Usecase     string    `json:"usecase" validate:"required"`
	Currency    string    `json:"currency" validate:"required"`
}

type BulkCreateAccountRequest struct {
	MerchantID string `json:"merchantId" validate:"required,uuid"`
	Usecase    string `json:"usecase" validate:"required"`
	Currency   string `json:"currency" validate:"required"`
}

type GetEntityWithoutAccountRequest struct {
	MerchantID string
	Usecase    string
	Limit      int
}

func NewAccount(request *NewAccountRequest) (*Account, error) {
	account := &Account{
		UUID:                uuid.New(),
		ReferenceID:         request.ReferenceID,
		EODBalance:          0,
		Currency:            request.Currency,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
		LastUpdateBalanceAt: time.Now().UTC(),
	}

	name := GetAccountNameByUsecase(request.Usecase)
	if name == "" {
		return nil, constant.ErrInvalidUsecase
	}
	account.Name = name

	userType := GetUserType(request.UserType)
	if userType == "" {
		return nil, constant.ErrInvalidUserType
	}
	account.UserType = userType

	ledgerType := GetLedgerType(request.UserType)
	account.Type = ledgerType

	return account, nil
}

func (b *Account) ToResponse() *AccountResponse {
	return &AccountResponse{
		UUID:                    b.UUID,
		ReferenceID:             b.ReferenceID,
		Name:                    b.Name,
		EODBalance:              b.EODBalance,
		Currency:                b.Currency,
		LastUpdateBalanceAt:     b.LastUpdateBalanceAt,
		CreatedAt:               b.CreatedAt,
		UpdatedAt:               b.UpdatedAt,
		CurrentBalance:          b.CurrentBalance,
		CurrentBalanceCheckTime: b.CurrentBalanceCheckTime,
		UserType:                b.UserType,
		Type:                    b.Type,
	}
}

func (b *Account) ToWalletResponse() *WalletAccountResponse {
	return &WalletAccountResponse{
		UUID:        b.UUID,
		ReferenceID: b.ReferenceID,
		Name:        b.Name,
		Currency:    b.Currency,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

// Should be deprioritized
func GetAccountName(transactionType string) string {
	name := transactionType
	switch transactionType {
	case constant.TypeTopUp, constant.TypeManualAdjust, constant.TypeAccountInquiryFee, constant.TypeDisbursement:
		name = constant.TypeDisbursement
	case constant.TypePayment:
		name = constant.TypePayment
	case constant.TypeWallet:
		name = constant.TypeWallet
	default:
		name = ""
	}

	return name
}

func GetAccountNameByUsecase(usecase string) string {
	switch usecase {
	case constant.TypeWallet:
		return constant.TypeWallet
	case constant.TypePayment, constant.TypeRefund:
		return constant.TypePayment
	case constant.TypeVirtualTerminal:
		return constant.TypeVirtualTerminal
	case constant.TypePaymentFundedPayout:
		return constant.TypePaymentFundedPayout
	case constant.TypeInvalidUsecase:
		return ""
	}

	return constant.TypeDisbursement
}

func GetUserType(userType string) string {
	switch userType {
	case constant.UserTypeMerchant, constant.UserTypeSubMerchant:
		return constant.UserTypeMerchant
	case constant.UserTypeCustomer:
		return constant.UserTypeCustomer
	default:
		return ""
	}
}

func GetLedgerType(userType string) string {
	switch userType {
	case constant.UserTypeMerchant:
		return constant.TypeGeneralLedger
	case constant.UserTypeSubMerchant, constant.UserTypeCustomer:
		return constant.TypeLedger
	default:
		return ""
	}
}
