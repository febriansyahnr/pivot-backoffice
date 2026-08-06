package config

import (
	"strings"
	"time"
)

type Config struct {
	Environment      string     `mapstructure:"ENVIRONMENT"`
	ServiceName      string     `mapstructure:"SERVICE_NAME"`
	ServiceVersion   string     `mapstructure:"SERVICE_VERSION"`
	Host             HostConfig `mapstructure:"HOST"`
	GracefulWaitTime int        `mapstructure:"GRACEFUL_WAIT_TIME"`
	ServerHost       string     `mapstructure:"SERVER_HOST"`
	ServerPort       string     `mapstructure:"SERVER_PORT"`

	MySQLConfig                   MySQLConfig                    `mapstructure:"DATABASE"`
	RedisConfig                   RedisConfig                    `mapstructure:"REDIS"`
	RabbitMQConfig                RabbitMQConfig[string]         `mapstructure:"RABBITMQ"`
	RabbitMQStream                RabbitMQConfig[int]            `mapstructure:"RABBITMQ_STREAM"`
	MongoDBConfig                 MongoDBConfig                  `mapstructure:"MONGODB"`
	AppConfig                     AppConfig                      `mapstructure:"APP"`
	SnapCoreConfig                SnapCoreConfig                 `mapstructure:"SNAP_CORE"`
	GCSConfig                     GCSConfig                      `mapstructure:"GCS"`
	BigQueryConfig                BigQueryConfig                 `mapstructure:"BIGQUERY"`
	PermissionConfig              PermissionConfig               `mapstructure:"PERMISSION"`
	SlackConfig                   SlackConfig                    `mapstructure:"SLACK"`
	DictionaryConfig              DictionaryConfig               `mapstructure:"DICTIONARY"`
	UserOTPConfig                 UserOTPConfig                  `mapstructure:"USER_OTP_CONFIG"`
	PaperCommunication            PaperCommunication             `mapstructure:"PAPER_COMMUNICATION"`
	CreditcardConfig              CreditcardConfig               `mapstructure:"CREDIT_CARD"`
	MerchantPortalConfig          MerchantPortalConfig           `mapstructure:"MERCHANT_PORTAL"`
	FeatureFlagConfig             FeatureFlagConfig              `mapstructure:"FEATURE_FLAG"`
	WorkerPoolConfig              WorkerPoolConfig               `mapstructure:"WORKER_POOL"`
	OTLPConfig                    OTLPConfig                     `mapstructure:"OTLP"`
	XbCoreProcessorConfig         XbCoreProcessorConfig          `mapstructure:"XB_CORE_PROCESSOR"`
	CreditcardCoreProcessorConfig CreditcardCoreProcessorConfig  `mapstructure:"CREDITCARD_CORE_PROCESSOR"`
	PaymentUIConfig               PaymentUIConfig                `mapstructure:"PAYMENT_UI"`
	PaymentSettlementConfig       PaymentSettlementConfig        `mapstructure:"PAYMENT_SETTLEMENT"`
	MerchantConfig                MerchantConfig                 `mapstructure:"MERCHANT"`
	WithdrawalConfig              WithdrawalConfig               `mapstructure:"WITHDRAWAL"`
	DisbursementConfig            DisbursementConfig             `mapstructure:"DISBURSEMENT"`
	UnifiedPaymentConfig          UnifiedPaymentConfig           `mapstructure:"UNIFIED_PAYMENT_CONFIG"`
	MonitoringConfig              MonitoringConfig               `mapstructure:"MONITORING"`
	RateLimit                     RateLimitConfig                `mapstructure:"RATE_LIMIT"`
	WalletBackendConfig           WalletBackendConfig            `mapstructure:"WALLET_BACKEND"`
	FraudNetConfig                FraudNetConfig                 `mapstructure:"FRAUD_NET"`
	Sokratech                     SokratechConfig                `mapstructure:"SOKRATECH"`
	CreditCardReferences          CreditCardReferences           `mapstructure:"CREDIT_CARD_REFERENCES"`
	InstallmentFee                InstallmentFee                 `mapstructure:"INSTALLMENT_FEE"`
	FdsConfig                     FdsConfig                      `mapstructure:"FDS"`
	AdvanceAIConfig               AdvanceAIConfig                `mapstructure:"ADVANCE_AI"`
	WhitelistedIPs                []string                       `mapstructure:"WHITELISTED_IPS"`
	PaymentSimulationConfig       PaymentSimulationConfig        `mapstructure:"PAYMENT_SIMULATION"`
	GCPConfig                     GCPConfig                      `mapstructure:"GCP_CONFIG"`
	PaymentFeeDefaults            PaymentFeeDefaultConfig        `mapstructure:"PAYMENT_FEE_DEFAULTS"`
	ReconConfig                   ReconConfig                    `mapstructure:"RECON"`
	Dukcapil                      DukcapilConfig                 `mapstructure:"DUKCAPIL"`
	ReorganizeTablePartition      ReorganizeTablePartitionConfig `mapstructure:"REORGANIZE_TABLE_PARTITION"`
	Vault                         VaultConfig                    `mapstructure:"VAULT"`
	MultiFactorAuth               MultiFactorAuthConfig          `mapstructure:"MULTI_FACTOR_AUTHENTICATION"`
	Conductor                     ConductorConfig                `mapstructure:"CONDUCTOR"`
	MerchantCallbackTask          MerchantCallbackTaskConfig     `mapstructure:"MERCHANT_CALLBACK_TASK"`
	ShortLinkRedirection          ShortLinkRedirectionConfig     `mapstructure:"SHORT_LINK_REDIRECTION"`
	Investigation                 InvestigationConfig            `mapstructure:"INVESTIGATION"`
	VccSlackRecipient             VccSlackRecipientConfig        `mapstructure:"VCC_SLACK_RECIPIENT"`
	VccTerminal                   VccTerminalConfig              `mapstructure:"VCC_TERMINAL"`
	ReportingConsumers            ReportingConsumers             `mapstructure:"REPORTING_CONSUMERS"`
	CardFundedPayout              CardFundedPayoutConfig         `mapstructure:"CARD_FUNDED_PAYOUT"`
	AutoSplitPayment              AutoSplitPaymentConfig         `mapstructure:"AUTO_SPLIT_PAYMENT"`
	HTTPClients                   HTTPClientsConfig              `mapstructure:"HTTP_CLIENTS"`
}

