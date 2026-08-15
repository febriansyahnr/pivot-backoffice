package customerModel

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

type Customer struct {
	UUID             string
	MerchantID       string
	Email            string
	PhoneCountryCode string
	PhoneNumber      string
	FirstName        string
	LastName         string
	BusinessName     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time

	City         string
	Country      string
	AddressLine1 string
	AddressLine2 string
	PostalCode   string
	State        string

	IsBlocked   bool
	BlockReason string

	Metadata map[string]interface{}

	OriginalPhoneNumber string `json:"-"`
}

type CustomerDBModel struct {
	UUID             string         `db:"uuid"`
	MerchantID       string         `db:"merchant_id"`
	Email            sql.NullString `db:"email"`
	PhoneCountryCode sql.NullString `db:"phone_country_code"`
	PhoneNumber      string         `db:"phone_number"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	DeletedAt        sql.NullTime   `db:"deleted_at"`
	FirstName        string         `db:"first_name"`
	LastName         sql.NullString `db:"last_name"`
	BusinessName     sql.NullString `db:"business_name"`

	City         sql.NullString `db:"city"`
	Country      sql.NullString `db:"country"`
	AddressLine1 sql.NullString `db:"address_line1"`
	AddressLine2 sql.NullString `db:"address_line2"`
	PostalCode   sql.NullString `db:"postal_code"`
	State        sql.NullString `db:"state"`

	IsBlocked   bool           `db:"is_blocked"`
	BlockReason sql.NullString `db:"block_reason"`

	Metadata []byte `db:"metadata"`
}

func (c *CustomerDBModel) ToCustomerModel() *Customer {
	var metadata map[string]interface{}
	if len(c.Metadata) > 0 {
		_ = json.Unmarshal(c.Metadata, &metadata)
	}
	var deletedTime *time.Time = nil
	if c.DeletedAt.Valid {
		deletedTime = &c.DeletedAt.Time
	}
	customer := &Customer{
		UUID:                c.UUID,
		MerchantID:          c.MerchantID,
		PhoneNumber:         c.PhoneNumber,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
		DeletedAt:           deletedTime,
		FirstName:           c.FirstName,
		Metadata:            metadata,
		OriginalPhoneNumber: c.PhoneNumber,
	}

	if c.Email.Valid {
		customer.Email = c.Email.String
	}

	if c.LastName.Valid {
		customer.LastName = c.LastName.String
	}

	if c.BusinessName.Valid {
		customer.BusinessName = c.BusinessName.String
	}

	if c.City.Valid {
		customer.City = c.City.String
	}

	if c.Country.Valid {
		customer.Country = c.Country.String
	}

	if c.AddressLine1.Valid {
		customer.AddressLine1 = c.AddressLine1.String
	}

	if c.AddressLine2.Valid {
		customer.AddressLine2 = c.AddressLine2.String
	}

	if c.PostalCode.Valid {
		customer.PostalCode = c.PostalCode.String
	}

	if c.State.Valid {
		customer.State = c.State.String
	}

	customer.IsBlocked = c.IsBlocked
	if c.BlockReason.Valid {
		customer.BlockReason = c.BlockReason.String
	}

	// handle indonesian phone number
	if c.PhoneCountryCode.Valid {
		customer.PhoneCountryCode = c.PhoneCountryCode.String
		if c.PhoneCountryCode.String == constant.DefaultPhoneCountryCode {
			customer.PhoneNumber = "0" + customer.PhoneNumber
		} else {
			customer.PhoneNumber = c.PhoneCountryCode.String + customer.PhoneNumber
		}
	}

	return customer
}

func (c *Customer) ToCustomerDBModel() *CustomerDBModel {
	if c.Metadata == nil {
		c.Metadata = map[string]interface{}{}
	}
	metadata, _ := json.Marshal(c.Metadata)
	deletedTime := sql.NullTime{}
	if c.DeletedAt != nil {
		deletedTime = sql.NullTime{Time: *c.DeletedAt, Valid: true}
	}
	return &CustomerDBModel{
		UUID:       c.UUID,
		MerchantID: c.MerchantID,
		Email: sql.NullString{
			String: c.Email,
			Valid:  c.Email != "",
		},
		PhoneCountryCode: sql.NullString{
			String: c.PhoneCountryCode,
			Valid:  c.PhoneCountryCode != "",
		},
		PhoneNumber: c.PhoneNumber,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		DeletedAt:   deletedTime,
		FirstName:   c.FirstName,
		LastName: sql.NullString{
			String: c.LastName,
			Valid:  c.LastName != "",
		},
		BusinessName: sql.NullString{
			String: c.BusinessName,
			Valid:  c.BusinessName != "",
		},
		City: sql.NullString{
			String: c.City,
			Valid:  c.City != "",
		},
		Country: sql.NullString{
			String: c.Country,
			Valid:  c.Country != "",
		},
		AddressLine1: sql.NullString{
			String: c.AddressLine1,
			Valid:  c.AddressLine1 != "",
		},
		AddressLine2: sql.NullString{
			String: c.AddressLine2,
			Valid:  c.AddressLine2 != "",
		},
		PostalCode: sql.NullString{
			String: c.PostalCode,
			Valid:  c.PostalCode != "",
		},
		State: sql.NullString{
			String: c.State,
			Valid:  c.State != "",
		},
		IsBlocked: c.IsBlocked,
		BlockReason: sql.NullString{
			String: c.BlockReason,
			Valid:  c.BlockReason != "",
		},
		Metadata: metadata,
	}
}

func (c *CustomerDBModel) ToCardFundedPayoutSavedCardList() *cardFundedPayoutModel.GetSavedCardResponse {
	var metadata map[string]interface{}

	if len(c.Metadata) > 0 {
		_ = json.Unmarshal(c.Metadata, &metadata)
	}

	paymentMethods, ok := metadata["paymentMethods"].([]interface{})
	if !ok || len(paymentMethods) == 0 {
		return nil
	}

	pm := paymentMethods[0].(map[string]interface{})

	card, ok := pm["card"].(map[string]interface{})
	if !ok {
		return nil
	}

	customer := &cardFundedPayoutModel.GetSavedCardResponse{
		ID:             c.UUID,
		CardName:       util.GetValueAsString(card, "cardName", ""),
		CardOrigin:     util.GetValueAsString(card, "cardOrigin", ""),
		IssuingBank:    util.GetValueAsString(card, "issuingBank", ""),
		Last4:          util.GetValueAsString(card, "last4", ""),
		ExpiryYear:     util.GetValueAsString(card, "expYear", ""),
		ExpiryMonth:    util.GetValueAsString(card, "expMonth", ""),
		PaymentChannel: util.GetValueAsString(pm, "paymentChannel", ""),
	}

	return customer
}

func FullNameToFirstNameAndLastName(s string) (string, string) {
	lastSpaceIndex := strings.LastIndex(s, " ")
	if lastSpaceIndex == -1 {
		return s, ""
	}
	beforeLastSpace := s[:lastSpaceIndex]
	afterLastSpace := s[lastSpaceIndex+1:]
	return beforeLastSpace, afterLastSpace
}

func FirstNameAndLastNameToFullName(firstName, lastName string) string {
	if lastName == "" {
		return firstName
	}
	return firstName + " " + lastName
}

func NewString(s string) *string {
	return &s
}

func CreateCustomer(c *CreateCustomerRequest) *Customer {
	now := time.Now()
	uuid, _ := uuid.NewV7()
	customer := &Customer{
		MerchantID:       c.MerchantID,
		Email:            c.Email,
		PhoneCountryCode: c.PhoneCountryCode,
		PhoneNumber:      util.CleanUpIDNPhoneNumber(c.PhoneNumber),
		FirstName:        c.FirstName,
		LastName:         c.LastName,
		BusinessName:     c.BusinessName,
		City:             c.City,
		Country:          c.Country,
		AddressLine1:     c.AddressLine1,
		AddressLine2:     c.AddressLine2,
		PostalCode:       c.PostalCode,
		State:            c.State,
		IsBlocked:        c.IsBlocked,
		BlockReason:      c.BlockReason,
		Metadata:         c.Metadata,
		UUID:             uuid.String(),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if customer.PhoneCountryCode == "" {
		customer.PhoneCountryCode = constant.DefaultPhoneCountryCode
	}

	return customer
}

func (c *Customer) Update(u *UpdateCustomerRequest) {
	now := time.Now()
	if u.Email != nil {
		c.Email = *u.Email
	}
	if u.PhoneNumber != nil {
		c.PhoneNumber = util.CleanUpIDNPhoneNumber(*u.PhoneNumber)
	}

	if u.PhoneCountryCode != nil {
		c.PhoneCountryCode = *u.PhoneCountryCode
	}

	if u.FirstName != nil {
		c.FirstName = *u.FirstName
	}
	if u.LastName != nil {
		c.LastName = *u.LastName
	}
	if u.BusinessName != nil {
		c.BusinessName = *u.BusinessName
	}
	if u.City != nil {
		c.City = *u.City
	}
	if u.Country != nil {
		c.Country = *u.Country
	}
	if u.AddressLine1 != nil {
		c.AddressLine1 = *u.AddressLine1
	}
	if u.AddressLine2 != nil {
		c.AddressLine2 = *u.AddressLine2
	}
	if u.PostalCode != nil {
		c.PostalCode = *u.PostalCode
	}
	if u.State != nil {
		c.State = *u.State
	}
	if u.Metadata != nil {
		refundPref, _ := util.ConvertToStruct[*unifiedPaymentModel.UnifiedPaymentRefundPreference]((u.Metadata)["refundPreference"])
		if refundPref == nil && c.Metadata != nil {
			// Notes:
			// Keep "refundPreference" always available in the customer metadata to avoid deletion by update
			u.Metadata["refundPreference"] = c.Metadata["refundPreference"]
		}
		c.Metadata = u.Metadata
	}
	if u.IsBlocked != nil {
		c.IsBlocked = *u.IsBlocked
	}
	if u.BlockReason != nil {
		c.BlockReason = *u.BlockReason
	}

	c.UpdatedAt = now
}

func (c *Customer) SetUnifiedPaymentCustomerInfo(unifiedPaymentResp *unifiedPaymentModel.UnifiedPaymentSessionResponse) {
	unifiedPaymentResp.CustomerId = &c.UUID
	unifiedPaymentResp.CustomerInformation = c.ToUnifiedPaymentCustomerResponse()
}
