package fdscommon

import (
	"time"

	card "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"

	"github.com/shopspring/decimal"
)

type CheckTransactionRequest struct {
	CardData           *card.CardDataRequest                              `json:"cardData,omitempty"`
	AuthorizationData  *card.PaymentNotificationAuthorizationDataRequest  `json:"authorizationData,omitempty"`
	AuthenticationData *card.PaymentNotificationAuthenticationDataRequest `json:"authenticationData,omitempty"`
	Device             *DeviceCheck                                       `json:"device,omitempty"`
	MidNumber          *string                                            `json:"midNumber,omitempty"`
	AcquiringName      *string                                            `json:"acquiringName,omitempty"`
	BillingInformation BillingInformation                                 `json:"billingInformation"`
	ThreeDsMethod      string                                             `json:"threeDsMethod,omitempty"`
}

func (c *CheckTransactionRequest) FromCcMetadata(ccMetadata *card.CreditcardMetadata) {
	c.CardData = ccMetadata.CardData
	c.AuthenticationData = ccMetadata.AuthenticationData
	c.AuthorizationData = ccMetadata.AuthorizationData
}

type CheckRequest struct {
	Account     AccountCheck          `json:"account"`
	Customer    CustomerCheck         `json:"customer"`
	Payment     PaymentCheck          `json:"payment"`
	Partner     PartnerCheck          `json:"partner"`
	Transaction TransactionCheck      `json:"transaction"`
	Device      DeviceCheck           `json:"device"`
	IB          IntermediaryBankCheck `json:"ib"`
	CP          CounterpartyBankCheck `json:"cp"`
	Custom      *CustomCheck          `json:"custom,omitempty"`
}

type CustomCheck struct {
	Number        *string `json:"mid.number,omitempty"`
	Type          *string `json:"mid.type,omitempty"`
	AcquiringName *string `json:"mid.acquiring_name,omitempty"`
}

type AccountCheck struct {
	BankCode         *string          `json:"bank_code,omitempty"`          // The bank code of correspondent bank
	BankCodeType     *string          `json:"bank_code_type,omitempty"`     // BIC, IBAN, Routing Number, SWIFT Code, custom
	AccountID        *string          `json:"account_id"`                   // Account number used for transaction
	CurrentBalance   *decimal.Decimal `json:"current_balance,omitempty"`    // Account balance at the time of current transaction
	DaysLate         *int             `json:"days_late,omitempty"`          // Number of days late for payment
	DueOn            *time.Time       `json:"due_on,omitempty"`             // Date the next account payment is due
	IncCountry       *string          `json:"inc_country,omitempty"`        // Country the business account was incorporated
	IncOn            *time.Time       `json:"inc_on,omitempty"`             // Date the business account was incorporathed
	IncRegion        *string          `json:"inc_region,omitempty"`         // Region or State the business account was incorporated
	Label            *string          `json:"label,omitempty"`              // Merchant specific name of the type of account
	LastBalance      *decimal.Decimal `json:"last_balance,omitempty"`       // Account balance as of last statement
	LastLogin        *time.Time       `json:"last_login,omitempty"`         // Date of the last account login
	LastMinDue       *decimal.Decimal `json:"last_min_due,omitempty"`       // Minimum amount due on last account statement
	LastPaymentAmt   *decimal.Decimal `json:"last_payment_amt,omitempty"`   // Last payment amount for the account
	LastPaymentOn    *time.Time       `json:"last_payment_on,omitempty"`    // Date of the last account payment
	LastStatementOn  *time.Time       `json:"last_statement_on,omitempty"`  // Date of the last account statement
	MinDue           *decimal.Decimal `json:"min_due,omitempty"`            // Minimum amount due on account at the time of current transaction
	OpenedOn         *time.Time       `json:"opened_on,omitempty"`          // UTC date and time the transaction placed
	TaxID            *string          `json:"tax_id,omitempty"`             // Account's tax identification number
	Type             *string          `json:"type,omitempty"`               // Categorization of the account type
	UnbilledCharges  *decimal.Decimal `json:"unbilled_charges,omitempty"`   // Fees on account that have not been paid
	ExposureAmt      *decimal.Decimal `json:"exposure_amt,omitempty"`       // Total exposure of the account
	AvailFunds       *decimal.Decimal `json:"avail_funds,omitempty"`        // Available funds for the account
	ClosedOn         *time.Time       `json:"closed_on,omitempty"`          // Date the account was closed
	CreditLimit      *decimal.Decimal `json:"credit_limit,omitempty"`       // Credit limit for the account
	IsActive         *bool            `json:"is_active,omitempty"`          // Is the account active?
	IsFraud          *bool            `json:"is_fraud,omitempty"`           // Was the account closed for fraud reasons?
	LateStatusLabel  *string          `json:"late_status_label,omitempty"`  // Merchant specific status for account delinquency
	Status           *string          `json:"status,omitempty"`             // Merchant specific term for account's current status
	PinChangeOn      *time.Time       `json:"pin_change_on,omitempty"`      // Date of PIN change
	EmailChangeOn    *time.Time       `json:"email_change_on,omitempty"`    // Date of email change
	AddressChangeOn  *time.Time       `json:"address_change_on,omitempty"`  // Date of address change
	PasswordChangeOn *time.Time       `json:"password_change_on,omitempty"` // Date of password change
	PhoneChangeOn    *time.Time       `json:"phone_change_on,omitempty"`    // Date of phone change
}