type SlackConfig struct {
	PaymentNotifWebhookURL                         string `mapstructure:"PAYMENT_NOTIF_WEBHOOK_URL"`
	ManualTopupNotifWebhookURL                     string `mapstructure:"MANUAL_TOPUP_NOTIF_WEBHOOK_URL"`
	XBPayoutStatusUpdateWebhookURL                 string `mapstructure:"XB_PAYOUT_STATUS_UPDATE_WEBHOOK_URL"`
	PayoutCutOffTimeWebHookURL                     string `mapstructure:"PAYOUT_CUT_OFF_TIME_WEBHOOK_URL"`
	BeneficiaryPayoutLimitWebHookURL               string `mapstructure:"BENEFICIARY_PAYOUT_LIMIT_WEBHOOK_URL"`
	FDSAlertWebhookURL                             string `mapstructure:"FDS_ALERT_WEBHOOK_URL"`
	PayoutAlertWebHookURL                          string `mapstructure:"PAYOUT_ALERT_WEBHOOK_URL"`
	PGAlertWebHookURL                              string `mapstructure:"PG_ALERT_WEBHOOK_URL"`
	WithdrawalAlertWebHookURL                      string `mapstructure:"WITHDRAWAL_ALERT_WEBHOOK_URL"`
	TopUpNotifWebhookURL                           string `mapstructure:"TOPUP_NOTIF_WEBHOOK_URL"`
	VccSettlementTransactionInquiryAlertWebhookURL string `mapstructure:"VCC_SETTLEMENT_TRANSACTION_INQUIRY_ALERT_WEBHOOK_URL"`
}

type HostConfig struct {
	Local      string `mapstructure:"LOCAL"`
	Staging    string `mapstructure:"STAGING"`
	Production string `mapstructure:"PRODUCTION"`
}

type MySQLConfig struct {
	Dialect      string `mapstructure:"DIALECT"`
	Host         string `mapstructure:"HOST"`
	Port         string `mapstructure:"PORT"`
	MaxIdleConns int    `mapstructure:"MAX_IDLE_CONNS"`
	MaxOpenConns int    `mapstructure:"MAX_OPEN_CONNS"`
	MaxIdleTime  int    `mapstructure:"MAX_IDLE_TIME"`
	MaxLifeTime  int    `mapstructure:"MAX_LIFE_TIME"`
	SlaveHost    string `mapstructure:"SLAVE_HOST"`
	SlavePort    string `mapstructure:"SLAVE_PORT"`
}

type RedisConfig struct {
	Host            string `mapstructure:"HOST"`
	Port            string `mapstructure:"PORT"`
	CacheDB         int    `mapstructure:"CACHE_DB"`
	MaxRetries      int    `mapstructure:"MAX_RETRIES"`
	MinRetryBackoff int    `mapstructure:"MIN_RETRY_BACKOFF"`
	MaxRetryBackoff int    `mapstructure:"MAX_RETRY_BACKOFF"`
	DialTimeout     int    `mapstructure:"DIAL_TIMEOUT"`
	ReadTimeout     int    `mapstructure:"READ_TIMEOUT"`
	WriteTimeout    int    `mapstructure:"WRITE_TIMEOUT"`
	PoolSize        int    `mapstructure:"POOL_SIZE"`
	PoolTimeout     int    `mapstructure:"POOL_TIMEOUT"`
}

