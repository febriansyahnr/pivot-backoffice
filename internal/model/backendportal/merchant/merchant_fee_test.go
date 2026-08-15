package merchant

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerchantFeeToResponse(t *testing.T) {
	now := time.Now()
	maxFeeAmount, deductionDay := 20_000.0, int16(20)

	merchantFee := &MerchantFee{
		UUID:          "uuid-uuid-uuid",
		MerchantID:    "merchant-id",
		AmountType:    constant.MerchantFeeAmountPercentageType,
		Percentage:    0.9,
		MaxFeeAmount:  &maxFeeAmount,
		Reference:     constant.ReferencePlatformTransaction,
		DeductionType: constant.MerchantFeeDeductionTypeAutomated,
		DeductionDay:  &deductionDay,
		CreatedAt:     now, UpdatedAt: now,
	}

	response := &MerchantFeeResponse{
		UUID:          "uuid-uuid-uuid",
		MerchantID:    "merchant-id",
		Percentage:    0.9,
		AmountType:    constant.MerchantFeeAmountPercentageType,
		MaxFeeAmount:  &maxFeeAmount,
		Reference:     constant.ReferencePlatformTransaction,
		DeductionType: constant.MerchantFeeDeductionTypeAutomated,
		DeductionDay:  &deductionDay,
		CreatedAt:     now, UpdatedAt: now,
	}

	testCases := []struct {
		Name     string
		Input    *MerchantFee
		Expected *MerchantFeeResponse
	}{
		{
			Name:     "it should return merchant fee response",
			Input:    merchantFee,
			Expected: response,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newMerchantResponse := tc.Input.ToResponse()
			require.Equal(t, tc.Expected, newMerchantResponse)
		})
	}
}

