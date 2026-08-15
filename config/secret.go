package config

type Secret struct {
	AISecret                      AISecret                      `mapstructure:"AI_SECRET"`
	MySQLSecret                   Databases                     `mapstructure:"DATABASE"`
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
	CreditcardCoreProcessorSecret CreditcardCoreProcessorSecret `mapstructure:"CREDIT_CARD_CORE"`
	Vault                         VaultSecret                   `mapstructure:"VAULT"`
	Conductor                     ConductorSecret               `mapstructure:"CONDUCTOR"`
	Payment                       PaymentSecret                 `mapstructure:"PAYMENT"`
}

type AISecret struct {
	AnthropicAPIKey string `mapstructure:"ANTHROPIC_API_KEY"`
	OpenAIKey       string `mapstructure:"OPENAI_KEY"`
	ZaiKey          string `mapstructure:"ZAI_KEY"`
}

type Databases struct {
	Service       MySQLSecret `mapstructure:"SERVICE"`
	BackendPortal MySQLSecret `mapstructure:"BACKEND_PORTAL"`
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

type CrmSecret struct {
	ApiKey string `mapstructure:"API_KEY"`
}

type ConsulSecret struct {
	Token string `mapstructure:"TOKEN"`
}

type XbCoreProcessorSecret struct {
	InternalServiceKey string `mapstructure:"INTERNAL_SERVICE_KEY"`
}

type WalletBackendSecret struct {
	InternalServiceKey string `mapstructure:"INTERNAL_SERVICE_KEY"`
}

type CreditcardCoreProcessorSecret struct {
	InternalServiceKey        string `mapstructure:"INTERNAL_SERVICE_KEY"`
	EncryptionPublicKeySecret string `mapstructure:"ENCRYPTION_PUBLIC_KEY_SECRET"`
	EncryptionPublicKeyIV     string `mapstructure:"ENCRYPTION_PUBLIC_KEY_IV"`
}

type VaultSecret struct {
	Token string `mapstructure:"TOKEN"`
}

type ConductorSecret struct {
	BasicAuth *ConductorSecretBasicAuth `mapstructure:"BASIC_AUTH"`
}

type ConductorSecretBasicAuth struct {
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
}

type PaymentSecret struct {
	KeyEncryptionKey string `mapstructure:"KEY_ENCRYPTION_KEY"`
}