type RabbitMQConfig[T string | int] struct {
	ServiceName        string `mapstructure:"-"`
	Host               string `mapstructure:"HOST"`
	Port               T      `mapstructure:"PORT"`
	VHost              string `mapstructure:"VHOST"`
	HeartbeatInSeconds int    `mapstructure:"HEARTBEAT_IN_SECONDS"`
}

type MongoDBConfig struct {
	Host string `mapstructure:"HOST"`
	Port string `mapstructure:"PORT"`
}

type AppConfig struct {
	PaginationPerPage                   int64  `mapstructure:"PAGINATION_PER_PAGE"`
	MaxPaginationPerPage                int    `mapstructure:"MAX_PAGINATION_PER_PAGE"`
	BulkDisbursementExpireLockMinute    int64  `mapstructure:"BULK_DISBURSEMENT_EXPIRE_LOCK_MINUTE"`
	DisbursementProcessExpireLockSecond int64  `mapstructure:"DISBURSEMENT_PROCESS_EXPIRE_LOCK_SECOND"`
	MaskingSensitiveData                string `mapstructure:"MASKING_SENSITIVE_DATA"`
	PdkLoggerUsed                       string `mapstructure:"PDK_LOGGER_USED"` // Oneof: ZAP or SLOGGER. Default: ZAP
	UseOverFetchPagination              bool   `mapstructure:"USE_OVER_FETCH_PAGINATION"`
	InitialPageWindow                   int64  `mapstructure:"INITIAL_PAGE_WINDOW"`
}

type SnapCoreConfig struct {
	BaseUrl string `mapstructure:"BASE_URL"`
}

type XbCoreProcessorConfig struct {
	BaseUrl         string  `mapstructure:"BASE_URL"`
	BaseUSDRate     float64 `mapstructure:"DEFAULT_USD_RATE"`
	DefaultLocalFee float64 `mapstructure:"DEFAULT_LOCAL_FEE"`
	DefaultSwiftFee float64 `mapstructure:"DEFAULT_SWIFT_FEE"`
	// default extended expired at in minutes
	ExtendedExpireAt time.Duration `mapstructure:"EXTENDED_EXPIRED_AT"`
}

// GetExtendedExpireAt returns extended expire at in time.Duration. Default is 5 minutes
func (c *XbCoreProcessorConfig) GetExtendedExpireAt() time.Duration {
	if c.ExtendedExpireAt == 0 {
		return 5 * time.Minute
	}
	return c.ExtendedExpireAt * time.Minute
}

type CreditcardCoreProcessorConfig struct {
	BaseUrl string `mapstructure:"BASE_URL"`
}

type CardFundedPayoutConfig struct {
	ExpiryAfterMinutes    int                             `mapstructure:"EXPIRY_AFTER_MINUTES"`
	PartnerProcessorLimit CardPartnerProcessorLimitConfig `mapstructure:"PARTNER_PROCESSOR_LIMIT"`
}

type CardPartnerProcessorLimitConfig struct {
	MPGS float64 `mapstructure:"MPGS"`
	CYBS float64 `mapstructure:"CYBS"`
}

type PaymentUIConfig struct {
	PaymentLinkURL       string `mapstructure:"PAYMENT_LINK_URL"`
	PaymentSuccessURL    string `mapstructure:"PAYMENT_SUCCESS_URL"`
	PaymentFailedURL     string `mapstructure:"PAYMENT_FAILED_URL"`
	PaymentProcessingURL string `mapstructure:"PAYMENT_PROCESSING_URL"`
}

type PaymentAutoInquiryConfig struct {
	CooldownSeconds int `mapstructure:"COOLDOWN_SECONDS"`
}

type MonitoringConfig struct {
	IsEnabled             bool `mapstructure:"IS_ENABLED"`
	MaxBytesPerPayload    int  `mapstructure:"MAX_BYTES_PER_PAYLOAD"`
	MaxMessagesPerPayload int  `mapstructure:"MAX_MESSAGES_PER_PAYLOAD"`
	WriteTimeout          int  `mapstructure:"WRITE_TIMEOUT"`
}

var installmentDefaultChannelFeeConfig map[string]InstallmentDefaultFeeConfig

func GetInstallmentDefaultChannelFeeConfig() map[string]InstallmentDefaultFeeConfig {
	return installmentDefaultChannelFeeConfig
}

