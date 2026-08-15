package routingProcessorModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankTransfer"
)

type BankTransferRequest struct {
	HeaderRequest        snapCoreModel.BankTransferHeaderRequest
	PartnerReferenceNo   string
	Beneficiary          SubjectRequest
	Source               SubjectRequest
	Amount               commonModel.Amount
	Currency             string
	Remark               string
	PurposeOfTransaction string
	TransactionDate      time.Time
	AdditionalInfo       map[string]any
}

type SubjectRequest struct {
	Name              string
	BankCode          string
	AccountNo         string
	AccountName       string
	Email             string
	CustomerResidence string
	CustomerType      string
	CitizenStatus     string
	BICCode           string
	PlaceOfBirth      string
	Address           string
	IdentityNo        string
	Job               string
}
