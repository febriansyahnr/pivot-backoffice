package merchant

type MerchantNotificationConfig struct {
	Transaction *MerchantNotificationTransactionConfig `json:"transaction,omitempty"`
	Balance     *MerchantNotificationBalanceConfig     `json:"balance,omitempty"`
}

type MerchantNotificationTransactionConfig struct {
	Recipient MerchantNotificationRecipient `json:"recipient"`
	Active    bool                          `json:"active"`
	Events    []string                      `json:"events"`
}

type MerchantNotificationBalanceConfig struct {
	Recipient MerchantNotificationRecipient `json:"recipient"`
	Threshold int                           `json:"threshold" validate:"min=0"`
}

type MerchantNotificationRecipient struct {
	Email []*MerchantNotificationEmailRecipient `json:"email" validate:"dive"`
	// placeholder for whatsapp and webhook
}

type MerchantNotificationEmailRecipient struct {
	Email string `json:"email" validate:"required,email"`
	Type  string `json:"type" validate:"required,oneof=PRIMARY CC"` // Primary, Optional
}
