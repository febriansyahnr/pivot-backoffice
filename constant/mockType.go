package constant

import (
	"fmt"

	"github.com/stretchr/testify/mock"
)

const (
	MockTypeBackgroundContext        = "context.backgroundCtx"
	MockTypeValueContextReference    = "*context.valueCtx"
	MockTypeCancelContextReference   = "*context.cancelCtx"
	MockTypeTime                     = "time.Time"
	MockTypeTimeReference            = "*time.Time"
	MockTypeInt64Reference           = "*int64"
	MockTypeIntReference             = "*int"
	MockTypeMapStringStringReference = "map[string]string"

	// Model
	MockTypeAccountInquiriesReference              = "*accountInquiries.AccountInquiries"
	MockTypeAccountTransactionWithUseCaseReference = "*orchestrator_model.AccountTransactionWithUseCase"
	MockTypeDisbursementWithTransactionReference   = "*disbursementModel.DisbursementWithTransaction"
	MockTypeManualAdjustmentHistoryReference       = "*adjustment.ManualAdjustmentHistory"
	MockTypeDisbursementRetryBulkRequest           = "disbursementModel.RetryBulkRequest"
	MockTypeDisbursementRetryBulkRequestReference  = "*disbursementModel.RetryBulkRequest"
	MockTypeDisbursementFilterRequestReference     = "*disbursementModel.GetDisbursementFilterRequest"
	MockTypeRetryDisbursementFromOpenAPIResponse   = "disbursementModel.RetryDisbursementFromOpenApiResponse"
	MockTypeBankAccountReference                   = "*bankAccount.BankAccount"
)

const (
	SkipTestingInShortMode = "Skipping testing in short mode"
	SkipIntegrationTest    = "Skipping integration test"
)

func TimeMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType(MockTypeTime)
}

func TimeReferenceMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType(MockTypeTimeReference)
}

func ValueCtxMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType(MockTypeValueContextReference)
}

func CancelCtxMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType(MockTypeCancelContextReference)
}

func BackgroundCtxMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType(MockTypeBackgroundContext)
}

func StringMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("string")
}
func Float64MockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("float64")
}

func PtrStringMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*string")
}

func ArrayStringMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("[]string")
}

func ZapFieldMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("zapcore.Field")
}

func LoggerFieldMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("logger.Field")
}

func MapStrValStringMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("map[string]string")
}

func MapStrValPtrMultipartFileHeaderMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("map[string]*multipart.FileHeader")
}

func PtrCallbackLogMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*callback_model.CallbackLog")
}

func PtrCallbackLogWithMasterMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*callback_model.CallbackLogWithMaster")
}

func PtrGetListCallbackLogFilterRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*callback_model.GetListCallbackLogFilterRequest")
}

func PtrAccountMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*account_model.Account")
}

func CustomerMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("customerModel.Customer")
}

func PtrMenuAndPermissionIDsMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*menuModel.MenuAndPermissionIDs")
}

func PtrMerchantFeeMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*merchant.MerchantFee")
}

func PtrMerchantStatusResponseMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*merchant.MerchantStatusResponse")
}

func PtrRequestAccountInquiriesMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*requestAccountInquiries.RequestAccountInquiries")
}

func UuidMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("uuid.UUID")
}

func DurationMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("time.Duration")
}

func BoolMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("bool")
}

func PtrBoolMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*bool")
}

func DecimalMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("decimal.Decimal")
}

func PtrCreateAccTransactionReqMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*orchestrator_model.CreateAccountTransactionRequest")
}

func FileHeaderMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*multipart.FileHeader")
}

func PtrRoleMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*role.Role")
}

func PtrUserMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*user.User")
}

func PtrUserLoggedInDeviceMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*userLoggedInDeviceModel.UserLoggedInDevice")
}

func WrapErrApiRespForTest(code int, errType, msg string) string {
	return fmt.Sprintf(`{"code":"%d","data":null,"error":{"details":[],"traceId":"","type":"%s"},"message":"%s"}`, code, errType, msg)
}

func WrapErrRespForTest(code int, msg string) string {
	return fmt.Sprintf(`{"code":"%d","errors":"%s"}`, code, msg)
}

func WrapErrOpenApiSnapForTest(code, msg string) string {
	return fmt.Sprintf(`{"responseCode":"%s","responseMessage":"%s"}`, code, msg)
}

func Int64MockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("int64")
}

func PtrFileHeaderMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*multipart.FileHeader")
}

func Uint16MockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("uint16")
}

func PtrGetPaymentMethodFilterRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paymentModel.GetPaymentMethodFilterRequest")
}

func PtrQrMpmPaymentSimulationRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*snapCoreModel.QrMpmPaymentSimulationRequest")
}

func PtrGetFeeRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*feeModel.GetFeeRequest")
}

func PtrFeeHistoryMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*feeModel.FeeHistory")
}

func PtrVANotificationRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paymentModel.VirtualAccountPaymentNotificationRequest")
}

func PtrMerchantTransactionConfigMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*merchant.TransactionConfigs")
}

func PtrMerchantSettlementConfigMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*merchant.SettlementConfig")
}

func PtrNullJSONText() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*types.NullJSONText")
}

func NullJSONText() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("types.NullJSONText")
}

func JSONTextMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("types.JSONText")
}

func PtrBytesBufferMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*bytes.Buffer")
}

func XlsxOptionsMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("excelize.Options")
}

func PtrProcessSettlementRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*settlementModel.ProcessSettlementRequest")
}

func PtrFeeMetadataObjectMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*feeModel.FeeMetadataObject")
}

func PtrCreateVirtualAccountConfigRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*snapCoreModel.CreateVirtualAccountConfigRequest")
}

func PtrUpdateVirtualAccountConfigPrefixRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*snapCoreModel.UpdateVirtualAccountConfigPrefixRequest")
}

func PtrAccountTransactionMetadataObjectMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*settlementModel.AccountTransactionMetadataObject")
}

func PtrPaymentMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paymentModel.Payment")
}

func PtrPaymentWithPaymentMethodDTOMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paymentModel.PaymentWithPaymentMethodDTO")
}

func PtrPaperCommEmailRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paperCommunication.Email")
}

func PtrDisbursementWithTransactionMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType(MockTypeDisbursementWithTransactionReference)
}

func SliceUint8MockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("[]uint8")
}

func SliceByteMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("[]byte")
}

func PtrRateLimitConfiguration() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*ratelimiter.RateLimitConfiguration")
}

func PtrPaymentMethodWithPivot() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paymentModel.PaymentMethodWithPivot")
}

func PtrSetupPaymentMethodConfigRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paymentMethodModel.SetupPaymentMethodConfigRequest")
}

func PtrGetListFilterRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paymentModel.GetListFilterRequest")
}

func PtrFilterChargeRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*unifiedPaymentModel.FilterChargeRequest")
}

func PtrGetUnifiedPaymentChargeRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*unifiedPaymentModel.GetUnifiedPaymentChargeRequest")
}

func PtrCreateUnifiedPaymentSessionRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*unifiedPaymentModel.CreateUnifiedPaymentSessionRequest")
}

func PtrConfirmUnifiedPaymentSessionRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest")
}

func PtrGetUnifiedPaymentSessionRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*unifiedPaymentModel.GetUnifiedPaymentSessionRequest")
}

func PtrCreateRefundRequest() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*refundModel.CreateRefundRequest")
}

func FilterRefundRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("refundModel.FilterRefundRequest")
}

func PtrPushNotificationMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*notification.PushNotification")
}

func PtrPostCreateLedgerRequestMockType() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("*paymentModel.PostCreateLedgerRequest")
}
