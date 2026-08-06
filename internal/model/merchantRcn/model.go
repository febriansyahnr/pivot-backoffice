package merchantRcn

import (
	"time"

	"github.com/google/uuid"
)

type MerchantRcn struct {
	ID                uuid.UUID  `db:"uuid" json:"id"`
	MerchantID        uuid.UUID  `db:"merchant_id" json:"merchantId"`
	PrincipalIssuer   string     `db:"principal_issuer" json:"principalIssuer"`
	RealCardNumber    string     `db:"real_card_number" json:"realCardNumber"`
	EncryptKMSVersion string     `db:"encrypt_kms_version" json:"encryptKMSVersion"`
	CreatedAt         time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt         *time.Time `db:"deleted_at" json:"-"`
}

type MerchantRcnDetail struct {
	ID              uuid.UUID `json:"id"`
	MerchantID      uuid.UUID `json:"merchantId"`
	PrincipalIssuer string    `json:"principalIssuer"`
	CardNumber      string    `json:"realCardNumber"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (m *MerchantRcnDetail) EraseSensitiveData() {
	m.CardNumber = ""
}
