package merchant

import (
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

var asiaJakartaLoc, _ = time.LoadLocation(constant.TimeLoc)

type EventMerchantActionRequest struct {
	Event string    `json:"event"`
	Data  *Merchant `json:"data"`
}

type UploadDocumentReq struct {
	MerchantId string                `form:"-" validate:"required,uuid"`
	Type       string                `form:"type" validate:"required,max=50,oneof=NationalIdentityCard BusinessLicense TaxIdentification BusinessRegistration CertificateIncorporation CertificateNo40 CertificateLastAmendment CertificateDeedAmendment CertificateAmendmentAct CertificateEstablishment CertificateTaxRegistration BusinessEnvironmentPhoto BusinessProfile RegulatoryApprovalLicense AdminNationalIdentityCard ShareholderIdentityCard OwnerTaxIdentification ShareholderStructureDeed CertificateIncorporationApproval DirectorNationalIdentityCard PlaceOfEstablishment DateOfEstablishment BoardOfManagement "`
	Identifier string                `form:"identifier" validate:"max=32,required_if=Type NationalIdentityCard,required_if=Type BusinessLicense,required_if=Type TaxIdentification,required_if=Type BusinessRegistration"`
	File       *multipart.FileHeader `form:"file" validate:"-"`
	Hash       string                `form:"-" validate:"-"`
	Notes      string                `form:"notes"`
	CreatedBy  string                `form:"createdBy" validate:"max=50,required"`
}

func (d *UploadDocumentReq) ToInsertData(loc *DocLocation) *Document {
	id, _ := uuid.NewV7()

	doc := &Document{
		Id:         id.String(),
		MerchantId: d.MerchantId,
		Type:       d.Type,
		Identifier: d.Identifier,
		Hash:       d.Hash,
		Status:     constant.StatusApproved,
		Notes:      d.Notes,
		CreatedBy:  d.CreatedBy,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	doc.Location, _ = json.Marshal(loc)

	// Note: It will be deleted if no longer use the internal dashboard
	doc.ApprovedBy = d.CreatedBy
	doc.ApprovedAt = sql.NullTime{
		Valid: true,
		Time:  time.Now().UTC(),
	}

	return doc
}

type UpsertBoardOfDirectorReq struct {
	Method         string                `form:"-" validate:"oneof=POST PUT"`
	Id             string                `form:"-" validate:"required_if=Method PUT,omitempty,uuid"`
	MerchantId     string                `form:"-" validate:"required,uuid"`
	Position       string                `form:"position" validate:"required"`
	Name           string                `form:"name" validate:"required,max=255"`
	IdentityNumber string                `form:"identityNumber"`
	IdentityFile   *multipart.FileHeader `form:"identityFile"`
	Shares         string                `form:"shares"`
	Hash           string                `form:"-" validate:"-"`
	PositionLong   string                `form:"positionLong" validate:"omitempty,max=100"`
	CreatedBy      string                `form:"createdBy" validate:"required_if=Method POST,omitempty,max=50"`
}

func (r *UpsertBoardOfDirectorReq) ValidateRequest() error {
	switch r.Position {
	case constant.MerchantBODPositionDirector,
		constant.MerchantBODPositionPresidentDirector,
		constant.MerchantBODPositionCommissioner,
		constant.MerchantBODPositionPresidentCommissioner,
		constant.MerchantBODPositionShareholder:
	default:
		return constant.ErrInvalidMerchantBODPosition
	}

	if r.Position == constant.MerchantBODPositionShareholder {
		if r.Method == constant.ActionPost {
			if r.Shares == "" {
				return constant.ErrMerchantBODMandatoryShares
			}

			share, err := strconv.ParseFloat(r.Shares, 64)
			if err != nil || share <= 0 || share > 100 {
				return constant.ErrMerchantBODInvalidShares
			}
		}
		if r.Method == constant.ActionPut {
			if r.Shares != "" {
				share, err := strconv.ParseFloat(r.Shares, 64)
				if err != nil || share < 0 || share > 100 {
					return constant.ErrMerchantBODInvalidShares
				}
			}
		}
	} else {
		if r.Method == constant.ActionPost && (r.IdentityFile == nil || r.IdentityNumber == "") {
			return constant.ErrMerchantBODMandatoryIdentity
		}

		if r.Shares != "" {
			share, err := strconv.ParseFloat(r.Shares, 64)
			if err != nil || share < 0 || share > 100 {
				return constant.ErrMerchantBODInvalidShares
			}
		}
	}
	return nil
}

func (d *UpsertBoardOfDirectorReq) ToUpsertData(loc *DocLocation) *BoardOfDirector {
	data := &BoardOfDirector{
		Id:             d.Id,
		MerchantId:     d.MerchantId,
		Position:       d.Position,
		Name:           d.Name,
		IdentityNumber: d.IdentityNumber,
		Hash:           d.Hash,
		PositionLong:   d.PositionLong,
		UpdatedAt:      time.Now().UTC(),
	}
	if loc != nil {
		data.IdentityFile, _ = json.Marshal(loc)
	}

	if d.Method == constant.ActionPost {
		id, _ := uuid.NewV7()

		data.Id = id.String()
		data.CreatedBy = d.CreatedBy
		data.CreatedAt = time.Now().UTC()

		// Note: It will be deleted if no longer use the internal dashboard
		data.Status = constant.StatusApproved
		data.ApprovedBy = d.CreatedBy
		data.ApprovedAt = sql.NullTime{
			Valid: true, Time: time.Now().UTC(),
		}
	}

	if d.Shares != "" {
		share, _ := strconv.ParseFloat(d.Shares, 64)
		data.Shares = &share
	}

	return data
}

type BODValidation struct {
	Valid           bool           `db:"valid"`
	IdentityFile    types.JSONText `db:"identity_file"`
	Hash            string         `db:"hash"`
	IsCreate        bool           `db:"-"`
	ObjIdentityFile DocLocation    `db:"-"`
}

type GenSignatureReq struct {
	MerchantId string `json:"-" validate:"required,uuid"`
	Timestamp  string `json:"-" validate:"required,datetime=2006-01-02T15:04:05+07:00"`
	PrivateKey string `json:"privateKey" validate:"required"`
	GrantType  string `json:"grantType"`
}

type SubMerchantListFilter struct {
	ParentId       string     `json:"parentID"`
	MID            string     `json:"mid"`
	Name           string     `json:"name"`
	ShortName      string     `json:"shortName"`
	Status         string     `json:"status"`
	Email          string     `json:"email"`
	StartCreatedAt *time.Time `json:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt"`
	Keywords       string     `json:"keywords"`
}

type UpdateMerchantOpenApiRequest struct {
	ID            string `json:"id" validate:"required"`
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description"`
	Address       string `json:"address" validate:"required"`
	PostCode      string `json:"postCode" validate:"required"`
	Logo          string `json:"logo" validate:"required"`
	MerchantEmail string `json:"merchantEmail" validate:"required,email"`
	MerchantPhone string `json:"merchantPhone" validate:"required"`

	RequesterID   string `json:"-"`
	RequesterType string `json:"-"`
	ParentId      string `json:"-"`
}

type UpdateMerchantStatus struct {
	ID           string `json:"id" validate:"required"`
	Status       string `json:"status" validate:"required"`
	ReasonStatus string `json:"reasonStatus"`
}

type AutoWithdrawalSettingRequest struct {
	UserId     string `json:"-" validate:"required,uuid"`
	MerchantId string `json:"-" validate:"required,uuid"`
	Status     string `json:"status" validate:"required,oneof=ON OFF"`
}

type CreateFeeConfigOnBehalfRequest struct {
	MerchantId    string  `json:"merchantId" validate:"required,uuid"`
	Type          string  `json:"type" validate:"required,oneof=ALL DEFAULT DIRECT"`
	SubMerchantId *string `json:"subMerchantId" validate:"required_if=Type DIRECT,omitempty,uuid"`
	Reference     string  `json:"reference" validate:"required,oneof=PAYMENT DISBURSEMENT ACCOUNT_INQUIRY WALLET REFUND DISBURSEMENT_VA"`
	ReferenceType string  `json:"referenceType" validate:"required_if=Reference WALLET,omitempty,oneof=TOP_UP TRANSFER BANK_TRANSFER BILL MERCHANT_PAYMENT"`
	PaymentMethod *string `json:"paymentMethod" validate:"required_if=Reference PAYMENT,omitempty,oneof=VIRTUAL_ACCOUNT QRIS CREDIT_CARD"`
	AmountType    string  `json:"amountType" validate:"required,oneof=AMOUNT PERCENTAGE AMOUNT_PERCENTAGE"`
	Amount        float64 `json:"amount" validate:"min=0"`
	Percentage    float64 `json:"percentage" validate:"min=0.00,max=100.00"`
}

func (r *CreateFeeConfigOnBehalfRequest) ToOnBehalfFeeConfig() *OnBehalfFeeConfig {
	id, _ := uuid.NewV7()

	feeConfig := &OnBehalfFeeConfig{
		Id:            id.String(),
		MerchantId:    r.MerchantId,
		Type:          r.Type,
		Reference:     r.Reference,
		ReferenceType: r.ReferenceType,
		AmountType:    r.AmountType,
		Amount:        r.Amount,
		Percentage:    r.Percentage,
		CreatedAt:     time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if r.Type == constant.FeeOnBehalfTypeDirect {
		feeConfig.SubMerchantId = r.SubMerchantId
	}
	if r.Reference == constant.TypePayment {
		feeConfig.PaymentMethod = r.PaymentMethod
	}
	return feeConfig
}

type GetFeeConfigOnBehalfRequest struct {
	Reference     string  `json:"reference"`
	ReferenceType string  `json:"referenceType"`
	PaymentMethod *string `json:"paymentMethod"`
	MerchantId    string  `json:"merchantId"`
}

type UpdateFeeConfigOnBehalfRequest struct {
	AmountType string  `json:"amountType" validate:"required,oneof=AMOUNT PERCENTAGE AMOUNT_PERCENTAGE"`
	Amount     float64 `json:"amount" validate:"min=0"`
	Percentage float64 `json:"percentage" validate:"min=0.00,max=100.00"`
}

type UpdateMerchantKYCRequest struct {
	MerchantID     string  `json:"merchantID" validate:"required,uuid"`
	KYCStatus      string  `json:"status" validate:"required,oneof=WAITING_FOR_DOCUMENT REJECTED IN_REVIEW APPROVED NEED_RESUBMISSION REQUIRE_DOCUMENT_UPDATE_SUBMISSION NOT_REQUIRED"`
	MerchantStatus string  `json:"-"`
	MID            *string `json:"-"`
}

type BeneficiaryLimitConfigRequest struct {
	BeneficiaryBankCode        string                             `json:"beneficiaryBankCode,omitempty" validate:"omitempty"`
	BeneficiaryAccountNo       string                             `json:"beneficiaryAccountNo,omitempty" validate:"omitempty"`
	BeneficiaryPayoutLimitRule *BeneficiaryLimitConfigRuleRequest `json:"beneficiaryPayoutLimitRule" validate:"omitempty"`

	MerchantID string `json:"-" validate:"required"`
}

type BeneficiaryLimitConfigRuleRequest struct {
	Velocity        int64   `json:"velocity" validate:"required,min=0"`
	Timeframe       string  `json:"timeframe" validate:"required,oneof=DAILY"` // Do not read it for now, default = DAILY
	AmountThreshold float64 `json:"amountThreshold" validate:"required,min=0"`
}

type MerchantDocumentFilterRequest struct {
	MerchantID     string     `json:"-"`
	DocumentID     string     `json:"documentID"`
	DocumentType   string     `json:"documentType"`
	Identifier     string     `json:"identifier"`
	StartCreatedAt *time.Time `json:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt"`

	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
	SortBy  string `json:"sortBy"`
	Sort    string `json:"sort"`
}

type BillingFeeRequest struct {
	MerchantId string `validate:"required,uuid"`
	Status     string `validate:"omitempty,oneof=paid unpaid"`
	*BillingDateRangeRequest
}

type PayBillingFeeRequest struct {
	MerchantId string `json:"-" validate:"required,uuid"`
	Username   string `json:"username" validate:"required"`
	*BillingDateRangeRequest
}

type BillingDateRangeRequest struct {
	StrStartDate string    `json:"startDate" validate:"required,datetime=2006-01-02"`
	StrEndDate   string    `json:"endDate" validate:"required,datetime=2006-01-02"`
	StartDate    time.Time `json:"-" validate:"-"`
	EndDate      time.Time `json:"-" validate:"-"`
}

func (r *BillingDateRangeRequest) ParseDateRangeRequestFromAsiaJakartaToUtc() (err error) {
	if r.StartDate, err = time.ParseInLocation(time.DateOnly, r.StrStartDate, asiaJakartaLoc); err != nil {
		return err
	}
	if r.EndDate, err = time.ParseInLocation(time.DateOnly, r.StrEndDate, asiaJakartaLoc); err != nil {
		return err
	}

	r.StartDate = r.StartDate.UTC()
	r.EndDate = r.EndDate.AddDate(0, 0, 1).Add(-time.Millisecond).UTC()

	if r.StartDate.After(r.EndDate) {
		return constant.ErrInvalidDateRange
	}
	return nil
}

type FDSConfigRequest struct {
	FDSConfig
}

type PaymentInvestigationConfigRequest PaymentInvestigationConfig
