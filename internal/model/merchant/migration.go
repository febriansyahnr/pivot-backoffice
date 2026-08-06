package merchant

type UnencryptedMerchantSecretsForMigration struct {
	MerchantId            string `db:"uuid"`
	MerchantName          string `db:"name"`
	CallbackApiKey        string `db:"callback_api_key"`
	CallbackApiKeyVersion uint   `db:"callback_api_key_version"`
	JITApiKey             string `db:"jit_api_key"`
	JITApiKeyVersion      uint   `db:"jit_api_key_version"`
	Secret                string `db:"secret"`
	SecretVersion         uint   `db:"secret_version"`
}

type MigrateMerchantSecretsToEncryption UnencryptedMerchantSecretsForMigration
