package platform

import (
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type TransactionRequest struct {
	UUID             string
	ParentMerchantId string
	MerchantId       string
	Reference        string // Disbursement, Payment, Transfer, Withdrawal, TopUp
	ReferenceType    string // Disbursement: Single/Bulk
	ReferenceID      string
	StartDate        time.Time
	EndDate          time.Time
	PaymentStartDate time.Time
	PaymentEndDate   time.Time
	Status           string
	ApprovalStatus   string
	PaymentMethod    string
	Keyword          string

	SortBy    string
	SortOrder string

	Page    int64
	PerPage int64
}

func (r *TransactionRequest) Validate() error {
	reference := strings.ToUpper(r.Reference)
	if reference != constant.ReferenceDisbursement &&
		reference != constant.ReferencePayment &&
		reference != constant.ReferencePlatformTransfer &&
		reference != constant.ReferenceWithdrawal &&
		reference != constant.ReferenceTopUp &&
		reference != constant.ReferenceCharge {
		return constant.ErrInvalidReference
	}

	if r.SortOrder != "" {
		if err := commonModel.ValidateSortOrder(r.SortOrder); err != nil {
			return err
		}
	}

	return nil
}