var installmentDefaultFeeConfig InstallmentDefaultFeeConfig

func GetInstallmentDefaultFeeConfig() InstallmentDefaultFeeConfig {
	return installmentDefaultFeeConfig
}

type InstallmentFee struct {
	Default InstallmentDefaultFeeConfig            `mapstructure:"DEFAULT"`
	Channel map[string]InstallmentChannelFeeConfig `mapstructure:"CHANNEL"`
}

type InstallmentDefaultFeeConfig struct {
	Amount     float64 `mapstructure:"AMOUNT"`
	Percentage float64 `mapstructure:"PERCENTAGE"`
}

type InstallmentChannelFeeConfig struct {
	Tenor      []int     `mapstructure:"TENOR"`
	Amount     []float64 `mapstructure:"AMOUNT"`
	Percentage []float64 `mapstructure:"PERCENTAGE"`
}

type ReportingConsumers struct {
	ReportBalanceHistory ReportingConsumerConfig `mapstructure:"REPORT_BALANCE_HISTORY"`
}

type ReportingConsumerConfig struct {
	Enabled         bool          `mapstructure:"ENABLED"`
	StreamQueueName string        `mapstructure:"STREAM_QUEUE_NAME"`
	ConsumerName    string        `mapstructure:"CONSUMER_NAME"`
	RetryCount      int           `mapstructure:"RETRY_COUNT"`
	RetryDelay      time.Duration `mapstructure:"RETRY_DELAY"`
	CommitSize      int           `mapstructure:"COMMIT_SIZE"`
	CommitInterval  time.Duration `mapstructure:"COMMIT_INTERVAL"`
	ReconnectDelay  time.Duration `mapstructure:"RECONNECT_DELAY"`
}

type HTTPClientsConfig struct {
	MerchantCallback HTTPClientConfig `mapstructure:"MERCHANT_CALLBACK"`
}

type HTTPClientConfig struct {
	DialTimeout         time.Duration `mapstructure:"DIAL_TIMEOUT"`
	RequestTimeout      time.Duration `mapstructure:"REQUEST_TIMEOUT"`
	MaxIdleConns        int           `mapstructure:"MAX_IDLE_CONNS"`
	MaxIdleConnsPerHost int           `mapstructure:"MAX_IDLE_CONNS_PER_HOST"`
	IdleConnTimeout     time.Duration `mapstructure:"IDLE_CONN_TIMEOUT"`
	ProbeInterval       time.Duration `mapstructure:"PROBE_INTERVAL"`
	MaxProbes           int           `mapstructure:"MAX_PROBES"`
}

type SnapCoreSecret struct {
	InternalServiceKey string `mapstructure:"INTERNAL_SERVICE_KEY"`
}

type InternalApiKeySecret struct {
	Salt                 string `mapstructure:"SALT"`
	HashResult           string `mapstructure:"HASH_RESULT"`
	AutomatedTest        string `mapstructure:"AUTOMATED_TEST"`
	Simulation           string `mapstructure:"SIMULATION"`
	CardEncryptionSecret string `mapstructure:"CARD_ENCRYPTION_SECRET"`
}

type JWTSignatureKey struct {
	UserKey            string          `mapstructure:"USER"`
	MerchantKey        string          `mapstructure:"MERCHANT"`
	MerchantRSAKey     *RSAKeyPairFile `mapstructure:"MERCHANT_RSA_KEY"`
	TokenOTPKey        string          `mapstructure:"TOKEN_OTP_KEY"`
	TokenOTPFeatureKey string          `mapstructure:"TOKEN_OTP_FEATURE_KEY"`
	PaymentToken       string          `mapstructure:"PAYMENT_TOKEN"`
}

type RSAKeyPairFile struct {
	PrivateKeyFile string `mapstructure:"PRIVATE_KEY_FILE"`
	PublicKeyFile  string `mapstructure:"PUBLIC_KEY_FILE"`
}

type GCSConfig struct {
	ServiceBucketName          string `mapstructure:"SERVICE_BUCKET_NAME"`
	ReportingBucketName        string `mapstructure:"REPORTING_BUCKET_NAME"`
	BulkDisbursementBucketName string `mapstructure:"BULK_DISBURSEMENT_BUCKET_NAME"`
	ProofOfTransferFolderName  string `mapstructure:"PROOF_OF_TRANSFER_FOLDER_NAME"`
	MerchantDocumentFolderName string `mapstructure:"MERCHANT_DOCUMENT_FOLDER_NAME"`
	MerchantReservedSortName   string `mapstructure:"MERCHANT_RESERVED_SHORT_NAME"`
	MerchantLogoBucketName     string `mapstructure:"MERCHANT_LOGO_BUCKET_NAME"`
	InvestigationPoPFolderName string `mapstructure:"INVESTIGATION_POP_FOLDER_NAME"`
	TNCDocumentFolderName      string `mapstructure:"TNC_DOCUMENT_FOLDER_NAME"`
}

