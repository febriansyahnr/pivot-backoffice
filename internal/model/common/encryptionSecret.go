package commonModel

type EncryptionSecret struct {
	Payment EncryptionSecretPayment `json:"payment"`
}

type EncryptionSecretPayment struct {
	KeyEncryptionKey string `json:"keyEncryptionKey"`
}