func TestNewMerchantFee(t *testing.T) {
	contentConfig := `
CREDIT_CARD_REFERENCES:
  CARD_BRANDS:
    - VISA

`
	f, err := os.CreateTemp(os.TempDir(), "*.yaml")
	require.NoError(t, err)
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	_, _ = f.WriteString(contentConfig)

	_, _, err = config.LoadConfig(f.Name(), f.Name())
	require.NoError(t, err)

	merchantId := uuid.New().String()
	timestamp := time.Now().UTC()
	qrisPaymentMethod := "QRIS"

	testCases := []struct {
		Name     string
		Input    *NewMerchantFeeRequest
		Expected *MerchantFee
		WantErr  bool
	}{
		{
			Name: "New merchant fee for installment",
			Input: &NewMerchantFeeRequest{
				MerchantID:       merchantId,
				AmountType:       constant.MerchantFeeAmountType,
				Amount:           400,
				Reference:        constant.ReferencePayment,
				PaymentMethod:    paymentConstant.PAYMENT_METHOD_INSTALLMENT,
				InstallmentTenor: 12,
				Channel:          "BCA",
			},
			Expected: &MerchantFee{
				MerchantID:    merchantId,
				AmountType:    constant.MerchantFeeAmountType,
				Amount:        400,
				Reference:     constant.ReferencePayment,
				PaymentMethod: util.ValueToPtr(paymentConstant.PAYMENT_METHOD_INSTALLMENT),
				Channel:       util.ValueToPtr("BCA_12M"),
			},
			WantErr: false,
		},
		{
			Name: "New merchant fee for platform transfer",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				AmountType: constant.MerchantFeeAmountType,
				Amount:     400,
				Reference:  constant.ReferencePlatformTransfer,
			},
			Expected: &MerchantFee{
				MerchantID: merchantId,
				AmountType: constant.MerchantFeeAmountType,
				Amount:     400,
				Reference:  constant.ReferencePlatformTransfer,
			},
			WantErr: false,
		},
		{
			Name: "New merchant fee for platform transaction",
			Input: &NewMerchantFeeRequest{
				MerchantID:    merchantId,
				AmountType:    constant.MerchantFeePercentageType,
				Percentage:    1.2,
				MaxFeeAmount:  25_000,
				Reference:     constant.ReferencePlatformTransaction,
				DeductionType: constant.MerchantFeeDeductionTypeAutomated,
				DeductionDay:  15,
			},
			Expected: &MerchantFee{
				MerchantID:    merchantId,
				AmountType:    constant.MerchantFeePercentageType,
				Percentage:    1.2,
				MaxFeeAmount:  util.ValueToPtr(25_000.00),
				Reference:     constant.ReferencePlatformTransaction,
				DeductionType: constant.MerchantFeeDeductionTypeAutomated,
				DeductionDay:  util.ValueToPtr(int16(15)),
			},
			WantErr: false,
		},
		{
			Name: "New merchant fee disbursement with amount fee type",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				Amount:     1000,
				AmountType: constant.MerchantFeeAmountType,
				Reference:  constant.ReferenceDisbursement,
			},
			Expected: &MerchantFee{
				MerchantID: merchantId,
				Amount:     1000,
				AmountType: constant.MerchantFeeAmountType,
				Reference:  constant.ReferenceDisbursement,
			},
			WantErr: false,
		},
		{
			Name: "New merchant fee account inquiry with amount fee type",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				Amount:     1000,
				AmountType: constant.MerchantFeeAmountType,
				Reference:  constant.ReferenceAccountInquiry,
			},
			Expected: &MerchantFee{
				MerchantID: merchantId,
				Amount:     1000,
				AmountType: constant.MerchantFeeAmountType,
				Reference:  constant.ReferenceAccountInquiry,
			},
			WantErr: false,
		},
		{
			Name: "New merchant fee payment with amount fee type",
			Input: &NewMerchantFeeRequest{
				MerchantID:    merchantId,
				Amount:        1000,
				AmountType:    constant.MerchantFeeAmountType,
				Reference:     constant.ReferencePayment,
				PaymentMethod: qrisPaymentMethod,
				Channel:       "BNC",
			},
			Expected: &MerchantFee{
				MerchantID:    merchantId,
				Amount:        1000,
				AmountType:    constant.MerchantFeeAmountType,
				Reference:     constant.ReferencePayment,
				PaymentMethod: &qrisPaymentMethod,
				Channel:       util.ValueToPtr("BNC"),
			},
			WantErr: false,
		},
		{
			Name: "New merchant fee payment with percentage fee type",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				Percentage: 10,
				AmountType: constant.MerchantFeePercentageType,
				Reference:  constant.ReferencePayment,
			},
			Expected: &MerchantFee{
				MerchantID: merchantId,
				Percentage: 10,
				AmountType: constant.MerchantFeePercentageType,
				Reference:  constant.ReferencePayment,
			},
			WantErr: false,
		},
		{
			Name: "New merchant fee payment with percentage fee type",
			Input: &NewMerchantFeeRequest{
				MerchantID:    merchantId,
				Reference:     constant.ReferencePayment,
				PaymentMethod: constant.ChannelCreditCard,
				Channel:       "LOCAL_VISA", // NOSONAR
				AmountType:    constant.MerchantFeePercentageType,
				Percentage:    0.75,
			},
			Expected: &MerchantFee{
				MerchantID:    merchantId,
				Reference:     constant.ReferencePayment,
				PaymentMethod: util.ValueToPtr(constant.ChannelCreditCard),
				Channel:       util.ValueToPtr("LOCAL_VISA"), // NOSONAR
				AmountType:    constant.MerchantFeePercentageType,
				Percentage:    0.75,
			},
		},
		{
			Name: "Error create new merchant fee",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				Percentage: 10,
				AmountType: constant.MerchantFeePercentageType,
				Reference:  "reference",
			},
			Expected: &MerchantFee{
				MerchantID: merchantId,
				Percentage: 10,
				AmountType: constant.MerchantFeePercentageType,
				Reference:  constant.ReferencePayment,
			},
			WantErr: true,
		},
		{
			Name: "Error non payment reference fill channel attribute",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				Reference:  constant.ReferencePlatformTransaction,
				Channel:    "PERMATA",
			},
			WantErr: true,
		},
		{
			Name: "Invalid credit card channel",
			Input: &NewMerchantFeeRequest{
				MerchantID:    merchantId,
				Reference:     constant.ReferencePayment,
				PaymentMethod: constant.ChannelCreditCard,
				Channel:       "PERMATA", // NOSONAR
			},
			WantErr: true,
		},
		{
			Name: "Error invalid disbursement channel",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				Reference:  constant.ReferenceDisbursement,
				Channel:    "XXXX", // NOSONAR
			},
			WantErr: true,
		},
		{
			Name: "Error invalid reference",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				Reference:  "XXXX", // NOSONAR
			},
			WantErr: true,
		},
		{
			Name: "New merchant fee for TopUp reference with channel",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				AmountType: constant.MerchantFeeAmountType,
				Amount:     500,
				Reference:  constant.ReferenceTopUp,
				Channel:    "OVO",
			},
			Expected: &MerchantFee{
				MerchantID: merchantId,
				AmountType: constant.MerchantFeeAmountType,
				Amount:     500,
				Reference:  constant.ReferenceTopUp,
				Channel:    util.ValueToPtr("OVO"),
			},
			WantErr: false,
		},
		{
			Name: "New merchant fee for TopUp reference without channel",
			Input: &NewMerchantFeeRequest{
				MerchantID: merchantId,
				AmountType: constant.MerchantFeePercentageType,
				Percentage: 1.5,
				Reference:  constant.ReferenceTopUp,
			},
			Expected: &MerchantFee{
				MerchantID: merchantId,
				AmountType: constant.MerchantFeePercentageType,
				Percentage: 1.5,
				Reference:  constant.ReferenceTopUp,
			},
			WantErr: false,
		},
	}

	for _, test := range testCases {
		output, err := NewMerchantFee(test.Input)

		if test.WantErr {
			assert.NotNil(t, err)
			continue
		}
		assert.Nil(t, err)

		assert.Equal(t, test.Expected.MerchantID, output.MerchantID)
		assert.Equal(t, test.Expected.Amount, output.Amount)
		assert.Equal(t, test.Expected.AmountType, output.AmountType)
		assert.Equal(t, test.Expected.Reference, output.Reference)
		assert.Equal(t, test.Expected.MaxFeeAmount, output.MaxFeeAmount)
		assert.Equal(t, test.Expected.DeductionDay, output.DeductionDay)
		assert.Equal(t, test.Expected.DeductionType, output.DeductionType)
		assert.Equal(t, util.ValueOfPtr(test.Expected.Channel), util.ValueOfPtr(output.Channel))
		assert.NotEqual(t, uuid.Max, output.UUID)
		assert.NotEqual(t, timestamp, output.CreatedAt)
		assert.NotEqual(t, timestamp, output.DeletedAt)
		assert.NotEqual(t, timestamp, output.UpdatedAt)

	}
}

