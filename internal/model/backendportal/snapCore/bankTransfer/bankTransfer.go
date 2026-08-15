package banktransfer

import (
	"database/sql"
	"encoding/json"
	"time"

	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore"
)

type BankTransferMetadata struct {
	AsyncConfig  BankTransferMetadaAsynConfig     `json:"asyncConfig"`
	TransferInfo BankTransferMetadataTransferInfo `json:"transferInfo"`
}

type BankTransferMetadaAsynConfig struct {
	Routes                 []string `json:"routes"`
	NextRoutes             []string `json:"nextRoutes"`
	Status                 string   `json:"status"`
	LatestRoute            string   `json:"latestRoute"`
	IsReconConfirmedFailed bool     `json:"isReconConfirmedFailed,omitempty"`
}

type BankTransferMetadataTransferInfo struct {
	LatestResponseCode    string `json:"latestResponseCode"`
	ShouldSkipCheckStatus bool   `json:"shouldSkipCheckStatus"`
	AllowRetryTransfer    bool   `json:"allowRetryTransfer"`
}

type BankTransfer struct {
	UUID                 string                      `db:"uuid"`
	BankReferenceNo      string                      `db:"bank_reference_no"`
	BeneficiaryID        string                      `db:"beneficiary_id"`
	ExternalID           string                      `db:"external_id"`
	MerchantID           sql.NullString              `db:"merchant_id"`
	PartnerReferenceNo   sql.NullString              `db:"partner_reference_no"`
	Amount               sql.NullString              `db:"amount"`
	SourceAccountNo      sql.NullString              `db:"source_account_no"`
	SourceAccountName    sql.NullString              `db:"source_account_name"`
	Currency             sql.NullString              `db:"currency"`
	CustomerReference    sql.NullString              `db:"customer_reference"`
	Remark               sql.NullString              `db:"remark"`
	PurposeOfTransaction sql.NullString              `db:"purpose_of_transaction"`
	AdditionalInfo       sql.NullString              `db:"additional_info"`
	Status               string                      `db:"status"`
	TransferType         sql.NullString              `db:"transfer_type"`
	BankAcquirer         sql.NullString              `db:"bank_acquirer"`
	TransactionDate      sql.NullTime                `db:"transaction_date"`
	CreatedAt            time.Time                   `db:"created_at"`
	UpdatedAt            sql.NullTime                `db:"updated_at"`
	AdditionalInfoObj    *BankTransferAdditionalInfo `db:"-"`
	MetadataObj          BankTransferMetadata        `db:"-"`
}

func (bt *BankTransfer) BuildMetadataObj() {
	if bt.AdditionalInfo.Valid {
		metadataStr := bt.AdditionalInfo.String
		metadataObj := BankTransferMetadata{}

		if err := json.Unmarshal([]byte(metadataStr), &metadataObj); err != nil {
			return
		}
		bt.MetadataObj = metadataObj
	}
}

type BankTransferAdditionalInfo struct {
	TransactionDate       *time.Time             `json:"transactionDate,omitempty"`
	CustomerReference     string                 `json:"customerReference,omitempty"`
	PartnerExternalID     string                 `json:"partnerExternalId,omitempty"`
	PartnerReferenceNo    string                 `json:"partnerReferenceNo,omitempty"`
	SendEmailNotification bool                   `json:"sendEmailNotification,omitempty"`
	TransactionStatusDesc string                 `json:"transactionStatusDesc,omitempty"`
	Remark                string                 `json:"remark,omitempty"`
	SourceAccountNo       string                 `json:"sourceAccountNo,omitempty"`
	ReconInfo             *BankTransferReconInfo `json:"reconInfo,omitempty"`
	Action                string                 `json:"action,omitempty"`
	TraceNo               string                 `json:"traceNo,omitempty"`
	SharedBiller          *SharedBillerInfo      `json:"sharedBiller,omitempty"`
}

type SharedBillerInfo struct {
	InquiryRequestId      string          `json:"inquiryRequestId"`
	VirtualAccountTrxType string          `json:"virtualAccountTrxType,omitempty"`
	BillDetails           json.RawMessage `json:"billDetails,omitempty"`
	BcaAdditionalInfo     map[string]any  `json:"bcaAdditionalInfo,omitempty"`
	PaymentRequestId      string          `json:"paymentRequestId,omitempty"`
	PaymentFlagStatus     string          `json:"paymentFlagStatus,omitempty"`
}

type BankTransferReconInfo struct {
	Status   string                `json:"status"`
	Reason   string                `json:"reason"`
	Amount   *snapCoreModel.Amount `json:"amount,omitempty"`
	CoreSync bool                  `json:"coreSync"`
}
