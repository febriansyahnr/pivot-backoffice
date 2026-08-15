package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/account"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/accountInquiries"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/activity"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/adjustment"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/amlProcessor"
	bankAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/bankAccount"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/beneficiaryAccount"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/callback"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cardFundedPayout"
	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cdc"
	cimbProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cimbProcessor"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	countryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/country"
	credModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/credential"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcardCoreProcessor"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/customer"
	dailyAccountTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/dailyAccountTransaction"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursementDashboard"
	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/dukcapil"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fdsProcessor/fdsCommon"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fraudRules"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/inbound"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/industry"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/installmentPlan"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ipWhitelist"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/liveFeature"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/location"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	merchantForbiddenUseCaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantForbiddenUsecase"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantRcn"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantTopUp"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/outbound"
	paperCommModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paperCommunication"
	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/partition"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/passwordHistories"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentCapture"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentMethod"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payoutManualProcessingAccount"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/permission"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/product"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/qris"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/rateLimiter"
	reconciliationModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/reconciliation"
	recurringContractModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/recurringContract"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/refund"
	reportingModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/reporting"
	requestAccountInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/requestAccountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/role"
	roleMenuPermissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/roleMenuPermission"
	routingProcessorModelInquiry "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/routingProcessor/accountInquiry"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/routingProcessor/bankTransfer"
	routingProcessorModelEscrowBalance "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/routingProcessor/escrowBalance"
	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ruleEvaluations"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/settlementHolds"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/shortLink"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore"
	snapCoreBankAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankAccount"
	snapCoreBankConfigModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankConfig"
	snapCoreBankTransferModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/ewallet"
	snapPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/payment"
	snapCoreQRModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qr"
	snapQrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qris"
	snapCoreTopUpSimulationModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/topUpSimulation"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/virtualAccount"
	statusHistoriesModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/statusHistory"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/tnc"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/transfer"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/user"
	userLoggedInDeviceModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/userLoggedInDevice"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/userRole"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/vccSettlement"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/vendor"
	walletTransactionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/wallet/transaction"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/withdrawal"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/xbCoreProcessor"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

type IBasicSQL interface {
	BeginTransaction(ctx context.Context) (context.Context, error)
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error
}

type IHealthCheckRepository interface {
	CheckDB(ctx context.Context) error
	CheckRedis(ctx context.Context) string
	CheckRabbitMQ(ctx context.Context) (bool, error)
}

type IUserRepository interface {
	ListUsers(ctx context.Context, limit, offset int) ([]*userModel.User, error)
	ListUsersByMerchantID(
		ctx context.Context,
		filter *userModel.ListUsersByMerchantIDRequest,
		page, perPage int64) (*commonModel.PaginationResponse, error)
	FindUserByID(ctx context.Context, id string) (*userModel.User, error)
	FindUserByEmail(ctx context.Context, email string) (*userModel.User, error)
	Create(ctx context.Context, user *userModel.User) error
	Update(ctx context.Context, user *userModel.User) error
	UpdateRefreshToken(ctx context.Context, id, token string) error
	ChangePassword(ctx context.Context, id string, password string) (bool, error)

	UpdatePin(ctx context.Context, userID, hashedPin string) error
	BlockUser(ctx context.Context, id string, blocked sql.NullTime) (err error)

	// Multi Factor Authentication
	FindUserTOTPDataByID(ctx context.Context, userId string) (*userModel.UserTOTPData, error)
	UpdateUserTOTPData(ctx context.Context, request *userModel.UpdateUserTOTPDataRequest) error
	UpdateUserPreferred2FAMethod(ctx context.Context, userId, preferred2FAMethod string) error
}

type IUserLoggedInDeviceRepository interface {
	GetAllByUserID(ctx context.Context, userID string) ([]*userLoggedInDeviceModel.UserLoggedInDevice, error)
	FindByUserAndDevice(ctx context.Context, userID, deviceIdentifier string) (*userLoggedInDeviceModel.UserLoggedInDevice, error)

	Create(ctx context.Context, data *userLoggedInDeviceModel.UserLoggedInDevice) error
	SetRememberDevice(ctx context.Context, userId, deviceIdentifier, data string) error
}

type IMerchantRepository interface {
	IBasicSQL

	FindMerchantByID(ctx context.Context, id string) (*merchant.Merchant, error)
	Create(ctx context.Context, merchant *merchant.Merchant) error
	GenerateNewMID(ctx context.Context) (string, error)
	GetMerchantAuthByMerchantId(ctx context.Context, merchantID string) (*merchant.MerchantAuth, error)
	GetMerchantSnapPKCS8KeyByMerchantID(ctx context.Context, merchantID string) (*merchant.MerchantAuth, error)
	DeterminePaymentFeeByMerchantIdMethodAndChannel(ctx context.Context, request *feeModel.GetFeeRequest) (*merchant.MerchantFee, error)
	DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel(ctx context.Context, merchantId, method, channel, settlementType string) (*merchant.MerchantFee, error)
	DetermineRefundFeeByMerchantIdAndReferenceType(ctx context.Context, merchantId, referenceType string) (*merchant.MerchantFee, error)
	GetListOfMerchantsWhoHaveSubMerchant(ctx context.Context) ([]merchant.MerchantWithSubMerchantList, error)
	GetDepositSetting(ctx context.Context, merchantId string) (*merchant.DepositSettingResponse, error)

	// Merchant Fees
	GetMerchantFeeByID(ctx context.Context, id string) (*merchant.MerchantFee, error)
	GetMerchantFeeByMerchantIDAndType(ctx context.Context, merchantID, feeType string) (*merchant.MerchantFee, error)
	GetMerchantFeeListForBalanceDeduction(ctx context.Context) ([]merchant.MerchantFeeForBalanceDeduction, error)
	DeterminePayoutFeeByMerchantIdAndChannel(ctx context.Context, merchantId, channel, reference string) (*merchant.MerchantFee, error)
	GetMerchantFeeXB(ctx context.Context, q *merchant.MerchantFeeXBQuery) (*merchant.MerchantFee, error)
	DetermineTopupFeeByMerchantIdMethodAndChannel(ctx context.Context, merchantId, method, channel string) (*merchant.MerchantFee, error)

	// Settlement Config
	GetSettlementConfig(ctx context.Context, request merchant.GetSettlementConfigRequest) (*merchant.SettlementConfig, error)

	Update(ctx context.Context, merchant *merchant.Merchant) error
	UpdateCallbackApiKey(ctx context.Context, merchantId, apiKey string, version uint) error
	CreateMerchantAuth(ctx context.Context, merchantAuth *merchant.MerchantAuth) error
	UpdateMerchantAuth(ctx context.Context, merchantAuth *merchant.MerchantAuth) error
	CreateMerchantFee(ctx context.Context, merchantFee *merchant.MerchantFee) error
	UpdateMerchantFee(ctx context.Context, merchantFee *merchant.MerchantFee) error
	GetMerchantFeeByRequest(ctx context.Context, request *merchant.GetMerchantFeeRequest) (*merchant.MerchantFee, error)
	UpdateTransactionConfig(ctx context.Context, merchantId string, config *merchant.TransactionConfigs) error
	UpdatePaymentInvestigationConfig(ctx context.Context, merchantID string, config merchant.PaymentInvestigationConfigRequest) error
	UpdateFDSConfig(ctx context.Context, merchantID string, config merchant.FDSConfigRequest) error
	GetTransactionConfig(ctx context.Context, merchantId string) (*merchant.TransactionConfigResp, error)
	GetFDSConfig(ctx context.Context, merchantID string) (*merchant.GetFDSConfigResponse, error)
	IsInvestigationFlowEnabled(ctx context.Context, merchantID string) (enabled bool, err error)
	UpdateMerchantFeeLastDeductionDate(ctx context.Context, merchantId, reference string, date time.Time) error
	UpdateFeeTieringConfig(ctx context.Context, request *merchant.FeeTieringRequest) error
	AppliedFeeFromTiers(ctx context.Context, feeId string, appliedFee *merchant.FeeTieringConfig) error
	SetAutoWithdrawal(ctx context.Context, request *merchant.AutoWithdrawalSettingRequest) error
	FindStatusByID(ctx context.Context, id string) (*merchant.MerchantStatusResponse, error)
	UpdateStatusByID(ctx context.Context, status, reasonStatus, id string) error
	UpdateThirdPartyScreeningData(ctx context.Context, merchantID string, screeningData types.NullJSONText) error

	FindDocumentIdByType(ctx context.Context, merchantId, docType string) (id string, err error)
	FindDocumentByType(ctx context.Context, merchantId, docType string) (doc *merchant.Document, err error)
	CreateDocument(ctx context.Context, document *merchant.Document) error
	UpdateDocument(ctx context.Context, document *merchant.Document) error
	GetDocuments(ctx context.Context, req *merchant.MerchantDocumentFilterRequest) (*commonModel.PaginationResponse, error)

	ValidateMerchantBODData(ctx context.Context, req *merchant.UpsertBoardOfDirectorReq) (result *merchant.BODValidation, err error)
	UpsertMerchantBOD(ctx context.Context, action string, data *merchant.BoardOfDirector) error
	GetListMerchantBODs(ctx context.Context, merchantId string) (resp []merchant.BoardOfDirectorResp, err error)

	FindMerchantForQrRegistration(ctx context.Context, merchantId, acquirer string) (resp *merchant.QrisMerchant, err error)

	GetSubmerchantsByIDs(ctx context.Context, parentMerchantID string, submerchantIDs []string) ([]*merchant.Merchant, error)
	GetMerchantsByIDs(ctx context.Context, merchantIDs []string) ([]*merchant.Merchant, error)
	ListSubMerchantByParentID(
		ctx context.Context,
		filter *merchant.SubMerchantListFilter,
		page, perPage int64) (*commonModel.PaginationResponse, error)
	GetListMerchantFeeThatUseTiers(ctx context.Context) (map[string][]merchant.MerchantFeeThatUseTier, error)
	GetSubMerchantIdListByParentId(ctx context.Context, parentId string) ([]string, error)
	GetSubMerchantsByParentID(ctx context.Context, parentMerchantID string) ([]*merchant.Merchant, error)

	GetAllActiveMerchantIDs(ctx context.Context) ([]string, error)
	GetListOfMerchantsWithActiveAutoWithdrawalStatus(ctx context.Context) ([]merchant.MerchantWithActiveAutoWithdrawalStatus, error)
	GetListOfMerchantsToForceTheAutoWithdrawalProcess(ctx context.Context) ([]merchant.MerchantWithdrawalDetails, error)

	ValidateCreateFeeConfigOnBehalf(ctx context.Context, request *merchant.CreateFeeConfigOnBehalfRequest) (bool, error)
	CreateFeeConfigOnBehalf(ctx context.Context, data *merchant.OnBehalfFeeConfig) error
	GetFeeConfigOnBehalf(ctx context.Context, request *merchant.GetFeeConfigOnBehalfRequest) ([]merchant.FeeConfigOnBehalfResponse, error)
	UpdateFeeConfigOnBehalf(ctx context.Context, id string, request *merchant.UpdateFeeConfigOnBehalfRequest) error
	GetTransactionFeeOnBehalf(ctx context.Context, merchantId, subMerchantId, reference, paymentMethod, referenceType string) (*merchant.TransactionFeeOnBehalf, error)
	GetDisbursementMerchantConfig(ctx context.Context, merchantId string) (*merchant.DisbursementMerchantConfig, error)

	UpdateKYC(ctx context.Context, payload merchant.UpdateMerchantKYCRequest) error

	// Billing
	GetBillingFees(ctx context.Context, request merchant.BillingFeeRequest) (*merchant.BillingFeeResponse, error)
	PayBillingFees(ctx context.Context, request merchant.PayBillingFeeRequest) error

	// Migration
	MigrateMerchantSecretsToEncryption(ctx context.Context, merchant merchant.MigrateMerchantSecretsToEncryption) error
	ListUnencryptedMerchantSecretsForMigration(ctx context.Context) ([]merchant.UnencryptedMerchantSecretsForMigration, error)
}

