package config

type FDSFeaturesConfig struct {
	ProofOfPayment FDSFeatureProofOfPaymentConfig `mapstructure:"PROOF_OF_PAYMENT"`
}

type FDSFeatureProofOfPaymentConfig struct {
	Velocity FDSRuleVelocityConfig `mapstructure:"VELOCITY"`
}

type FDSRuleVelocityConfig struct {
	Enabled   bool               `mapstructure:"ENABLED"`
	Window    FDSWindowConfig    `mapstructure:"WINDOW"`
	Threshold FDSThresholdConfig `mapstructure:"THRESHOLD"`
	Action    string             `mapstructure:"ACTION"`
}

type FDSWindowConfig struct {
	Interval int    `mapstructure:"INTERVAL"`
	Unit     string `mapstructure:"UNIT"`
}

type FDSThresholdConfig struct {
	Count int `mapstructure:"COUNT"`
}

type SokratechConfig struct {
	BaseURL        string                  `mapstructure:"BASE_URL"`
	Workflow       SokratechWorkflowConfig `mapstructure:"WORKFLOW"`
	TimeoutSeconds int                     `mapstructure:"TIMEOUT_SECONDS"`
}

type SokratechWorkflowConfig struct {
	PayoutTransactionID  string `mapstructure:"PAYOUT_TRANSACTION_ID"`
	PaymentTransactionID string `mapstructure:"PAYMENT_TRANSACTION_ID"`
}

type SokratechSecret struct {
	AccessSecret string `mapstructure:"ACCESS_SECRET"`
}

type FraudNetConfig struct {
	BaseURL string `mapstructure:"BASE_URL"`
}

type FraudNetSecret struct {
	AccessKey    string `mapstructure:"ACCESS_KEY"`
	AccessSecret string `mapstructure:"ACCESS_SECRET"`
}

type FdsConfig struct {
	ScoreThreshold int64             `mapstructure:"SCORE_THRESHOLD"`
	Timeout        int64             `mapstructure:"TIMEOUT"`
	BinLength      int64             `mapstructure:"BIN_LENGTH"`
	Features       FDSFeaturesConfig `mapstructure:"FEATURES"`
}

type AdvanceAIConfig struct {
	BaseURL   string `mapstructure:"BASE_URL"`
	JourneyID string `mapstructure:"JOURNEY_ID"`
}

type AdvanceAISecret struct {
	ApiKey string `mapstructure:"API_KEY"`
}
