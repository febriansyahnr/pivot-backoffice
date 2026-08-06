package service

import (
	"bytes"
	"context"
	"crypto/rsa"
	"io"
	"mime/multipart"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	bankAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/country"
	credModel "github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/dukcapil"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/inbound"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/liveFeature"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	merchantforbiddenusecaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantRcn"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/paperCommunication"
	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/partition"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/passwordHistories"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platformFee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	rateLimiterModel "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	reconciliationModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	reportingModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reporting"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/roleMenuPermission"
	routingProcessorModelInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	routingProcessorModelEscrowBalance "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/escrowBalance"
	processorPriorityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/processorPriority"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	snapCoreTopUpSimulationModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/topUpSimulation"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	walletTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/walletInsights"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"

	"github.com/google/uuid"
	inboundPdk "github.com/paper-indonesia/pdk/v2/chiExt/inbound"
)

type IUserService interface {
	ListUsers(ctx context.Context, limit, offset int) ([]*userModel.User, error)
	ListUsersByMerchantID(
		ctx context.Context,
		filter *userModel.ListUsersByMerchantIDRequest,
		page, perPage int64) (*commonModel.PaginationResponse, error)
	FindUserByID(ctx context.Context, id string) (*userModel.User, error)
	FindUserByEmail(ctx context.Context, email string) (*userModel.User, error)
	Create(ctx context.Context, user *userModel.User) error
	CreateMerchantUser(ctx context.Context, req *userModel.MerchantUserRequest) (*userModel.User, error)
	Update(ctx context.Context, user *userModel.User) error
	ActivateUser(ctx context.Context, request *userModel.ActivateUserRequest) (*userModel.UserLoggedInResponse, error)
	UnblockUser(ctx context.Context, email string) error

	Login(ctx context.Context, request *userModel.UserLoginRequest) (user *userModel.User, signedToken string, err error)
	LoginWithOTP(ctx context.Context, email, password string) (token string, err error)
	GenerateTokenFromLogin2FA(ctx context.Context, id string) (user *userModel.User, signedToken string, err error)
	Logout(ctx context.Context, email string) (err error)
	ForgotPassword(ctx context.Context, email string) (token string, err error)
	ResetPassword(ctx context.Context, id, password string) error
	Refresh(ctx context.Context, email, refreshToken string) (user *userModel.User, token string, err error)

	CheckCurrentPassword(ctx context.Context, userID string, password string) error
	ChangePassword(ctx context.Context, userID string, OldPassword string, NewPassword string) (*userModel.ChangePasswordResponse, error)
	GenerateRandomPassword(ctx context.Context, userData *userModel.User) (*userModel.GenerateRandomPasswordResponse, error)
	UserDetail(ctx context.Context, id string) (*userModel.UserResponse, error)
	SendGeneratedInvitationURL(ctx context.Context, request *userModel.SendGeneratedInvitationRequest) error
	ValidateInvitationToken(ctx context.Context, token string) (*userModel.ValidateInvitationResponse, error)
	GetInvitationURL(ctx context.Context, merchantId, email string) (string, error)

	CreatePin(ctx context.Context, userID, pin string) error
	CheckCurrentPin(ctx context.Context, userID, pin string) error
	ChangePin(ctx context.Context, userID, pin, newPin string) error
	ResetPIN(ctx context.Context, id, pin string) error

	// Multi Factor Authentication (MFA)
	FindUserTOTPDataByID(ctx context.Context, userId string) (*userModel.UserTOTPData, error)
	EnrollTOTP(ctx context.Context, request userModel.EnrollTOTPRequest) (*userModel.EnrollTOTPResponse, error)
	ConfirmTOTP(ctx context.Context, request userModel.ConfirmTOTPRequest) (bool, error)
	SetPreferred2FAMethod(ctx context.Context, request userModel.SetPreferred2FAMethodRequest) (*userModel.SetPreferred2FAMethodResponse, error)
}

type IUserLoggedInDeviceService interface {
	Validate(ctx context.Context, userID, deviceIdentifier string, isRemember bool) error
}

type IMerchantService interface {
	FindMerchantByID(ctx context.Context, id string) (*merchantModel.Merchant, error)
	GetMerchantsByIDs(ctx context.Context, merchantIDs []string) ([]*merchantModel.Merchant, error)
	GetSubmerchantsByIDs(ctx context.Context, parentMerchantID string, submerchantIDs []string) ([]*merchantModel.Merchant, error)
	Create(ctx context.Context, merchant *merchantModel.Merchant, userId string) error
	Update(ctx context.Context, merchant *merchantModel.UpdateMerchantRequest) (*merchantModel.Merchant, error)
	UploadMerchantLogo(ctx context.Context, merchantID string, file *multipart.FileHeader) (string, error)
	CreatePKCS8SecretKey(ctx context.Context, merchantID string) (*merchantModel.PKCS8SecretKeyResponse, error)
	GetAccessTokenB2b(ctx context.Context, clientID, clientSecret string) (*string, error)
	ValidateAccessTokenB2b(ctx context.Context, request *merchantModel.ValidateAccessTokenB2bRequest) (*merchantModel.MerchantAuthTokenClaims, error)
	GetOrGenerateCallbackApiKey(ctx context.Context, id string) (*string, error)
	GetOrGenerateJITApiKey(ctx context.Context, merchantId string) (string, error)
	GetPKCS8SecretKey(ctx context.Context, merchantID string) (*merchantModel.PKCS8SecretKeyResponse, error)
	SetMerchantPublicKey(ctx context.Context, merchantId string, publicKey string) error
	UtilEncryptingKey(ctx context.Context, key string, data string) (string, error)
	FindMerchantFeeByID(ctx context.Context, id string) (*merchantModel.MerchantFee, error)
	FindMerchantFeeByMerchantIDAndType(
		ctx context.Context, merchantId, feeType string) (*merchantModel.MerchantFee, error)
	GetCachedMerchantStatus(ctx context.Context, id string) (*merchantModel.MerchantStatusResponse, error)
	UpdateStatusByID(ctx context.Context, status, reasonStatus, id string) error
	BlockMerchant(ctx context.Context, merchantId string) (*merchantModel.BlockMerchantResponse, error)
	UnblockMerchant(ctx context.Context, merchantId string) (*merchantModel.UnblockMerchantResponse, error)

	// Cron
	DormantMerchant(ctx context.Context, date time.Time) error

	// Merchant Fee
	CreateMerchantFee(ctx context.Context, request *merchantModel.NewMerchantFeeRequest) (*merchantModel.MerchantFeeResponse, error)
	UpdateMerchantFee(ctx context.Context, request *merchantModel.UpdateMerchantFeeRequest) error
	UpdateFeeTieringConfig(ctx context.Context, request *merchantModel.FeeTieringRequest) (*merchantModel.FeeTieringResponse, error)
	CreateFeeConfigOnBehalf(ctx context.Context, request *merchantModel.CreateFeeConfigOnBehalfRequest) (id string, err error)
	GetFeeConfigOnBehalf(ctx context.Context, request *merchantModel.GetFeeConfigOnBehalfRequest) ([]merchantModel.FeeConfigOnBehalfResponse, error)
	UpdateFeeConfigOnBehalf(ctx context.Context, id string, request *merchantModel.UpdateFeeConfigOnBehalfRequest) error

	GetSnapPrivateKey(ctx context.Context, merchantId string) (*rsa.PrivateKey, error)
	CreateSubMerchant(ctx context.Context, request *merchantModel.MerchantRequest) (*merchantModel.Merchant, error)
	UpdateSubMerchant(ctx context.Context, request *merchantModel.UpdateMerchantRequest) (*merchantModel.Merchant, error)
	UpdateSubMerchantOpenApi(ctx context.Context, request *merchantModel.UpdateMerchantOpenApiRequest) (*merchantModel.SubMerchantResponse, error)
	AssignSubMerchantAdmin(ctx context.Context, request *merchantModel.SubMerchantAdminRequest) error
	CreateMerchantAuth(ctx context.Context, merchantID string) error
	ValidateSubMerchantParent(ctx context.Context, parentMerchantID, merchantID string) error
	BulkCreateSubMerchant(ctx context.Context, request *merchantModel.BulkCreateSubMerchantRequest) (*merchantModel.BulkCreateSubMerchantResponse, error)
	ProcessBulkCreateSubMerchant(ctx context.Context, request *merchantModel.ProcessBulkCreateSubMerchantRequest) error
	GetBulkCreateSubMerchantSummary(ctx context.Context, request *merchantModel.GetBulkCreateSubMerchantSummaryRequest) (*merchantModel.BulkCreateSubMerchantResponse, error)
	// Merchant Config
	TransactionConfig(ctx context.Context, merchantId string, config *merchantModel.TransactionConfigs) error
	FDSConfig(ctx context.Context, merchantID string, config merchantModel.FDSConfigRequest) (*merchantModel.FDSConfigResponse, error)
	PaymentInvestigationConfig(ctx context.Context, merchantID string, config merchantModel.PaymentInvestigationConfigRequest) (*merchantModel.PaymentInvestigationConfigResponse, error)
	GetTransactionConfig(ctx context.Context, merchantId string) (*merchantModel.TransactionConfigResp, error)
	GetFDSConfig(ctx context.Context, merchantID string) (*merchantModel.GetFDSConfigResponse, error)
	GetDepositSetting(ctx context.Context, merchantId string) (*merchantModel.DepositSettingResponse, error)
	SetAutoWithdrawal(ctx context.Context, request *merchantModel.AutoWithdrawalSettingRequest) error
	GetMerchantIdForConfigs(ctx context.Context, merchantId string, setMerchantCtx bool) (context.Context, *merchantModel.MerchantIdForConfigs, error)
	EnableAllPaymentMethod(ctx context.Context, merchant *merchantModel.Merchant) error

	// Settlement Config
	UpdateSettlementConfig(ctx context.Context, merchantFeeId string, config *merchantModel.SettlementConfig) error

	// Merchant For KYC/KYB
	UploadDocument(ctx context.Context, document *merchantModel.UploadDocumentReq) (id string, err error)
	GetDocuments(ctx context.Context, request *merchantModel.MerchantDocumentFilterRequest) (resp *commonModel.PaginationResponse, err error)
	FindDocumentByType(ctx context.Context, merchantId, docType string) (doc *merchantModel.Document, err error)
	UpsertMerchantBOD(ctx context.Context, request *merchantModel.UpsertBoardOfDirectorReq) (id string, err error)
	GetListMerchantBODs(ctx context.Context, merchantId string) (resp []merchantModel.BoardOfDirectorResp, err error)

	// Open API
	GenOpenAPISignature(ctx context.Context, req *merchantModel.GenSignatureReq) (string, error)
	GetSNAPAccessTokenB2B(ctx context.Context, request *merchantModel.SNAPAccessTokenB2BReq) (*merchantModel.SNAPAccessTokenB2BResp, error)
	ValidateSNAPAccessTokenRequestSignature(ctx context.Context, request *merchantModel.SNAPValidateB2b2cTokenSignatureRequest) error

	ValidateSnapRequestSignature(ctx context.Context, req *merchantModel.ValidateSnapSignatureRequest) error
	GenerateSnapRequestSignature(ctx context.Context, req *merchantModel.GenerateSnapSignatureRequest) (string, error)

	ListSubMerchantByParentID(
		ctx context.Context,
		filter *merchantModel.SubMerchantListFilter,
		page, perPage int64) (*commonModel.PaginationResponse, error)
	SubMerchantResendInvitation(ctx context.Context, request *merchantModel.ResendInvitationRequest) error

	CloseMerchant(ctx context.Context, req *merchantModel.UpdateMerchantStatus) error

	// Merchant KYC
	UpdateKYC(ctx context.Context, payload merchantModel.UpdateMerchantKYCRequest) (*merchantModel.UpdateMerchantKYCResponse, error)

	SetCustomLimitConfig(ctx context.Context, request merchantModel.BeneficiaryLimitConfigRequest) error
	UploadReservedShortName(ctx context.Context, request *merchantModel.ReservedMerchantShortNameRequest) error

	// Merchant Billing
	GetBillingFees(ctx context.Context, request merchantModel.BillingFeeRequest) (*merchantModel.BillingFeeResponse, error)
	PayBillingFees(ctx context.Context, request merchantModel.PayBillingFeeRequest) (*merchantModel.BillingFeeResponse, error)

	// Migration
	MigrateMerchantSecretsToEncryption(ctx context.Context) error

	GetNotificationConfig(ctx context.Context, merchantID string) (*merchantModel.MerchantNotificationConfig, error)
	UpdateNotificationConfig(ctx context.Context, merchantID string, req *merchantModel.MerchantNotificationConfig) (*merchantModel.MerchantNotificationConfig, error)
}