type CustomerCheck struct {
	Address1         *string          `json:"address1,omitempty"`           // Building number and street address
	Address2         *string          `json:"address2,omitempty"`           // Apartment/suite/unit number
	City             *string          `json:"city,omitempty"`               // Address city
	Company          *string          `json:"company,omitempty"`            // Company name, for business addresses
	Country          *string          `json:"country,omitempty"`            // Two-digit country code
	Email            *string          `json:"email,omitempty"`              // Email address
	FirstName        *string          `json:"first_name,omitempty"`         // Customer first name
	LastName         *string          `json:"last_name,omitempty"`          // Customer last name
	Phone            *string          `json:"phone,omitempty"`              // Primary phone number
	PostalCode       *string          `json:"postal_code,omitempty"`        // Zip or postal code
	Region           *string          `json:"region,omitempty"`             // State or province code
	ID               string           `json:"id,omitempty"`                 // Customer ID
	Phone2           *string          `json:"phone_2,omitempty"`            // Secondary phone number
	Phone3           *string          `json:"phone_3,omitempty"`            // Tertiary phone number
	StartedOn        *time.Time       `json:"started_on,omitempty"`         // Date/Time customer started
	Type             *string          `json:"type,omitempty"`               // Categorization of the account type
	DOB              *time.Time       `json:"dob,omitempty"`                // Date of birth
	AnnualIncome     *decimal.Decimal `json:"annual_income,omitempty"`      // Annual income
	EmailChangeOn    *time.Time       `json:"email_change_on,omitempty"`    // Date of email change
	AddressChangeOn  *time.Time       `json:"address_change_on,omitempty"`  // Date of address change
	PasswordChangeOn *time.Time       `json:"password_change_on,omitempty"` // Date of password change
	PhoneChangeOn    *time.Time       `json:"phone_change_on,omitempty"`    // Date of phone change
}

