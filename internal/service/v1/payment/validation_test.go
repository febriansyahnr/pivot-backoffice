package paymentService

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/stretchr/testify/assert"
)

func TestPaymentService_ValidatePaymentExpiry(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	merchantID := "test-merchant-id"

	// Helper function to create time pointer
	timePtr := func(t time.Time) *time.Time {
		return &t
	}

	testCases := []struct {
		name          string
		setupService  func() *PaymentService
		cmd           paymentModel.PaymentRequestExpiryValidation
		expectedError bool
		errorContains string
	}{
		{
			name: "success - expiry request is zero, should return nil immediately",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: nil, // nil expiry date
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{},
			},
			expectedError: false,
		},
		{
			name: "success - virtual account with valid expiry within default config",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(12 * time.Hour)), // 12 hours, within 24 hours limit
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			expectedError: false,
		},
		{
			name: "success - qris with valid expiry within default config",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							QrConfig: &config.UnifiedPaymentQrConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_QRIS,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					Qris: &paymentModel.PaymentMetadataQris{
						ValidityPeriod: 900, // 15 minutes in seconds (15 * 60), within 30 minutes limit
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			expectedError: false,
		},
		{
			name: "success - validation skipped when ShouldValidateExpiry returns false (mode not set)",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: "", // Empty mode, should skip validation
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(48 * time.Hour)), // exceeds limit but validation should be skipped
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			expectedError: false,
		},
		{
			name: "error - partial mode with non-excluded merchant exceeding limit",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode:              paymentConstant.UnifiedPaymentExpiryModePartial,
								ExcludedMerchants: []string{"other-merchant-id"}, // Different merchant, so this one WILL be validated
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(48 * time.Hour)), // exceeds limit and validation should occur
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			expectedError: true,
			errorContains: "max expiry time is",
		},
		{
			name: "success - database config overrides default config",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(36 * time.Hour)), // 36 hours, exceeds default 24 hours but within db config
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 48,
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "success - database config with days unit",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     1,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitDays,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(36 * time.Hour)), // 1.5 days, within 3 days from db
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 3,
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitDays,
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "error - expiry request exceeds default config for virtual account",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(48 * time.Hour)), // 48 hours, exceeds 24 hours limit
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			expectedError: true,
			errorContains: "max expiry time is",
		},
		{
			name: "error - expiry request exceeds default config for qris",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							QrConfig: &config.UnifiedPaymentQrConfig{
								MaxExpiryDuration:     30,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitMinutes,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_QRIS,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					Qris: &paymentModel.PaymentMetadataQris{
						ValidityPeriod: 3600, // 60 minutes in seconds (60 * 60), exceeds 30 minutes limit
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			expectedError: true,
			errorContains: "max expiry time is",
		},
		{
			name: "error - expiry request exceeds database config",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(25 * time.Hour)), // 25 hours, exceeds db config of 12 hours
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 12,
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
						},
					},
				},
			},
			expectedError: true,
			errorContains: "max expiry time is",
		},
		{
			name: "success - partial mode with excluded merchant skips validation",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode:              paymentConstant.UnifiedPaymentExpiryModePartial,
								ExcludedMerchants: []string{merchantID}, // Excluded merchant, validation should be skipped
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(48 * time.Hour)), // 48 hours, exceeds 24 hours limit but validation is skipped
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			expectedError: false,
		},
		{
			name: "success - database config with empty unit should not be used",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(12 * time.Hour)), // Within default 24 hours
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 48,
								Unit:     "", // Empty unit, should use default config
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "success - database config with zero duration should not be used",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(12 * time.Hour)), // Within default 24 hours
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: &paymentModel.PaymentMethodConfig{
							ExpiryConfig: paymentModel.PaymentMethodExpiryConfig{
								Duration: 0, // Zero duration, should use default config
								Unit:     paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "success - nil config object should use default config",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
								MaxExpiryDuration:     24,
								MaxExpiryDurationUnit: paymentConstant.UnifiedPaymentExpiryUnitHours,
							},
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(12 * time.Hour)), // Within default 24 hours
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{
						ConfigObj: nil, // Nil config object, should use default config
					},
				},
			},
			expectedError: false,
		},
		{
			name: "success - different payment method (not VA or QRIS), empty validation config",
			setupService: func() *PaymentService {
				return &PaymentService{
					config: &config.Config{
						UnifiedPaymentConfig: config.UnifiedPaymentConfig{
							ExpiryConfig: &config.UnifiedPaymentExpiryConfig{
								Mode: paymentConstant.UnifiedPaymentExpiryModeFull,
							},
						},
					},
				}
			},
			cmd: paymentModel.PaymentRequestExpiryValidation{
				Method:     paymentConstant.PAYMENT_METHOD_CREDIT_CARD, // Different payment method
				MerchantID: merchantID,
				Request: &paymentModel.PaymentRequest{
					VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
						ExpiredDate: timePtr(now.Add(1000 * time.Hour)), // Any future date should pass with empty config
					},
				},
				PaymentMethod: &paymentModel.PaymentMethodWithPivot{
					PaymentMethod: paymentModel.PaymentMethod{},
				},
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			svc := tc.setupService()

			// Execute
			err := svc.ValidatePaymentExpiry(ctx, tc.cmd)

			// Assert
			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