func TestValidate(t *testing.T) {

	testCases := []struct {
		Name     string
		Input    MerchantFee
		Expected error
	}{
		{
			Name: "Validate negative amount",
			Input: MerchantFee{
				Amount: -1000,
			},
			Expected: constant.ErrNegativeValue,
		},
		{
			Name: "Invalid max fee amount",
			Input: MerchantFee{
				Reference:    constant.ReferenceDisbursement,
				MaxFeeAmount: util.ValueToPtr(25_000.00),
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("not allowed to have maximum fee for requested reference & referenceType")), // NOSONAR
		},
		{
			Name: "Platform Transaction Max fee amount is required field",
			Input: MerchantFee{
				Reference: constant.ReferencePlatformTransaction,
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("maximum fee amount is required field")), // NOSONAR
		},
		{
			Name: "Platform Transaction Invalid amount type",
			Input: MerchantFee{
				Reference:    constant.ReferencePlatformTransaction,
				MaxFeeAmount: util.ValueToPtr(15_000.00),
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("platform transaction fee type must use PERCENTAGE type")), // NOSONAR
		},
		{
			Name: "Invalid deduction day",
			Input: MerchantFee{
				Reference:     constant.ReferenceDisbursement,
				DeductionType: constant.MerchantFeeDeductionTypeDirect,
				DeductionDay:  util.ValueToPtr(int16(20)),
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("deduction day only for automated deduction type or platform activity")), // NOSONAR
		},
		{
			Name: "Platform Activity/Invalid amount type",
			Input: MerchantFee{
				Reference:  constant.ReferencePlatformActivity,
				AmountType: constant.MerchantFeeAmountPercentageType,
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("platform activity fee type must use AMOUNT type")), // NOSONAR
		},
		{
			Name: "Platform Activity/Invalid deduction type",
			Input: MerchantFee{
				Reference:     constant.ReferencePlatformActivity,
				AmountType:    constant.MerchantFeeAmountType,
				DeductionType: constant.MerchantFeeDeductionTypeDirect,
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("for this reference must use indirect deduction")), // NOSONAR
		},
		{
			Name: "Validate invalid fee type AMOUNT",
			Input: MerchantFee{
				AmountType: constant.MerchantFeeAmountType,
				Percentage: .07,
				Amount:     1_000,
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("percentage must be zero if fee type is AMOUNT")), // NOSONAR
		},
		{
			Name: "Validate invalid fee type PERCENTAGE",
			Input: MerchantFee{
				AmountType: constant.MerchantFeePercentageType,
				Percentage: .07,
				Amount:     1_000,
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("amount must be zero if fee type is PERCENTAGE")), // NOSONAR
		},
		{
			Name: "Validate input",
			Input: MerchantFee{
				Amount:     1000,
				AmountType: constant.MerchantFeeAmountType,
				Reference:  constant.ReferenceDisbursement,
			},
		},
		{
			Name: "Non payment reference fill channel attribute",
			Input: MerchantFee{
				Reference: constant.ReferenceAccountInquiry,
				Channel:   util.ValueToPtr("JAGO"), // NOSONAR
			},
			Expected: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("channel attribute is only applicable for the PAYMENT and PAYOUT reference")),
		},
		{
			Name: "TopUp reference with channel is valid",
			Input: MerchantFee{
				Reference:  constant.ReferenceTopUp,
				Channel:    util.ValueToPtr("OVO"), // NOSONAR
				AmountType: constant.MerchantFeeAmountType,
				Amount:     500,
			},
			Expected: nil,
		},
		{
			Name: "TopUp reference without channel is valid",
			Input: MerchantFee{
				Reference:  constant.ReferenceTopUp,
				AmountType: constant.MerchantFeePercentageType,
				Percentage: 1.5,
			},
			Expected: nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			output := test.Input.validate()
			assert.Equal(t, test.Expected, output)
		})
	}

}