type InvestigationConfig struct {
	MaxFileSizeMB   int64    `mapstructure:"MAX_FILE_SIZE_MB"`
	AllowedFileExts []string `mapstructure:"ALLOWED_FILE_EXTS"`
}

type VccSlackRecipientConfig struct {
	DefaultRecipient []string `mapstructure:"DEFAULT_RECIPIENT"`
}

type BigQueryConfig struct {
	ProjectID                  string `mapstructure:"PROJECT_ID"`
	Location                   string `mapstructure:"LOCATION"`
	QueryTimeoutSeconds        int    `mapstructure:"QUERY_TIMEOUT_SECONDS"`
	MaxRetries                 int    `mapstructure:"MAX_RETRIES"`
	PayoutSuccessMetricsTable  string `mapstructure:"PAYOUT_SUCCESS_METRICS_TABLE"`
	PaymentSuccessMetricsTable string `mapstructure:"PAYMENT_SUCCESS_METRICS_TABLE"`
}

type PermissionConfig struct {
	Path string `mapstructure:"PATH"`
}

type DictionaryConfig struct {
	Path string `mapstructure:"PATH"`
}

type UserOTPConfig struct {
	MaxSendResetPwd  int `mapstructure:"MAX_SEND_RESET_PWD"`
	MaxSendResetPIN  int `mapstructure:"MAX_SEND_RESET_PIN"`
	MaxSendChangePwd int `mapstructure:"MAX_SEND_CHANGE_PWD"`
	MaxSendUserLogin int `mapstructure:"MAX_SEND_USER_LOGIN"`

	MaxFailedVerifyResetPwd  int `mapstructure:"MAX_FAILED_VERIFY_RESET_PWD"`
	MaxFailedVerifyResetPIN  int `mapstructure:"MAX_FAILED_VERIFY_RESET_PIN"`
	MaxFailedVerifyChangePwd int `mapstructure:"MAX_FAILED_VERIFY_CHANGE_PWD"`
	MaxFailedVerifyUserLogin int `mapstructure:"MAX_FAILED_VERIFY_USER_LOGIN"`

	ExpirationSecondsForgotPassword int `mapstructure:"EXPIRATION_SECONDS_FORGOT_PASSWORD"`
	ExpirationSecondsResetPIN       int `mapstructure:"EXPIRATION_SECONDS_RESET_PIN"`
	ExpirationSecondsChangePassword int `mapstructure:"EXPIRATION_SECONDS_CHANGE_PASSWORD"`
	ExpirationSecondsUserLogin      int `mapstructure:"EXPIRATION_SECONDS_USER_LOGIN"`
	ExpirationSecondsFirstTimeLogin int `mapstructure:"EXPIRATION_SECONDS_FIRST_TIME_LOGIN"`

	ResendDelaySecondsForgotPassword int `mapstructure:"RESEND_DELAY_SECONDS_FORGOT_PASSWORD"`
	ResendDelaySecondsResetPIN       int `mapstructure:"RESEND_DELAY_SECONDS_RESET_PIN"`
	ResendDelaySecondsChangePassword int `mapstructure:"RESEND_DELAY_SECONDS_CHANGE_PASSWORD"`
	ResendDelaySecondsUserLogin      int `mapstructure:"RESEND_DELAY_SECONDS_USER_LOGIN"`
	ResendDelaySecondsFirstTimeLogin int `mapstructure:"RESEND_DELAY_SECONDS_FIRST_TIME_LOGIN"`
	ResendDelaySecondsDefault        int `mapstructure:"RESEND_DELAY_SECONDS_DEFAULT"`

	UserLoginWaitAfterSend    int `mapstructure:"USER_LOGIN_WAIT_AFTER_SEND"`
	UserLoginWaitTimeMinute   int `mapstructure:"USER_LOGIN_WAIT_TIME_MINUTE"`
	UserLoginRememberInMinute int `mapstructure:"USER_LOGIN_REMEMBER_IN_MINUTE"`

	FirstTimeLoginMaxSend         int `mapstructure:"FIRST_TIME_LOGIN_MAX_SEND"`
	FirstTimeLoginMaxFailedVerify int `mapstructure:"FIRST_TIME_LOGIN_MAX_FAILED_VERIFY"`
	FirstTimeLoginWaitAfterSend   int `mapstructure:"FIRST_TIME_LOGIN_WAIT_AFTER_SEND"`
	FirstTimeLoginWaitTimeMinute  int `mapstructure:"FIRST_TIME_LOGIN_WAIT_TIME_MINUTE"`
}

