package merchant

import (
	"database/sql"
	"encoding/json"
	"errors"
	"mime/multipart"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/proto/messages/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Merchant struct {
	UUID              string         `json:"uuid" db:"uuid"`
	ExternalId        string         `json:"externalId" db:"external_id"`
	ParentID          sql.NullString `json:"parentId" db:"parent_id"`
	Name              string         `json:"name" db:"name"`
	ShortName         string         `json:"shortName" db:"short_name"`
	Description       string         `json:"description" db:"description"`
	Website           string         `json:"website" db:"website"`
	Address           string         `json:"address" db:"address"`
	DistrictId        uint16         `json:"districtId" db:"district_id"`
	PostCode          string         `json:"postcode" db:"postcode"`
	Logo              string         `json:"logo" db:"logo"`
	MerchantEmail     string         `json:"merchantEmail" db:"merchant_email"`
	MerchantPhone     string         `json:"merchantPhone" db:"merchant_phone"`
	PICEmail          string         `json:"picEmail" db:"pic_email"`
	PICPhone          string         `json:"picPhone" db:"pic_phone"`
	PICName           sql.NullString `json:"picName" db:"pic_name"`
	PICJobTitle       sql.NullString `json:"picJobTitle" db:"pic_job_title"`
	MID               sql.NullString `json:"mid" db:"mid"`
	BusinessType      sql.NullString `json:"businessType" db:"business_type"`
	BusinessStructure sql.NullString `json:"businessStructure" db:"business_structure"`
	BusinessCountry   sql.NullString `json:"businessCountry" db:"business_country"`
	ParentIndustry    sql.NullString `json:"parentIndustry" db:"parent_industry"`
	ChildIndustry     sql.NullString `json:"childIndustry" db:"child_industry"`
	MCC               sql.NullString `json:"mcc" db:"mcc"`
	CountryOfEntity   sql.NullString `json:"countryOfEntity" db:"country_of_entity"`
	DigitalStatus     sql.NullString `json:"digitalStatus" db:"digital_status"`
	KYCStatus         sql.NullString `json:"kycStatus" db:"kyc_status"`
	Status            string         `json:"status" db:"status"`
	RiskLevel         sql.NullString `json:"riskLevel" db:"risk_level"`
	ReasonStatus      string         `json:"reasonStatus" db:"reason_status"`
	CreatedAt         time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time      `json:"updatedAt" db:"updated_at"`
	DeletedAt         sql.NullTime   `json:"deletedAt" db:"deleted_at"`

	CallbackApiKey          *string            `json:"-" db:"callback_api_key"`
	CallbackApiKeyVersion   uint               `json:"-" db:"callback_api_key_version"`
	JITApiKey               string             `json:"-" db:"jit_api_key"` // just in time api key, wallet purpose
	JITApiKeyVersion        uint               `json:"-" db:"jit_api_key_version"`
	AddrDetail              types.JSONText     `json:"-" db:"address_detail"`
	TransactionConfigs      types.NullJSONText `json:"-" db:"transaction_configs"`
	PICInvitation           string             `json:"-" db:"pic_invitation"`
	ThirdPartyScreeningData types.NullJSONText `json:"-" db:"third_party_screening_data"`
	Metadata                types.NullJSONText `json:"-" db:"metadata"`
	// Additional Info After Create Merchant / Sub-Merchant
	BankAccount *MerchantBankAccountResponse `json:"bankAccount,omitempty" db:"-"`
	IndustryID  string                       `json:"-"`
	KYMNotes    string                       `json:"-"`
}

type MerchantMetadata struct {
	KYMNotes                   string                             `json:"kymNotes,omitempty"`
	BeneficiaryPayoutLimitRule *BeneficiaryLimitConfigRuleRequest `json:"beneficiaryPayoutLimitRule,omitempty"`
	NotificationConfig         *MerchantNotificationConfig        `json:"notificationConfig,omitempty"`
	TNCSigningMetadata         *TNCSigningMetadata                `json:"tncSigningMetadata,omitempty"`
}

// TNCSigningMetadata tracks the latest TNC signing state for a merchant.
type TNCSigningMetadata struct {
	IsSigned      bool       `json:"isSigned"`
	SignedVersion string     `json:"signedVersion"`
	SignedPath    string     `json:"signedPath,omitempty"`
	SignedAt      *time.Time `json:"signedAt,omitempty"`
	SignedBy      string     `json:"signedBy,omitempty"`
}

func (m *Merchant) GetMetadata() (*MerchantMetadata, error) {
	if m.Metadata.Valid {
		var metadata MerchantMetadata
		err := m.Metadata.Unmarshal(&metadata)
		if err != nil {
			return nil, err
		}
		return &metadata, nil
	}
	return nil, nil
}

func (d *Merchant) ToProtoDataEvent() *pb.Merchant {
	dst := &pb.Merchant{
		UUID:          d.UUID,
		ExternalId:    d.ExternalId,
		Name:          d.Name,
		ShortName:     d.ShortName,
		Description:   d.Description,
		Address:       d.Address,
		DistrictId:    util.MustUint16ToUint32(d.DistrictId),
		PostCode:      d.PostCode,
		Logo:          d.Logo,
		MerchantEmail: d.MerchantEmail,
		MerchantPhone: d.MerchantPhone,
		PICEmail:      d.PICEmail,
		PICPhone:      d.PICPhone,
		MID:           d.MID.String,
		Status:        d.Status,
		ReasonStatus:  d.ReasonStatus,
		CreatedAt:     timestamppb.New(d.CreatedAt),
		UpdatedAt:     timestamppb.New(d.UpdatedAt),
	}

	if d.ParentID.Valid {
		dst.ParentId = &d.ParentID.String
	}
	if d.PICName.Valid {
		dst.PICName = &d.PICName.String
	}
	if d.PICJobTitle.Valid {
		dst.PICJobTitle = &d.PICJobTitle.String
	}
	if d.BusinessType.Valid {
		dst.BusinessType = &d.BusinessType.String
	}
	if d.BusinessStructure.Valid {
		dst.BusinessStructure = &d.BusinessStructure.String
	}
	if d.BusinessCountry.Valid {
		dst.BusinessCountry = &d.BusinessCountry.String
	}
	if d.ParentIndustry.Valid {
		dst.ParentIndustry = &d.ParentIndustry.String
	}
	if d.ChildIndustry.Valid {
		dst.ChildIndustry = &d.ChildIndustry.String
	}
	if d.MCC.Valid {
		dst.Mcc = &d.MCC.String
	}
	if d.CountryOfEntity.Valid {
		dst.CountryOfEntity = &d.CountryOfEntity.String
	}
	if d.KYCStatus.Valid {
		dst.KycStatus = &d.KYCStatus.String
	}
	return dst
}

type CRMCreateMerchantRequest struct {
	Name              string  `json:"name" validate:"required"`
	MerchantStatus    string  `json:"merchantStatus" validate:"required"`
	ShortName         string  `json:"shortName" validate:"required,max=25"`
	Description       string  `json:"description"`
	Website           string  `json:"website" validate:"required,url"`
	Address           string  `json:"address" validate:"required,max=254"`
	DistrictId        uint16  `json:"districtId" validate:"required"`
	PostCode          string  `json:"postcode" validate:"required,numeric"`
	Logo              string  `json:"logo" validate:"required"`
	MerchantEmail     string  `json:"merchantEmail" validate:"required,email"`
	MerchantPhone     string  `json:"merchantPhone" validate:"required"`
	BusinessType      string  `json:"businessType" validate:"required"`
	BusinessStructure string  `json:"businessStructure" validate:"required"`
	BusinessCountry   string  `json:"businessCountry" validate:"required"`
	ParentIndustry    string  `json:"parentIndustry" validate:"required"`
	ChildIndustry     string  `json:"childIndustry" validate:"required"`
	MCC               string  `json:"mcc" validate:"required"`
	CountryOfEntity   string  `json:"countryOfEntity" validate:"required,max=2"`
	DigitalStatus     string  `json:"digitalStatus" validate:"required,oneof=Digital Non-digital"`
	PICInvitation     bool    `json:"picInvitation" validate:"-"`
	PICName           string  `json:"picName" validate:"required"`
	PICEmail          string  `json:"picEmail" validate:"required,email"`
	PICPhone          string  `json:"picPhone" validate:"required"`
	PICJobTitle       string  `json:"picJobTitle"`
	AutoWithdrawal    *string `json:"autoWithdrawal" validate:"omitempty,oneof=OFF ON"`
	KYCStatus         string  `json:"kycStatus" validate:"omitempty,oneof=KYC NON_KYC"`
	RiskLevel         string  `json:"riskLevel" validate:"omitempty,oneof=LOW LOW_MID MID MID_HIGH HIGH"`
	UserID            string  `json:"userId"`
}

type CRMUpdateMerchantRequest struct {
	Name                 string `json:"name" validate:"required"`
	ShortName            string `json:"shortName" validate:"omitempty,max=25"`
	Description          string `json:"description"`
	Address              string `json:"address" validate:"required,max=254"`
	DistrictId           uint16 `json:"districtId" validate:"required"`
	PostCode             string `json:"postcode" validate:"required,numeric"`
	Website              string `json:"website" validate:"required"`
	Logo                 string `json:"logo" validate:"omitempty"`
	MerchantEmail        string `json:"merchantEmail" validate:"required,email"`
	MerchantPhone        string `json:"merchantPhone" validate:"required"`
	MerchantStatus       string `json:"merchantStatus"`
	MerchantReasonStatus string `json:"merchantReasonStatus"`
	BusinessType         string `json:"businessType" validate:"required"`
	BusinessStructure    string `json:"businessStructure" validate:"required"`
	BusinessCountry      string `json:"businessCountry" validate:"required"`
	PICName              string `json:"picName" validate:"required"`
	PICEmail             string `json:"picEmail" validate:"required,email"`
	PICPhone             string `json:"picPhone" validate:"required"`
	PICJobTitle          string `json:"picJobTitle"`
	MerchantID           string `json:"merchantId"`
	UserID               string `json:"userId"`
	IndustryID           string `json:"industryId" validate:"omitempty,uuid"`
	CountryOfEntity      string `json:"countryOfEntity"`
	DigitalStatus        string `json:"digitalStatus"`
	RiskLevel            string `json:"riskLevel" validate:"omitempty,oneof=LOW LOW_MID MID MID_HIGH HIGH"`
	KYMNotes             string `json:"kymNotes"`
}

type CreateSubMerchantRequest struct {
	Name              string  `json:"name" validate:"required"`
	MerchantStatus    string  `json:"merchantStatus" validate:"required"`
	ShortName         string  `json:"shortName" validate:"required,max=25"`
	Description       string  `json:"description"`
	Website           string  `json:"website" validate:"required"`
	Address           string  `json:"address" validate:"required,max=254"`
	DistrictId        uint16  `json:"districtId" validate:"required"`
	PostCode          string  `json:"postcode" validate:"required,numeric"`
	Logo              string  `json:"logo" validate:"required"`
	MerchantEmail     string  `json:"merchantEmail" validate:"required,email"`
	MerchantPhone     string  `json:"merchantPhone" validate:"required"`
	BusinessType      string  `json:"businessType" validate:"required"`
	BusinessStructure string  `json:"businessStructure" validate:"required"`
	BusinessCountry   string  `json:"businessCountry" validate:"required"`
	ParentIndustry    string  `json:"parentIndustry" validate:"required"`
	ChildIndustry     string  `json:"childIndustry" validate:"required"`
	MCC               string  `json:"mcc" validate:"required"`
	CountryOfEntity   string  `json:"countryOfEntity" validate:"required,max=2"`
	DigitalStatus     string  `json:"digitalStatus" validate:"required,oneof=Digital Non-digital"`
	PICInvitation     bool    `json:"picInvitation" validate:"-"`
	PICName           string  `json:"picName" validate:"required"`
	PICEmail          string  `json:"picEmail" validate:"required,email"`
	PICPhone          string  `json:"picPhone" validate:"required"`
	PICJobTitle       string  `json:"picJobTitle"`
	AutoWithdrawal    *string `json:"autoWithdrawal" validate:"omitempty,oneof=OFF ON"`
	KYCStatus         string  `json:"kycStatus" validate:"omitempty,oneof=KYC NON_KYC"`
	RiskLevel         string  `json:"riskLevel" validate:"omitempty,oneof=LOW LOW_MID MID MID_HIGH HIGH"`
	ParentID          string

	RequesterID   string
	RequesterType string
	MerchantID    string `json:"merchantId"`
	UserID        string `json:"userId"`

	BankAccount *MerchantBankAccountRequest `json:"bankAccount" validate:"omitempty,required"`
}

type MerchantRequest struct {
	Name              string  `json:"name" validate:"required"`
	MerchantStatus    string  `json:"merchantStatus" validate:"required"`
	ShortName         string  `json:"shortName" validate:"required,max=25"`
	Description       string  `json:"description"`
	Website           string  `json:"website" validate:"required_if=EnforceMandatoryFields true"`
	Address           string  `json:"address" validate:"required,max=254"`
	DistrictId        uint16  `json:"districtId" validate:"required_if=EnforceMandatoryFields true"`
	PostCode          string  `json:"postcode" validate:"required,numeric"`
	Logo              string  `json:"logo" validate:"required"`
	MerchantEmail     string  `json:"merchantEmail" validate:"required,email"`
	MerchantPhone     string  `json:"merchantPhone" validate:"required"`
	BusinessType      string  `json:"businessType" validate:"required"`
	BusinessStructure string  `json:"businessStructure" validate:"required"`
	BusinessCountry   string  `json:"businessCountry" validate:"required"`
	ParentIndustry    string  `json:"parentIndustry" validate:"required_if=EnforceMandatoryFields true"`
	ChildIndustry     string  `json:"childIndustry" validate:"required_if=EnforceMandatoryFields true"`
	MCC               string  `json:"mcc" validate:"required_if=EnforceMandatoryFields true"`
	CountryOfEntity   string  `json:"countryOfEntity" validate:"required_if=EnforceMandatoryFields true,omitempty,max=2"`
	DigitalStatus     string  `json:"digitalStatus" validate:"required_if=EnforceMandatoryFields true,omitempty,oneof=Digital Non-digital"`
	PICInvitation     bool    `json:"picInvitation" validate:"-"`
	PICName           string  `json:"picName" validate:"required"`
	PICEmail          string  `json:"picEmail" validate:"required,email"`
	PICPhone          string  `json:"picPhone" validate:"required"`
	PICJobTitle       string  `json:"picJobTitle"`
	AutoWithdrawal    *string `json:"autoWithdrawal" validate:"omitempty,oneof=OFF ON"`
	KYCStatus         string  `json:"kycStatus" validate:"omitempty,oneof=KYC NON_KYC"`
	RiskLevel         string  `json:"riskLevel" validate:"omitempty,oneof=LOW LOW_MID MID MID_HIGH HIGH"`
	SubAccountType    string  `json:"subAccountType" validate:"omitempty,oneof=KYC NON_KYC"` // Only to create sub-accounts via open-api
	ParentID          string

	RequesterID   string
	RequesterType string
	MerchantID    string `json:"merchantId"`
	UserID        string `json:"userId"`

	BankAccount *MerchantBankAccountRequest `json:"bankAccount" validate:"omitempty,required"`

	EnforceMandatoryFields bool `json:"-"` // Internal flag to enforce mandatory fields validation, for backward compatibility
}

type MerchantBankAccountRequest struct {
	ChannelCode   string `json:"channelCode" validate:"required"`
	AccountNumber string `json:"accountNumber" validate:"required,numeric"`
}

type MerchantBankAccountResponse struct {
	ChannelCode   string `json:"channelCode"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}

type UpdateMerchantRequest struct {
	ID                string          `json:"id" validate:"required"`
	Name              string          `json:"name" validate:"required"`
	ShortName         string          `json:"shortName" validate:"required,max=25"`
	Description       string          `json:"description"`
	Address           string          `json:"address" validate:"required,max=254"`
	DistrictId        uint16          `json:"districtId" validate:"omitempty,min=1"`
	PostCode          string          `json:"postcode" validate:"required,numeric"`
	Website           string          `json:"website"`
	Logo              string          `json:"logo" validate:"omitempty"`
	LogoFile          *LogoFileUpload `json:"-"` // Optional file upload for logo
	MerchantEmail     string          `json:"merchantEmail" validate:"required"`
	MerchantPhone     string          `json:"merchantPhone" validate:"required"`
	PICName           string          `json:"picName" validate:"required"`
	PICEmail          string          `json:"picEmail" validate:"required"`
	PICPhone          string          `json:"picPhone" validate:"required"`
	PICJobTitle       string          `json:"picJobTitle"`
	PICInvitation     bool            `json:"picInvitation"`
	BusinessType      string
	BusinessStructure string `json:"businessStructure" validate:"required"`
	BusinessCountry   string
	Status            string `json:"status"`
	ReasonStatus      string
	// Industry classification fields
	IndustryID      string
	ParentIndustry  string `json:"parentIndustry"`
	ChildIndustry   string `json:"childIndustry"`
	MCC             string `json:"mcc"`
	CountryOfEntity string `json:"countryOfEntity"`
	DigitalStatus   string `json:"digitalStatus"`
	RiskLevel       string `json:"riskLevel" validate:"omitempty,oneof=LOW LOW_MID MID MID_HIGH HIGH"`

	// KYM
	KYMNotes string

	RequesterID   string
	RequesterType string
}

type LogoFileUpload struct {
	FileHeader *multipart.FileHeader
}

type MerchantAssignRequest struct {
	UserID     string `json:"userId" validate:"required"`
	MerchantID string `json:"merchantId" validate:"required"`
}

type MerchantResponse struct {
	UUID              string         `json:"uuid"`
	ExternalId        string         `json:"externalId"`
	Name              string         `json:"name"`
	ShortName         string         `json:"shortName"`
	Description       string         `json:"description"`
	Website           string         `json:"website"`
	Address           string         `json:"address"`
	DistrictId        uint16         `json:"districtId,omitempty"`
	AddrDetail        *AddressDetail `json:"addressDetail,omitempty"`
	PostCode          string         `json:"postcode"`
	Logo              string         `json:"logo"`
	MID               string         `json:"mid"`
	MerchantEmail     string         `json:"merchantEmail"`
	MerchantPhone     string         `json:"merchantPhone"`
	PICEmail          string         `json:"picEmail"`
	PICPhone          string         `json:"picPhone"`
	PICName           string         `json:"picName"`
	PICJobTitle       string         `json:"picJobTitle"`
	PICInvitation     string         `json:"picInvitation,omitempty"`
	BusinessType      string         `json:"businessType"`
	BusinessStructure string         `json:"businessStructure"`
	BusinessCountry   string         `json:"businessCountry"`
	ParentIndustry    string         `json:"parentIndustry"`
	ChildIndustry     string         `json:"childIndustry"`
	MCC               string         `json:"mcc"`
	CountryOfEntity   string         `json:"countryOfEntity"`
	DigitalStatus     string         `json:"digitalStatus"`
	KYCStatus         string         `json:"kycStatus"`
	Status            string         `json:"status"`
	RiskLevel         string         `json:"riskLevel"`
	ReasonStatus      string         `json:"reasonStatus"`
	ParentID          string         `json:"parentId"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

func (m *Merchant) ToResponse() *MerchantResponse {
	data := &MerchantResponse{
		UUID:              m.UUID,
		ExternalId:        m.ExternalId,
		Name:              m.Name,
		ShortName:         m.ShortName,
		Description:       m.Description,
		Website:           m.Website,
		Address:           m.Address,
		DistrictId:        m.DistrictId,
		PostCode:          m.PostCode,
		Logo:              m.Logo,
		MID:               m.MID.String,
		MerchantEmail:     m.MerchantEmail,
		MerchantPhone:     m.MerchantPhone,
		PICName:           m.PICName.String,
		PICJobTitle:       m.PICJobTitle.String,
		PICEmail:          m.PICEmail,
		PICPhone:          m.PICPhone,
		PICInvitation:     m.PICInvitation,
		BusinessType:      m.BusinessType.String,
		BusinessStructure: m.BusinessStructure.String,
		BusinessCountry:   m.BusinessCountry.String,
		ParentIndustry:    m.ParentIndustry.String,
		ChildIndustry:     m.ChildIndustry.String,
		MCC:               m.MCC.String,
		CountryOfEntity:   m.CountryOfEntity.String,
		DigitalStatus:     m.DigitalStatus.String,
		KYCStatus:         m.KYCStatus.String,
		Status:            m.Status,
		RiskLevel:         m.RiskLevel.String,
		ReasonStatus:      m.ReasonStatus,
		ParentID:          m.ParentID.String,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
	if len(m.AddrDetail) > 0 {
		data.AddrDetail = &AddressDetail{}
		_ = json.Unmarshal(m.AddrDetail, data.AddrDetail)
	}

	return data
}

func (m *MerchantRequest) NewSubMerchant(callbackApiKey *string, callbackApiKeyVersion uint) (*Merchant, error) {
	if m.ParentID == "" {
		return nil, errors.New("parent id is required")
	}

	merchant := &Merchant{
		UUID:          uuid.New().String(),
		ExternalId:    util.GenerateULID(),
		Name:          m.Name,
		ShortName:     m.ShortName,
		Description:   m.Description,
		Website:       m.Website,
		Address:       m.Address,
		PostCode:      m.PostCode,
		Logo:          m.Logo,
		MerchantEmail: m.MerchantEmail,
		MerchantPhone: m.MerchantPhone,
		PICEmail:      m.PICEmail,
		PICPhone:      m.PICPhone,
		PICName: sql.NullString{
			String: m.PICName,
			Valid:  true,
		},
		PICJobTitle: sql.NullString{
			String: m.PICJobTitle,
			Valid:  true,
		},
		BusinessType: sql.NullString{
			String: m.BusinessType,
			Valid:  true,
		},
		BusinessStructure: sql.NullString{
			String: m.BusinessStructure,
			Valid:  true,
		},
		BusinessCountry: sql.NullString{
			String: m.BusinessCountry,
			Valid:  true,
		},
		ParentIndustry: sql.NullString{
			String: m.ParentIndustry,
			Valid:  m.ParentIndustry != "",
		},
		ChildIndustry: sql.NullString{
			String: m.ChildIndustry,
			Valid:  m.ChildIndustry != "",
		},
		MCC: sql.NullString{
			String: m.MCC,
			Valid:  m.MCC != "",
		},
		CountryOfEntity: sql.NullString{
			String: m.CountryOfEntity,
			Valid:  m.CountryOfEntity != "",
		},
		DigitalStatus: sql.NullString{
			String: m.DigitalStatus,
			Valid:  m.DigitalStatus != "",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Status:    constant.MerchantStatusCreated,
		ParentID: sql.NullString{
			String: m.ParentID,
			Valid:  true,
		},
		CallbackApiKey:        callbackApiKey,
		CallbackApiKeyVersion: callbackApiKeyVersion,
	}

	if m.KYCStatus != "" {
		merchant.KYCStatus = sql.NullString{
			String: m.KYCStatus,
			Valid:  true,
		}
	}

	if m.RiskLevel != "" {
		merchant.RiskLevel = sql.NullString{
			String: m.RiskLevel,
			Valid:  true,
		}
	}

	if m.KYCStatus == constant.KYCStatusNotRequired || m.KYCStatus == constant.KYCStatusApproved {
		merchant.Status = constant.MerchantStatusActive
	}

	if m.DistrictId > 0 {
		merchant.DistrictId = m.DistrictId
	}
	transactionConfig := map[string]interface{}{
		"autoWithdrawal": "ON",
	}
	if m.AutoWithdrawal != nil {
		transactionConfig["autoWithdrawal"] = *m.AutoWithdrawal
	}
	m.AutoWithdrawal = util.ValueToPtr(transactionConfig["autoWithdrawal"].(string))

	merchant.TransactionConfigs.Valid = true
	merchant.TransactionConfigs.JSONText, _ = json.Marshal(transactionConfig)

	return merchant, nil
}

func (m *Merchant) UpdateMerchant(r *UpdateMerchantRequest) error {
	if r.Name != "" {
		m.Name = r.Name
	}
	if r.ShortName != "" {
		m.ShortName = r.ShortName
	}
	if r.Description != "" {
		m.Description = r.Description
	}
	if r.Website != "" {
		m.Website = r.Website
	}
	if r.Address != "" {
		m.Address = r.Address
	}
	if r.DistrictId > 0 {
		m.DistrictId = r.DistrictId
	}
	if r.Status != "" {
		m.Status = r.Status
	}
	if r.ReasonStatus != "" {
		m.ReasonStatus = r.ReasonStatus
	}
	if r.PostCode != "" {
		m.PostCode = r.PostCode
	}
	if r.Logo != "" {
		m.Logo = r.Logo
	}
	if r.MerchantEmail != "" {
		m.MerchantEmail = r.MerchantEmail
	}
	if r.MerchantPhone != "" {
		m.MerchantPhone = r.MerchantPhone
	}
	if r.PICEmail != "" {
		m.PICEmail = r.PICEmail
	}
	if r.PICPhone != "" {
		m.PICPhone = r.PICPhone
	}
	if r.PICName != "" {
		m.PICName.String = r.PICName
		m.PICName.Valid = true
	}
	if r.PICJobTitle != "" {
		m.PICJobTitle.String = r.PICJobTitle
		m.PICJobTitle.Valid = true
	}
	if r.BusinessStructure != "" {
		m.BusinessStructure.String = r.BusinessStructure
		m.BusinessStructure.Valid = true
	}
	if r.BusinessCountry != "" {
		m.BusinessCountry.String = r.BusinessCountry
		m.BusinessCountry.Valid = true
	}
	if r.BusinessType != "" {
		m.BusinessType.String = r.BusinessType
		m.BusinessType.Valid = true
	}

	// Update industry-related fields
	if r.ParentIndustry != "" {
		m.ParentIndustry.String = r.ParentIndustry
		m.ParentIndustry.Valid = true
	}
	if r.ChildIndustry != "" {
		m.ChildIndustry.String = r.ChildIndustry
		m.ChildIndustry.Valid = true
	}
	if r.MCC != "" {
		m.MCC.String = r.MCC
		m.MCC.Valid = true
	}

	if r.CountryOfEntity != "" {
		m.CountryOfEntity.String = r.CountryOfEntity
		m.CountryOfEntity.Valid = true
	}
	if r.DigitalStatus != "" {
		m.DigitalStatus.String = r.DigitalStatus
		m.DigitalStatus.Valid = true
	}
	if r.RiskLevel != "" {
		m.RiskLevel.String = r.RiskLevel
		m.RiskLevel.Valid = true
	}

	if r.ReasonStatus != "" {
		m.ReasonStatus = r.ReasonStatus
	}
	if r.KYMNotes != "" {
		m.KYMNotes = r.KYMNotes
		if err := m.UpdateKYMNotes(); err != nil {
			return err
		}
	}
	m.UpdatedAt = time.Now()

	return nil
}

func BuildUpdateMerchantRequest(subMerchant *Merchant) UpdateMerchantRequest {
	return UpdateMerchantRequest{
		ID:                subMerchant.UUID,
		Name:              subMerchant.Name,
		ShortName:         subMerchant.ShortName,
		Description:       subMerchant.Description,
		Address:           subMerchant.Address,
		DistrictId:        subMerchant.DistrictId,
		PostCode:          subMerchant.PostCode,
		Website:           subMerchant.Website,
		Logo:              subMerchant.Logo,
		MerchantEmail:     subMerchant.MerchantEmail,
		MerchantPhone:     subMerchant.MerchantPhone,
		PICName:           subMerchant.PICName.String,
		PICEmail:          subMerchant.PICEmail,
		PICPhone:          subMerchant.PICPhone,
		PICJobTitle:       subMerchant.PICJobTitle.String,
		BusinessStructure: subMerchant.BusinessStructure.String,
		Status:            subMerchant.Status,
	}
}

func SetActiveMerchant(merchant *Merchant) {
	merchant.Status = constant.MerchantStatusActive
}

// SetInactiveMerchant sets the status of the merchant to inactive.
// Deprecated: Use DeactivateMerchant instead.
func SetInactiveMerchant(merchant *Merchant) {
	merchant.Status = constant.MerchantStatusInactive
}

func DeactivateMerchant(merchant *Merchant) {
	merchant.Status = constant.MerchantStatusDeactivated
}

func (m *Merchant) SetBlocked() {
	m.Status = constant.MerchantStatusBlocked
}

func (m *Merchant) Unblock() {
	m.Status = constant.MerchantStatusActive
	m.UpdatedAt = time.Now().UTC()
}

type AddressDetail struct {
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
}

type TransactionConfigs struct {
	Disbursement               DisbursementConfig          `json:"disbursement"`
	Withdrawal                 WithdrawalConfig            `json:"withdrawal"`
	DailyDisbursement          *DailyDisbursementConfig    `json:"dailyDisbursement,omitempty"`
	PaymentInvestigationConfig *PaymentInvestigationConfig `json:"paymentInvestigation,omitempty"`
}

func (t TransactionConfigs) Validate() error {
	return validatorExt.New().Struct(t) // Singleton
}

type WithdrawalConfig AmountConfig
type DisbursementConfig AmountConfig

type AmountConfig struct {
	MinAmount float64 `json:"minAmount" validate:"min=10000"`                   // Ref: constant.DisbursementMinAmount
	MaxAmount float64 `json:"maxAmount" validate:"min=10000,gtfield=MinAmount"` // Ref: constant.DisbursementMaxAmount
}

type DailyDisbursementConfig struct {
	Merchant         float64  `json:"merchant"`
	MerchantPlatform *float64 `json:"merchantPlatform,omitempty"`
}

type MerchantIdForConfigs struct {
	MerchantType              string `json:"merchantType"`
	MerchantTransactionConfig string `json:"merchantTransactionConfig"` // Merchant ID
}

type PaymentInvestigationConfig struct {
	Enabled             bool    `json:"enabled" validate:"-"`
	PivotPercentageLoss float64 `json:"pivotPercentageLoss" validate:"required,min=1,max=100"`
	PivotMaxLoss        float64 `json:"pivotMaxLoss" validate:"required,min=1"`
}

type SubAccountRegistrationCallback struct {
	SubAccountId        string    `json:"subAccountId"`
	SubAccountStatus    string    `json:"subAccountStatus"`
	SubAccountKycStatus string    `json:"subAccountKycStatus"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type BulkCreateSubMerchantRequest struct {
	MerchantId         string
	KYCType            string
	FileName           string
	IsInvitePIC        bool
	SubmerchantDetails [][]string
}

type ProcessBulkCreateSubMerchantRequest struct {
	ID                 string                         `json:"id"`
	MerchantId         string                         `json:"merchantId"`
	KYCType            string                         `json:"kycType"`
	SubmerchantDetails []BulkSubMerchantDetailRequest `json:"submerchantDetails"`
	Batch              int                            `json:"batch"`
	FileName           string                         `json:"fileName"`
}

type BulkSubMerchantDetailRequest struct {
	Row    int             `json:"row"`
	Detail MerchantRequest `json:"detail"`
}

type BulkCreateSubMerchantResponse struct {
	ID           string                                `json:"id"`
	FileName     string                                `json:"fileName"`
	TotalSuccess int                                   `json:"totalSuccess"`
	TotalFailed  int                                   `json:"totalFailed"`
	Results      []BulkCreateSubMerchantDetailResponse `json:"results"`
}

type BulkCreateSubMerchantDetailResponse struct {
	Row          int    `json:"row"`
	MerchantID   string `json:"merchantId,omitempty"`
	MerchantName string `json:"merchantName,omitempty"`
	Error        string `json:"error,omitempty"`
}

type GetBulkCreateSubMerchantSummaryRequest struct {
	MerchantId string `json:"merchantId"`
	ID         string `json:"id"`
}

func (m *Merchant) UpdateKYMNotes() error {
	metadata, err := m.GetMetadata()
	if err != nil {
		return err
	}
	if metadata == nil {
		metadata = &MerchantMetadata{}
	}
	metadata.KYMNotes = m.KYMNotes
	b, _ := json.Marshal(metadata)
	m.Metadata = types.NullJSONText{Valid: true, JSONText: b}
	return nil
}

func (m *Merchant) UpdateBeneficiaryPayoutLimitRule(rule *BeneficiaryLimitConfigRuleRequest) error {
	metadata, err := m.GetMetadata()
	if err != nil {
		return err
	}
	if metadata == nil {
		metadata = &MerchantMetadata{}
	}
	metadata.BeneficiaryPayoutLimitRule = rule
	b, _ := json.Marshal(metadata)
	m.Metadata = types.NullJSONText{Valid: true, JSONText: b}
	return nil
}

func (m *Merchant) UpdateNotificationConfig(req *MerchantNotificationConfig) error {
	metadata, err := m.GetMetadata()
	if err != nil {
		return err
	}
	if metadata == nil {
		metadata = &MerchantMetadata{}
	}
	if metadata.NotificationConfig == nil {
		metadata.NotificationConfig = &MerchantNotificationConfig{}
	}

	// Merge logic: only update if the specific config section is provided (not nil)
	if req.Transaction != nil {
		metadata.NotificationConfig.Transaction = req.Transaction
	}
	if req.Balance != nil {
		metadata.NotificationConfig.Balance = req.Balance
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	m.Metadata = types.NullJSONText{Valid: true, JSONText: b}
	return nil
}

// UpdateTNCSigningStatus records the latest TNC signing event in merchant metadata.
func (m *Merchant) UpdateTNCSigningStatus(status *TNCSigningMetadata) error {
	metadata, err := m.GetMetadata()
	if err != nil {
		return err
	}
	if metadata == nil {
		metadata = &MerchantMetadata{}
	}
	metadata.TNCSigningMetadata = status
	b, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	m.Metadata = types.NullJSONText{Valid: true, JSONText: b}
	return nil
}

type CRMMerchantResponse struct {
	UUID              string         `json:"uuid"`
	ExternalId        string         `json:"externalId"`
	Name              string         `json:"name"`
	ShortName         string         `json:"shortName"`
	Description       string         `json:"description"`
	Website           string         `json:"website"`
	Address           string         `json:"address"`
	DistrictId        uint16         `json:"districtId,omitempty"`
	AddrDetail        *AddressDetail `json:"addressDetail,omitempty"`
	PostCode          string         `json:"postcode"`
	Logo              string         `json:"logo"`
	MID               string         `json:"mid"`
	MerchantEmail     string         `json:"merchantEmail"`
	MerchantPhone     string         `json:"merchantPhone"`
	PICEmail          string         `json:"picEmail"`
	PICPhone          string         `json:"picPhone"`
	PICName           string         `json:"picName"`
	PICJobTitle       string         `json:"picJobTitle"`
	PICInvitation     string         `json:"picInvitation,omitempty"`
	BusinessType      string         `json:"businessType"`
	BusinessStructure string         `json:"businessStructure"`
	BusinessCountry   string         `json:"businessCountry"`
	ParentIndustry    string         `json:"parentIndustry"`
	ChildIndustry     string         `json:"childIndustry"`
	MCC               string         `json:"mcc"`
	CountryOfEntity   string         `json:"countryOfEntity"`
	DigitalStatus     string         `json:"digitalStatus"`
	KYCStatus         string         `json:"kycStatus"`
	KYMNotes          string         `json:"kymNotes,omitempty"`
	Status            string         `json:"status"`
	RiskLevel         string         `json:"riskLevel"`
	ReasonStatus      string         `json:"reasonStatus"`
	ParentID          string         `json:"parentId"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

func (m *Merchant) ToCRMMerchantResponse() *CRMMerchantResponse {
	data := &CRMMerchantResponse{
		UUID:              m.UUID,
		ExternalId:        m.ExternalId,
		Name:              m.Name,
		ShortName:         m.ShortName,
		Description:       m.Description,
		Address:           m.Address,
		DistrictId:        m.DistrictId,
		PostCode:          m.PostCode,
		Website:           m.Website,
		Logo:              m.Logo,
		MID:               m.MID.String,
		MerchantEmail:     m.MerchantEmail,
		MerchantPhone:     m.MerchantPhone,
		PICName:           m.PICName.String,
		PICJobTitle:       m.PICJobTitle.String,
		PICEmail:          m.PICEmail,
		PICPhone:          m.PICPhone,
		PICInvitation:     m.PICInvitation,
		BusinessType:      m.BusinessType.String,
		BusinessStructure: m.BusinessStructure.String,
		BusinessCountry:   m.BusinessCountry.String,
		ParentIndustry:    m.ParentIndustry.String,
		ChildIndustry:     m.ChildIndustry.String,
		MCC:               m.MCC.String,
		CountryOfEntity:   m.CountryOfEntity.String,
		DigitalStatus:     m.DigitalStatus.String,
		KYCStatus:         m.KYCStatus.String,
		Status:            m.Status,
		RiskLevel:         m.RiskLevel.String,
		ReasonStatus:      m.ReasonStatus,
		ParentID:          m.ParentID.String,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
	if len(m.AddrDetail) > 0 {
		data.AddrDetail = &AddressDetail{}
		_ = json.Unmarshal(m.AddrDetail, data.AddrDetail)
	}

	metadata, _ := m.GetMetadata()
	if metadata != nil {
		data.KYMNotes = metadata.KYMNotes
	}

	return data
}