type IPasswordHistoriesService interface {
	FindByUserID(ctx context.Context, userID string) ([]*passwordHistories.PasswordHistories, error)
	InsertByUserID(ctx context.Context, userID string, password string) error
}

type IActivityService interface {
	Create(ctx context.Context, activity *activityModel.Activity) error
	GetList(ctx context.Context, filter activityModel.ActivityFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
}

type IRoleService interface {
	FindRoleBySlug(ctx context.Context, slug string) (*role.Role, error)
	FindRoleById(ctx context.Context, slug string) (*role.Role, error)
	Create(ctx context.Context, role *role.Role) error
	Update(ctx context.Context, user *role.Role) error
	GetList(ctx context.Context, filter *role.GetRoleFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	Delete(ctx context.Context, merchantID, roleID string) (err error)

	CreateRoleAndPermissions(ctx context.Context, payload *role.CreateRoleRequest) (*role.RoleMenuResponse, error)
	UpdateRoleAndPermissions(ctx context.Context, payload *role.UpdateRoleRequest) (*role.RoleMenuResponse, error)
	AddDefaultRolePermissions(ctx context.Context, payload *role.CRMUpdateDefaultRolePermissionsRequest) (*role.RoleMenuResponse, error)
	DeleteDefaultRolePermissions(ctx context.Context, payload *role.CRMUpdateDefaultRolePermissionsRequest) (*role.RoleMenuResponse, error)
}

type IUserRoleService interface {
	FindUserRoleByUserID(ctx context.Context, userId string) (*userRole.UserRole, error)
	Create(ctx context.Context, userRole *userRole.UserRole) error
	UpdateByUserID(ctx context.Context, userRole *userRole.UserRole) error
}

type IPermissionService interface {
	FindBySlug(ctx context.Context, slug string) (*permissionModel.Permission, error)
	FindByRoleId(ctx context.Context, roleId string) ([]*permissionModel.Permission, error)

	Create(ctx context.Context, permission *permissionModel.Permission) error
	Update(ctx context.Context, permission *permissionModel.Permission) error
	GetCachedPermissionsByRoleId(ctx context.Context, roleId string) ([]string, error)
}

type IMenuService interface {
	FindBySlug(ctx context.Context, slug string) (*menuModel.Menu, error)
	GetAll(ctx context.Context, excludeHome bool) ([]*menuModel.MenuResponse, error)
	GetByRole(ctx context.Context, roleID string, isMenuFormatting bool) ([]*menuModel.MenuResponse, error)

	Create(ctx context.Context, menu *menuModel.Menu) error
	Update(ctx context.Context, menu *menuModel.Menu) error
	IsShouldUpdate(ctx context.Context, existingMenu *menuModel.Menu, newMenu roleMenuPermissionModel.MenuPermissionFromFileRequest) bool
}

type IRoleMenuPermissionService interface {
	Create(ctx context.Context, pivot *roleMenuPermissionModel.RoleMenuPermission) error
}

type IOrchestratorService interface {
	CreateAccountTransaction(ctx context.Context, request *orchestratorModel.CreateAccountTransactionRequest) error
	UpdateStatusAccountTransaction(ctx context.Context, id string, status string, reasonType, reasonDescription *string) error
	UpdateReasonOnly(ctx context.Context, id string, reasonType, reasonDescription *string) error
	UpdateStatusAccountTransactionByReferenceID(ctx context.Context, id string, status string, reasonType, reasonDescription *string) error
	GetAvailableMerchantBalance(ctx context.Context, merchantID, balanceName string) (float64, error)
	GetMerchantBulkBalances(ctx context.Context, request *account_model.GetBulkBalanceRequest) (map[string]*account_model.AvailableBalanceResponse, error)
	GetList(ctx context.Context, filter *orchestratorModel.TransactionHistoryFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetDetailById(ctx context.Context, merchantId, id string) (*orchestratorModel.TransactionHistoryDetailResp, error)
	FindByReference(ctx context.Context, referenceID, referenceType string) (*orchestratorModel.AccountTransactionWithUseCase, error)
	FindByID(ctx context.Context, id string) (*orchestratorModel.AccountTransactionWithUseCase, error)
	GetReferenceIdByTransactionIdAndType(ctx context.Context, transactionId, transactionType string) (referenceId string, err error)
	PostAccountTransaction(ctx context.Context, request *orchestratorModel.CreateAccountTransactionRequest) error
	GenExcelForTransactionHistories(ctx context.Context, w *orchestratorModel.FileWriter, req *orchestratorModel.TransactionHistoryFilterRequest) error
	GetPendingBalance(ctx context.Context, subMerchantId, balanceName string) (float64, error)
	UpdateAdditionalInfoByID(ctx context.Context, id string, additionalInfo []byte) error
	UpdateStatusAccountAndAdditionalInfoTransaction(ctx context.Context, id string, status string, reasonType string, reasonDescription string, additionalInfo []byte) error
	UpdateProcessorAndReconReferenceByID(ctx context.Context, id string, processorReferenceName, processorReferenceId, reconReference string) error
	UpdateTransactionTimestamp(ctx context.Context, id string, transactionTimestamp time.Time) error
	UpdateTransaction(ctx context.Context, request *orchestratorModel.UpdateTransactionRequest) error
	GetWalletCustomersTotalBalance(ctx context.Context, request *orchestratorModel.GetWalletTotalBalanceRequest) (float64, error)
	GetMerchantBalance(ctx context.Context, request orchestratorModel.GetMerchantBalanceRequest) (*orchestratorModel.GetMerchantBalanceResponse, error)

	VoidCreditcardTransaction(ctx context.Context, request *orchestratorModel.VoidTransactionRequest) error
}

type IPaymentService interface {
	CreatePayment(
		ctx context.Context,
		merchantID string, paymentRequest paymentModel.PaymentRequest) (*paymentModel.PaymentResponse, error)
	UpdatePayment(
		ctx context.Context, req *paymentModel.PaymentUpdateRequest) (*paymentModel.PaymentResponse, error)
	FindPaymentById(ctx context.Context, id, merchantID string) (*paymentModel.PaymentResponse, error)
	GetAndUpdateVirtualAccountPayment(
		ctx context.Context,
		request *paymentModel.VirtualAccountPaymentNotificationRequest) (*paymentModel.Payment, error)

	GetQrMpmDynamic(ctx context.Context, uuid string, referenceId string, merchantId string) (*paymentModel.PaymentResponse, error)
	GetQrMpmStatic(ctx context.Context, request *paymentModel.QueryQrMpmStaticRequest, merchantId string) (*paymentModel.PaymentResponse, error)
	ProcessQrisPayment(ctx context.Context, request *paymentModel.QrisPaymentNotificationRequest) error

	FindPaymentForSimulationByID(ctx context.Context, id string) (*paymentModel.PaymentResponse, error)
	ProcessPaymentForSimulationByID(ctx context.Context, id string, paymentAmount commonModel.Amount, status string) error

	ProcessVirtualAccountPayment(ctx context.Context, request *paymentModel.VirtualAccountPaymentNotificationRequest) (err error)

	GetTotalPaymentBalance(ctx context.Context, merchantID uuid.UUID) (*commonModel.Amount, error)
	FilterPaymentHistory(ctx context.Context, opt paymentModel.FilterPaymentHistoryOption) (*commonModel.PaginationResponse, error)
	GetInvestigationSummary(ctx context.Context, opt paymentModel.GetInvestigationSummaryOption) (*paymentModel.InvestigationSummaryResponse, error)
	GetPaymentHistoryDetail(ctx context.Context, opt paymentModel.PaymentHistoryDetailOption) (*paymentModel.PaymentHistoryDetailResponse, error)
	GetTodayPaymentInsight(ctx context.Context, opt paymentModel.PaymentInsightOption) (*paymentModel.PaymentInsightItem, error)
	Export(ctx context.Context, request *paymentModel.PaymentDownloadHistoryRequest) (*paymentModel.PaymentDownloadHistoryResponse, error)
	GetListForInternalDashboard(ctx context.Context, request *paymentModel.GetListFilterRequest) (*commonModel.PaginationResponse, error)
	GetDetailByID(ctx context.Context, id string) (*paymentModel.Payment, error)
	InquiryPayment(ctx context.Context, request *paymentModel.InquiryPaymentRequest) (*paymentModel.Payment, error)
	GetSplitRoutingByTransferID(ctx context.Context, paymentID, transferID string) (*splitRoutingPaymentModel.SplitRoutingPaymentDetailResponse, error)
	DeterminePaymentFee(ctx *context.Context, payment *paymentModel.Payment) error
	GetEncryptionKey(ctx context.Context) (*paymentModel.GetEncryptionKeyResponse, error)
	VCCTerminalBatchCharge(ctx context.Context, request *paymentModel.VCCTerminalChargeRequest) (*paymentModel.VCCTerminalBatchChargeResponse, error)
	VCCTerminalSubmitCharge(ctx context.Context, request paymentModel.VCCTerminalChargeMessage) error

	// Payment UI
	GetPaymentDetailForPaymentUI(ctx context.Context, paymentID string) (*paymentModel.PaymentDetailForPaymentUIResponse, error)
	PublishPaymentExpirationMessage(ctx context.Context) error

	ExpirePayment(ctx context.Context, request paymentModel.ExpiringPayment) error
	GetImages(ctx context.Context) (paymentModel.ImageResponse, error)
	GetPaymentInstructions(ctx context.Context, paymentMethod string) ([]paymentModel.InstructionResponse, error)
	GeneratePaymentToken(ctx context.Context, paymentID string, expiryAt time.Time) (string, error)
	HandleStrictExpiry(ctx context.Context, paymentID string) error

	// Unified Payment
	CreateUnifiedPayment(ctx context.Context, request *paymentModel.CreateUnifiedPaymentRequest) (*paymentModel.CreateUnifiedPaymentResponse, error)
	GetPaymentByReferenceId(ctx context.Context, referenceId string, merchantID string) (*paymentModel.UnifiedPaymentResponse, error)
	UpdateUnifiedPayment(ctx context.Context, request *paymentModel.UpdateUnifiedPaymentRequest) (*paymentModel.UpdateUnifiedPaymentResponse, error)
	GetActivePaymentByProcessorReferenceNumber(ctx context.Context, request *paymentModel.GetActivePaymentByProcessorReferenceNumberRequest) (*paymentModel.Payment, error)

	// Payment Ledger
	IPaymentLedgerService

	// Split and Route Payment
	ProcessSplitRoute(ctx context.Context, paymentID string) error

	// Static QRIS Dashboard
	FilterStaticQrisList(ctx context.Context, opt paymentModel.StaticQrisFilterRequest) (*commonModel.PaginationResponse, error)
	GetStaticQrisDetail(ctx context.Context, opt paymentModel.StaticQrisDetailRequest) (*paymentModel.StaticQrisDetailResponse, error)
	GetStaticQrisTransactions(ctx context.Context, opt paymentModel.StaticQrisTransactionFilterRequest) (*commonModel.PaginationResponse, error)
	GetFirstActiveStaticQris(ctx context.Context, merchantID string, partnerReferenceNo string) (*paymentModel.Payment, error)
	GetMaxActiveStaticQRPerMerchant() int
	DeactivateStaticQris(ctx context.Context, paymentID string, merchantID string, request paymentModel.StaticQrisUpdateStatusRequest) error
	UpdatePaymentMetadataById(ctx context.Context, paymentID string, metadata paymentModel.UpdatePaymentMetadataRequest) error

	// Static VA Dashboard
	FilterStaticVaList(ctx context.Context, opt paymentModel.StaticVaFilterRequest) (*commonModel.PaginationResponse, error)
	GetStaticVaDetail(ctx context.Context, opt paymentModel.StaticVaDetailRequest) (*paymentModel.StaticVaDetailResponse, error)
	GetStaticVaTransactions(ctx context.Context, opt paymentModel.StaticVaTransactionFilterRequest) (*commonModel.PaginationResponse, error)
	DeactivateStaticVa(ctx context.Context, paymentID string, merchantID string, request paymentModel.StaticVaUpdateStatusRequest) error

	// Status History
	RecordPaymentStatusHistory(ctx context.Context, paymentID, actor, statusType string)
	// CRM
	CRMRetryNotification(ctx context.Context, request *paymentModel.CRMRetryNotificationRequest) error
	CRMStaticVARetryNotification(ctx context.Context, request *paymentModel.CRMStaticVARetryNotificationRequest) error

	// Investigation
	GetInvestigatedPayments(ctx context.Context, filter *paymentModel.GetInvestigatedPaymentsFilterRequest) (*commonModel.PaginationResponse, error)
	GetInvestigationProofOfPayment(ctx context.Context, request paymentModel.GetInvestigationProofOfPaymentRequest) (*paymentModel.GetInvestigationProofOfPaymentResponse, error)
	UpdateInvestigationStatus(ctx context.Context, paymentID string, request *paymentModel.UpdateInvestigationRequest) (*paymentModel.UpdateInvestigationResponse, error)
	ProcessInvestigationMonthlyReconciliation(ctx context.Context, request paymentModel.MonthlyReconciliationRequest) error
	ExportInvestigatedPayments(ctx context.Context, request *paymentModel.InvestigationDownloadHistoryRequest) (*paymentModel.InvestigationDownloadHistoryResponse, error)

	// Payment Insights
	GetPaymentDashboardInsights(ctx context.Context, request paymentModel.GetPaymentDashboardInsightRequest) (*paymentModel.PaymentDashboardInsights, error)

	// VCC Terminal
	GetVCCTerminalList(ctx context.Context, request *paymentModel.GetVCCTerminalListFilterRequest) (*commonModel.PaginationResponse, error)

	// Receipt
	GetReceiptByID(ctx context.Context, request *paymentModel.GetPaymentReceiptRequest) (*paymentModel.GetPaymentReceiptResponse, error)
}

type IRefundService interface {
	Create(ctx context.Context, request *refundModel.CreateRefundRequest) (*refundModel.RefundResponse, error)
	FindByID(ctx context.Context, refundID string) (*refundModel.Refund, error)
	GetRefundList(ctx context.Context, request refundModel.FilterRefundRequest) (*commonModel.PaginationResponse, error)
	GetRefundDetail(ctx context.Context, request refundModel.FilterRefundRequest) (*refundModel.RefundResponse, error)
	GetRefundDetailWithStatusHistories(ctx context.Context, request refundModel.FilterRefundRequest) (*refundModel.RefundResponse, error)
	GetExistingRefundList(ctx context.Context, request refundModel.GetExistingRefundListRequest) ([]refundModel.RefundResponse, error)
	GetTotalRefundedAmount(ctx context.Context, paymentID string) (float64, error)
	SendCallback(ctx context.Context, refundID string, merchantID string) error
	RecordRefundStatusHistory(ctx context.Context, refundID, actor, statusType string)
	GetReceipt(ctx context.Context, request *refundModel.GetRefundReceiptRequest) (*refundModel.GetRefundReceiptResponse, error)
}

type IRecurringContractService interface {
	Create(ctx context.Context, request recurringContractModel.CreateRecurringContractRequest) (*recurringContractModel.CreateRecurringContractResponse, error)
	Cancel(ctx context.Context, request recurringContractModel.CancelRecurringContractRequest) error
	GetRecurringByID(ctx context.Context, request recurringContractModel.GetRecurringByIDRequest) (*recurringContractModel.GetRecurringByIDDashboardResponse, error)
	UpdateRecurringPayment(ctx context.Context, request recurringContractModel.UpdateRecurringPaymentRequest) error
}

type IRefundProcessorService interface {
	Process(ctx context.Context, request *refundModel.RefundProcessRequest) error
	ProcessUpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error
}

type IPaymentLedgerService interface {
	PostCreateLedger(ctx context.Context, payment *paymentModel.Payment, request *paymentModel.PostCreateLedgerRequest) error
	UpdatePendingLedger(ctx context.Context, payment *paymentModel.Payment, request orchestratorModel.UpdatePaymentTransactionRequest) error
	PostCreateFeeTransaction(ctx context.Context, payment *paymentModel.Payment, request *paymentModel.PostCreateFeeTransactionRequest) error
	DeterminePaymentFee(ctx *context.Context, payment *paymentModel.Payment) error
}

type IPaymentInternalDirectFunc interface {
	CreateUnifiedPayment(ctx context.Context, request *paymentModel.CreateUnifiedPaymentRequest) (*paymentModel.CreateUnifiedPaymentResponse, error)
	UpdatePayment(ctx context.Context, req *paymentModel.PaymentUpdateRequest) (*paymentModel.PaymentResponse, error)
	CreatePayment(ctx context.Context, merchantID string, paymentRequest paymentModel.PaymentRequest) (*paymentModel.PaymentResponse, error)
	DeterminePaymentFee(ctx *context.Context, payment *paymentModel.Payment) error
	DecryptRequest(ctx context.Context, data *encryption.DataEncryption, dst any) error

	FilterPaymentHistory(ctx context.Context, opt paymentModel.FilterPaymentHistoryOption) (*commonModel.PaginationResponse, error)
	GetCacheDownloadHistory(ctx context.Context, hashFilterKey string) (url string, err error)
	ExportToExcel(ctx context.Context, request *paymentModel.PaymentDownloadHistoryRequest, transactions []paymentModel.PaymentHistoryItem, wr io.Writer) error
}

type IUnifiedPaymentService interface {
	// Case: Create payment link from merchant dashboard
	CreateDashboardPaymentLink(ctx context.Context, request *unifiedPaymentModel.DashboardPaymentLinkCreateRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error)
	CreateSession(ctx context.Context, request *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error)
	ConfirmSession(ctx context.Context, request *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error)
	CancelSession(ctx context.Context, request *unifiedPaymentModel.CancelUnifiedPaymentSessionRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error)
	GetSessionDetail(ctx context.Context, request *unifiedPaymentModel.GetUnifiedPaymentSessionRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error)
	GetSessionList(ctx context.Context, request *paymentModel.GetListFilterRequest) (*commonModel.PaginationResponse, error)

	GetChargeList(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest) (*commonModel.PaginationResponse, error)
	GetChargeDetail(ctx context.Context, request *unifiedPaymentModel.GetUnifiedPaymentChargeRequest) (*unifiedPaymentModel.ChargeResponse, error)
	ExportCharge(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest) (*commonModel.ExportResponse, error)

	ProcessorInitialization(ctx context.Context, request *unifiedPaymentModel.BaseProcessorRequest) (resp *unifiedPaymentModel.ChargePaymentMethodDetails, err error)
	InquiryEWalletPayment(ctx context.Context, payment *paymentModel.Payment) (*paymentModel.Payment, error)
	InquiryCardPayment(ctx context.Context, payment *paymentModel.Payment) (*paymentModel.Payment, error)
	UpdateEWalletPaymentSession(ctx context.Context, paymentID string) (*paymentModel.Payment, error)

	ProcessNotification(ctx context.Context, request *unifiedPaymentModel.PaymentNotificationRequest) error

	// Payment Method
	GetPaymentMethodConfig(ctx context.Context, merchantId string) (*unifiedPaymentModel.GetPaymentMethodConfigResponse, error)

	// Payment Simulation via Open API
	SimulatePayment(ctx context.Context, request *unifiedPaymentModel.SimulatePaymentRequest) error

	// Card Encryption
	EncryptCard(ctx context.Context, request *unifiedPaymentModel.EncryptCardRequest) (*unifiedPaymentModel.EncryptedCardResponse, error)
	GetEncryptedCard(ctx context.Context, merchantId, cardId string) (*unifiedPaymentModel.EncryptedCardResponse, error)
	// BIN Lookup
	GetCardBinDetail(ctx context.Context, request unifiedPaymentModel.GetBinDetailRequest) (*unifiedPaymentModel.GetBinDetailResponse, error)

	Capture(ctx context.Context, request *unifiedPaymentModel.CaptureRequest) (*unifiedPaymentModel.CaptureResponse, error)
	ProcessCapture(ctx context.Context, request *unifiedPaymentModel.ProcessCaptureRequest) error

	// Callback
	SendCallback(ctx context.Context, payment *paymentModel.Payment)
	ResendPaymentCallback(ctx context.Context, request *callbackModel.ResendCallbackRequest) error

	// Investigation
	UploadProofOfPayment(ctx context.Context, request *unifiedPaymentModel.UploadProofOfPaymentRequest) (*unifiedPaymentModel.UploadProofOfPaymentResponse, error)
}

// Background: Internal service to reduce test creation complexity
type IInternalUnifiedPaymentService interface {
	CreateSession(ctx context.Context, request *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error)
	PrepareRecurringPaymentRequest(ctx context.Context, request *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) error

	// Auto Split Payment
	InitiateSplitPayment(ctx context.Context, request *paymentModel.ProcessSplitPaymentRequest) error
	ContinueSplitPaymentExecution(ctx context.Context, request *paymentModel.ProcessSplitPaymentRequest) error
	EvaluateSplitPaymentOutcome(ctx context.Context, payment *paymentModel.Payment) error
	FinalizeSplitPayment(ctx context.Context, request *paymentModel.ProcessSplitPaymentRequest) error
	GetAutoSplitPaymentDetail(ctx context.Context, request *paymentModel.GetAutoSplitPaymentSummaryRequest) (*unifiedPaymentModel.AutoSplitPaymentSummary, error)
	AbortSplitPaymentOnCITFailure(ctx context.Context, request *paymentModel.ProcessSplitPaymentRequest) error
}

type IPaymentMethodService interface {
	Create(ctx context.Context, payload *paymentMethodModel.CreatePaymentMethodRequest) error
	FindPaymentMethodByCategory(ctx context.Context, category string) ([]*paymentModel.PaymentMethod, error)
	FindPaymentMethodByIdAndMerchant(ctx context.Context, paymentMethodID, merchantID string) (*paymentModel.PaymentMethodWithPivot, error)
	GetPaymentMethodByMerchant(ctx context.Context, filter *paymentModel.GetPaymentMethodFilterRequest) ([]*paymentModel.PaymentMethodWithPivot, error)
	GetStaticVAPaymentMethodByMerchant(ctx context.Context, filter *paymentModel.GetPaymentMethodFilterRequest) ([]*paymentModel.PaymentMethodWithPivot, error)
	GetStaticQRPaymentMethodByMerchant(ctx context.Context, filter *paymentModel.GetPaymentMethodFilterRequest) (*paymentModel.PaymentMethodWithPivot, error)
	GetActivePaymentMethodDetailForPaymentRequest(ctx context.Context, request paymentModel.GetActivePaymentMethodRequest) (*paymentModel.PaymentMethodWithPivot, error)

	Deactivate(ctx context.Context, paymentMethodMerchant *paymentModel.PaymentMethodWithPivot) error
	Activate(ctx context.Context, paymentMethodMerchant *paymentModel.PaymentMethodWithPivot) error

	SetupConfig(ctx context.Context, request *paymentMethodModel.SetupPaymentMethodConfigRequest) error
	ChangeActivationStatus(ctx context.Context, request *paymentMethodModel.ChangeActivationStatusRequest) error
	GetRequiredMerchantDocuments(ctx context.Context, request *paymentMethodModel.GetRequiredMerchantDocumentsRequest) (*[]paymentMethodModel.MerchantRequiredDocumentsResponse, error)
}

type ICallbackService interface {
	RegisterCallback(ctx context.Context, request *callbackModel.RegisterCallbackRequest) (*callbackModel.RegisterCallbackResponse, error)
	ProcessCallback(ctx context.Context, request *callbackModel.ProcessCallbackRequest) error
	GetList(ctx context.Context, filter *callbackModel.GetListCallbackFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetCallbackURLByMerchantId(ctx context.Context, request *callbackModel.CallbackURLSettingReq) (any, error)
	GetCallbackAPIKeyByMerchantId(ctx context.Context, request *callbackModel.CallbackURLSettingReq) (any, error)
	GetCallbackLogList(ctx context.Context, filter *callbackModel.GetListCallbackLogFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetCallbackLogDetail(ctx context.Context, id, merchantID string) (*callbackModel.CallbackLogWithMaster, error)
	FindCallbackByMerchantIdAndCallbackName(ctx context.Context, merchantId uuid.UUID, callbackName string) (*callbackModel.Callback, error)
	ResendCallback(ctx context.Context, id, merchantID string) error
	ResendSNAPCallback(ctx context.Context, id, merchantID string) error
	SendMerchantCallback(ctx context.Context, request callbackModel.SendMerchantCallbackRequest) (*callbackModel.SendMerchantCallbackResponse, error)
	TestAndSaveCallbackURL(ctx context.Context, request *callbackModel.TestAndSaveCallbackURLReq) (*callbackModel.TestAndSaveCallbackURLResp, error)
	TestAndSaveB2b(ctx context.Context, request *callbackModel.TestAndSaveCallbackURLReq) (*callbackModel.TestAndSaveCallbackURLResp, error)
	TestAndSaveSnapPayment(ctx context.Context, request *callbackModel.TestAndSaveCallbackURLReq) (*callbackModel.TestAndSaveCallbackURLResp, error)
	WriteCallbackLogFromWorkflowTask(ctx context.Context, log callbackModel.WorkflowWriteLogRequest) (*callbackModel.WorkflowWriteLogResponse, error)
	GetCallbackEvents(ctx context.Context) ([]callbackModel.CallbackEvent, error)
}

type IInboundService interface {
	inboundPdk.Recorder

	GetList(ctx context.Context, filter *inboundModel.GetInboundFilterRequest) (*commonModel.PaginationResponse, error)
	GetByID(ctx context.Context, id string) (*inboundModel.InboundResponse, error)
	GetSnapVersionByID(ctx context.Context, id string) (*inboundModel.InboundSnapVersionResponse, error)
}

type IMerchantTopUpCallbackService interface {
	SendCallback(ctx context.Context, event string, request *merchantTopUp.MerchantTopUpCallbackRequest) error
}

type IMerchantTopUpService interface {
	CreateTopupSimulation(ctx context.Context, req snapCoreTopUpSimulationModel.TopupSimulationRequest) (*snapCoreTopUpSimulationModel.TopupSimulationResponseData, error)
	ProcessMerchantTopUpWithVirtualAccount(ctx context.Context, request *paymentModel.VirtualAccountPaymentNotificationRequest) error

	FindOrCreate(ctx context.Context, merchantId, accountName, paymentMethodId string) (*merchantTopUp.MerchantTopUp, error)
	FindByMerchantAccountNameAndPaymentMethodId(ctx context.Context, merchantId, accountName, paymentMethodId string) (*merchantTopUp.MerchantTopUp, error)
	FindByReferenceNumber(ctx context.Context, referenceNumber string) (*merchantTopUp.MerchantTopUp, error)

	IMerchantTopUpCallbackService

	GetList(ctx context.Context, request *merchantTopUp.TopUpTransactionListRequest) (*commonModel.PaginationResponse, error)
}

type IBeneficiaryAccountService interface {
	FindByBankCodeAndAccountNo(
		ctx context.Context,
		req *beneficiaryAccountModel.CheckAccountRequest) (*beneficiaryAccountModel.Account, error)
	GetList(
		ctx context.Context,
		filter *beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest,
		page, perPage int64) (*commonModel.PaginationResponse, error)
}

type IDisbursementDashboardService interface {
	Get(ctx context.Context, request disbursementDashboardModel.GetDisbursementDashboardFilter) (*disbursementDashboardModel.DisbursementDashboardResponse, error)
	GetApprovalDashboard(ctx context.Context, merchantID uuid.UUID) (*disbursementDashboardModel.DisbursementApprovalDashboardResponse, error)
}

type IDisbursementService interface {
	WPRelease()

	// Get
	GetList(ctx context.Context, filter *disbursementModel.GetDisbursementFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	ExportToExcel(ctx context.Context, filter *disbursementModel.GetDisbursementFilterRequest) (*disbursementModel.ExportDisbursementListResponse, error)
	GetListBulk(ctx context.Context, filter *disbursementModel.GetBulkDisbursementFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetBulkDisbursementDetail(ctx context.Context, id string) (*disbursementModel.BulkDisbursementDetail, error)
	GetBulkDisbursementForOpenApiByID(ctx context.Context, filter *disbursementModel.GetDisbursementFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetBulkDisbursementForOpenApiByReferenceID(ctx context.Context, bulkID, referenceID, merchantID string) (*disbursementModel.GetBulkDisbursementForOpenApiByReferenceIDResponse, error)
	GetDisbursementByReferenceID(ctx context.Context, referenceID, merchantID string) (*disbursementModel.PayoutObject, error)
	FindByID(ctx context.Context, id string) (*disbursementModel.DisbursementWithTransaction, error)
	GetReceiptByID(ctx context.Context, request *disbursementModel.GetDisbursementReceiptRequest) (*disbursementModel.GetDisbursementReceiptResponse, error)
	GetTransactionConfig(ctx context.Context, merchantId string) (result *disbursementModel.TransactionConfig, err error)
	GetDailyTransactionLimit(ctx context.Context, merchantId, merchantType string) (*disbursementModel.DailyTransactionLimitResponse, error)
	ReportAfterPayoutCutOffTime(ctx context.Context, startTime, endTime time.Time) (disbursementModel.AfterPayoutCutOffTimeSummary, error)
	ReportAfterPayoutCutOffTimeByPartnerWindow(ctx context.Context, req *disbursementModel.PartnerWindowCutOffReportRequest) (disbursementModel.AfterPayoutCutOffTimeSummary, error)
	GetDisbursementInsight(ctx context.Context, filter disbursementModel.GetDisbursementInsightFilter) (*disbursementModel.DisbursementInsightResponse, error)

	// Create
	CreateSingle(ctx context.Context, request *disbursementModel.CreateSingleRequest) (*disbursementModel.Disbursement, error)
	CreateBulk(ctx context.Context, request *disbursementModel.CreateBulkDisbursementRequest) (*disbursementModel.BulkDisbursement, error)
	BatchCreateDisbursement(ctx context.Context, request *disbursementModel.BatchCreateDisbursementRequest) (err error)
	Reversal(ctx context.Context, request *disbursementModel.ReversalTransactionReq) (*disbursementModel.ReversalTransactionResp, error)
	BulkCreate(ctx context.Context, request *disbursementModel.BulkCreateRequest) (*disbursementModel.BulkCreateResponse, error)

	// Update
	ProcessUpdateTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error

	// Action
	ApprovalAction(ctx context.Context, request *disbursementModel.ApprovalActionsRequest) (*disbursementModel.ApprovalActionsResponse, error)
	Approve(ctx context.Context, request *disbursementModel.ApproveRequest) error
	Reject(ctx context.Context, request *disbursementModel.RejectRequest) (string, error)
	Process(ctx context.Context, id string, isRetryTransfer bool) error
	BatchProcessDisbursement(ctx context.Context, request *disbursementModel.BatchProcessDisbursementRequest) (err error)
	RetrySingle(ctx context.Context, request *disbursementModel.RetrySingleRequest) error
	RetryBulk(ctx context.Context, request *disbursementModel.RetryBulkRequest) error
	GenerateExcelAndUpdateInvalidBulkDisbursement(ctx context.Context, bulkID string, rows []*disbursementModel.BulkPreviewResponse) (string, error)
	GenerateExcelAndUpdateRejectedBulkDisbursement(ctx context.Context, bulkID string, rows []*disbursementModel.BulkPreviewResponse) (string, error)
	RetryDueToInsufficientEscrowFund(ctx context.Context, request *disbursementModel.RetryTransaction) error
	InquiryTransaction(ctx context.Context, request *disbursementModel.InquiryTransaction) (*disbursementModel.DisbursementWithTransaction, error)
	BulkPreview(ctx context.Context, request *disbursementModel.BulkPreviewRequest) ([]disbursementModel.BulkPreviewResponse, error)
	BulkValidate(ctx context.Context, request *disbursementModel.BulkPreviewRequest) ([]disbursementModel.BulkPreviewResponse, error)
	RetryInquirePendingTransactions(ctx context.Context, start, end time.Time) (*disbursementModel.RetryInquireDisbuesementSummary, error)
	Cancel(ctx context.Context, request *disbursementModel.CancelPayoutRequest) ([]string, error)
	ValidateBatchPayoutItems(ctx context.Context, request *disbursementModel.ApprovalActionsRequest) (*disbursementModel.ApprovalActionsRequest, error)

	// Validate
	ValidateBalance(ctx context.Context, request *disbursementModel.ValidateBalanceRequest) bool
	ValidateDailyTransactionLimit(ctx context.Context, merchantId string, totalAmount float64) (ITransactionCloser, error)
	IsExistReferenceID(ctx context.Context, merchantID, referenceID string) bool

	// change transaction status for internal tool
	ChangeDisbursementTransactionStatus(ctx context.Context, request disbursementModel.ChangeDisbursementTransactionStatusRequest) []disbursementModel.ChangeDisbursementTransactionStatusResponse

	// check status
	CheckTransactionStatus(ctx context.Context, request *disbursementModel.CheckDisbursementTransactionStatusRequest) ([]*disbursementModel.CheckDisbursementTransactionStatusResponse, error)
	GetCutOffTimeStatus(ctx context.Context, now time.Time, merchantId string, windowConfig *config.DisbursementCutOffTimeWindow) (*disbursementModel.CutOffTimeStatusResponse, error)
	GetPayoutStatusAndRouting(ctx context.Context, request *disbursementModel.CRMSinglePayoutStatusRequest) (*disbursementModel.CRMPayoutStatusResponse, error)
	GetBatchPayoutStatusAndRouting(ctx context.Context, request *disbursementModel.CRMBatchPayoutStatusRequest) (*disbursementModel.CRMBatchPayoutStatusResponse, error)
	IBankTransferConfig
	IsMerchantAllowedToUseBeneficiaryCustomRule(ctx context.Context, merchantId string, isCustomRule bool) bool
	IsMerchantAllowedExcludeBeneficiaryRules(ctx context.Context, merchantId string, amount float64) (maxAmount float64, isValid bool)

	// alert
	ProcessPayoutAlert(ctx context.Context, request *disbursementModel.PayoutTransactionAlertRequest) error

	// Callback
	ResendDisbursementCallback(ctx context.Context, request *callbackModel.ResendCallbackRequest) error
}

type IBankTransferConfig interface {
	IsBankcodeOverbookingChannelAllowed(ctx context.Context, bankcode, merchantId string) bool
}

type IDisbursementInternalService interface {
	ValidateDailyTransactionLimit(ctx context.Context, merchantId string, totalAmount float64) (ITransactionCloser, error)
	DecrDailyTransactionLimit(ctx context.Context, merchantId string, totalAmount float64) error
	DeleteDailyTransactionLimit(ctx context.Context, merchantId string) error
	GetCutOffTimeStatus(ctx context.Context, now time.Time, merchantId string, windowConfig *config.DisbursementCutOffTimeWindow) (*disbursementModel.CutOffTimeStatusResponse, error)

	ValidateBeneficiaryPayoutDefaultRule(
		ctx context.Context, merchantId, bankCode, accountNo, accountName string, amount float64, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig) error
	ValidateBeneficiaryPayoutCustomRule(
		ctx context.Context, merchantId, bankCode, accountNo, accountName string, amount float64, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig,
	) error
	ValidateBeneficiaryPayoutMerchantPolicyRule(
		ctx context.Context, merchantId, bankCode, accountNo, accountName string, amount float64, rule *disbursementModel.BeneficiaryPayoutLimitRuleConfig,
	) error
	DecrBeneficiaryPayoutLimit(ctx context.Context, merchantID, bankCode, accountNo string, totalAmount float64) error

	CreateBankTransfer(ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction) error
	Approve(ctx context.Context, request *disbursementModel.ApproveRequest) error
	CreatePendingOrchestratorTransaction(ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction) (transactionId, feeId string, err error)
	ValidateBankAccountAndUpdateTransaction(ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction, orchestratorTransaction *orchestratorModel.TransactionAndFeeObject) error
	ProcessBankTransferAndUpdateTransaction(ctx context.Context, disbursement *disbursementModel.DisbursementWithTransaction, orchestratorTransaction *orchestratorModel.TransactionAndFeeObject) error

	ExternalFDS(ctx context.Context, payout *disbursementModel.DisbursementWithTransaction, ledger *orchestratorModel.TransactionAndFeeObject) error
	FailTransactionByFDSResult(ctx context.Context, payoutID string, ledger *orchestratorModel.TransactionAndFeeObject) error
}

type ITransactionCloser interface {
	Close(ctx context.Context, status bool) error
}

type IAccountService interface {
	GetAccount(ctx context.Context, accountId uuid.UUID) (*accountModel.Account, error)
	GetMerchantAccounts(ctx context.Context, merchantIDs []uuid.UUID, usecase string) (map[uuid.UUID]*accountModel.Account, error)
	GetAccountByReferenceIDAndUsecase(ctx context.Context, referenceID uuid.UUID, usecase string, userType string) (*accountModel.Account, error)
	GetWalletMerchantAccount(ctx context.Context, parentMerchantId, merchantId uuid.UUID) (*accountModel.Account, error)
	GetWalletCustomerAccount(ctx context.Context, req *accountModel.GetCustomerAccountRequest) (*accountModel.Account, error)
	CalculateAccountEodBalance(ctx context.Context) error
	CalculateDailyAccountTransaction(ctx context.Context, location *time.Location) error
	CreateMerchantAccount(ctx context.Context, merchantId, userType string) error
	CreateAccount(ctx context.Context, request *accountModel.NewAccountRequest) (*accountModel.AccountResponse, error)
	BulkCreateAccount(ctx context.Context, request *accountModel.BulkCreateAccountRequest) error
}

type IOTP interface {
	SendGenerateOTPCode(ctx context.Context, request *otpModel.GenerateOTPCodeRequest) (string, error)
	ValidateOTPCode(ctx context.Context, data *otpModel.VerifyOTP) (token string, err error)
	ValidateTOTPCode(ctx context.Context, request *otpModel.VerifyTOTPRequest) (bool, error)

	IOTPGenerator
}

type IOTPGenerator interface {
	GenerateOTPCode(ctx context.Context, id, email string, feature constant.OTPIdentifier) (token string, err error)
	GenerateTOTPVerifyToken(ctx context.Context, request otpModel.GenerateTOTPVerifyTokenRequest) (token string, err error)
}

type IAdjustmentService interface {
	CreateManualTopup(ctx context.Context, req *adjustModel.ManualTopupRequest) (id string, err error)
	CreateBalanceAdjustmentFromManualTopUp(ctx context.Context, req *adjustModel.BalanceAdjustmentRequest) (id string, err error)
	CreateMerchantBalanceAdjustment(ctx context.Context, req *adjustModel.MerchantBalanceAdjustmentRequest) (*adjustModel.ManualAdjustmentHistory, error)
	HoldMerchantBalance(ctx context.Context, req *adjustModel.HoldMerchantBalanceRequest) (*adjustModel.HoldMerchantBalanceResponse, error)
	ReleaseHoldedMerchantBalance(ctx context.Context, req *adjustModel.HoldMerchantBalanceRequest) (*adjustModel.HoldMerchantBalanceResponse, error)
	GetHoldedMerchantBalance(ctx context.Context, req *adjustModel.GetHoldedMerchantBalanceRequest) (*adjustModel.GetHoldedMerchantBalanceResponse, error)
}

type IRateLimiter interface {
	RateLimitFailedAttempt(ctx context.Context, req *rateLimiterModel.RateLimit) error
	ValidateMerchantRateLimit(ctx context.Context, req rateLimiterModel.MerchantRateLimitRequest) (*rateLimiterModel.MerchantRateLimitHeaderMetadata, error)
	CacheMerchantRateLimitConfig(ctx context.Context, merchantID string) (*[]rateLimiterModel.MerchantRateLimitConfig, error)
	List(ctx context.Context, req *rateLimiterModel.MerchantRateLimitRequest) (*commonModel.PaginationResponse, error)
	Detail(ctx context.Context, merchantId, uuid string) (*rateLimiterModel.RateLimitConfiguration, error)
	Create(ctx context.Context, req *rateLimiterModel.CreateRateLimitConfiguration) (*rateLimiterModel.RateLimitConfiguration, error)
	Update(ctx context.Context, req *rateLimiterModel.UpdateRateLimitConfiguration) (*rateLimiterModel.RateLimitConfiguration, error)
}

type ICredentialService interface {
	Get(ctx context.Context, request *credModel.CredentialDashboardReq) (*credModel.CredentialDashboardResp, error)
	ClientSecretById(ctx context.Context, request *credModel.ClientSecretReq) (*credModel.ClientSecretResp, error)
}

type IAccountInquiryService interface {
	RequestAccountInquiry(ctx context.Context, req requestAccountInquiries.RequestAccountInquiriesHttpRequest) (*requestAccountInquiries.RequestAccountInquiriesHttpResponse, error)
	CheckStatusRequestInquiry(ctx context.Context, merchantID, inquiryID string) (*requestAccountInquiries.RequestAccountInquiriesHttpResponse, error)
	FindLatestByInquiryID(ctx context.Context, inquiryID, merchantID string) (*requestAccountInquiries.RequestAccountInquiryWithMaster, error)
}

type ICreditCardService interface {
	CreatePayment(
		ctx context.Context,
		request creditcardModel.CreateCardPaymentRequest,
	) (*creditcardModel.CreateCardPaymentResponse, error)
	PaymentNotification(
		ctx context.Context,
		request creditcardModel.CardPaymentNotificationRequest,
	) error
	// TODO: find better alternative
	PaymentNotificationFDS(
		ctx context.Context,
		request creditcardModel.CardPaymentNotificationRequest,
	) error
	GetPaymentById(
		ctx context.Context,
		merchantID, uuid string,
	) (*paymentModel.Payment, error)
	Void(
		ctx context.Context,
		request *creditcardModel.VoidRequest,
	) (*creditcardModel.VoidResponse, error)
	GetTransactionList(
		ctx context.Context,
		request *creditcardModel.GetTransactionListRequest,
	) (*commonModel.PaginationResponse, error)
	CreateEncryptedCardAuthenticationLink(ctx context.Context, request *creditcardModel.EncryptedCardAuthenticationRequest) (*creditcardModel.EncryptedCardAuthenticationResponse, error)
	GetMIDList(ctx context.Context, request *creditcardModel.GetMIDListRequest) (*commonModel.PaginationResponse, error)
	GetMIDMapList(ctx context.Context, limit, page int, merchantId string) (*commonModel.PaginationResponse, error)
	GetMIDDetail(ctx context.Context, midId string) (*creditcardCoreProcessorModel.MIDResponseData, error)
	CreateMID(ctx context.Context, request *creditcardModel.CreateMIDRequest) error
	UpdateMID(ctx context.Context, request *creditcardModel.UpdateMIDRequest) error
	InquiryTransaction(ctx context.Context, payload *creditcardModel.InquiryTransactionRequest) (*creditcardModel.PaymentNotificationDataRequest, error)
	GetStoredCardByCustomerID(
		ctx context.Context,
		merchantID, customerID string,
	) ([]*unifiedPaymentModel.CustomerPaymentMethodResponse, error)
	BlockCard(
		ctx context.Context,
		request *creditcardModel.BlockCardRequest,
	) error
	RemoveCardTokenization(ctx context.Context, request unifiedPaymentModel.RemoveCardTokenizationRequest) error
	ValidateMIDInstallmentBins(ctx context.Context, request *creditcardModel.ValidateMIDInstallmentBinsRequest) error
	GetCardEncryptionPublicKey(ctx context.Context, merchantID string) ([]byte, error)
	Authentication(ctx context.Context, request creditcardCoreProcessorModel.AuthenticationRequest) (*creditcardCoreProcessorModel.AuthenticationResponse, error)
}

type IMerchantForbiddenUseCaseService interface {
	BlockUseCase(ctx context.Context, request *merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest) error
	UnblockUseCase(ctx context.Context, request *merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest) error
	CheckUseCase(ctx context.Context, merchantId, useCase string) error
}

type ILedgerService interface {
	RecordTransaction(ctx context.Context, merchantId string, request *ledger_model.CreateNewLedgerEntryRequest) error
	GetLedgerBalance(ctx context.Context, accountId uuid.UUID) (*ledger_model.LedgerBalance, error)
	CalculateBulkLedgerBalance(ctx context.Context, request *accountModel.CalculateBulkLedgerBalanceRequest) (*ledger_model.LedgerBalance, error)
	UpdateTransaction(ctx context.Context, request *ledger_model.UpdateLedgerEntryRequest) error
	BulkUpdateLedgerEntry(ctx context.Context, bulkRequest *ledger_model.BulkUpdateLedgerEntryRequest) error
	GetLedgerTransactions(ctx context.Context, request *ledger_model.GetLedgerTransactionRequest, pagination *commonModel.Meta) (*commonModel.PaginationResponse, error)
	GetLedgerDetail(ctx context.Context, referenceId string) ([]orchestratorModel.AccountTransaction, error)
	ILedgerValidatorService
}

type ILedgerValidatorService interface {
	ValidateTransaction(ctx context.Context, merchantId string, request *ledger_model.CreateNewLedgerEntryRequest) error
}

type ILedgerMoneyFlowService interface {
	CreateTransactions(ctx context.Context, request *ledger_model.CreateNewLedgerEntryRequest) error
}

type IAddrLocationService interface {
	Get(ctx context.Context, req *location.LocationReq) (*location.LocationResp, error)
}

type IIndustryService interface {
	GetAllIndustries(ctx context.Context, request *industryModel.SearchIndustryRequest) ([]*industryModel.Industry, error)
	GetUniqueParentIndustries(ctx context.Context) ([]string, error)
	GetChildIndustries(ctx context.Context, parentIndustry string) ([]string, error)
	GetMCCForIndustry(ctx context.Context, parentIndustry, childIndustry string) (string, error)
	IsValidMCC(ctx context.Context, mcc string) (bool, error)
	ValidateIndustry(ctx context.Context, parentIndustry, childIndustry string) (bool, error)
	ValidateIndustryMCCCombination(ctx context.Context, parentIndustry, childIndustry, mcc string) error
	GetIndustryByID(ctx context.Context, id string) (*industryModel.Industry, error)
	CreateIndustry(ctx context.Context, req industryModel.CreateIndustryRequest) (*industryModel.Industry, error)
	UpdateIndustry(ctx context.Context, req industryModel.UpdateIndustryRequest) (*industryModel.Industry, error)
	DeleteIndustry(ctx context.Context, uuid string) error
}

type IQrisService interface {
	Registration(ctx context.Context, request *qris.RegistrationReq) (id string, err error)
	RegistrationCallback(ctx context.Context, request *qris.RegistrationCallback) error
	ReuploadDocument(ctx context.Context, request *qris.ReuploadDocumentReq) (resp *qris.ReuploadDocumentResp, err error)

	RegistrationList(ctx context.Context, merchantId string) ([]qris.RegistrationListResp, error)
	FindQrRegistrationByExternalID(ctx context.Context, externalID string) (*qris.Registration, error)
	FindQrRegistrationByExternalIDAndAcquirer(ctx context.Context, externalID string, acquirer string) (*qris.Registration, error)
	DuplicateRegistration(ctx context.Context, request *qris.DuplicateRegistrationReq) (id string, err error)

	CreateManualRegistration(ctx context.Context, request *qris.RegistrationReq) (id string, err error)
	UpdateQrRegistration(ctx context.Context, id string, acquirerMerchantId string, acquirerTerminalId string) error
}

type ICustomerService interface {
	CreateCustomer(ctx context.Context, request customerModel.CreateCustomerRequest) (*customerModel.GeneralCustomerResponse, error)
	CreateUnfiedPaymentCustomer(ctx context.Context, request customerModel.CreateUnifiedPaymentCustomerRequest) (*customerModel.GeneralCustomerResponse, error)
	GetMerchantCustomersByID(ctx context.Context, merchantId string, customerIds []string) ([]*customerModel.Customer, error)
	GetCustomerById(ctx context.Context, id, merchantId string) (*customerModel.GeneralCustomerResponse, error)
	GetCustomerByPhoneNumber(ctx context.Context, phoneNumber, merchantId string) (*customerModel.GeneralCustomerResponse, error)
	GetCustomerList(ctx context.Context, merchantId, phoneNumber string, page, perPage int64) (*commonModel.PaginationResponse, error)
	DeleteCustomer(ctx context.Context, id, merchantId string) (*customerModel.GeneralCustomerResponse, error)
	UpdateCustomer(ctx context.Context, request customerModel.UpdateCustomerRequest) (*customerModel.GeneralCustomerResponse, error)
	FindCustomerByID(ctx context.Context, id string) (*customerModel.GeneralCustomerResponse, error)
	GetCustomerByIDForUnifiedPayment(ctx context.Context, id, merchantId string) (*unifiedPaymentModel.CustomerInformationResponse, error)
	GetCardFundedPayoutSavedCardList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetSavedCardList) (*commonModel.PaginationResponse, error)
	GetCardFundedPayoutSavedCardDetail(ctx context.Context, request cardFundedPayoutModel.GetSavedCardDetailRequest) (*cardFundedPayoutModel.GetSavedCardResponse, error)
}

type IFeeService interface {
	GetFeeCalculationAndDetail(ctx context.Context, request *feeModel.GetFeeRequest) (fee float64, detail *feeModel.FeeMetadataObject, err error)
	GetTransactionFeeOnBehalf(ctx context.Context, request *feeModel.GetTrxFeeOnBehalfRequest) (*feeModel.TrxFeeOnBehalfMetadata, error)
	GetPayoutTransactionFee(ctx context.Context, request feeModel.GetPayoutTrxFeeRequest) (feeModel.FeeResponseder, error)
	IncrementLadderCounter(ctx context.Context, redisKey string, increment int64)
	GetXbFeeConfigs(ctx context.Context, merchantID string) (*merchantModel.XbFeeConfigResponse, error)

	IFeeCalculator

	// Cron Functions
	PlatformActivitiesFee(ctx context.Context, date time.Time) error
	DeductBalanceForIndirectFeeType(ctx context.Context, date time.Time) error
	DetermineFeeTierLvlFromMonthlyTPV(ctx context.Context, date time.Time) error
}

type IFeeCalculator interface {
	CalculateFee(ctx context.Context, request *feeModel.GetFeeRequest, feeDetail *feeModel.FeeMetadataObject) (fee float64, tax float64)
}

type IXbPayoutService interface {
	GetFxRate(ctx context.Context, request *xbModel.GetFxRateRequest) (*xbModel.GetFxRateResponse, error)
	CreateSession(ctx context.Context, request *xbModel.CreatePayoutSessionRequest) (*xbModel.CreatePayoutSessionResponse, error)
	UploadUnderlyingDocument(ctx context.Context, request *xbModel.UploadUnderlyingDocumentRequest) (*xbModel.UploadUnderlyingDocumentResponse, error)
	Confirm(ctx context.Context, request *xbModel.ConfirmPayoutRequest) (*xbModel.ConfirmPayoutResponse, error)
	ReConfirm(ctx context.Context, request *xbModel.ConfirmPayoutRequest) (*xbModel.ReConfirmEvent, error)

	// Beneficiary
	CreateBeneficiary(ctx context.Context, request *xbModel.CreateBeneficiaryRequest) (*xbModel.CreateBeneficiaryResponse, error)
	GetListBeneficiary(ctx context.Context, request *xbModel.GetListBeneficiaryRequest) (*xbModel.PaginationResponse, error)
	GetBeneficiaryById(ctx context.Context, request *xbModel.GetBeneficiaryByIdRequest) (*xbModel.CreateBeneficiaryResponse, error)
	UpdateBeneficiary(ctx context.Context, request *xbModel.UpdateBeneficiaryRequest) (*xbModel.CreateBeneficiaryResponse, error)
	DeactivateBeneficiary(ctx context.Context, request *xbModel.GetBeneficiaryByIdRequest) (*xbModel.CreateBeneficiaryResponse, error)

	// Sender
	CreateSender(ctx context.Context, request *xbModel.CreateSenderRequest) (*xbModel.CreateSenderResponse, error)
	GetListSender(ctx context.Context, request *xbModel.GetListSenderRequest) (*xbModel.PaginationResponse, error)
	GetSenderById(ctx context.Context, request *xbModel.GetSenderByIdRequest) (*xbModel.CreateSenderResponse, error)
	UpdateSender(ctx context.Context, request *xbModel.UpdateSenderRequest) (*xbModel.CreateSenderResponse, error)
	DeactivateSender(ctx context.Context, request *xbModel.GetSenderByIdRequest) (*xbModel.CreateSenderResponse, error)

	// Payout
	GetList(ctx context.Context, filter *xbModel.GetPayoutFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetPayoutById(ctx context.Context, request *xbModel.GetPayoutRequest) (*xbModel.GetPayoutResponse, error)
	GetRfiDetails(ctx context.Context, request *xbModel.GetRfiDetailsRequest) (*xbModel.GetRfiDetailsResponse, error)
	SubmitRfiDetails(ctx context.Context, request *xbModel.SubmitRfiDetailsRequest) (*xbModel.SubmitRfiDetailsResponse, error)
	UpdateStatusFromProcessor(ctx context.Context, request *xbModel.ConsumePayoutStatusChangeRequest) error
	ExportToExcel(ctx context.Context, request *xbModel.ExportXbPayoutRequest) (*xbModel.ExportXbPayoutResponse, error)

	// Master
	GetListMasterCountry(ctx context.Context, request *xbModel.GetListMasterCountryRequest) (*xbModel.PaginationResponse, error)
	GetListMasterCurrency(ctx context.Context, request *xbModel.GetListMasterCurrencyRequest) (*xbModel.PaginationResponse, error)
	GetListMasterCurrencyMapping(ctx context.Context, request *xbModel.GetListMasterCurrencyMappingRequest) (*xbModel.PaginationResponse, error)
	GetListMasterIdentificationType(ctx context.Context, request *xbModel.GetListMasterIdentificationTypeRequest) (*xbModel.PaginationResponse, error)
	GetListMasterAccountType(ctx context.Context, request *xbModel.GetListMasterAccountTypeRequest) (*xbModel.PaginationResponse, error)
	GetListMasterPurpose(ctx context.Context, request *xbModel.GetListMasterPurposeRequest) (*xbModel.PaginationResponse, error)
	GetListMasterState(ctx context.Context, request *xbModel.GetListMasterStateRequest) (*xbModel.PaginationResponse, error)
	GetListMasterCity(ctx context.Context, request *xbModel.GetListMasterCityRequest) (*xbModel.PaginationResponse, error)
	GetListMasterTransferMethod(ctx context.Context, request *xbModel.GetListMasterTransferMethodRequest) (*xbModel.PaginationResponse, error)
	GetListMasterSourceOfIncome(ctx context.Context, request *xbModel.GetListMasterSourceOfIncomeRequest) (*xbModel.PaginationResponse, error)

	// Config
	GetListConfigSpread(ctx context.Context, request *xbModel.GetListConfigSpreadRequest) (*xbModel.PaginationResponse, error)
	GetConfigSpreadDetailByID(ctx context.Context, id string) (*xbModel.GetConfigSpreadDetailResponse, error)
	CreateConfigSpread(ctx context.Context, request *xbModel.CreateConfigSpreadRequest) (*xbModel.CreateConfigSpreadResponse, error)
	UpdateConfigSpread(ctx context.Context, request *xbModel.UpdateConfigSpreadRequest) (*xbModel.UpdateConfigSpreadResponse, error)

	// Insights
	GetXbPayoutDashboardInsights(ctx context.Context, request disbursementModel.GetXbPayoutDashboardInsightRequest) (*disbursementModel.XbPayoutDashboardInsights, error)

	// Fee Config
	GetFeeConfig(ctx context.Context, merchantID string) (*merchantModel.XbFeeConfigResponse, error)
}

type ISettlementService interface {
	ProcessSettlement(ctx context.Context, request *settlementModel.ProcessSettlementRequest) error
	ProcessSettlementTransactionFee(ctx context.Context, merchantId, transactionFeeId string) error
	ProcessSettlementHoldOrRelease(ctx context.Context, request *settlementModel.ProcessHoldReleaseSettlementRequest) error
}

type ITransferService interface {
	Transfer(ctx context.Context, request *transfer.TransferRequest) (*transfer.Transfer, error)
	ReverseTransfer(ctx context.Context, request *transfer.ReverseTransferRequest) (*transfer.Transfer, error)
	UpdateTransferStatus(ctx context.Context, merchantId, transferId, status string, description *string) error

	GetById(ctx context.Context, id, merchantId string) (*transfer.Transfer, error)
	GetList(ctx context.Context, req *transfer.GetTransferListRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetTransferTransaction(ctx context.Context, req transfer.GetTransferTransactionRequest) (*transfer.TransferTransactionDetail, error)
}

type IBankAccountService interface {
	Create(ctx context.Context, request *bankAccountModel.CreateBankAccountRequest) (*bankAccountModel.BankAccountResponse, error)
	Update(ctx context.Context, request *bankAccountModel.UpdateBankAccountRequest) (*bankAccountModel.BankAccountResponse, error)
	GetByMerchantID(ctx context.Context, request *bankAccountModel.GetMerchantBankAccountRequest) (*bankAccountModel.BankAccount, error)
}

type IWithdrawalService interface {
	Preparation(ctx context.Context, request *withdrawal.PreparationRequest) (*withdrawal.PreparationResponse, error)
	Create(ctx context.Context, request *withdrawal.WithdrawalRequest) (*withdrawal.WithdrawalProcessResponse, error)
	InquiryTransaction(ctx context.Context, request *withdrawal.InquiryTransactionRequest) (*withdrawal.InquiryTransactionResponse, error)
	RetryTransaction(ctx context.Context, request *withdrawal.RetryTransactionRequest) error
	TransferBalance(ctx context.Context, request *withdrawal.WithdrawalTransferBalanceRequest) (*withdrawal.WithdrawalTransferBalanceResponse, error)
	UpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error
	ChangeStatusWithdrawal(ctx context.Context, request *withdrawal.WithdrawalChangeStatusRequest) (*withdrawal.WithdrawalChangeStatusResponse, error)

	// CronJob
	TriggeringAutoWithdrawalProcess(ctx context.Context) (messages int64, err error)
	ForceAutoWithdrawalProcess(ctx context.Context, date time.Time) (*merchantModel.ForceAutoWithdrawalProcessResponse, error)

	GetList(ctx context.Context, request *withdrawal.WithdrawalHistoryRequest) (*commonModel.PaginationResponse, error)
	GetById(ctx context.Context, request *withdrawal.WithdrawalDetailRequest) (*withdrawal.WithdrawalDetailResponse, error)
	GetTodayWithdrawalInsight(ctx context.Context, opt withdrawal.WithdrawalInsightRequest) (*withdrawal.WithdrawalInsightResponse, error)
	Export(ctx context.Context, request *withdrawal.WithdrawalListRequest) (*withdrawal.WithdrawalDownloadResponse, error)
}

type IWithdrawalExporterService interface {
	GetCacheDownloadHistory(ctx context.Context, hashFilterKey string) (url string, err error)
	GetList(ctx context.Context, request *withdrawal.WithdrawalHistoryRequest) (list *commonModel.PaginationResponse, err error)
	ExportToExcel(ctx context.Context, req *withdrawal.WithdrawalListRequest, transactions []withdrawal.WithdrawalHistoryResponse, wr io.Writer) error
}

type IPlatformFeeService interface {
	ApplyMerchantTransferFee(ctx context.Context, req platformFee.PlatformFeeRequest) error
	ApplyMerchantTransactionFee(ctx context.Context, req platformFee.PlatformFeeRequest) error
	ReverseMerchantFee(ctx context.Context, req platformFee.PlatformReversalFeeRequest) error
}

type IProductService interface {
	GetProductList(ctx context.Context) ([]*product.Product, error)
	UpdateProductAvailability(ctx context.Context, request *product.UpdateProductRequest) error
	AddMerchantSelectedProduct(ctx context.Context, request *product.AddMerchantProductRequest) error
	GetMerchantSelectedProducts(ctx context.Context, merchantId string) ([]*product.MerchantWithProductName, error)
	GetMerchantActiveProducts(ctx context.Context, merchantID string) ([]*product.MerchantWithProductName, error)
	UpdateMerchantProductAvailability(ctx context.Context, request *product.UpdateMerchantSelectedProductAvailabilityRequest) error
	ValidateMerchantProductAvailability(ctx context.Context, request *product.ValidateMerchantProductAvailability) error
}

type ICommService interface {
	PostEmailService(ctx context.Context, from string, data *paperCommunication.Email) error
}

type ILiveFeaturesService interface {
	GetList(ctx context.Context) ([]liveFeature.LiveFeature, error)
	GetAppVersion(ctx context.Context) (liveFeature.AppVersion, error)
	PollForChanges(ctx context.Context, interval time.Duration, config *config.Config)
}

type IPlatformService interface {
	GetMerchantTransactionList(ctx context.Context, request *platform.TransactionRequest) (*commonModel.PaginationResponse, error)
	GetSubMerchantUserList(ctx context.Context, request *platform.GetSubMerchantUsersRequest) (*commonModel.PaginationResponse, error)
	GetSubMerchantBalances(ctx context.Context, request *platform.GetBulkBalanceRequest) (*commonModel.PaginationResponse, error)
}

type IRoutingProcessorService interface {
	// BankTransfer Usecase
	AccountInquiry(ctx context.Context, request *routingProcessorModelInquiry.InquiryAccountRequest) (*routingProcessorModelInquiry.InquiryAccountResponseData, error)
	BankTransfer(
		ctx context.Context,
		request *routingProcessorModel.BankTransferRequest,
		routeConfigs ...processorPriorityModel.ProcessorPriority,
	) (*routingProcessorModel.BankTransferResponseData, error)
	GetProcessorList(ctx context.Context, merchantID string) []processorPriorityModel.ProcessorPriority
	GetTransferByID(ctx context.Context, payload *orchestratorModel.AccountTransactionWithUseCase, forceFailed bool) (*routingProcessorModel.BankTransferResponseData, error)

	ProcessAccountInquiryCallback(ctx context.Context, payload *routingProcessorModelInquiry.InquiryAccountResponseData) error
	AddressingReplyToAccountInquiry(ctx context.Context, payload *routingProcessorModelInquiry.InquiryAccountResponseData) error

	// flip processor
	GetFlipEscrowBalance(ctx context.Context, processorReference string) (res *routingProcessorModelEscrowBalance.EscrowBalanceResponse, err error)

	GetDanaEscrowBalance(ctx context.Context) (res *routingProcessorModelEscrowBalance.EscrowBalanceResponse, err error)
}

type IReconciliationService interface {
	// UploadFile uploads reconciliation file to gcs and validate header file
	UploadFile(ctx context.Context, transactionType, createdBy string, fileRecon io.Reader, fileHeader *multipart.FileHeader) (*string, error)
	ListRecon(ctx context.Context, request *reconciliationModel.ReconciliationFilterRequest) (*commonModel.PaginationResponse, error)
	ProcessFile(ctx context.Context, id string) error
	UpdateReconDetail(ctx context.Context, id string, payload *reconciliationModel.ReconDetail) error
	DownloadResult(ctx context.Context, id string) (string, error)
	ProcessPayoutRecon(ctx context.Context, req *reconciliationModel.ReconciliationPayout) error
}

type IIPWhitelistService interface {
	Create(ctx context.Context, req *ipwhitelistModel.CreateIPWhitelistConfiguration) (*ipwhitelistModel.IPWhitelistConfiguration, error)
	Update(ctx context.Context, req *ipwhitelistModel.UpdateIPWhitelistConfiguration) (*ipwhitelistModel.IPWhitelistConfiguration, error)
	List(ctx context.Context, req *ipwhitelistModel.GetIPWhitelistConfiguration) (*commonModel.PaginationResponse, error)
	Detail(ctx context.Context, merchantId, uuid string) (*ipwhitelistModel.IPWhitelistConfiguration, error)
	Delete(ctx context.Context, merchantId, uuid string) error
	ValidateIP(ctx context.Context, merchantId, ip string) error
}

type IWalletInsightService interface {
	TotalBalance(ctx context.Context, merchantId string, hardRefresh bool) (*walletInsights.MerchantTotalBalance, error)
}

type IWalletTransactionService interface {
	GetMerchantTransactionHistoryList(ctx context.Context, req walletTransactionModel.MerchantTransactionHistoryListReq) (*commonModel.PaginationResponse, error)
	GetMerchantTransactionDetail(ctx context.Context, merchantId, id string) (*walletTransactionModel.MerchantTransactionDetailResp, error)
	ExportMerchantTransactionHistoryList(ctx context.Context, req walletTransactionModel.MerchantTransactionHistoryListReq) (*commonModel.ExportResponse, error)
}

type IWalletTransactionInternalService interface {
	ExportExcelMerchantTransactionHistoryList(
		ctx context.Context,
		request walletTransactionModel.MerchantTransactionHistoryListReq,
		transactions []walletTransactionModel.MerchantTransactionHistoryListResp,
	) (*bytes.Buffer, error)
}

type IMerchantRcnService interface {
	FindByIDAndMerchantID(ctx context.Context, id string, merchantId string) (*merchantRcn.MerchantRcnResponse, error)
	GetRcnDetail(ctx context.Context, id string, merchantId string) (*merchantRcn.MerchantRcnDetail, error)
}

type IFdsService interface {
	CheckTransaction(ctx context.Context, id string, request *fdscommon.CheckTransactionRequest) (*fdscommon.CheckTransactionResponse, error)
	UpdateTransaction(ctx context.Context, id string, request *fdscommon.UpdateRequest) (*[]fdscommon.UpdateResponse, error)
}

type IFraudRuleService interface {
	Create(ctx context.Context, request *fraudrulesmodel.CreateFraudRuleRequest) (*fraudrulesmodel.FraudRulesResponse, error)
	Update(ctx context.Context, request *fraudrulesmodel.UpdateFraudRuleRequest) (*fraudrulesmodel.FraudRulesResponse, error)
	Delete(ctx context.Context, uuid string) error
	Detail(ctx context.Context, uuid string) (*fraudrulesmodel.FraudRules, error)
	List(ctx context.Context, req *fraudrulesmodel.FraudRulesQuery) (*commonModel.PaginationResponse, error)
}

type IAmlService interface {
	Screening(ctx context.Context, request *amlcommon.CheckRequest, provider string, merchantID string) (*amlcommon.ScreeningResponse, error)
	Profile(ctx context.Context, request *amlcommon.CheckRequest, provider string, merchantID string, profileID string) (*amlcommon.ProfileDetailResponse, error)
	UpdateDetailStatusByProfileId(ctx context.Context, profileID string, merchantID string, request *amlcommon.UpdateDetailStatusRequest) error
}

type ICountryService interface {
	GetAll(ctx context.Context, filter *countryModel.SearchFilterRequest) ([]*countryModel.Country, error)
	FindByCode(ctx context.Context, code string) (*countryModel.Country, error)
}

type IDukcapilService interface {
	VerifyIdentity(ctx context.Context, request *dukcapilmodel.IdentityVerificationRequest) (*dukcapilmodel.IdentityVerificationResponse, error)
}

type ITablePartitionService interface {
	ReorganizeMonthlyRangePartition(ctx context.Context, request partitionModel.ReorganizeRangePartitionRequest) error
}

type IInstallmentPlanService interface {
	Create(ctx context.Context, request *installmentPlanModel.CreateInstallmentPlanRequest) (*installmentPlanModel.InstallmentPlan, error)
	Update(ctx context.Context, request *installmentPlanModel.UpdateInstallmentPlanRequest) (*installmentPlanModel.InstallmentPlan, error)
	List(ctx context.Context, req *installmentPlanModel.FilterRequest) ([]*installmentPlanModel.InstallmentPlan, int64, error)
}

type IShortLinkService interface {
	Create(ctx context.Context, request *shortLinkModel.CreateShortLink) (*shortLinkModel.ShortLink, error)
	GetByCode(ctx context.Context, code string) (*shortLinkModel.ShortLink, error)
}

type INotificationService interface {
	SendFailedWithdrawalAlert(ctx context.Context, request *withdrawal.FailedWithdrawalAlertRequest) error
	SendVccSettlementTransactionAlert(ctx context.Context, request *vccSettlement.VccTransactionInquiryAlert) error
}

type IVccSettlementService interface {
	RcnTransactionInquiry(ctx context.Context, request *vccSettlement.VccTransactionInquiryRequest) (*vccSettlement.VccTransactionInquiryResponse, error)
	ProcessRcnTransactionInquiry(ctx context.Context, request *vccSettlement.VccTransactionInquiryRequest) error
}

type ISettlementHoldService interface {
	CreateUpdate(ctx context.Context, request *settlementHold.CreateUpdateSettlementHoldRequest) (*settlementHold.CreateUpdateSettlementHoldResponse, error)
	GetSettlementHoldByPaymentID(ctx context.Context, paymentID string) (*settlementHold.SettlementHold, error)
}

type IReportingService interface {
	UpsertBalanceHistory(ctx context.Context, request *reportingModel.UpsertBalanceHistoryRequest) error
	ListBalanceHistory(ctx context.Context, filters *orchestratorModel.TransactionHistoryFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	MigrateBalanceHistoryToDataReporting(ctx context.Context, startDate, endDate time.Time) error
}

type ICardFundedPayoutService interface {
	// Saved Cards Group
	CreateSavedCard(ctx context.Context, request *cardFundedPayoutModel.CreateSavedCardRequest) (*cardFundedPayoutModel.CreateSavedCardResponse, error)
	GetSavedCardList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetSavedCardList) (*commonModel.PaginationResponse, error)
	// Payout Actions
	CreatePayout(ctx context.Context, request cardFundedPayoutModel.CreatePayoutRequest) (*cardFundedPayoutModel.PayoutActionResponse, error)
	ApprovePayout(ctx context.Context, request cardFundedPayoutModel.ApprovePayoutRequest) (*cardFundedPayoutModel.PayoutActionResponse, error)
	RejectPayout(ctx context.Context, request cardFundedPayoutModel.RejectPayoutRequest) (*cardFundedPayoutModel.PayoutActionResponse, error)
	UpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error
	UpdatePayoutTransactionStatus(ctx context.Context, request cardFundedPayoutModel.PatchPayoutTransactionStatusRequest) (*cardFundedPayoutModel.PayoutActionResponse, error)
	// Payout List
	GetPayoutList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetPayoutList) (*commonModel.PaginationResponse, error)
	// Payout Detail
	GetPayoutDetail(ctx context.Context, filter *cardFundedPayoutModel.GetPayoutDetailRequest) (*cardFundedPayoutModel.GetPayoutDetailResponse, error)
	// Export Payout List
	ExportPayoutList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetPayoutList) (*cardFundedPayoutModel.ExportPayoutListResponse, error)
	// Payout Insights
	GetPayoutInsights(ctx context.Context, filter *cardFundedPayoutModel.FilterGetPayoutInsights) (*cardFundedPayoutModel.GetPayoutInsightsResponse, error)
	// Process Payout
	ProcessPendingSubsequentPayments(ctx context.Context, merchantID, referenceID string) error
	// Handle Failed Payout
	ProcessInitialCardFundedPayoutAuthFailure(ctx context.Context, merchantID, referenceID string) error

	// Finish Settlement Action
	ProcessFinishCardFundedPayoutSettlement(ctx context.Context, request *cardFundedPayoutModel.ProcessFinishCardFundedPayoutSettlementRequest) error
	// Receipt
	GetReceipt(ctx context.Context, request *cardFundedPayoutModel.GetReceiptRequest) (*cardFundedPayoutModel.GetReceiptResponse, error)
	// Transaction List
	GetPayoutTransactionList(ctx context.Context, request cardFundedPayoutModel.GetPayoutTransactionListRequest) ([]cardFundedPayoutModel.GetPayoutTransactionListResponse, error)
}

type IVendorService interface {
	Create(ctx context.Context, request *vendorModel.CreateVendorRequest) (*vendorModel.VendorResponse, error)
	Update(ctx context.Context, request *vendorModel.UpdateVendorRequest) (*vendorModel.VendorResponse, error)
	Delete(ctx context.Context, uuid string) error
	Detail(ctx context.Context, uuid string) (*vendorModel.Vendor, error)
	List(ctx context.Context, req *vendorModel.VendorQuery) (*commonModel.PaginationResponse, error)
}

type IPayoutManualProcessingAccountService interface {
	Create(
		ctx context.Context,
		request *payoutManualProcessingAccountModel.CreatePayoutManualProcessingAccountRequest,
	) (*payoutManualProcessingAccountModel.PayoutManualProcessingAccountResponse, error)
	Update(
		ctx context.Context,
		request *payoutManualProcessingAccountModel.UpdatePayoutManualProcessingAccountRequest,
	) (*payoutManualProcessingAccountModel.PayoutManualProcessingAccountResponse, error)
	List(ctx context.Context, req *payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery) (*commonModel.PaginationResponse, error)
}

type ITNCService interface {
	// CRM management
	CreateTNCVersion(ctx context.Context, req *tncModel.CreateTNCVersionRequest) (*tncModel.TNCVersionResponse, error)
	ActivateTNCVersion(ctx context.Context, id string) (*tncModel.TNCVersionResponse, error)
	DeactivateTNCVersion(ctx context.Context, id string) (*tncModel.TNCVersionResponse, error)
	ListTNCVersions(ctx context.Context, q *tncModel.TNCVersionQuery) (*commonModel.PaginationResponse, error)
	GetTNCVersion(ctx context.Context, id string) (*tncModel.TNCVersionResponse, error)

	// Merchant signing
	SignTNC(ctx context.Context, req *tncModel.SignTNCRequest) (*tncModel.MerchantTNCSigningHistoryResponse, error)
	GetTNCSigningStatus(ctx context.Context, merchantID string) (*tncModel.TNCSigningStatus, error)
	GetSigningHistory(ctx context.Context, q *tncModel.SigningHistoryQuery) (*commonModel.PaginationResponse, error)
}