type PaperCommunication struct {
	BaseURL      string `mapstructure:"BASE_URL"`
	PlatformName string `mapstructure:"PLATFORM_NAME"`
	SenderOrigin string `mapstructure:"SENDER_ORIGIN"`
	EmailSender  string `mapstructure:"EMAIL_SENDER"`
	EmailLogoURL string `mapstructure:"EMAIL_LOGO_URL"`
}

type MerchantPortalConfig struct {
	UserInvitationURL           string `mapstructure:"USER_INVITATION_URL"`
	LoginURL                    string `mapstructure:"LOGIN_URL"`
	LogoURL                     string `mapstructure:"LOGO_URL"`
	DashboardGuideURL           string `mapstructure:"DASHBOARD_GUIDE_URL"`
	PaymentSimulationPatternURL string `mapstructure:"PAYMENT_SIMULATION_PATTERN_URL"`
	PaymentReceiptBackgroundURL string `mapstructure:"PAYMENT_RECEIPT_BACKGROUND_URL"`
	ReceiptSenderName           string `mapstructure:"RECEIPT_SENDER_NAME"`
	CardFundedPayoutURL         string `mapstructure:"CARD_FUNDED_PAYOUT_URL"`
}

type FeatureFlagConfig struct {
	LogLevel                  string `mapstructure:"LOG_LEVEL"`
	PollingInterval           int    `mapstructure:"POLLING_INTERVAL"`
	FileFormat                string `mapstructure:"FILE_FORMAT"`
	StartWithRetrieverError   bool   `mapstructure:"START_WITH_RETRIEVER_ERROR"`
	Offline                   bool   `mapstructure:"OFFLINE"`
	ConsulAddr                string `mapstructure:"CONSUL_ADDR"`
	ConsulConfigPath          string `mapstructure:"CONSUL_CONFIG_PATH"`
	ConsulPaymentImages       string `mapstructure:"CONSUL_PAYMENT_IMAGES"`
	ConsulPaymentInstructions string `mapstructure:"CONSUL_PAYMENT_INSTRUCTIONS"`
	ConsulAppVersion          string `mapstructure:"CONSUL_APP_VERSION"`
	SlackWebhookURL           string `mapstructure:"SLACK_WEBHOOK_URL"`
	ExporterSlackWebhookURL   string `mapstructure:"EXPORTER_SLACK_WEBHOOK_URL"`
}

type WorkerPoolConfig struct {
	Disbursement int `mapstructure:"DISBURSEMENT"`
}

type MerchantConfig struct {
	DormantValidation            bool                   `mapstructure:"DORMANT_VALIDATION"`
	CacheStatusDurationInMinutes int                    `mapstructure:"CACHE_STATUS_DURATION_IN_MINUTES"`
	DormantCondition             DormantConditionConfig `mapstructure:"DORMANT_CONDITION"`
}