func TestUpdateMerchantFee(t *testing.T) {
	testUUID := uuid.NewString()
	merchantUUID := uuid.NewString()
	tests := []struct {
		name        string
		initialFee  MerchantFee
		updateReq   UpdateMerchantFeeRequest
		expected    MerchantFee
		expectedErr error
	}{
		{
			name: "Valid update",
			initialFee: MerchantFee{
				UUID:       testUUID,
				MerchantID: merchantUUID,
				Percentage: 0.6,
				AmountType: constant.MerchantFeePercentageType,
				Reference:  constant.ReferencePlatformTransaction,
				CreatedAt:  time.Now().UTC().Add(-1 * time.Hour),
			},
			updateReq: UpdateMerchantFeeRequest{
				ID:            testUUID,
				MerchantID:    merchantUUID,
				Percentage:    0.5,
				AmountType:    constant.MerchantFeePercentageType,
				DeductionType: constant.MerchantFeeDeductionTypeAutomated,
				DeductionDay:  30,
				MaxFeeAmount:  20_000,
			},
			expected: MerchantFee{
				UUID:          testUUID,
				MerchantID:    merchantUUID,
				Percentage:    0.5,
				AmountType:    constant.MerchantFeePercentageType,
				Reference:     constant.ReferencePlatformTransaction,
				DeductionType: constant.MerchantFeeDeductionTypeAutomated,
				DeductionDay:  util.ValueToPtr(int16(30)),
				MaxFeeAmount:  util.ValueToPtr(20_000.00),
			},
			expectedErr: nil,
		},
		{
			name: "Invalid amount",
			initialFee: MerchantFee{
				UUID:       testUUID,
				MerchantID: merchantUUID,
				Amount:     10,
				AmountType: constant.MerchantFeeAmountType,
				Reference:  constant.ReferenceDisbursement,
			},
			updateReq: UpdateMerchantFeeRequest{
				ID:         testUUID,
				MerchantID: merchantUUID,
				Amount:     -5,
				AmountType: constant.MerchantFeeAmountType,
			},
			expected:    MerchantFee{},
			expectedErr: constant.ErrNegativeValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.initialFee.UpdateMerchantFee(&tt.updateReq)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, result)

			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.UUID, result.UUID)
				assert.Equal(t, tt.expected.MerchantID, result.MerchantID)
				assert.Equal(t, tt.expected.Amount, result.Amount)
				assert.Equal(t, tt.expected.Percentage, result.Percentage)
				assert.Equal(t, tt.expected.MaxFeeAmount, result.MaxFeeAmount)
				assert.Equal(t, tt.expected.Reference, result.Reference)
				assert.Equal(t, tt.expected.DeductionType, result.DeductionType)
				assert.Equal(t, tt.expected.DeductionDay, result.DeductionDay)
				assert.WithinDuration(t, time.Now().UTC(), result.UpdatedAt, time.Second)
			}
		})
	}
}

