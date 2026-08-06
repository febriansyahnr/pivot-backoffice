package config

type Secret struct {
	MySQLSecret                   MySQLSecret                   `mapstructure:"DATABASE"`
	RedisSecret                   RedisSecret                   `mapstructure:"REDIS"`
	RabbitMQSecret                RabbitMQSecret                `mapstructure:"RABBITMQ"`
	RabbitMQStream                RabbitMQSecret                `mapstructure:"RABBITMQ_STREAM"`
	SecuritySecret                SecuritySecret                `mapstructure:"SECURITY"`
	MongoDBSecret                 MongoDBSecret                 `mapstructure:"MONGODB"`
	SnapCoreSecret                SnapCoreSecret                `mapstructure:"SNAP_CORE"`
	InternalApiKeySecret          InternalApiKeySecret          `mapstructure:"INTERNAL_API_KEY"`
	JWTSignatureKey               JWTSignatureKey               `mapstructure:"JWT_SIGNATURE_KEY"`
	CrmSecret                     CrmSecret                     `mapstructure:"CRM"`
	ConsulSecret                  ConsulSecret                  `mapstructure:"CONSUL"`
	XbCoreProcessorSecret         XbCoreProcessorSecret         `mapstructure:"XB_CORE_PROCESSOR"`
	NewRelicLicenseKey            string                        `mapstructure:"NEW_RELIC_LICENSE_KEY"`
	StatsdHost                    string                        `mapstructure:"STATSD_HOST"`
	StatsdPort                    string                        `mapstructure:"STATSD_PORT"`
	WalletBackendSecret           WalletBackendSecret           `mapstructure:"WALLET_BACKEND"`
	GoogleKMSKey                  string                        `mapstructure:"GOOGLE_KMS_KEY"`
	FraudNetSecret                FraudNetSecret                `mapstructure:"FRAUD_NET"`
	Sokratech                     SokratechSecret               `mapstructure:"SOKRATECH"`
	CreditcardCoreProcessorSecret CreditcardCoreProcessorSecret `mapstructure:"CREDIT_CARD_CORE"`
	AdvanceAISecret               AdvanceAISecret               `mapstructure:"ADVANCE_AI"`
	Dukcapil                      DukcapilSecret                `mapstructure:"DUKCAPIL"`
	Vault                         VaultSecret                   `mapstructure:"VAULT"`
	Conductor                     ConductorSecret               `mapstructure:"CONDUCTOR"`
	Payment                       PaymentSecret                 `mapstructure:"PAYMENT"`
}

type MySQLSecret struct {
	Database string `mapstructure:"DB_NAME"`
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
}

type RedisSecret struct {
	Password string `mapstructure:"PASSWORD"`
}

type RabbitMQSecret struct {
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
}

type SecuritySecret struct {
	CardSecretKey  string `mapstructure:"CARD_SECRET_KEY"`
	CardIV         string `mapstructure:"CARD_IV"`
	RespEncryptKey string `mapstructure:"RESPONSE_ENCRYPT_KEY"`
}

type MongoDBSecret struct {
	Database string `mapstructure:"DATABASE"`
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
}