type DormantConditionConfig struct {
	NoActivitiesInMonths int `mapstructure:"NO_ACTIVITIES_IN_MONTHS"`
	NoTransactionInDays  int `mapstructure:"NO_TRANSACTION_IN_DAYS"`
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

type WithdrawalConfig struct {
	MinAmount                                  float64                 `mapstructure:"MIN_AMOUNT"`
	MaxAmount                                  float64                 `mapstructure:"MAX_AMOUNT"`
	LimitOverbooking                           float64                 `mapstructure:"LIMIT_OVERBOOKING"`
	LimitNonOverbooking                        float64                 `mapstructure:"LIMIT_NON_OVERBOOKING"`
	ActiveConsumer                             int                     `mapstructure:"ACTIVE_CONSUMER"`
	AutoWithdrawalWorker                       int                     `mapstructure:"AUTO_WITHDRAWAL_WORKER"`
	AutoWithdrawalDefaultState                 string                  `mapstructure:"AUTO_WITHDRAWAL_DEFAULT_STATE"`
	AutoWithdrawalDefaultSchedulingTime        string                  `mapstructure:"AUTO_WITHDRAWAL_DEFAULT_SCHEDULING_TIME"`
	AutoWithdrawalBankCodeToBeExecAfterTrigger []string                `mapstructure:"AUTO_WITHDRAWAL_BANK_CODE_TO_BE_EXEC_AFTER_TRIGGER"`
	PaymentBalanceConfig                       BalanceWithdrawalConfig `mapstructure:"PAYMENT_BALANCE_CONFIG"`
	DisbursementBalanceConfig                  BalanceWithdrawalConfig `mapstructure:"DISBURSEMENT_BALANCE_CONFIG"`
	VirtualTerminalBalanceConfig               BalanceWithdrawalConfig `mapstructure:"VIRTUAL_TERMINAL_BALANCE_CONFIG"`
}

type BalanceWithdrawalConfig struct {
	WithdrawalAfterInactivityDays   int `mapstructure:"WITHDRAWAL_AFTER_INACTIVITY_DAYS"`
	NotificationAfterInactivityDays int `mapstructure:"NOTIFICATION_AFTER_INACTIVITY_DAYS"`
}

type OTLPConfig struct {
	Host            string           `mapstructure:"HOST"`
	Insecure        bool             `mapstructure:"INSECURE"`
	TLSClientConfig *TLSClientConfig `mapstructure:"TLS_CLIENT_CONFIG"`
	MetricConfig    OTLPMetricConfig `mapstructure:"METRIC"`
}

type OTLPMetricConfig struct {
	Interval      int `mapstructure:"INTERVAL"`
	ExportTimeout int `mapstructure:"EXPORT_TIMEOUT"`
}

type RateLimitConfig struct {
	Limit int64  `mapstructure:"LIMIT"`
	Time  string `mapstructure:"TIME_OPTION"`
}

type WalletBackendConfig struct {
	Host              string `mapstructure:"HOST"`
	InternalPrefixUrl string `mapstructure:"INTERNAL_PREFIX_URL"`
}

type TLSClientConfig struct {
	InsecureSkipVerify bool `mapstructure:"INSECURE_SKIP_VERIFY"`
}

type PaymentSimulationConfig struct {
	Cards PaymentSimulationCardConfig `mapstructure:"CARDS"`
}

type PaymentSimulationCardConfig struct {
	BlockedByFds       []string `mapstructure:"BLOCKED_BY_FDS"`
	RequireReviewByFds []string `mapstructure:"REQUIRE_REVIEW_BY_FDS"`
}

type GCPConfig struct {
	ProjectId     string                 `mapstructure:"PROJECT_ID"`
	SecretManager GCPSecretManagerConfig `mapstructure:"SECRET_MANAGER"`
}

type GCPSecretManagerConfig struct {
	EncryptionSecretName string `mapstructure:"ENCRYPTION_SECRET_NAME"`
}

type PaymentFeeDefaultConfig struct {
	EWallet map[string]PaymentFeeDefault `mapstructure:"EWALLET"`
	Other   PaymentFeeDefault            `mapstructure:"OTHER"`
}

type ReconConfig struct {
	ScanningToleranceInDays int `mapstructure:"SCANNING_TOLERANCE_IN_DAYS"`
}

type VaultConfig struct {
	Address string             `mapstructure:"ADDRESS"`
	Secrets VaultSecretsAsList `mapstructure:"SECRETS"`
}

type VaultSecretsAsList struct {
	UserEncryptionKey    VaultSecretConfig  `mapstructure:"USER_ENCRYPTION_KEY"`
	MerchantCredentials  VaultTransitConfig `mapstructure:"MERCHANT_CREDENTIALS"`
	PaymentEncryptionKey VaultSecretConfig  `mapstructure:"PAYMENT_ENCRYPTION_KEY"`
}

type VaultSecretConfig struct {
	MountPath  string `mapstructure:"MOUNT_PATH"`
	SecretPath string `mapstructure:"SECRET_PATH"`
	KeyName    string `mapstructure:"KEY_NAME"`
}

type VaultTransitConfig struct {
	SecretPath string `mapstructure:"SECRET_PATH"`
	SecretKey  string `mapstructure:"SECRET_KEY"`
}

type MultiFactorAuthConfig struct {
	TimeBasedOTP TimeBasedOTPConfig `mapstructure:"TIME_BASED_OTP"`
}

type TimeBasedOTPConfig struct {
	TOTPIssuer          string              `mapstructure:"TOTP_ISSUER"`
	TOTPSecretSize      uint                `mapstructure:"TOTP_SECRET_SIZE"`
	TOTPPeriodInSeconds uint                `mapstructure:"TOTP_PERIOD_IN_SECONDS"`
	TOTPRateLimit       TOTPRateLimitConfig `mapstructure:"TOTP_RATE_LIMIT"`
}

type TOTPRateLimitConfig struct {
	RequestLimit  int           `mapstructure:"REQUEST_LIMIT"`
	RequestWindow time.Duration `mapstructure:"REQUEST_WINDOW"`
}

type PaymentFeeDefault struct {
	Type       string  `mapstructure:"TYPE"`
	Amount     float64 `mapstructure:"AMOUNT"`
	Percentage float64 `mapstructure:"PERCENTAGE"`
}

type DukcapilConfig struct {
	URL             string                  `mapstructure:"URL"`
	FieldThresholds DukcapilFieldThresholds `mapstructure:"FIELD_THRESHOLDS"`
}

type ReorganizeTablePartitionConfig struct {
	AccountTransactions RangeTablePartitionConfig `mapstructure:"ACCOUNT_TRANSACTIONS"`
}

type RangeTablePartitionConfig struct {
	Partitions   []RangePartitionConfig `mapstructure:"RANGE_PARTITIONS"`
	AnalyzeTable bool                   `mapstructure:"ANALYZE_TABLE"`
}

type RangePartitionConfig struct {
	PartitionName     string `mapstructure:"PARTITION_NAME"`
	DataOlderThanDays string `mapstructure:"DATA_OLDER_THAN_DAYS"`
}

type ConductorConfig struct {
	Address string `mapstructure:"ADDRESS"`
}

type MerchantCallbackTaskConfig struct {
	Preparation         TaskRunnerConfig `mapstructure:"PREPARATION"`
	SendCallback        TaskRunnerConfig `mapstructure:"SEND_CALLBACK"`
	WriteCallbackLog    TaskRunnerConfig `mapstructure:"WRITE_CALLBACK_LOG"`
	WriteCallbackMetric TaskRunnerConfig `mapstructure:"WRITE_CALLBACK_METRIC"`
}

type ShortLinkRedirectionConfig struct {
	InvalidURL string `mapstructure:"INVALID_URL"`
	URLFormat  string `mapstructure:"URL_FORMAT"`
}

type TaskRunnerConfig struct {
	BatchSize       int           `mapstructure:"BATCH_SIZE"`
	PollingInterval time.Duration `mapstructure:"POLLING_INTERVAL_DURATION"`
	PollingTimeout  time.Duration `mapstructure:"POLLING_TIMEOUT_DURATION"`
}

type VccTerminalConfig struct {
	DefaultConfig      VCCTerminalDefaultConfig `mapstructure:"DEFAULT_CONFIG"`
	TravelAgents       MStrStr                  `mapstructure:"TRAVEL_AGENTS"`
	ExpiryAfterMinutes int                      `mapstructure:"EXPIRY_AFTER_MINUTES"`
	ConsumerCount      int                      `mapstructure:"CONSUMER_COUNT"`
	WorkerCount        int                      `mapstructure:"WORKER_COUNT"`
}

type VCCTerminalDefaultConfig struct {
	AcquirerMerchantID string    `mapstructure:"ACQUIRER_MERCHANT_ID"`
	AllowedBinNumbers  *[]string `mapstructure:"ALLOWED_BIN_NUMBERS"`
	CardTypes          []string  `mapstructure:"CARD_TYPES"`
	PrincipalAvailable []string  `mapstructure:"PRINCIPAL_AVAILABLE"`
}

type AutoSplitPaymentConfig struct {
	ProcessorLimitDefault float64 `mapstructure:"PROCESSOR_LIMIT_DEFAULT"`
}

type DukcapilFieldThresholds struct {
	Name     int `mapstructure:"NAME"`
	Gender   int `mapstructure:"GENDER"`
	DOB      int `mapstructure:"DOB"`
	POB      int `mapstructure:"POB"`
	Job      int `mapstructure:"JOB"`
	Address  int `mapstructure:"ADDRESS"`
	RT       int `mapstructure:"RT"`
	RW       int `mapstructure:"RW"`
	Village  int `mapstructure:"VILLAGE"`
	District int `mapstructure:"DISTRICT"`
	Regency  int `mapstructure:"REGENCY"`
	Province int `mapstructure:"PROVINCE"`
}

type DukcapilSecret struct {
	UserID   string `mapstructure:"USER_ID"`
	Password string `mapstructure:"PASSWORD"`
	IP       string `mapstructure:"IP"`
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

var defEmailSender string

func DefaultEmailSender() string {
	return defEmailSender
}

var otpConfig UserOTPConfig

func OTPConfig() UserOTPConfig {
	return otpConfig
}

var env string

func Environment() string {
	return env
}

var emailLogoURL string

func GetEmailLogoURL() string {
	return emailLogoURL
}

var gcpConfig GCPConfig

func GetGCPConfig() GCPConfig {
	return gcpConfig
}

var paymentFeeDefault PaymentFeeDefaultConfig

func GetPaymentEWalletFeeDefault(channel string) PaymentFeeDefault {
	if paymentFeeDefault.EWallet == nil {
		return paymentFeeDefault.Other
	}
	feeDefault, ok := paymentFeeDefault.EWallet[strings.ToLower(channel)]
	if !ok {
		return paymentFeeDefault.Other
	}
	return feeDefault
}