type PaymentCheck struct {
	Purpose           *string          `json:"purpose,omitempty"`        // The purpose of the payment or financial instrument
	ReferenceCode     *string          `json:"reference_code,omitempty"` // A reference number or identifier associated with the payment or financial instrument
	ChargeBearer      *string          `json:"charge_bearer,omitempty"`  // "BEN" for the beneficiary, "OUR" for the sender
	CardStatus        *string          `json:"card_status,omitempty"`    // The status of the card, eg. active, lost, stolen, etc.
	IsActive          *bool            `json:"is_active,omitempty"`
	PaymentStatus     *string          `json:"payment_status,omitempty"` // For fraudnet rules. Deprecated.
	ChargebackStatus  *string          `json:"chargeback_status,omitempty"`
	PaymentID         *string          `json:"payment_id,omitempty"`
	ThreeDsEci        *string          `json:"3DS_eci,omitempty"`
	ThreeDsVid        *string          `json:"3DS_vid,omitempty"`
	ThreeDsXid        *string          `json:"3DS_xid,omitempty"`
	ActualAmt         *decimal.Decimal `json:"actual_amt,omitempty"`
	ActualCcy         *string          `json:"actual_ccy,omitempty"`
	Arn               *string          `json:"arn,omitempty"`
	AuthAttempts      *int             `json:"auth_attempts,omitempty"`
	AuthFlag          *string          `json:"auth_flag,omitempty"`
	AuthResCode       *string          `json:"auth_res_code,omitempty"`
	AvsResultCode     *string          `json:"avs_result_code,omitempty"`
	BilledAmt         *decimal.Decimal `json:"billed_amt,omitempty"`
	BilledCcy         *string          `json:"billed_ccy,omitempty"`
	Bin               *string          `json:"bin,omitempty"`
	CardAccountID     *string          `json:"card_account_id,omitempty"`
	CardPresent       *string          `json:"card_present,omitempty"`
	CardProductType   *string          `json:"card_product_type,omitempty"`
	ChAuth            *string          `json:"ch_auth,omitempty"`
	ChPresent         *string          `json:"ch_present,omitempty"`
	CvvResultCode     *string          `json:"cvv_result_code,omitempty"`
	Eci               *string          `json:"eci,omitempty"`
	EmvAid            *string          `json:"emv_aid,omitempty"`
	EmvChipID         *string          `json:"emv_chip_id,omitempty"`
	EposRec           *string          `json:"epos_rec,omitempty"`
	ExpDate           *string          `json:"exp_date,omitempty"`
	GatewayMessage    *string          `json:"gateway_message,omitempty"`
	GiftCardsNumbers  *string          `json:"gift_cards_numbers,omitempty"`
	GiftCards         *[]string        `json:"gift_cards,omitempty"`
	InputCapabilities *string          `json:"input_capabilities,omitempty"`
	IssueNumber       *string          `json:"issue_number,omitempty"`
	Last4             string           `json:"last_4,omitempty"`
	First8            string           `json:"first_8,omitempty"`
	MerchantAccount   *string          `json:"merchant_account,omitempty"`
	Method            *string          `json:"method,omitempty"` // For fraudnet rules. Deprecated.
	ThreeDsMethod     string           `json:"threeDsMethod,omitempty"`
	PinStatus         *string          `json:"pin_status,omitempty"`
	ServiceCode       *string          `json:"service_code,omitempty"`
	TerminalID        *string          `json:"terminal_id,omitempty"`
	TerminalMethod    *string          `json:"terminal_method,omitempty"`
	TerminalOption    *string          `json:"terminal_option,omitempty"`
	TerminalType      *string          `json:"terminal_type,omitempty"`
	TokenID           *string          `json:"token_id,omitempty"`
	TransactionLabel  *string          `json:"transaction_label,omitempty"`
	TransactionType   *string          `json:"transaction_type,omitempty"`
	Type              *string          `json:"type,omitempty"` // For fraudnet rules. Deprecated.
	AuthCode          *string          `json:"auth_code,omitempty"`
	ActiveOn          *time.Time       `json:"active_on,omitempty"`
	Direction         *string          `json:"direction,omitempty"` // For fraudnet rules. Deprecated.
	// Internal
	MethodType       string `json:"-"`
	Fingerprint      string `json:"-"`
	MaskedCardNumber string `json:"-"`
	CardBrand        string `json:"-"`
	CardCountryCode  string `json:"-"`
	CardType         string `json:"-"`
	CardIssuing      string `json:"-"`
}

