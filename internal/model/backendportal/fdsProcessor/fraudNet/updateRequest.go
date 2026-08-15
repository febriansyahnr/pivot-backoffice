package fraudnetmodel

import (
	"time"

	"github.com/shopspring/decimal"
)

type MarketplaceUpdateRequest struct {
	OrderID        string           `json:"order_id"`                   // Required: Transaction/order ID
	AgentCode      *string          `json:"agent_code,omitempty"`       // Code of internal agent (nullable)
	AgentUpdatedOn *time.Time       `json:"agent_updated_on,omitempty"` // Time when agent updated the order
	AgentDept      *string          `json:"agent_dept,omitempty"`       // Department of the agent
	Event          *string          `json:"event,omitempty"`            // Event type (e.g., shipped, cancelled)
	IsFraud        *bool            `json:"is_fraud,omitempty"`         // Is the order fraudulent?
	IsLocked       *bool            `json:"is_locked,omitempty"`        // Is the order locked internally?
	OrderTotal     *decimal.Decimal `json:"order_total,omitempty"`      // Current order total
	UpdatedOn      *time.Time       `json:"updated_on,omitempty"`       // When the order status was updated
	Status         string           `json:"status"`                     // Order status (required, has defaults)
	FraudType      *string          `json:"fraud_type,omitempty"`       // Fraud category
	Note           *string          `json:"note,omitempty"`             // Additional note

	Payment *PaymentMarketplaceUpdate `json:"payment,omitempty"` // Payment info
	Account *AccountMarketplaceUpdate `json:"account,omitempty"` // Account info

	AdjAmt *decimal.Decimal `json:"adj_amt,omitempty"`
}

type PaymentMarketplaceUpdate struct {
	CardStatus       *string `json:"card_status,omitempty"`       // Status of card (e.g. stolen, lost)
	IsActive         *bool   `json:"is_active,omitempty"`         // Is card active?
	PaymentStatus    string  `json:"payment_status"`              // Status of payment (required)
	ChargebackStatus *string `json:"chargeback_status,omitempty"` // Chargeback status (nullable)
}

type AccountMarketplaceUpdate struct {
	AvailFunds       *decimal.Decimal `json:"avail_funds,omitempty"`       // Available funds
	ClosedOn         *time.Time       `json:"closed_on,omitempty"`         // Account closed date
	CreditLimit      *decimal.Decimal `json:"credit_limit,omitempty"`      // Credit limit
	IsActive         *bool            `json:"is_active,omitempty"`         // Is account active?
	IsFraud          *bool            `json:"is_fraud,omitempty"`          // Was account closed due to fraud?
	LateStatusLabel  *string          `json:"late_status_label,omitempty"` // Custom late status
	Status           *string          `json:"status,omitempty"`            // Current account status
	PINChangeOn      *time.Time       `json:"pin_change_on,omitempty"`
	EmailChangeOn    *time.Time       `json:"email_change_on,omitempty"`
	AddressChangeOn  *time.Time       `json:"address_change_on,omitempty"`
	PasswordChangeOn *time.Time       `json:"password_change_on,omitempty"`
	PhoneChangeOn    *time.Time       `json:"phone_change_on,omitempty"`
	CurrentBalance   *decimal.Decimal `json:"current_balance,omitempty"` // Current account balance
}