func TestFeeTieringConfigValidate(t *testing.T) {
	tests := []struct {
		feeDetail  MerchantFee
		feeTiering *FeeTieringConfig
		wantErr    error
	}{
		{
			feeDetail: MerchantFee{
				Reference: constant.ReferenceDisbursement,
				Amount:    4_000,
			},
			feeTiering: &FeeTieringConfig{
				MaxFeeAmount: util.ValueToPtr(10_000.00),
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("not allowed to have maximum fee for requested reference & referenceType")),
		},
		{
			feeDetail: MerchantFee{
				Reference: constant.ReferenceAccountInquiry,
				Amount:    400,
			},
			feeTiering: &FeeTieringConfig{
				AmountType: "AMOUNT",
				Amount:     450,
				TaxType:    "NON_PKP",
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		originalFeeAmount := test.feeDetail.Amount
		assert.Equal(t, test.wantErr, test.feeTiering.Validate(test.feeDetail))
		assert.Equal(t, originalFeeAmount, test.feeDetail.Amount)
	}
}

func TestSettlementConfigCutOffWindowIsCutOffTime(t *testing.T) {

	loc, err := time.LoadLocation(constant.TimeLoc)
	require.NoError(t, err)

	date := time.Date(2026, 2, 10, 11, 0, 0, 0, loc)

	sameDayConfig := SettlementConfigCutOffWindow{
		StartTime: "19:00:00",
		EndTime:   "23:59:59",
	}

	crossDayConfig := SettlementConfigCutOffWindow{
		StartTime: "23:00:00",
		EndTime:   "06:00:00",
	}

	tests := []struct {
		date       time.Time
		window     SettlementConfigCutOffWindow
		wantError  error
		wantResult bool
	}{
		{
			date:      date,
			window:    SettlementConfigCutOffWindow{},
			wantError: errors.New("parsing start time: parsing time \"2026-02-10 \" as \"2006-01-02 15:04:05\": cannot parse \"\" as \"15\""),
		},
		{
			date: date,
			window: SettlementConfigCutOffWindow{
				StartTime: "22:00:00",
			},
			wantError: errors.New("parsing end time: parsing time \"2026-02-10 \" as \"2006-01-02 15:04:05\": cannot parse \"\" as \"15\""),
		},
		{
			date:       time.Date(2026, 2, 10, 11, 0, 0, 0, loc),
			window:     sameDayConfig,
			wantResult: false,
		},
		{
			date:       time.Date(2026, 2, 10, 18, 59, 59, 690, loc),
			window:     sameDayConfig,
			wantResult: false,
		},
		{
			date:       time.Date(2026, 2, 10, 19, 00, 00, 0, loc),
			window:     sameDayConfig,
			wantResult: true,
		},
		{
			date:       time.Date(2026, 2, 10, 22, 59, 00, 0, loc),
			window:     sameDayConfig,
			wantResult: true,
		},
		{
			date:       time.Date(2026, 2, 10, 23, 59, 59, 500, loc),
			window:     sameDayConfig,
			wantResult: true,
		},
		{
			date:       time.Date(2026, 2, 10, 0, 0, 0, 0, loc),
			window:     crossDayConfig,
			wantResult: true,
		},
		{
			date:       time.Date(2026, 2, 10, 5, 40, 49, 0, loc),
			window:     crossDayConfig,
			wantResult: true,
		},
		{
			date:       time.Date(2026, 2, 10, 6, 0, 0, 0, loc),
			window:     crossDayConfig,
			wantResult: true,
		},
		{
			date:       time.Date(2026, 2, 10, 6, 0, 1, 0, loc),
			window:     crossDayConfig,
			wantResult: false,
		},
		{
			date:       time.Date(2026, 2, 10, 18, 59, 59, 690, loc),
			window:     crossDayConfig,
			wantResult: false,
		},
		{
			date:       time.Date(2026, 2, 10, 19, 00, 00, 0, loc),
			window:     crossDayConfig,
			wantResult: false,
		},
		{
			date:       time.Date(2026, 2, 10, 22, 59, 59, 600, loc),
			window:     crossDayConfig,
			wantResult: false,
		},
		{
			date:       time.Date(2026, 2, 10, 23, 00, 00, 0, loc),
			window:     crossDayConfig,
			wantResult: true,
		},
	}
	for _, test := range tests {
		result, err := test.window.IsCutOffTime(test.date)
		assert.Equal(t, test.wantError, err)
		assert.Equal(t, test.wantResult, result)
	}
}