type PartnerCheck struct {
	Address1    *string `json:"address1,omitempty"` // Building number and street address
	Address2    *string `json:"address2,omitempty"` // Apartment/suite/unit number
	City        *string `json:"city,omitempty"`     // Address city
	Company     *string `json:"company,omitempty"`  // Company name, for business addresses
	Country     *string `json:"country,omitempty"`  // Two-digit country code
	Email       *string `json:"email,omitempty"`    // Email address
	Industry    *string `json:"industry,omitempty"`
	MccID       *string `json:"mcc_id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Phone       *string `json:"phone,omitempty"`       // Primary phone number
	PostalCode  *string `json:"postal_code,omitempty"` // Zip or postal code
	Region      *string `json:"region,omitempty"`      // State or province code
	Sector      *string `json:"sector,omitempty"`
	Event       *string `json:"event,omitempty"`
	ID          string  `json:"id,omitempty"`
	TaxID       *string `json:"tax_id,omitempty"`
	AccountID   *string `json:"account_id,omitempty"`
	AccountType *string `json:"account_type,omitempty"`
	Location    *string `json:"location,omitempty"`
	RiskLevel   string  `json:"-"`
}

type TransactionCheck struct {
	OrderID           string           `json:"order_id,omitempty"`
	OrderedOn         *time.Time       `json:"ordered_on,omitempty"`
	Type              *string          `json:"type,omitempty"` // Default value: authentication (fraudnet rules). Deprecated.
	OrderIsDigital    *bool            `json:"order_is_digital,omitempty"`
	OrderTotal        *decimal.Decimal `json:"order_total,omitempty"`
	OrderCurrency     *string          `json:"order_currency,omitempty"`
	Status            *string          `json:"status,omitempty"` // Refer to transaction status = PENDING, SUCCESS, FAILED
	Event             *string          `json:"event,omitempty"`
	UserID            *string          `json:"user_id,omitempty"`
	OrderSource       *string          `json:"order_source,omitempty"`
	OrderCount        *int             `json:"order_count,omitempty"`
	TotalSpent        *decimal.Decimal `json:"total_spent,omitempty"`
	SessionID         *string          `json:"session_id,omitempty"`
	FirstPurchaseDate *time.Time       `json:"first_purchase_date,omitempty"`
	LastPurchaseDate  *time.Time       `json:"last_purchase_date,omitempty"`
	UserLocale        *string          `json:"user_locale,omitempty"`
	CouponCode        *string          `json:"coupon_code,omitempty"`
	OrderDiscount     *decimal.Decimal `json:"order_discount,omitempty"`
	OrderShipping     *decimal.Decimal `json:"order_shipping,omitempty"`
	OrderSubtotal     *decimal.Decimal `json:"order_subtotal,omitempty"`
	OrderTax          *decimal.Decimal `json:"order_tax,omitempty"`
	ShippedOn         *time.Time       `json:"shipped_on,omitempty"`
	AgentCode         *string          `json:"agent_code,omitempty"`
	AgentDept         *string          `json:"agent_dept,omitempty"`
	IdentID           *string          `json:"ident_id,omitempty"`
	IdentCountry      *string          `json:"ident_country,omitempty"`
	IdentType         *string          `json:"ident_type,omitempty"`
	Iban              *string          `json:"iban,omitempty"`
	TransactionID     *string          `json:"transaction_id,omitempty"`
	Fee               *decimal.Decimal `json:"fee,omitempty"`
	Geo               *string          `json:"geo,omitempty"`
	Reference         *string          `json:"reference,omitempty"`
	TransferType      *string          `json:"transfer_type,omitempty"`
	Frequency         *string          `json:"frequency,omitempty"`
	BatchCode         *string          `json:"batch_code,omitempty"`
	PaymentStatus     *string          `json:"payment_status,omitempty"`

	// Internal Used
	ID                string    `json:"-"`
	ClientReferenceID string    `json:"-"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

type DeviceCheck struct {
	IPAddress           *string `json:"ip_address,omitempty"`
	Resolution          *string `json:"resolution,omitempty"`
	UserAgent           *string `json:"user_agent,omitempty"`
	Service             *string `json:"service,omitempty"` // Allowed: fraudnet, infoscore, inauth, none, custom
	ClientID            *string `json:"client_id,omitempty"`
	SessionID           *string `json:"session_id,omitempty"`
	FingerprintID       *string `json:"fingerprint_id,omitempty"`
	IPType              string  `json:"ip_type"` // Allowed: v4, v6
	PluginHash          *string `json:"plugin_hash,omitempty"`
	TimeZone            *string `json:"time_zone,omitempty"`
	Language            *string `json:"language,omitempty"`
	IsProxy             *bool   `json:"is_proxy,omitempty"`
	HTTPReferer         *string `json:"http_referer,omitempty"`
	NumMIMETypes        *string `json:"num_mime_types,omitempty"`
	MIMETypesHash       *string `json:"mime_types_hash,omitempty"`
	NumFonts            *int    `json:"num_fonts,omitempty"`
	FontsHash           *string `json:"fonts_hash,omitempty"`
	NumPlugins          *int    `json:"num_plugins,omitempty"`
	PluginsHash         *string `json:"plugins_hash,omitempty"`
	ColorDepth          *int    `json:"color_depth,omitempty"`
	FontSmoothing       *bool   `json:"font_smoothing,omitempty"`
	JavaSupport         *bool   `json:"java_support,omitempty"`
	TouchSupport        *bool   `json:"touch_support,omitempty"`
	CookieSupport       *bool   `json:"cookie_support,omitempty"`
	CanvasFingerprintID *string `json:"canvas_fingerprint_id,omitempty"`
	CanvasHeight        *int    `json:"canvas_height,omitempty"`
	CanvasWidth         *int    `json:"canvas_width,omitempty"`
	ScreenHeight        *int    `json:"screen_height,omitempty"`
	ScreenWidth         *int    `json:"screen_width,omitempty"`
	IsTor               *bool   `json:"is_tor,omitempty"`
	Geo                 *string `json:"geo,omitempty"` // Format: "lat,lon"
	City                *string `json:"city,omitempty"`
	Country             *string `json:"country,omitempty"`
	PostalCode          *string `json:"postal_code,omitempty"`
	Timezone            *string `json:"timezone,omitempty"`
	ProxyType           *string `json:"proxy_type,omitempty"` // Allowed: vpn, tor, dch, pub, web, ses, ib
}

type BillingInformation struct {
	GivenName     string `json:"givenName,omitempty"`
	Surname       string `json:"surname,omitempty"`
	Email         string `json:"email,omitempty"`
	Country       string `json:"country,omitempty"`
	Address1      string `json:"address1,omitempty"`
	City          string `json:"city,omitempty"`
	ProvinceState string `json:"provinceState,omitempty"`
	PostalCode    string `json:"postalCode,omitempty"`
}

type IntermediaryBankCheck struct {
	ID         *string `json:"id,omitempty"`
	Code       *string `json:"code,omitempty"`
	CodeType   *string `json:"code_type,omitempty"` // BIC, IBAN, Routing Number, SWIFT Code, custom
	Name       *string `json:"name,omitempty"`
	Address    *string `json:"address,omitempty"`
	City       *string `json:"city,omitempty"`
	PostalCode *string `json:"postal_code,omitempty"`
	Region     *string `json:"region,omitempty"`
	Country    *string `json:"country,omitempty"`
}

type CounterpartyBankCheck struct {
	ID             *string `json:"id,omitempty"`
	Address1       *string `json:"address1,omitempty"`
	Address2       *string `json:"address2,omitempty"`
	City           *string `json:"city,omitempty"`
	Country        *string `json:"country,omitempty"`
	MCCID          *string `json:"mcc_id,omitempty"`
	Name           *string `json:"name,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	PostalCode     *string `json:"postal_code,omitempty"`
	Region         *string `json:"region,omitempty"`
	TaxID          *string `json:"tax_id,omitempty"`
	AccountType    *string `json:"account_type,omitempty"`
	Company        *string `json:"company,omitempty"`
	Email          *string `json:"email,omitempty"`
	Industry       *string `json:"industry,omitempty"`
	Sector         *string `json:"sector,omitempty"`
	AccountNumber  *string `json:"account_number,omitempty"`
	BankCode       *string `json:"bank_code,omitempty"`
	BankCodeType   *string `json:"bank_code_type,omitempty"` // BIC, IBAN, Routing Number, SWIFT Code, custom
	BankName       *string `json:"bank_name,omitempty"`
	BankAddress    *string `json:"bank_address,omitempty"`
	BankCity       *string `json:"bank_city,omitempty"`
	BankPostalCode *string `json:"bank_postal_code,omitempty"`
	BankRegion     *string `json:"bank_region,omitempty"`
	BankCountry    *string `json:"bank_country,omitempty"`
}