type IPasswordHistoriesRepository interface {
	FindByUserID(ctx context.Context, userID string, limit *int) ([]*passwordHistories.PasswordHistories, error)
	FindByPassHashAndUserID(ctx context.Context, userID, passHash string) (*passwordHistories.PasswordHistories, error)
	Insert(ctx context.Context, uuid string, userID string, password string) error
}

type IRoleRepository interface {
	IBasicSQL

	FindRoleByID(ctx context.Context, id string) (*role.Role, error)
	FindRoleBySlug(ctx context.Context, slug string) (*role.Role, error)
	TotalRoleByMerchantID(ctx context.Context, merchantID string) (total uint64, err error)
	CheckAvailableRoleName(ctx context.Context, merchantID, roleName string) (avail bool, err error)
	Create(ctx context.Context, role *role.Role) error
	GetList(
		ctx context.Context,
		filter *role.GetRoleFilterRequest,
		page, perPage int64) (*commonModel.PaginationResponse, error)
	Update(ctx context.Context, role *role.Role) error
	Delete(ctx context.Context, id string) error
}

type IUserRoleRepository interface {
	FindUserRoleByUserID(ctx context.Context, userID string) (*userRole.UserRole, error)
	TotalActiveUsersByRoleID(ctx context.Context, id string) (total uint64, err error)
	Create(ctx context.Context, userRole *userRole.UserRole) error
	UpdateByUserID(ctx context.Context, ur *userRole.UserRole) error
}

type IPermissionRepository interface {
	FindBySlug(ctx context.Context, slug string) (*permissionModel.Permission, error)
	FindByRoleId(ctx context.Context, roleId string) ([]*permissionModel.Permission, error)

	Create(ctx context.Context, permission *permissionModel.Permission) error
	Update(ctx context.Context, permission *permissionModel.Permission) error
}

type IMenuRepository interface {
	FindBySlug(ctx context.Context, slug string) (*menuModel.Menu, error)
	FindBySlugWithPermissions(ctx context.Context, slug string) (*menuModel.MenuResponse, error)
	GetAll(ctx context.Context, filter *menuModel.GetAllFilterRequest) ([]*menuModel.MenuResponse, error)
	GetMenuAndPermissionIDs(ctx context.Context, slug string, permissionSlug ...string) (res *menuModel.MenuAndPermissionIDs, err error)

	Create(ctx context.Context, menu *menuModel.Menu) error
	Update(ctx context.Context, menu *menuModel.Menu) error
}

type IRoleMenuPermissionRepository interface {
	Create(ctx context.Context, pivot *roleMenuPermissionModel.RoleMenuPermission) error
	Delete(ctx context.Context, roleID string) (err error)
	GetByRoleID(ctx context.Context, roleID string) ([]*roleMenuPermissionModel.RoleMenuPermission, error)
	DeleteByMenuAndPermissions(ctx context.Context, roleID, menuID string, permissionIDs []string) error
}

type IActivityRepository interface {
	Create(ctx context.Context, model *activityModel.Activity) error
	GetList(ctx context.Context, filter activityModel.ActivityFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)

	FindLastMerchantActivityDate(ctx context.Context, merchantID string) (time.Time, error)
}

