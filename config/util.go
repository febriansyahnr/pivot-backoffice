package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadConfig(configPath, secretPath string) (*Config, *Secret, error) {
	// Load Config
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("error reading config file: %w", err)
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	viper.Reset()

	// Load Secret
	viper.SetConfigFile(secretPath)
	var secret Secret
	if err := viper.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("error reading secret file: %w", err)
	}
	if err := viper.Unmarshal(&secret); err != nil {
		return nil, nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	env = config.Environment
	otpConfig = config.UserOTPConfig
	defEmailSender = config.PaperCommunication.EmailSender
	emailLogoURL = config.PaperCommunication.EmailLogoURL
	creditCardReferences = config.CreditCardReferences
	gcpConfig = config.GCPConfig
	paymentFeeDefault = config.PaymentFeeDefaults
	installmentDefaultFeeConfig = config.InstallmentFee.Default
	installmentDefaultChannelFeeConfig = buildInstallmentDefaultChannelFeeConfig(&config)
	config.VccTerminal.TravelAgents.KeyToUpperCase()

	return &config, &secret, nil
}

func buildInstallmentDefaultChannelFeeConfig(config *Config) map[string]InstallmentDefaultFeeConfig {
	installmentTenorFeeMap := map[string]InstallmentDefaultFeeConfig{}
	defaultFee := config.InstallmentFee.Default
	for channel, feeDetail := range config.InstallmentFee.Channel {
		for i, tenor := range feeDetail.Tenor {
			channelFeeKey := fmt.Sprintf("%s_%dM", channel, tenor)
			channelFeeConfig := InstallmentDefaultFeeConfig{}
			if i < len(feeDetail.Amount) {
				channelFeeConfig.Amount = feeDetail.Amount[i]
			} else {
				channelFeeConfig.Amount = defaultFee.Amount
			}

			if i < len(feeDetail.Percentage) {
				channelFeeConfig.Percentage = feeDetail.Percentage[i]
			} else {
				channelFeeConfig.Percentage = defaultFee.Percentage
			}

			installmentTenorFeeMap[channelFeeKey] = channelFeeConfig
		}
	}
	return installmentTenorFeeMap
}