type IAccountTransactionRepository interface {
	IBasicSQL

	Create(ctx context.Context, accountTransaction *orchestratorModel.AccountTransaction) error
	UpdateStatusAccountTransaction(ctx context.Context, id string, status string, reasonType, reasonDescription *string) error
	UpdateReasonOnly(ctx context.Context, id string, reasonType, reasonDescription *string) error
	UpdateStatusAccountTransactionByReferenceID(ctx context.Context, id string, status string, reasonType, reasonDescription *string) error
	UpdateAdditionalInfoByID(ctx context.Context, id string, additionalInfo types.NullJSONText) error
	UpdateSettlementStatusAndSettlementAtByID(ctx context.Context, id string, settlementStatus string, settlementAt time.Time) error
	CancelIndirectTransactionFee(ctx context.Context, id string, date time.Time) error
	DeductBalanceForIndirectFeeType(ctx context.Context, merchantId string, ids []string) error
	UpdateProcessorAndReconReference(ctx context.Context, id string, processorReferenceName, processorReferenceId, reconReferenceNo string) error
	UpdateTransactionTimestamp(ctx context.Context, id string, transactionTimestamp time.Time) error
	RearrangeUpdatedAtForTransactionWithPendingStatus(ctx context.Context, referenceIds []string, updatedAt time.Time) error
	UpdateTransactionWithPendingStatusByReferenceIdTypeAndChannel(ctx context.Context, referenceId, typ, channel string, data orchestratorModel.UpdateTransactionWithPendingStatus) error
	UpdatePaymentTransactionStatusAndMetadataByID(ctx context.Context, request orchestratorModel.UpdatePaymentTransactionRequest, metadata orchestratorModel.MetadataPayment[any]) error
	UpdateCreditDebitByID(ctx context.Context, id string, credit, debit *float64) error
	UpdateFDSRiskAssessmentResultByID(ctx context.Context, id string, data fdscommon.RiskAssessmentResult) error
	UpdateSettlementDetailByIDs(ctx context.Context, ids []string, request orchestratorModel.UpdateSettlementDetailRequest) error
	UpdateSettlementHoldByReferenceID(ctx context.Context, referenceId string, flag bool) error
	UpdateTransactionDetail(ctx context.Context, request orchestratorModel.UpdateTransactionRequest) error
	GetPastDueSettlementTransactions(ctx context.Context, request *orchestratorModel.GetPastDueSettlementTransactionsRequest) ([]*orchestratorModel.AccountTransaction, error)

	GetAggregateTransactions(ctx context.Context, request *orchestratorModel.GetAggregateRequest) (*orchestratorModel.AggregateResponse, error)
	GetBulkAggregateTransactions(ctx context.Context, request *orchestratorModel.BulkGetAggregateRequest) ([]*orchestratorModel.BulkAggregateResponse, error)
	GetAggregateTransactionByReference(
		ctx context.Context, request *orchestratorModel.GetSummaryTransactionByReferenceRequest) (*orchestratorModel.AggregateResponse, error)
	FindByID(ctx context.Context, id string) (*orchestratorModel.AccountTransactionWithUseCase, error)
	FindByReference(ctx context.Context, referenceID, referenceType string) (*orchestratorModel.AccountTransactionWithUseCase, error)
	BulkInsert(ctx context.Context, accountTransactions []*orchestratorModel.AccountTransaction) error
	UpdateTransactionsStatus(ctx context.Context, request *ledger_model.UpdateLedgerEntryRequest) error
	UpdateTransactionsStatusAndAdditionalInfoByID(ctx context.Context, id string, status string, reasonType string, reasonDescription string, additionalInfo types.NullJSONText) error
	CalculatePendingBalance(ctx context.Context, request *orchestratorModel.GetAggregateRequest) (float64, error)
	GetEarliestUpdatedAt(ctx context.Context, request *orchestratorModel.GetAggregateRequest) (time.Time, error)

	GetLedgerRecords(ctx context.Context, filter *ledger_model.GetLedgerTransactionRequest, pagination *commonModel.Meta) ([]*ledger_model.GetLedgerTransactionData, int, error)
	GetLedgerDetail(ctx context.Context, referenceId string) ([]orchestratorModel.AccountTransaction, error)

	GetList(ctx context.Context, filter *orchestratorModel.TransactionHistoryFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetDetailById(ctx context.Context, merchantId, id string) (*orchestratorModel.TransactionHistoryDetailResp, error)
	GetListTransactionHistories(ctx context.Context, filter *orchestratorModel.TransactionHistoryFilterRequest) (result []orchestratorModel.TransactionHistory, err error)
	GetPlatformTransactionActivities(ctx context.Context, ids []string, startDate, endDate time.Time) ([]orchestratorModel.TransactionActivity, error)
	GetAccumulateTransactionFees(ctx context.Context, merchantId, reference, method string, startDate, endDate time.Time) (*orchestratorModel.AccumulateTransactionFees, error)
	CalculatingMerchantTPVToDetermineFeeTierLevel(ctx context.Context, merchantId string, startDate, endDate time.Time) (map[string]orchestratorModel.CalculatingMerchantTPVSummary, error)
	CalculatingMerchantTPVForLadderTiering(ctx context.Context, merchantId string, startDate, endDate time.Time) (map[string]orchestratorModel.CalculatingMerchantTPVSummary, error)
	CalculatingTPVForPlatformActivitiesToDetermineFeeTierLevel(
		ctx context.Context,
		merchantIDs []string,
		startDate, endDate time.Time,
	) (map[string]orchestratorModel.CalculatingMerchantTPVSummary, error)

	FindLastMerchantTransactionDate(ctx context.Context, merchantID string) (*time.Time, error)
	GetLastTransactionByAccountName(ctx context.Context, merchantId, accountName string) (*time.Time, error)
	GetListOfTransactionReferenceIdsWithPendingStatus(ctx context.Context, merchantId, accountId string, startTime, endTime time.Time) ([]string, error)
	GetTransactionByReferenceIdAndProcessorId(ctx context.Context, referenceId, processorId string) (*orchestratorModel.AccountTransaction, error)
	GetReferenceIdByTransactionIdAndType(ctx context.Context, transactionId, transactionType string) (string, error)

	// recon
	GetTransactionForRecon(ctx context.Context, params *reconciliationModel.ReconTransactionQuery) (*reconciliationModel.ReconTransactionModel, error)
	GetTransactionByProcessorID(ctx context.Context, trxType, processor, processorID string) (*reconciliationModel.ReconTransactionModel, error)
	UpdateBulkReconStatus(ctx context.Context, params *reconciliationModel.BulkUpatedStatus) error
	UpdateReconDetail(ctx context.Context, id string, params *reconciliationModel.ReconDetail) error
	SetAdditionalInfoReconciliation(ctx context.Context, id string, params *reconciliationModel.ReconDetail) error
	// GetTotalPaymentAmount get total amount of payment by processor reference number
	GetTotalPaymentAmount(ctx context.Context, params *reconciliationModel.PaymentTotalAmountQuery) (*reconciliationModel.PaymentTotalAmountResult, error)

	// Wallet
	GetWalletCustomersTotalBalance(ctx context.Context, request *orchestratorModel.GetWalletTotalBalanceRequest) (float64, error)
	BulkUpdateTransactions(ctx context.Context, request []*orchestratorModel.AccountTransaction) error

	// Void
	VoidTransaction(ctx context.Context, request *orchestratorModel.VoidTransactionRequest) error
}

type IAccountRepository interface {
	Create(ctx context.Context, account *account_model.Account) error
	GetByUUID(ctx context.Context, accountId uuid.UUID) (*account_model.Account, error)
	GetByIDs(ctx context.Context, accountId []string) ([]*account_model.Account, error)
	GetEntityAccounts(ctx context.Context, entityIDs []uuid.UUID, userType, name string) (map[uuid.UUID]*account_model.Account, error)
	GetByReferenceIDAndUsecase(ctx context.Context, referenceID uuid.UUID, usecase string, userType string) (*account_model.Account, error)
	FindMerchantAccountByName(ctx context.Context, merchantID uuid.UUID, balanceType string) (*account_model.Account, error)
	FindAll(ctx context.Context) ([]*account_model.Account, error)
	UpdateAccount(ctx context.Context, account *account_model.Account) error
	UpdateHoldedBalance(ctx context.Context, account *account_model.Account) error
	GetMerchantsWithoutAccount(ctx context.Context, request account_model.GetEntityWithoutAccountRequest) ([]*merchant.Merchant, error)
	GetCustomersWithoutAccount(ctx context.Context, request account_model.GetEntityWithoutAccountRequest) ([]*customerModel.Customer, error)
	BulkInsert(ctx context.Context, accounts []*account_model.Account) error
}

type IDailyAccountTransactionRepository interface {
	IBasicSQL

	Upsert(ctx context.Context, dailyAccountTransaction *dailyAccountTransactionModel.DailyAccountTransaction) error
	FindLatestByAccountIDAndTimezone(
		ctx context.Context,
		accountID, timezone string,
	) (*dailyAccountTransactionModel.DailyAccountTransaction, error)
}

type IPaymentRepository interface {
	IBasicSQL

	CreatePayment(ctx context.Context, paymentDTO *paymentModel.PaymentDTO) error
	GetPaymentById(ctx context.Context, id string) (*paymentModel.Payment, error)
	GetPaymentItemsByPaymentId(ctx context.Context, paymentID string) ([]*paymentModel.PaymentItem, error)
	UpdatePaymentItemsFromPaymentResponseItem(ctx context.Context, paymentID string, paymentRespItems []paymentModel.PaymentResponseItem) error

	CreatePaymentItem(ctx context.Context, paymentItemDTO *paymentModel.PaymentItemDTO) error
	BeginTransaction(ctx context.Context) (context.Context, error)
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error
	GetActivePaymentByProcessorReferenceNumber(ctx context.Context, request *paymentModel.GetActivePaymentByProcessorReferenceNumberRequest) (*paymentModel.Payment, error)
	GetPaymentByMerchantAndReferenceId(ctx context.Context, merchantId, referenceId string) (*paymentModel.Payment, error)
	GetPaymentByIdAndMerchantId(
		ctx context.Context, id, merchantId string) (*paymentModel.Payment, error)
	GetPaymentQrStaticByMerchantId(ctx context.Context, merchantId string, subMerchantId string, paymentMethodId string) (*paymentModel.Payment, error)
	// GetPaymentReceiptData fetches payment + merchant + payment_method in single query for receipt generation
	GetPaymentReceiptData(ctx context.Context, paymentID, referenceID, merchantID string) (*paymentModel.PaymentReceiptDTO, error)

	UpdatePaymentStatus(ctx context.Context, id string, merchantId string, status string, updatedAt time.Time) error
	UpdatePayment(ctx context.Context, id string, amount, totalAmount decimal.Decimal, metadata string, customerId string, expiredAt time.Time) error
	UpdatePaymentData(ctx context.Context, payment *paymentModel.PaymentDTO) error
	UpdatePaymentMetadataById(ctx context.Context, id string, metadata paymentModel.UpdatePaymentMetadataRequest) error
	UpdatePaymentForInvestigation(ctx context.Context, request paymentModel.UpdatePaymentForInvestigationRequest) error
	UpdatePaymentStatusWithReason(ctx context.Context, id string, request paymentModel.UpdatePaymentStatusWithReasonRequest) error
	FilterPaymentHistory(ctx context.Context, opt paymentModel.FilterPaymentHistoryOption) (*commonModel.PaginationResponse, error)
	GetInvestigationSummary(ctx context.Context, opt paymentModel.GetInvestigationSummaryOption) (*paymentModel.InvestigationSummaryResponse, error)
	GetTodayPaymentStatusInsight(ctx context.Context, opt paymentModel.PaymentInsightOption) (*paymentModel.PaymentInsightItem, error)
	GetExpiringPayments(ctx context.Context, start time.Time, end time.Time) ([]*paymentModel.ExpiringPayment, error)
	RetrieveImages(ctx context.Context) (paymentModel.ImageResponse, error)
	RetrieveInstructions(ctx context.Context) ([]paymentModel.InstructionResponse, error)
	ChangePaymentMethod(ctx context.Context, payment *paymentModel.PaymentDTO) error
	GetList(ctx context.Context, filter *paymentModel.GetListFilterRequest) (*commonModel.PaginationResponse, error)

	WithConfig(config *config.Config)
	WithSecret(secret *config.Secret)

	GetChargeList(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest) (*commonModel.PaginationResponse, error)
	GetChargeByID(ctx context.Context, id string) (*unifiedPaymentModel.ChargeResponse, error)
	GetCharges(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest) ([]unifiedPaymentModel.ChargeResponse, error)

	CountActiveStaticPayment(ctx context.Context, merchantID, paymentMethodID string) (int, error)

	// Static QRIS Dashboard methods
	FilterStaticQrisList(ctx context.Context, opt paymentModel.StaticQrisFilterRequest) (*commonModel.PaginationResponse, error)
	GetStaticQrisDetail(ctx context.Context, opt paymentModel.StaticQrisDetailRequest) (*paymentModel.StaticQrisDetailResponse, error)
	GetStaticQrisTransactions(ctx context.Context, opt paymentModel.StaticQrisTransactionFilterRequest) (*commonModel.PaginationResponse, error)
	GetFirstActiveStaticQrisByMerchant(ctx context.Context, merchantID, partnerReferenceNo string) (*paymentModel.Payment, error)

	// Static VA Dashboard methods
	FilterStaticVaList(ctx context.Context, opt paymentModel.StaticVaFilterRequest) (*commonModel.PaginationResponse, error)
	GetStaticVaDetail(ctx context.Context, opt paymentModel.StaticVaDetailRequest) (*paymentModel.StaticVaDetailResponse, error)
	GetStaticVaTransactions(ctx context.Context, opt paymentModel.StaticVaTransactionFilterRequest) (*commonModel.PaginationResponse, error)

	// Payment Insight
	GetPaymentDashboardInsights(ctx context.Context, request paymentModel.GetPaymentDashboardInsightRequest) (*paymentModel.PaymentDashboardInsights, error)

	// Investigation flow
	GetInvestigatedPayments(ctx context.Context, filter *paymentModel.GetInvestigatedPaymentsFilterRequest) (*commonModel.PaginationResponse, error)
	UpdateInvestigationStatus(ctx context.Context, request paymentModel.UpdateInvestigationStatusRequest) error
	CalculateInvestigationMonthlyReconciliation(ctx context.Context, request paymentModel.MonthlyReconciliationRequest) ([]paymentModel.CalculateInvestigationMonthlyReconciliation, error)
	InsertInvestigationMonthlyReconciliation(ctx context.Context, data paymentModel.PaymentInvestigationMonthlyReconciliation) error
	UpdatePaymentInvestigationReconciliation(ctx context.Context, data paymentModel.PaymentInvestigationMonthlyReconciliation) error
	// VCC Terminal
	GetVCCTerminalList(ctx context.Context, filter *paymentModel.GetVCCTerminalListFilterRequest) (*commonModel.PaginationResponse, error)
	// Card-funded Payout
	FindPendingSubsequentCardFundedPayout(ctx context.Context, merchantID, referenceID string) ([]cardFundedPayoutModel.CardFundedPayment, error)
	GetCardFundedPayoutFundingSummary(ctx context.Context, merchantID, referenceID string, maxCreatedDays int) (*cardFundedPayoutModel.CardFundedPayoutFundingSummary, error)
	HardDeleteCardFundedPayoutPayments(ctx context.Context, merchantID, referenceID string) error

	GetAutoSplitSubPayments(ctx context.Context, request *paymentModel.GetSubPaymentsRequest) ([]*paymentModel.Payment, error)
	GetSummaryAutoSplitPayment(ctx context.Context, request *paymentModel.GetAutoSplitPaymentSummaryRequest) (*paymentModel.AutoSplitPaymentSummary, error)
	HardDeleteAutoSplitSubPayments(ctx context.Context, merchantID, referenceID string) error
}

type IRefundRepository interface {
	IBasicSQL

	Insert(ctx context.Context, refund *refundModel.Refund) error
	FindByID(ctx context.Context, id string) (*refundModel.Refund, error)
	UpdateData(ctx context.Context, refund *refundModel.Refund) error
	ExistsByClientReferenceAndMerchantID(ctx context.Context, clientReferenceID, merchantID string) (bool, error)
	GetRefundList(ctx context.Context, request refundModel.FilterRefundRequest) (*commonModel.PaginationResponse, error)
	GetRefundByID(ctx context.Context, refundID, merchantID string) (*refundModel.RefundResponse, error)
	FindRefundByChargeID(ctx context.Context, chargeID string) (*refundModel.Refund, error)
	GetTotalRefundedAmount(ctx context.Context, paymentID string) (float64, error)
	ListByPaymentID(ctx context.Context, paymentID string, request refundModel.ListByPaymentIDRequest) (result []refundModel.RefundResponse, err error)
}

type IRecurringContractRepository interface {
	Insert(ctx context.Context, data recurringContractModel.RecurringContract) error
	GetDetailByID(ctx context.Context, merchantID, uuid string) (*recurringContractModel.RecurringContractDetail, error)

	UpdateRecurringContractStatus(ctx context.Context, uuid, status, updatedBy string) error
	UpdateRecurringContract(ctx context.Context, payload recurringContractModel.UpdateRecurringContractRequest) error
}

type IPaymentMethodRepository interface {
	GetPaymentMethodById(ctx context.Context, paymentMethodId string) (*paymentModel.PaymentMethod, error)
	GetActivePaymentMethodByRequest(
		ctx context.Context, request *paymentModel.GetPaymentMethodFilterRequest) (*paymentModel.PaymentMethodWithPivot, error)
	GetAllPaymentMethodByCategory(
		ctx context.Context, category string) ([]*paymentModel.PaymentMethod, error)
	GetPaymentMethodByType(
		ctx context.Context, tipe string) ([]*paymentModel.PaymentMethod, error)

	GetListPaymentMethodMerchant(ctx context.Context, filter *paymentModel.GetPaymentMethodFilterRequest) ([]*paymentModel.PaymentMethodWithPivot, error)
	FindPaymentMethodByIdAndMerchant(ctx context.Context, paymentMethodId, merchantId string) (*paymentModel.PaymentMethodWithPivot, error)
	GetPaymentMethodByCategoryTypeAndAcquirer(ctx context.Context, category, typ, acquirer string) (*paymentModel.PaymentMethod, error)

	UpsertPaymentMethodMerchantByIdAndMerchant(ctx context.Context, paymentMethodMerchant *paymentModel.PaymentMethodWithPivot) error
	CreatePaymentMethod(ctx context.Context, payload *paymentMethodModel.CreatePaymentMethodRequest) error
}

type ICallbackRepository interface {
	IBasicSQL
	CreateCallbackMaster(ctx context.Context, callbackMaster *callbackModel.CallbackMaster) error
	CreateCallback(ctx context.Context, callback *callbackModel.Callback) error
	CreateCallbackLog(ctx context.Context, callbackLog *callbackModel.CallbackLog) error
	UpdateCallbackLog(ctx context.Context, callbackLog *callbackModel.CallbackLog) error
	UpdateCallbackURLById(ctx context.Context, id, url string) error
	UpdateCallbackBaseURLById(ctx context.Context, id, url string) error
	FindCallbackMasterByName(ctx context.Context, name string) (*callbackModel.CallbackMaster, error)
	FindCallbackByNameAndMerchantID(ctx context.Context, name string, merchantId uuid.UUID) (*callbackModel.Callback, error)
	GetList(ctx context.Context, filter *callbackModel.GetListCallbackFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetCallbackURLByMerchantId(ctx context.Context, merchantID string, masterName string) ([]callbackModel.CallbackURLSettingResp, error)
	GetCallbackAPIKeyByMerchantId(ctx context.Context, merchantID string) (*callbackModel.CallbackAPIKeyResp, error)
	GetCallbackLogList(ctx context.Context, filter *callbackModel.GetListCallbackLogFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetCallbackLogByID(ctx context.Context, id string) (*callbackModel.CallbackLogWithMaster, error)
	GetCallbackIdByMerchantAndMasterCallbackId(ctx context.Context, merchantID, masterID string) (id string, err error)
	FindMerchantCallbackHistory(ctx context.Context, filter *callbackModel.GetListCallbackLogFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetCallbackEvents(ctx context.Context) ([]callbackModel.CallbackEvent, error)
}

type ICustomerRepository interface {
	GetCustomerList(ctx context.Context, merchantId, phoneNumber string, page, perPage int64) ([]customerModel.Customer, *commonModel.Meta, error)
	GetCustomerById(ctx context.Context, id, merchantId string) (*customerModel.Customer, error)
	GetCustomerByPhoneNumber(ctx context.Context, phoneNumber, merchantId string) (*customerModel.Customer, error)
	GetMerchantCustomerByEmail(ctx context.Context, req customerModel.GetMerchantCustomerRequest) (*customerModel.Customer, error)
	FindCustomerByEmail(ctx context.Context, email string) (*customerModel.Customer, error)
	FindCustomerById(ctx context.Context, id string) (*customerModel.Customer, error)
	Create(ctx context.Context, customer *customerModel.Customer) error
	Update(ctx context.Context, customer *customerModel.Customer) error
	Delete(ctx context.Context, id, merchantId string) error
	GetMerchantCustomersByID(ctx context.Context, merchantId string, customerIds []string) ([]*customerModel.Customer, error)
	RemovePaymentMethodFromCustomerByIDAndTokenID(ctx context.Context, id, tokenId string, paymentMethods []*unifiedPaymentModel.CustomerPaymentMethodResponse) error
	GetCardFundedPayoutSavedCardList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetSavedCardList) (*commonModel.PaginationResponse, error)
	GetCardFundedPayoutSavedCardDetail(ctx context.Context, request cardFundedPayoutModel.GetSavedCardDetailRequest) (*cardFundedPayoutModel.GetSavedCardResponse, error)
}

type IBeneficiaryAccountRepository interface {
	Create(
		ctx context.Context, account *beneficiaryAccountModel.BeneficiaryAccount) error
	Update(
		ctx context.Context, account *beneficiaryAccountModel.BeneficiaryAccount) error
	GetList(
		ctx context.Context,
		filter *beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest,
		page, perPage int64) (*commonModel.PaginationResponse, error)
	GetListOfDerived(
		ctx context.Context,
		filter *beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest,
		page, perPage int64) (*commonModel.PaginationResponse, error) // alternative func for non-kyc data
	GetByBankCodeAndAccountNo(
		ctx context.Context, merchantId, bankCode, AccountNo string) (*beneficiaryAccountModel.BeneficiaryAccount, error)
	Upsert(
		ctx context.Context, account *beneficiaryAccountModel.BeneficiaryAccount) error
	GetByID(
		ctx context.Context, id string) (*beneficiaryAccountModel.BeneficiaryAccount, error)
	GetByMerchantID(
		ctx context.Context, merchantId string) (*beneficiaryAccountModel.BeneficiaryAccount, error)
}

type IAccountInquiriesRepository interface {
	Create(ctx context.Context, account *accountInquiries.AccountInquiries) error
	GetByBankCodeAndAccountNo(
		ctx context.Context, bankCode, AccountNo string) (*accountInquiries.AccountInquiries, error)
	Update(ctx context.Context, account *accountInquiries.AccountInquiries) error
	GetByID(ctx context.Context, id string) (*accountInquiries.AccountInquiries, error)
}

type IDisbursementRepository interface {
	BeginTransaction(ctx context.Context) (context.Context, error)
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error

	// Select
	GetList(ctx context.Context, filter *disbursementModel.GetDisbursementFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	GetListBulk(ctx context.Context, filter *disbursementModel.GetBulkDisbursementFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	FindByID(ctx context.Context, id string) (*disbursementModel.DisbursementWithTransaction, error)
	GetByIDs(ctx context.Context, ids []string) ([]*disbursementModel.Disbursement, error)
	FindByMerchantAndReference(ctx context.Context, merchantID, referenceID string) (*disbursementModel.DisbursementWithTransaction, error)
	FindByReference(ctx context.Context, referenceID string) (*disbursementModel.DisbursementWithTransaction, error)
	FindBulkDisbursementByID(ctx context.Context, id string) (*disbursementModel.BulkDisbursement, error)
	GetBulkDisbursementDetailByID(ctx context.Context, id string) (*disbursementModel.BulkDisbursementDetail, error)
	GetAllDisbursementByBulkID(ctx context.Context, bulkID string) ([]*disbursementModel.DisbursementWithTransaction, error)
	GetTransactionConfig(ctx context.Context, merchantId string) (*disbursementModel.TransactionConfig, error)
	GetDailyTransactionLimit(ctx context.Context, merchantId, merchantType string) (*disbursementModel.DailyTransactionLimitResponse, error)
	FindForReversalDisbursementById(ctx context.Context, merchantId, id string) (result *disbursementModel.DisbursementForReversal, err error)
	FindByProcessorReferenceID(ctx context.Context, processorReferenceID string) (*disbursementModel.DisbursementWithTransaction, error)
	GetAvgDurationOfBankTransferProcessInMs(ctx context.Context, startTime, endTime time.Time) (ms float64, err error)
	GetSummaryOfDelayedTransactionBeforeProcessed(ctx context.Context, startTime, endTime time.Time) (disbursementModel.AfterPayoutCutOffTimeSummary, error)
	GetPendingTransactionsBetweenTimeForInquiryTransaction(ctx context.Context, from, to time.Time) ([]*disbursementModel.DisbursementWithTransaction, error)
	GetMerchantIDsForPayoutCallback(ctx context.Context, bulkId string) (merchantIds []string, err error)
	GetBeneficiaryTransactionLimit(ctx context.Context, merchantId, bankCode, accountNo string, startAt, endAt time.Time) (*disbursementModel.BeneficiaryPayoutLimitRuleLimit, error)
	GetCardFundedPayoutTransactionList(ctx context.Context, request cardFundedPayoutModel.GetPayoutTransactionListRequest) ([]cardFundedPayoutModel.GetPayoutTransactionListResponse, error)

	// Insert
	Insert(ctx context.Context, request *disbursementModel.Disbursement) error
	InsertBulkDisbursement(ctx context.Context, request *disbursementModel.BulkDisbursement) error

	// Update
	ApproveInBulk(ctx context.Context, ids []string, approvedBy string) error
	Reject(ctx context.Context, id, reasonType, reasonDescription, rejectedBy string) error
	UpdateProcessorReferenceIdAndBankReferenceNo(ctx context.Context, id, processorReferenceId, bankReferenceNo string) error
	UpdateBankReferenceNo(ctx context.Context, id, bankReferenceNo string) error
	UpdateReasonByIDs(ctx context.Context, ids []string, reasonType, reasonDescription string) error
	UpdateBulkDisbursementStatusByID(ctx context.Context, id, status string) error
	UpdateBulkDisbursementFailedFileByID(ctx context.Context, id, failedFilePath string) error
	UpdateBulkDisbursementRejectedFileByID(ctx context.Context, id, rejectedFilePath string) error
	UpdateReversalTransaction(ctx context.Context, id, reasonType, reasonDescription, createdBy string) error
	UpdateStatusAndReasonByID(ctx context.Context, id, status string, reasonType, reasonDescription *string) error
	UpdateProcessorReferenceByID(ctx context.Context, request *disbursementModel.Disbursement) error
	ReconfirmXB(ctx context.Context, request *disbursementModel.ReconfirmXBRequest) error

	// Sum
	SumAmountByIDs(ctx context.Context, ids []string) (*disbursementModel.SumAmountResponse, error)

	// Count
	CountByIDsAndMerchantID(ctx context.Context, ids []string, merchantID string) int
	CountByBulkID(ctx context.Context, bulkID string) int
	CountWaitingSingleDisbursement(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error)
	CountWaitingBulkDisbursement(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error)
	CountPendingSingleDisbursement(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error)
	CountPendingBulkDisbursement(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error)
	CountByMerchantAndReference(ctx context.Context, merchantID, referenceID string) int
	CountStatusInProgressByBulkID(ctx context.Context, bulkID string) int

	// Summary
	GetSummaryAll(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	GetSummarySuccess(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	GetSummaryFailed(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	GetSummaryInProgress(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	SummaryWaitingToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	SummarySingleWaitingToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	SummaryBulkWaitingToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	SummaryWaitingForTopUpToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	SummarySingleWaitingForTopUpToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	SummaryBulkWaitingForTopUpToday(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	GetSummaryRejected(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	GetSummaryApproved(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter) disbursementDashboardModel.SummaryTransactionDTO
	GetXbPayoutDashboardInsights(ctx context.Context, request disbursementModel.GetXbPayoutDashboardInsightRequest) (*disbursementModel.XbPayoutDashboardInsights, error)

	SummarySuccessByBulkID(ctx context.Context, bulkID string) disbursementDashboardModel.SummaryTransactionDTO
	SummaryFailedByBulkID(ctx context.Context, bulkID string) disbursementDashboardModel.SummaryTransactionDTO
	SummaryCancelledByBulkID(ctx context.Context, bulkID string) disbursementDashboardModel.SummaryTransactionDTO
	SummaryPendingByBulkID(ctx context.Context, bulkID string) disbursementDashboardModel.SummaryTransactionDTO
	GetSummaryByReasonType(
		ctx context.Context,
		filter disbursementDashboardModel.GetDisbursementDashboardFilter,
		transactionStatus string,
	) ([]disbursementDashboardModel.SummaryTransactionByReasonType, error)
	GetSummaryByDisbursementStatus(ctx context.Context, filter disbursementDashboardModel.GetDisbursementDashboardFilter, disbursementStatus string) disbursementDashboardModel.SummaryTransactionDTO
	GetActionTransactionSummary(ctx context.Context, merchantId string, disbursementIds []string) (*disbursementModel.ActionTransactionSummary, error)
	GetDetailForCardFundedPayoutByID(ctx context.Context, id string) (*disbursementModel.Disbursement, error)
	GetCardFundedPayoutList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetPayoutList) (*commonModel.PaginationResponse, error)
	GetCardFundedPayoutDetail(ctx context.Context, request *cardFundedPayoutModel.GetPayoutDetailRequest) (*cardFundedPayoutModel.GetPayoutDetailResponse, error)
	GetCardFundedPayoutInsights(ctx context.Context, filter *cardFundedPayoutModel.FilterGetPayoutInsights) (*cardFundedPayoutModel.GetPayoutInsightsDTO, error)
}

// ISnapCoreRepository SNAP CORE REPOSITORIES
type ISnapCoreRepository interface {
	// Bank Account
	GetBankAccountInquiry(
		ctx context.Context,
		request snapCoreBankAccountModel.InquiryAccountRequest) (*snapCoreBankAccountModel.InquiryAccountResponseData, error)

	CreateVirtualAccount(
		ctx context.Context,
		request snapCoreVAModel.CreateVirtualAccountRequest) (*snapCoreVAModel.CreateVirtualAccountResponseData, error)
	UpdateVirtualAccount(
		ctx context.Context,
		request snapCoreVAModel.UpdateVirtualAccountRequest) (*snapCoreVAModel.UpdateVirtualAccountResponseData, error)
	DeleteVirtualAccount(ctx context.Context, request *snapCoreVAModel.DeleteVirtualAccountRequest) (*snapCoreVAModel.DeleteVirtualAccountResponseData, error)
	BlockVirtualAccount(
		ctx context.Context,
		request *snapCoreVAModel.BlockVirtualAccountRequest,
	) ([]*snapCoreVAModel.BlockVirtualAccountResponseData, error)
	UnblockVirtualAccount(
		ctx context.Context,
		request *snapCoreVAModel.UnblockVirtualAccountRequest,
	) ([]*snapCoreVAModel.UnblockVirtualAccountResponseData, error)

	InquiryStatusVirtualAccount(ctx context.Context, request *snapCoreVAModel.InquiryStatusVARequest) (*snapCoreVAModel.InquiryStatusVAResponse, error)

	BankTransfer(
		ctx context.Context,
		request *snapCoreBankTransferModel.BankTransferRequest,
		headerRequest *snapCoreBankTransferModel.BankTransferHeaderRequest,
	) (*snapCoreBankTransferModel.BankTransferResponseData, error)

	FindBankTransferByExternalID(ctx context.Context, externalId string, forceFailed bool) (*snapCoreBankTransferModel.BankTransferResponseData, error)

	// Bank Config
	GetBankCodeList(ctx context.Context, filter *snapCoreBankConfigModel.GetBankCodeListRequest) (*snapCoreBankConfigModel.BankCodeListResponseData, error)

	// Qris
	GenerateQrMpm(ctx context.Context, request snapCoreQRModel.GenerateQrMpmRequest) (*snapCoreQRModel.GenerateQrMpmResponseData, error)
	QueryQrMpmDynamic(ctx context.Context, uuid string) (*snapCoreQRModel.QueryQrMpmDynamicResponseData, error)
	QueryQrMpmStatic(ctx context.Context, request snapCoreQRModel.QueryQrMpmStaticRequest) (*snapCoreQRModel.QueryQrMpmStaticResponseData, error)
	CancelQrMpm(ctx context.Context, qrisID string) (*snapQrisModel.CancelQrMpmResponseData, error)
	RefundQRMPM(ctx context.Context, req *snapQrisModel.QRMPMRefundRequest) (*snapQrisModel.RefundResponseData, error)
	InquiryStatusQris(ctx context.Context, request *snapCoreQRModel.InquiryStatusQrMpmRequest) (*snapCoreQRModel.QrisInquiryStatusResponse, error)

	TopUpSimulation(ctx context.Context, request snapCoreTopUpSimulationModel.TopupSimulationRequest) (*snapCoreTopUpSimulationModel.TopupSimulationResponseData, error)
	QrMpmPaymentSimulation(ctx context.Context, data *snapQrisModel.QrMpmPaymentSimulationRequest) error

	QrUploadDocument(ctx context.Context, data *snapQrisModel.UploadDocumentReq) (*snapQrisModel.UploadDocumentResp, error)
	QrFinalRegistration(ctx context.Context, data *snapQrisModel.RegistrationReq) error
	QrSyncRegistration(ctx context.Context, data *snapQrisModel.SyncRegistrationDataRequest) error

	// auto recon
	CheckReconTransaction(ctx context.Context, request *snapCoreModel.AutoReconTrxRequest) (*snapCoreModel.AutoReconTrxResponse, error)

	// internal-tool
	UpdateBankTransferStatus(ctx context.Context, request snapCoreBankTransferModel.UpdateBankTransferStatusRequest) error
	CheckStatusByExternalId(ctx context.Context, externalId string, checkBankStatement bool) (*snapCoreBankTransferModel.BankTransferCheckStatusResponseData, error)

	// Virtual account config
	CreateVirtualAccountConfig(ctx context.Context, request *snapCoreVAModel.CreateVirtualAccountConfigRequest) (*snapCoreVAModel.VirtualAccountConfigResponseData, error)
	GetVirtualAccountConfig(ctx context.Context, request *snapCoreVAModel.GetVirtualAccountConfigRequest) ([]*snapCoreVAModel.VirtualAccountConfigResponseData, error)
	UpdateVirtualAccountConfigPrefix(ctx context.Context, request *snapCoreVAModel.UpdateVirtualAccountConfigPrefixRequest) error

	// Create ewallet
	CreateEWalletPaymentLink(ctx context.Context, request *ewallet.EwalletPaymentRequest) (*ewallet.EwalletPaymentLinkResponse, error)
	InquiryStatusEWalletPayment(ctx context.Context, request *ewallet.EWalletInquiryStatusRequest) (*ewallet.EWalletInquiryStatusResponse, error)
	RefundEWallet(ctx context.Context, request *ewallet.EWalletRefundRequest) (*ewallet.EWalletRefundResponse, error)
	EWalletPaymentSimulation(ctx context.Context, request *ewallet.EWalletPaymentSimulationRequest) (*ewallet.EWalletPaymentSimulationResponse, error)

	PublishPayment(ctx context.Context, request snapPaymentModel.PublishRequest) error
	CheckAllowedToRetry(ctx context.Context, request snapCoreBankTransferModel.CheckAllowedToRetryRequest) (*snapCoreBankTransferModel.CheckAllowedToRetryResponse, error)
}

type IMerchantTopUpRepository interface {
	GetByReferenceNumber(ctx context.Context, referenceNumber string) (*merchantTopUp.MerchantTopUp, error)
	GetByMerchantAccountNameAndPaymentMethodId(ctx context.Context, merchantId, accountName, paymentMethodId string) (*merchantTopUp.MerchantTopUp, error)
	Create(ctx context.Context, data *merchantTopUp.MerchantTopUp) error
	GetList(ctx context.Context, request *merchantTopUp.TopUpTransactionListRequest) (*commonModel.PaginationResponse, error)
	CountActiveMerchantTopUpReferences(ctx context.Context, request *merchantTopUp.GetMerchantTopUpReferencesRequest) (int, error)
}

type IPaperCommunicationRepository interface {
	SendEmailV1(ctx context.Context, from string, data *paperCommModel.Email) error
}

type IAdjustmentRepository interface {
	IBasicSQL
	CreateAdjustment(ctx context.Context, data *adjustModel.ManualAdjustmentHistory) error
	FindByID(ctx context.Context, id string) (*adjustModel.ManualAdjustmentHistory, error)
	CalculateAmountBalanceAdjustmentForTopUp(ctx context.Context, relatedAdjustmentID string) (float64, error)
}

type ICredentialRepository interface {
	Get(ctx context.Context, merchantID string) (*credModel.CredentialDashboard, error)
	GetClientSecretById(ctx context.Context, merchantID, id string) (*credModel.ClientSecret, error)
	UpdateClientSecretById(ctx context.Context, merchantID, id string, data *credModel.ClientSecret) (affected bool, err error)
}

type IRequestAccountInquiryRepository interface {
	IBasicSQL

	Create(ctx context.Context, data *requestAccountInquiry.RequestAccountInquiries) error
	FindLatestWithMasterByInquiryID(
		ctx context.Context,
		inquiryID, merchantID string,
	) (*requestAccountInquiry.RequestAccountInquiryWithMaster, error) // TODO: To be deprecated when account_inquiries is deleted
	FindLatestByInquiryID(ctx context.Context, inquiryID, merchantID string) (*requestAccountInquiry.RequestAccountInquiryWithMaster, error)
	FindLatestByNumberAccount(ctx context.Context, accountNo, merchantID string) (*requestAccountInquiry.RequestAccountInquiries, error)
	FindByID(ctx context.Context, id string) (*requestAccountInquiry.RequestAccountInquiryWithMaster, error)
	Update(ctx context.Context, data *requestAccountInquiry.RequestAccountInquiryWithMaster) error
}

type IOutboundRepository interface {
	Insert(ctx context.Context, data *outbound.OutboundRequest) error
	FindByID(ctx context.Context, id string) (*outbound.Outbound, error)
	UpdateClient(ctx context.Context, id string, data *outbound.Client) error
}

type IInboundRepository interface {
	Insert(ctx context.Context, data *inboundModel.InboundRequest) error
	GetList(ctx context.Context, filter *inboundModel.GetInboundFilterRequest) (*commonModel.PaginationResponse, error)
	GetByID(ctx context.Context, id string) (*inboundModel.InboundResponse, error)
}

type IMerchantForbiddenUsecaseRepository interface {
	RegisterForbiddenUsecase(ctx context.Context, req *merchantForbiddenUseCaseModel.MerchantForbiddenUseCase) (*merchantForbiddenUseCaseModel.MerchantForbiddenUseCase, error)
	RemoveForbiddenUsecase(ctx context.Context, req *merchantForbiddenUseCaseModel.MerchantForbiddenUseCase) error
	GetForbiddenUsecase(ctx context.Context, req *merchantForbiddenUseCaseModel.GetMerchantForbiddenUseCaseRequest) ([]*merchantForbiddenUseCaseModel.MerchantForbiddenUseCase, error)
}

type IAddrLocationRepository interface {
	GetProvinces(ctx context.Context) ([]location.Province, error)
	GetCitiesByProvinceId(ctx context.Context, provinceId uint16) ([]location.City, error)
	GetDistrictsByCityId(ctx context.Context, cityId uint16) ([]location.District, error)
	GetDistrictById(ctx context.Context, id uint16) (*location.District, error)
}

type IQrisRepository interface {
	InitRegistration(ctx context.Context, data *qris.Registration) error

	UpdateUploadedDocument(ctx context.Context, id string, data *qris.UpdateDocument) error
	UpdateRegistrationStatus(ctx context.Context, id, status string) error
	UpdateCallbackQrRegistration(ctx context.Context, id string, data *qris.RegistrationCallback) error

	RegistrationList(ctx context.Context, merchantId string) (resp []qris.RegistrationListResp, err error)
	FindQrRegistrationForValidationById(ctx context.Context, id string) (result *qris.Registration, err error)
	FindQrRegistrationByExternalID(ctx context.Context, externalID string) (*qris.Registration, error)
	FindQrRegistrationByExternalIDAndAcquirer(ctx context.Context, externalID string, acquirer string) (*qris.Registration, error)
	FindRegistrationById(ctx context.Context, id string) (resp *qris.RegistrationMerchant, err error)

	UpdateQrRegistration(ctx context.Context, id string, acquirerMerchantId string, acquirerTerminalId string) error
}

type IFeeRepository interface {
	IBasicSQL
}

type IXbCoreProcessorRepository interface {
	GetFxRate(ctx context.Context, request *xbCoreProcessorModel.GetFxRateRequest) (*xbCoreProcessorModel.GetFxRateResponseData, error)
	CreatePayoutSession(ctx context.Context, request *xbCoreProcessorModel.CreatePayoutSessionRequest) (*xbCoreProcessorModel.CreatePayoutSessionResponseData, error)
	ConfirmPayout(ctx context.Context, request *xbCoreProcessorModel.ConfirmPayoutRequest) (*xbCoreProcessorModel.ConfirmPayoutResponseData, error)
	ReConfirmPayout(ctx context.Context, request *xbCoreProcessorModel.ConfirmPayoutRequest) (xbCoreProcessorModel.ReConfirmPayoutResponse, error)
	UploadUnderlyingDocument(ctx context.Context, request *xbCoreProcessorModel.UploadUnderlyingDocumentRequest) (*xbCoreProcessorModel.UploadUnderlyingDocumentResponseData, error)
	// Beneficiary
	CreateBeneficiary(ctx context.Context, request *xbCoreProcessorModel.CreateBeneficiaryRequest) (*xbCoreProcessorModel.CreateBeneficiaryData, error)
	GetListBeneficiary(ctx context.Context, request *xbCoreProcessorModel.GetListBeneficiaryRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetBeneficiaryById(ctx context.Context, request *xbCoreProcessorModel.GetBeneficiaryByIdRequest) (*xbCoreProcessorModel.CreateBeneficiaryData, error)
	UpdateBeneficiary(ctx context.Context, request *xbCoreProcessorModel.UpdateBeneficiaryRequest) (*xbCoreProcessorModel.CreateBeneficiaryData, error)
	DeactivateBeneficiary(ctx context.Context, request *xbCoreProcessorModel.GetBeneficiaryByIdRequest) (*xbCoreProcessorModel.CreateBeneficiaryData, error)
	// Sender
	CreateSender(ctx context.Context, request *xbCoreProcessorModel.CreateSenderRequest) (*xbCoreProcessorModel.CreateSenderData, error)
	GetListSender(ctx context.Context, request *xbCoreProcessorModel.GetListSenderRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetSenderById(ctx context.Context, request *xbCoreProcessorModel.GetSenderByIdRequest) (*xbCoreProcessorModel.CreateSenderData, error)
	UpdateSender(ctx context.Context, request *xbCoreProcessorModel.UpdateSenderRequest) (*xbCoreProcessorModel.CreateSenderData, error)
	DeactivateSender(ctx context.Context, request *xbCoreProcessorModel.GetSenderByIdRequest) (*xbCoreProcessorModel.CreateSenderData, error)
	// Payout
	GetPayoutById(ctx context.Context, request *xbCoreProcessorModel.GetPayoutRequest) (*xbCoreProcessorModel.GetPayoutResponseData, error)
	GetRfiDetails(ctx context.Context, request *xbCoreProcessorModel.GetRfiDetailsRequest) ([]*xbCoreProcessorModel.GetRfiDetailsResponseData, error)
	SubmitRfiDetails(ctx context.Context, request *xbCoreProcessorModel.SubmitRfiDetailsRequest) (*xbCoreProcessorModel.SubmitRfiDetailsResponseData, error)
	// Master Data
	GetListMasterCountry(ctx context.Context, request *xbCoreProcessorModel.GetListMasterCountryRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterCurrency(ctx context.Context, request *xbCoreProcessorModel.GetListMasterCurrencyRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterCurrencyMapping(ctx context.Context, request *xbCoreProcessorModel.GetListMasterCurrencyMappingRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterIdentificationType(ctx context.Context, request *xbCoreProcessorModel.GetListMasterIdentificationTypeRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterAccountType(ctx context.Context, request *xbCoreProcessorModel.GetListMasterAccountTypeRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterPurpose(ctx context.Context, request *xbCoreProcessorModel.GetListMasterPurposeRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterState(ctx context.Context, request *xbCoreProcessorModel.GetListMasterStateRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterCity(ctx context.Context, request *xbCoreProcessorModel.GetListMasterCityRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterTransferMethod(ctx context.Context, request *xbCoreProcessorModel.GetListMasterTransferMethodRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetListMasterSourceOfIncome(ctx context.Context, request *xbCoreProcessorModel.GetListMasterSourceOfIncomeRequest) (*xbCoreProcessorModel.PaginationData, error)
	// Config
	GetListConfigSpread(ctx context.Context, request *xbCoreProcessorModel.GetListConfigSpreadRequest) (*xbCoreProcessorModel.PaginationData, error)
	GetConfigSpreadDetailByID(ctx context.Context, id string) (*xbCoreProcessorModel.GetConfigSpreadDetailData, error)
	CreateConfigSpread(ctx context.Context, request *xbCoreProcessorModel.CreateConfigSpreadRequest) (*xbCoreProcessorModel.CreateConfigSpreadData, error)
	UpdateConfigSpread(ctx context.Context, request *xbCoreProcessorModel.UpdateConfigSpreadRequest) (*xbCoreProcessorModel.UpdateConfigSpreadData, error)
}

type ICreditcardCoreProcessorRepository interface {
	Void(
		ctx context.Context,
		request *creditcardCoreProcessorModel.VoidRequest,
	) (*creditcardCoreProcessorModel.VoidResponseData, error)
	GetTransactionList(
		ctx context.Context,
		request *creditcardCoreProcessorModel.GetTransactionListRequest,
	) (*creditcardCoreProcessorModel.GetTransactionDataList, error)
	Refund(
		ctx context.Context,
		request *creditcardCoreProcessorModel.RefundRequest,
	) (*creditcardCoreProcessorModel.RefundResponseData, error)
	Capture(
		ctx context.Context,
		request *creditcardCoreProcessorModel.CaptureRequest,
	) (*creditcardCoreProcessorModel.CaptureResponseData, error)
	GetMIDByAcquirerMID(
		ctx context.Context,
		acquirerMID string,
	) (*creditcardCoreProcessorModel.MIDResponseData, error)
	GetMID(
		ctx context.Context,
		midId string,
	) (*creditcardCoreProcessorModel.MIDResponseData, error)
	ValidateMidInstallmentBins(
		ctx context.Context,
		request *creditcardCoreProcessorModel.ValidateMIDInstallmentBinsRequest,
	) error
	CreateMID(
		ctx context.Context,
		request *creditcardCoreProcessorModel.CreateMIDRequest,
	) (*creditcardCoreProcessorModel.CreateMIDResponseData, error)
	UpdateMID(
		ctx context.Context,
		request *creditcardCoreProcessorModel.UpdateMIDRequest,
	) (*creditcardCoreProcessorModel.UpdateMIDResponseData, error)
	CreateMIDMap(
		ctx context.Context,
		request *creditcardCoreProcessorModel.CreateMIDMapRequest,
	) (*creditcardCoreProcessorModel.CreateMIDMapResponseData, error)
	GetMIDList(
		ctx context.Context,
		request *creditcardCoreProcessorModel.GetMIDListRequest,
	) (*creditcardCoreProcessorModel.MIDListResponseData, error)
	GetMIDMapList(
		ctx context.Context,
		request *creditcardCoreProcessorModel.GetMIDMapListRequest,
	) (*creditcardCoreProcessorModel.MIDMapListResponseData, error)
	FindMIDMapByMerchant(
		ctx context.Context,
		request *creditcardCoreProcessorModel.FindMIDMapByMerchantRequest,
	) (*creditcardCoreProcessorModel.MIDMapResponseData, error)
	UpdateMIDMapPriority(
		ctx context.Context,
		request creditcardCoreProcessorModel.UpdateMIDMapPriorityRequest,
	) (*creditcardCoreProcessorModel.UpdateMIDMapResponseData, error)
	EncryptCardData(ctx context.Context, request *creditcardCoreProcessorModel.EncryptCardRequest) (*creditcardCoreProcessorModel.EncryptedCardResponse, error)
	GetEncryptedCardData(ctx context.Context, merchantId, cardId string) (*creditcardCoreProcessorModel.EncryptedCardResponse, error)
	CreateEncryptedCardAuthenticationLink(
		ctx context.Context,
		request *creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest,
	) (*creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse, error)
	CheckReconTransaction(ctx context.Context, request *creditcardCoreProcessorModel.AutoReconTrxRequest) (*creditcardCoreProcessorModel.AutoReconTrxResponse, error)
	InquiryTransaction(ctx context.Context, payload *creditcardModel.InquiryTransactionRequest) (*creditcardModel.PaymentNotificationDataRequest, error)
	BlockCard(ctx context.Context, request *creditcardCoreProcessorModel.BlockCardRequest) error
	GetBinDetailByBinNumber(ctx context.Context, merchantId, binNumber string) (*creditcardCoreProcessorModel.GetBinDetailResponse, error)
	GetCardEncryptionPublicKey(ctx context.Context, merchantID string) ([]byte, error)
	Authentication(ctx context.Context, request creditcardCoreProcessorModel.AuthenticationRequest) (*creditcardCoreProcessorModel.AuthenticationResponse, error)
}

type IBankAccountRepository interface {
	Create(ctx context.Context, data *bankAccountModel.BankAccount) error
	Update(ctx context.Context, bk *bankAccountModel.BankAccount) error
	GetByMerchantID(ctx context.Context, merchantID string) (*bankAccountModel.BankAccount, error)
	GetListBankAccount(ctx context.Context, merchantId string) ([]bankAccountModel.BankAccountResponse, error)
	GetBankAccountValidation(ctx context.Context, merchantId, bankCode, accountNo string) (*bankAccountModel.BankAccountResponse, error)
	BankAccountHasBeenPrepared(ctx context.Context, merchantId string) (result bool, err error)
}

type IWithdrawalRepository interface {
	IBasicSQL
	Create(ctx context.Context, data *withdrawal.Withdrawal) error
	UpdateMetadataById(ctx context.Context, id string, metadata *withdrawal.Metadata) error

	FindById(ctx context.Context, id, merchantId string) (*withdrawal.Withdrawal, error)
	GetList(ctx context.Context, request *withdrawal.WithdrawalHistoryRequest) (*commonModel.PaginationResponse, error)
	GetById(ctx context.Context, request *withdrawal.WithdrawalDetailRequest) (*withdrawal.WithdrawalDetailResponse, error)
	GetByReferenceId(ctx context.Context, merchantId, referenceId string) (*withdrawal.WithdrawalDetailResponse, error)
	GetTodayWithdrawalInsight(ctx context.Context, opt withdrawal.WithdrawalInsightRequest) (*withdrawal.WithdrawalInsightResponse, error)

	GetTransactionConfig(ctx context.Context, merchantId string) (*merchant.WithdrawalConfig, error)
}

type ITransferRepository interface {
	Create(ctx context.Context, data *transfer.Transfer) error
	Update(ctx context.Context, data *transfer.Transfer) error
	GetByID(ctx context.Context, id, merchantId string) (*transfer.Transfer, error)
	GetByReferenceID(ctx context.Context, merchantId, recipientId, referenceId string) (*transfer.Transfer, error)
	GetList(ctx context.Context, req *transfer.GetTransferListRequest, page, perPage int64) (data []*transfer.Transfer, total int64, err error)
	GetTransferTransaction(ctx context.Context, req transfer.GetTransferTransactionRequest) (*transfer.TransferTransactionDetail, error)
}

type IProductRepository interface {
	GetProductList(ctx context.Context) ([]*product.Product, error)
	GetProductById(ctx context.Context, productId string) (*product.Product, error)
	UpdateProductAvailability(ctx context.Context, req *product.UpdateProductRequest) error
	AddMerchantSelectedProduct(ctx context.Context, req *product.MerchantSelectedProduct) error
	GetMerchantSelectedProducts(ctx context.Context, merchantId string) ([]*product.MerchantWithProductName, error)
	GetMerchantActiveProducts(ctx context.Context, merchantId string) ([]*product.MerchantWithProductName, error)
	UpdateMerchantProductAvailability(ctx context.Context, req *product.UpdateMerchantSelectedProductAvailabilityRequest) error
	GetMerchantSelectedProductByName(ctx context.Context, merchantId, productName string) (*product.MerchantWithProductName, error)
	GetMerchantSelectedProductById(ctx context.Context, merchantId, productId string) (*product.MerchantWithProductName, error)
}

type ILiveFeatureRepository interface {
	GetAll(ctx context.Context) ([]liveFeature.LiveFeature, error)
	RetrieveAppVersion(ctx context.Context) (liveFeature.AppVersion, error)

	WithConfig(config *config.Config)
	WithSecret(secret *config.Secret)
}
type IRoutingProcessorRepository interface {
	TriggerTransfer(ctx context.Context, request *routingProcessorModel.BankTransferRequest) (*routingProcessorModel.BankTransferResponseData, error)
	BankAccountInquiry(ctx context.Context, request *routingProcessorModelInquiry.InquiryAccountRequest) (*routingProcessorModelInquiry.InquiryAccountResponseData, error)
	GetTransferById(ctx context.Context, id string, forceFailed bool) (*routingProcessorModel.BankTransferResponseData, error)
}

type IFlipProcessorRepository interface {
	GetEscrowBalance(ctx context.Context) (*routingProcessorModelEscrowBalance.EscrowBalanceResponse, error) // when other processor already implement this, we need move interface to routingProcessor
}

type IDanaProcessorRepository interface {
	GetEscrowBalance(ctx context.Context) (*routingProcessorModelEscrowBalance.EscrowBalanceResponse, error) // when other processor already implement this, we need move interface to routingProcessor
}

type IReconciliationRepository interface {
	Create(ctx context.Context, data *reconciliationModel.Reconciliation) error
	GetByUUID(ctx context.Context, uuid string) (*reconciliationModel.Reconciliation, error)
	GetAll(ctx context.Context, filter *reconciliationModel.ReconciliationFilterRequest) (*commonModel.PaginationResponse, error)
	Update(ctx context.Context, data *reconciliationModel.Reconciliation) error
}

type IIPWhitelistRepository interface {
	Create(ctx context.Context, configuration *ipwhitelistModel.IPWhitelistConfiguration) error
	Update(ctx context.Context, configuration *ipwhitelistModel.IPWhitelistConfiguration) error
	List(ctx context.Context, req *ipwhitelistModel.GetIPWhitelistConfiguration) ([]*ipwhitelistModel.IPWhitelistConfiguration, int64, error)
	Detail(ctx context.Context, uuid string) (*ipwhitelistModel.IPWhitelistConfiguration, error)
	Delete(ctx context.Context, uuid string) error
}

type IRateLimiterRepository interface {
	List(ctx context.Context, req *ratelimiter.MerchantRateLimitRequest) ([]*ratelimiter.RateLimitConfiguration, int64, error)
	Detail(ctx context.Context, uuid string) (*ratelimiter.RateLimitConfiguration, error)
	Create(ctx context.Context, configuration *ratelimiter.RateLimitConfiguration) error
	Update(ctx context.Context, configuration *ratelimiter.RateLimitConfiguration) error
	GetMerchantConfigs(ctx context.Context, merchantID string) (*[]ratelimiter.MerchantRateLimitConfig, error)
}

type IWalletTransactionRepository interface {
	GetMerchantTransactionHistoryList(
		ctx context.Context,
		req walletTransactionModel.MerchantTransactionHistoryListReq,
	) (resp []walletTransactionModel.MerchantTransactionHistoryListResp, totalRows int64, err error)
	GetMerchantTransactionHistoryListForExport(ctx context.Context, req walletTransactionModel.MerchantTransactionHistoryListReq) ([]walletTransactionModel.MerchantTransactionHistoryListResp, error)
	GetMerchantTransactionDetail(ctx context.Context, merchantId, id string) (*walletTransactionModel.MerchantTransactionDetailResp, error)
}

type IMerchantRcnRepository interface {
	FindByIDAndMerchantID(ctx context.Context, id string, merchantId string) (*merchantRcn.MerchantRcn, error)
}
type ICimbProcessorRepository interface {
	InquiryCorporateCreditCard(ctx context.Context, request *cimbProcessorModel.InquiryCorporateCreditCardRequest) (*cimbProcessorModel.InquiryCorporateCreditCardResponse, error)
	InquiryTransactionCorporateCreditCard(ctx context.Context, request *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) (*vccSettlement.ProcessorVccTransactionInquiryResponse, error)
}

type IFraudRulesRepository interface {
	Create(ctx context.Context, rule *fraudrulesmodel.FraudRules) error
	Update(ctx context.Context, rule *fraudrulesmodel.FraudRules) error
	Delete(ctx context.Context, uuid string) error
	List(ctx context.Context, q *fraudrulesmodel.FraudRulesQuery) ([]*fraudrulesmodel.FraudRules, int, error)
	GetByID(ctx context.Context, id string) (*fraudrulesmodel.FraudRules, error)
}

type IRuleEvaluationsRepository interface {
	Create(ctx context.Context, eval *ruleevaluationsmodel.RuleEvaluations) error
	Update(ctx context.Context, eval *ruleevaluationsmodel.RuleEvaluations) error
	GetByID(ctx context.Context, id string) (*ruleevaluationsmodel.RuleEvaluations, error)
	GetByRefID(ctx context.Context, refID string) (*[]ruleevaluationsmodel.RuleEvaluations, error)
}

type IFdsProcessorRepository interface {
	Check(ctx context.Context, request *fdscommon.CheckRequest) (*fdscommon.CheckResponse, error)
	Update(ctx context.Context, request *fdscommon.UpdateRequest) (*fdscommon.UpdateResponse, error)
}

type IWorkflowFDSRepository interface {
	AssessPayoutTransaction(ctx context.Context, request fdscommon.AssessPayoutTransactionRequest) (*fdscommon.TransactionAssessmentResponse, error)
	NewFDSProcessor() IFdsProcessorRepository
}

type IIndustryRepository interface {
	GetAllIndustries(ctx context.Context, request *industryModel.SearchIndustryRequest) ([]*industryModel.Industry, error)
	GetUniqueParentIndustries(ctx context.Context) ([]string, error)
	GetChildIndustries(ctx context.Context, parentIndustry string) ([]string, error)
	GetMCCForIndustry(ctx context.Context, parentIndustry, childIndustry string) (string, error)
	IsValidMCC(ctx context.Context, mcc string) (bool, error)
	GetIndustryByID(ctx context.Context, id string) (*industryModel.Industry, error)
	Create(ctx context.Context, industry *industryModel.Industry) error
	Update(ctx context.Context, industry *industryModel.Industry) error
	Delete(ctx context.Context, uuid string) error
	GetByParentChildIndustry(ctx context.Context, parent, child string) (*industryModel.Industry, error)
	IsIndustryUsedByMerchants(ctx context.Context, parent, child string) (bool, error)
}

type IAmlProcessorRepository interface {
	Check(ctx context.Context, request *amlcommon.CheckRequest) (*amlcommon.CheckResponse, error)
	Inquiry(ctx context.Context, transactionID string) (*amlcommon.InquiryResponse, error)
	ProfileDetail(ctx context.Context, inquiryID string, profileID string) (*amlcommon.ProfileDetailResponse, error)
}

type ICountryRepository interface {
	GetAll(ctx context.Context, filter *countryModel.SearchFilterRequest) ([]*countryModel.Country, error)
	FindByCode(ctx context.Context, code string) (*countryModel.Country, error)
}

type IDukcapilGatewayRepository interface {
	VerifyIdentity(ctx context.Context, req *dukcapilmodel.VerifyRequest) (*dukcapilmodel.VerifyResult, error)
}

type IStatusHistoriesRepository interface {
	Insert(ctx context.Context, data *statusHistoriesModel.StatusHistory) error
	GetByReference(ctx context.Context, referenceType, referenceID string) ([]*statusHistoriesModel.StatusHistory, error)
}

type ITablePartitionRepository interface {
	GetPartitionNames(ctx context.Context, tableName, ordinalPartitionOrder string, limit int) ([]string, error)
	ReorganizeDailyRangePartition(ctx context.Context, request partitionModel.ReorganizeRangePartitionRequest) error
	ReorganizeMonthlyRangePartition(ctx context.Context, request partitionModel.ReorganizeRangePartitionRequest) error
}

type IPaymentCaptureRepository interface {
	Insert(ctx context.Context, capture *paymentCaptureModel.PaymentCapture) error
	Update(ctx context.Context, capture *paymentCaptureModel.PaymentCapture) error
	GetByID(ctx context.Context, id string) (*paymentCaptureModel.PaymentCapture, error)
	GetByPaymentID(ctx context.Context, paymentID string) ([]*paymentCaptureModel.PaymentCapture, error)
}

type IInstallmentPlanRepository interface {
	Create(ctx context.Context, plan *installmentPlanModel.InstallmentPlan) error
	Update(ctx context.Context, plan *installmentPlanModel.InstallmentPlan) error
	GetById(ctx context.Context, planId string) (*installmentPlanModel.InstallmentPlan, error)
	List(ctx context.Context, req *installmentPlanModel.FilterRequest) ([]*installmentPlanModel.InstallmentPlan, int64, error)
}

type IShortLinkRepository interface {
	Create(ctx context.Context, shortLink *shortLinkModel.ShortLink) error
	Update(ctx context.Context, shortLink *shortLinkModel.ShortLink) error
	GetByCode(ctx context.Context, code string) (*shortLinkModel.ShortLink, error)
}

type IVCCSettlementRepository interface {
	BulkInsert(ctx context.Context, data []*vccSettlement.VccSettlement) error
	Delete(ctx context.Context, rcnId string, postingDate time.Time) error
}

type ISettlementHoldRepository interface {
	Create(ctx context.Context, data *settlementHold.SettlementHold, history *settlementHold.SettlementHoldHistory) error
	Update(ctx context.Context, data *settlementHold.SettlementHold, history *settlementHold.SettlementHoldHistory) error
	GetByPaymentID(ctx context.Context, paymentId string) (*settlementHold.SettlementHold, error)
}

type IReportingRepository interface {
	PrepareAdvancedBalanceHistoryData(ctx context.Context, data *reportingModel.BalanceHistory) error
	ExportBalanceHistory(ctx context.Context, filter *orchestratorModel.TransactionHistoryFilterRequest) ([]orchestratorModel.TransactionHistory, error)
	ListBalanceHistory(ctx context.Context, filters *orchestratorModel.TransactionHistoryFilterRequest, page, perPage int64) (*commonModel.PaginationResponse, error)
	ListAccountTransactionsForMigration(ctx context.Context, startDate, endDate time.Time) ([]cdcModel.AccountTransaction, error)

	UpsertBalanceHistory(ctx context.Context, data reportingModel.BalanceHistory) error
	HardDeleteBalanceHistory(ctx context.Context, transactionID string) error
	SoftDeleteBalanceHistory(ctx context.Context, transactionID string, ingestedAt time.Time) error
	UpdateSettlementBalanceHistory(ctx context.Context, data reportingModel.BalanceHistory) error
}

type IVendorRepository interface {
	Create(ctx context.Context, vendor *vendorModel.Vendor) error
	Update(ctx context.Context, vendor *vendorModel.Vendor) error
	Delete(ctx context.Context, uuid string) error
	List(ctx context.Context, q *vendorModel.VendorQuery) ([]*vendorModel.Vendor, int, error)
	GetByID(ctx context.Context, id string) (*vendorModel.Vendor, error)
	GetByName(ctx context.Context, merchantID, name string) (*vendorModel.Vendor, error)
}

type IPayoutManualProcessingAccountRepository interface {
	Create(ctx context.Context, account *payoutManualProcessingAccountModel.PayoutManualProcessingAccount) error
	Update(ctx context.Context, account *payoutManualProcessingAccountModel.PayoutManualProcessingAccount) error
	GetByUUID(ctx context.Context, uuid string) (*payoutManualProcessingAccountModel.PayoutManualProcessingAccount, error)
	List(ctx context.Context, q *payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery) ([]*payoutManualProcessingAccountModel.PayoutManualProcessingAccount, int, error)
	IsManualProcessingAccount(ctx context.Context, merchantID, bankCode, accountNumber string) (bool, error)
}

type ITNCRepository interface {
	IBasicSQL

	// TNC Versions
	CreateTNCVersion(ctx context.Context, version *tncModel.TNC) error
	UpdateTNCVersion(ctx context.Context, version *tncModel.TNC) error
	DeactivateAllTNCVersions(ctx context.Context) error
	GetTNCVersionByID(ctx context.Context, id string) (*tncModel.TNC, error)
	GetTNCVersionByVersion(ctx context.Context, version string) (*tncModel.TNC, error)
	GetActiveTNCVersion(ctx context.Context) (*tncModel.TNC, error)
	ListTNCVersions(ctx context.Context, q *tncModel.TNCVersionQuery) ([]*tncModel.TNC, int, error)

	// Merchant TNC Signing Histories
	InsertSigningHistory(ctx context.Context, history *tncModel.MerchantTNCSigningHistory) error
	GetLatestSigningByMerchant(ctx context.Context, merchantID string) (*tncModel.MerchantTNCSigningHistory, error)
	GetSigningByMerchantAndVersion(ctx context.Context, merchantID, version string) (*tncModel.MerchantTNCSigningHistory, error)
	ListSigningHistories(ctx context.Context, q *tncModel.SigningHistoryQuery) ([]*tncModel.MerchantTNCSigningHistory, int, error)
}
