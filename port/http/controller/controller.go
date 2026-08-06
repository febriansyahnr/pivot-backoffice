package controller

import (
	"net/http"
)

type ContextKey string

type V1UserController interface {
	ListUsers(w http.ResponseWriter, r *http.Request)
	ListByMerchantID(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	FindByID(w http.ResponseWriter, r *http.Request)
	ActivateUser(w http.ResponseWriter, r *http.Request)
	DeactivateUser(w http.ResponseWriter, r *http.Request)
	UnblockUser(w http.ResponseWriter, r *http.Request)
	ResendInvitation(w http.ResponseWriter, r *http.Request)

	Login(w http.ResponseWriter, r *http.Request)
	LoginWithOTP(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	ForgotPassword(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	SessionFromLogin2FA(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	AddUserRole(w http.ResponseWriter, r *http.Request)
	UserDetail(w http.ResponseWriter, r *http.Request)
	UserProfile(w http.ResponseWriter, r *http.Request)
	GenerateRandomPassword(w http.ResponseWriter, r *http.Request)
	ChangePassword(w http.ResponseWriter, r *http.Request)
	CheckCurrentPassword(w http.ResponseWriter, r *http.Request)

	// PIN
	CreatePin(w http.ResponseWriter, r *http.Request)
	CheckCurrentPin(w http.ResponseWriter, r *http.Request)
	ChangePin(w http.ResponseWriter, r *http.Request)
	ResetPIN(w http.ResponseWriter, r *http.Request)

	ValidateInvitationToken(w http.ResponseWriter, r *http.Request)
	Activate(w http.ResponseWriter, r *http.Request)

	// Bypass Endpoint For Automated Test
	GetInvitationURL(w http.ResponseWriter, r *http.Request)

	// Multi Factor Authentication (MFA)
	EnrollTOTP(w http.ResponseWriter, r *http.Request)
	ConfirmTOTP(w http.ResponseWriter, r *http.Request)
	SetPreferred2FAMethod(w http.ResponseWriter, r *http.Request)
}

type V1MerchantController interface {
	Create(w http.ResponseWriter, r *http.Request)
	FindByMerchantID(w http.ResponseWriter, r *http.Request)
	SetPKCS8MerchantPublicKey(w http.ResponseWriter, r *http.Request)

	FindMerchantFeeByMerchantIDAndType(w http.ResponseWriter, r *http.Request)
	CreateMerchantFee(w http.ResponseWriter, r *http.Request)
	UpdateMerchantFee(w http.ResponseWriter, r *http.Request)

	UploadDocument(w http.ResponseWriter, r *http.Request)
	GetDocuments(w http.ResponseWriter, r *http.Request)
	CreateMerchantBOD(w http.ResponseWriter, r *http.Request)
	GetListMerchantBOD(w http.ResponseWriter, r *http.Request)
	UpdateMerchantBOD(w http.ResponseWriter, r *http.Request)

	GenOpenAPISignature(w http.ResponseWriter, r *http.Request)

	CloseMerchant(w http.ResponseWriter, r *http.Request)
	GetActiveProducts(w http.ResponseWriter, r *http.Request)

	GetNotificationConfig(w http.ResponseWriter, r *http.Request)
	UpdateNotificationConfig(w http.ResponseWriter, r *http.Request)
}

type V1SubMerchantController interface {
	ListSubMerchantByParentID(w http.ResponseWriter, r *http.Request)
	DetailSubMerchantByID(w http.ResponseWriter, r *http.Request)
	GetSubMerchantDailyLimit(w http.ResponseWriter, r *http.Request)
	GetSubMerchantBalance(w http.ResponseWriter, r *http.Request)
	ExportPaymentHistory(w http.ResponseWriter, r *http.Request)
	ExportDisbursementHistory(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Block(w http.ResponseWriter, r *http.Request)
	Unblock(w http.ResponseWriter, r *http.Request)
}

type V1RoleController interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	FindPermissionsByRoleId(w http.ResponseWriter, r *http.Request)
}

type V1MenuController interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	GetByRole(w http.ResponseWriter, r *http.Request)
	GetByRoleId(w http.ResponseWriter, r *http.Request)
}

type V1ActivityController interface {
	GetList(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
}

type V1InternalPaymentController interface {
	FindPaymentById(w http.ResponseWriter, r *http.Request)
	QueryQrMpmDynamic(w http.ResponseWriter, r *http.Request)
	QueryQrMpmStatic(w http.ResponseWriter, r *http.Request)
	SNAPQueryQrMpmStatic(w http.ResponseWriter, r *http.Request)
	SNAPQueryQrMpmDynamic(w http.ResponseWriter, r *http.Request)
	SNAPGenerateQRMpm(w http.ResponseWriter, r *http.Request)
	SNAPCreateVA(w http.ResponseWriter, r *http.Request)
	SNAPUpdateVA(w http.ResponseWriter, r *http.Request)
	SNAPGetVA(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type V1InternalMerchantAuthController interface {
	GetAuthInfo(w http.ResponseWriter, r *http.Request)
	GetAccessTokenB2b(w http.ResponseWriter, r *http.Request)
	ValidateAccessTokenB2b(w http.ResponseWriter, r *http.Request)
	ValidateSNAPAccessTokenB2b(w http.ResponseWriter, r *http.Request)
	GetPKCS8SecretKey(w http.ResponseWriter, r *http.Request)
	CreatePKCS8SecretKey(w http.ResponseWriter, r *http.Request)
	UtilEncryptingKey(w http.ResponseWriter, r *http.Request)
	GetSNAPAccessTokenB2B(w http.ResponseWriter, r *http.Request)
	GenerateB2BTokenSNAPSignature(w http.ResponseWriter, r *http.Request)
	ValidateSNAPSignature(w http.ResponseWriter, r *http.Request)
	GenerateSNAPSignature(w http.ResponseWriter, r *http.Request)
	ValidateB2B2CTokenSNAPSignature(w http.ResponseWriter, r *http.Request)
}

type V1InternalPayoutController interface {
	Create(w http.ResponseWriter, r *http.Request)
	FindByBulkId(w http.ResponseWriter, r *http.Request)
	RetryBulk(w http.ResponseWriter, r *http.Request)
}

type V1InternalAccountInquiryController interface {
	RequestAccountInquiry(w http.ResponseWriter, r *http.Request)
	CheckStatusRequestInquiry(w http.ResponseWriter, r *http.Request)
}

type V1InternalAccountController interface {
	CreateWalletCustomerAccount(w http.ResponseWriter, r *http.Request)
	GetWalletCustomerAccount(w http.ResponseWriter, r *http.Request)
	GetWalletMerchantAccount(w http.ResponseWriter, r *http.Request)
	GetBalance(w http.ResponseWriter, r *http.Request)
}

type V1InternalSubMerchantController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	ListSubMerchantByParentID(w http.ResponseWriter, r *http.Request)
	DetailSubMerchantByID(w http.ResponseWriter, r *http.Request)
	GetSubMerchantBalance(w http.ResponseWriter, r *http.Request)
	AssignAdmin(w http.ResponseWriter, r *http.Request)
	ResendInvitation(w http.ResponseWriter, r *http.Request)
}

type V1CallbackController interface {
	RegisterCallback(w http.ResponseWriter, r *http.Request)
	GetCallbackList(w http.ResponseWriter, r *http.Request)
	GetCallbackLogList(w http.ResponseWriter, r *http.Request)
	GetCallbackLogDetail(w http.ResponseWriter, r *http.Request)
	ResendCallbackByID(w http.ResponseWriter, r *http.Request)
	ResendSNAPCallbackByID(w http.ResponseWriter, r *http.Request)
	GetCallbackEvents(w http.ResponseWriter, r *http.Request)
}

type V1MerchantTopUpController interface {
	Topup(w http.ResponseWriter, r *http.Request)
	TopUpSimulation(w http.ResponseWriter, r *http.Request)
}

type V1DisbursementController interface {
	GetDisbursementDashboard(w http.ResponseWriter, r *http.Request)
	GetDisbursementApprovalDashboard(w http.ResponseWriter, r *http.Request)
	GetTransactionConfig(w http.ResponseWriter, r *http.Request)
	GetTransactionLimit(w http.ResponseWriter, r *http.Request)
	GetTransactionLimitSubMerchant(w http.ResponseWriter, r *http.Request)
	GetDailyTransactionLimit(w http.ResponseWriter, r *http.Request)
	GetCutOffTimeStatus(w http.ResponseWriter, r *http.Request)

	GetList(w http.ResponseWriter, r *http.Request)
	ExportToExcel(w http.ResponseWriter, r *http.Request)

	FindByID(w http.ResponseWriter, r *http.Request)
	GetReceiptByID(w http.ResponseWriter, r *http.Request)

	CreateSingle(w http.ResponseWriter, r *http.Request)
	RetrySingle(w http.ResponseWriter, r *http.Request)

	ApprovalActions(w http.ResponseWriter, r *http.Request)

	GetListBulkDisbursement(w http.ResponseWriter, r *http.Request)
	GetBulkDisbursementDetail(w http.ResponseWriter, r *http.Request)
	BulkPreview(w http.ResponseWriter, r *http.Request)
	BulkValidate(w http.ResponseWriter, r *http.Request)
	CreateBulk(w http.ResponseWriter, r *http.Request)
	RetryBulk(w http.ResponseWriter, r *http.Request)
	Cancel(w http.ResponseWriter, r *http.Request)

	GetDisbursementInsight(w http.ResponseWriter, r *http.Request)
}

type V1BeneficiaryAccountController interface {
	CheckBeneficiary(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)
}

type V1BankController interface {
	List(w http.ResponseWriter, r *http.Request)
}

type V1MasterPurposeController interface {
	List(w http.ResponseWriter, r *http.Request)
}

type V1PaymentMethodController interface {
	FindPaymentMethodByCategory(w http.ResponseWriter, r *http.Request)
}

type V1OrchestratorController interface {
	ExportToExcelTransactionHistory(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)
	GetDetailById(w http.ResponseWriter, r *http.Request)
	GetOpenApiBalanceHistories(w http.ResponseWriter, r *http.Request)
}

type V1AccountController interface {
	GetByUUID(w http.ResponseWriter, r *http.Request)
	GetBalance(w http.ResponseWriter, r *http.Request)
}

type V1OTPController interface {
	Send(w http.ResponseWriter, r *http.Request)
	Verify(w http.ResponseWriter, r *http.Request)
}

type V1CRMAdjustment interface {
	CreateManualTopup(w http.ResponseWriter, r *http.Request)
	CreateAdjustmentFromManualTopup(w http.ResponseWriter, r *http.Request)
	CreateBalanceAdjustment(w http.ResponseWriter, r *http.Request)
	HoldMerchantBalance(w http.ResponseWriter, r *http.Request)
	GetHoldedMerchantBalance(w http.ResponseWriter, r *http.Request)
}

type V1CRMUserController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	ResendInvitation(w http.ResponseWriter, r *http.Request)
}

type V1CRMMerchantController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Assign(w http.ResponseWriter, r *http.Request)
	TransactionConfig(w http.ResponseWriter, r *http.Request)
	FDSConfig(http.ResponseWriter, *http.Request)
	PaymentInvestigationConfig(http.ResponseWriter, *http.Request)
	GetTransactionConfig(w http.ResponseWriter, r *http.Request)
	GetFDSConfig(http.ResponseWriter, *http.Request)
	UpdateSettlementConfig(w http.ResponseWriter, r *http.Request)
	UpdateFeeTieringConfig(w http.ResponseWriter, r *http.Request)
	UploadReservedShortName(w http.ResponseWriter, r *http.Request)

	CreateFeeConfigOnBehalf(w http.ResponseWriter, r *http.Request)
	GetFeeConfigOnBehalf(w http.ResponseWriter, r *http.Request)
	UpdateFeeConfigOnBehalf(w http.ResponseWriter, r *http.Request)
	UpdateKYC(w http.ResponseWriter, r *http.Request)
	SetCustomLimitConfig(w http.ResponseWriter, r *http.Request)

	GetBillingFees(w http.ResponseWriter, r *http.Request)
	PayBillingFees(w http.ResponseWriter, r *http.Request)

	BulkCreateSubmerchant(w http.ResponseWriter, r *http.Request)
	GetBulkCreateSubmerchantSummary(w http.ResponseWriter, r *http.Request)
	BlockMerchant(w http.ResponseWriter, r *http.Request)
	UnblockMerchant(w http.ResponseWriter, r *http.Request)
	GetMerchantTNCHistory(w http.ResponseWriter, r *http.Request)
}

type V1CredentialSettingController interface {
	Get(w http.ResponseWriter, r *http.Request)
	GetClientSecretById(w http.ResponseWriter, r *http.Request)
	GenerateClientSecretById(w http.ResponseWriter, r *http.Request)
}

type V1CallbackSettingController interface {
	Get(w http.ResponseWriter, r *http.Request)
	GetApiKey(w http.ResponseWriter, r *http.Request)
	TestAndSaveCallbackURL(w http.ResponseWriter, r *http.Request)
	TestAndSaveSnapB2b(w http.ResponseWriter, r *http.Request)
	TestAndSaveSnapPayment(w http.ResponseWriter, r *http.Request)
}

type V1ApiLogsSettingController interface {
	GetList(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	GetSnapVersionByID(w http.ResponseWriter, r *http.Request)
}

type V1DepositSettingController interface {
	Get(w http.ResponseWriter, r *http.Request)
	SetAutoWithdrawal(w http.ResponseWriter, r *http.Request)
}

type V1CRMDisbursementController interface {
	RetryTransaction(w http.ResponseWriter, r *http.Request)
	InquiryTransaction(w http.ResponseWriter, r *http.Request)
	Reversal(w http.ResponseWriter, r *http.Request)
	ChangeStatus(w http.ResponseWriter, r *http.Request)
	CheckTransactionStatus(w http.ResponseWriter, r *http.Request)
	GetFlipEscrowBalance(w http.ResponseWriter, r *http.Request)
	GetDanaEscrowBalance(w http.ResponseWriter, r *http.Request)
	GetPayoutStatusAndRouting(w http.ResponseWriter, r *http.Request)
	GetBatchPayoutStatusAndRouting(w http.ResponseWriter, r *http.Request)
	GetReceipt(w http.ResponseWriter, r *http.Request)
}

type V1CreditCardController interface {
	CreatePayment(w http.ResponseWriter, r *http.Request)
	GetPaymentById(w http.ResponseWriter, r *http.Request)
	GetStoredCardByCustomerID(w http.ResponseWriter, r *http.Request)
	RemoveCardByCustomerIDAndTokenID(w http.ResponseWriter, r *http.Request)
	GeneratePaymentToken(w http.ResponseWriter, r *http.Request)
}

type V1CRMMerchantForbiddenUseCaseController interface {
	Block(w http.ResponseWriter, r *http.Request)
	Unblock(w http.ResponseWriter, r *http.Request)
}

type V1InternalMerchantController interface {
	Block(w http.ResponseWriter, r *http.Request)
	Unblock(w http.ResponseWriter, r *http.Request)
	Detail(w http.ResponseWriter, r *http.Request)
	GetJITApiKey(w http.ResponseWriter, r *http.Request)
}

type V1InternalMerchantRcnController interface {
	FindByIDAndMerchantID(w http.ResponseWriter, r *http.Request)
	InquiryTransactions(w http.ResponseWriter, r *http.Request)
}

type V2InternalLedgerController interface {
	GetLedgerDetail(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	GetTransactions(w http.ResponseWriter, r *http.Request)
	GetLedgerBalance(w http.ResponseWriter, r *http.Request)
	CalculateBulkLedgerBalance(w http.ResponseWriter, r *http.Request)
}

type V1AddrLocationController interface {
	Get(w http.ResponseWriter, r *http.Request)
}

type V1CRMPaymentMethodController interface {
	GetByMerchant(w http.ResponseWriter, r *http.Request)
	ActivatePaymentMethodMerchant(w http.ResponseWriter, r *http.Request)
	ActivateAllPaymentMethod(w http.ResponseWriter, r *http.Request)
	DeactivatePaymentMethodMerchant(w http.ResponseWriter, r *http.Request)
	SetupConfig(w http.ResponseWriter, r *http.Request)
	ChangeActivationStatus(w http.ResponseWriter, r *http.Request)
	GetRequiredMerchantDocumentList(w http.ResponseWriter, r *http.Request)
	GetStaticVAByMerchant(w http.ResponseWriter, r *http.Request)
	GetStaticQRByMerchant(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
}

type V1CRMProductController interface {
	GetProductList(w http.ResponseWriter, r *http.Request)
	UpdateProductAvailability(w http.ResponseWriter, r *http.Request)
	AddMerchantSelectedProduct(w http.ResponseWriter, r *http.Request)
	GetMerchantSelectedProducts(w http.ResponseWriter, r *http.Request)
	UpdateMerchantProductAvailability(w http.ResponseWriter, r *http.Request)
}

type V1CRMAccountController interface {
	GetBalance(w http.ResponseWriter, r *http.Request)
}

type V1XbPayoutController interface {
	GetFxRate(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)
	GetDetail(w http.ResponseWriter, r *http.Request)
	CreateSession(w http.ResponseWriter, r *http.Request)
	Confirm(w http.ResponseWriter, r *http.Request)
	UploadUnderlyingDocument(w http.ResponseWriter, r *http.Request)
	ExportToExcel(w http.ResponseWriter, r *http.Request)

	// Master
	GetListMasterCountry(w http.ResponseWriter, r *http.Request)
	GetListMasterCurrency(w http.ResponseWriter, r *http.Request)
	GetListMasterCurrencyMapping(w http.ResponseWriter, r *http.Request)
	GetListMasterIdentificationType(w http.ResponseWriter, r *http.Request)
	GetListMasterAccountType(w http.ResponseWriter, r *http.Request)
	GetListMasterPurpose(w http.ResponseWriter, r *http.Request)
	GetListMasterState(w http.ResponseWriter, r *http.Request)
	GetListMasterCity(w http.ResponseWriter, r *http.Request)
	GetListMasterTransferMethod(w http.ResponseWriter, r *http.Request)
	GetListMasterSourceOfIncome(w http.ResponseWriter, r *http.Request)

	// Insights
	GetXbPayoutDashboardInsights(w http.ResponseWriter, r *http.Request)

	// Config
	GetFeeConfig(w http.ResponseWriter, r *http.Request)
}

type V1QrisController interface {
	Registration(w http.ResponseWriter, r *http.Request)
	RegistrationList(w http.ResponseWriter, r *http.Request)
	ReuploadDocument(w http.ResponseWriter, r *http.Request)
	DuplicateRegistration(w http.ResponseWriter, r *http.Request)
}

type V1SimulationController interface {
	GetPaymentMethodForPayment(w http.ResponseWriter, r *http.Request)
	GetPaymentByID(w http.ResponseWriter, r *http.Request)
	ProcessPayment(w http.ResponseWriter, r *http.Request)
}

type V1BankAccountController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type V1WithdrawalController interface {
	Preparation(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	TransferBalance(w http.ResponseWriter, r *http.Request)

	GetList(w http.ResponseWriter, r *http.Request)
	GetById(w http.ResponseWriter, r *http.Request)
	GetInsights(w http.ResponseWriter, r *http.Request)
	Export(w http.ResponseWriter, r *http.Request)
}

type V1InternalCustomerController interface {
	Create(w http.ResponseWriter, r *http.Request)
	CreateWalletCustomer(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	GetById(w http.ResponseWriter, r *http.Request)
	GetByPhoneNumber(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)
	GetByIDForUnifiedPayment(w http.ResponseWriter, r *http.Request)
}

type V1InternalXbController interface {
	GetFxRate(w http.ResponseWriter, r *http.Request)
	// Beneficiary
	CreateBeneficiary(w http.ResponseWriter, r *http.Request)
	UpdateBeneficiary(w http.ResponseWriter, r *http.Request)
	DeactivateBeneficiary(w http.ResponseWriter, r *http.Request)
	GetListBeneficiary(w http.ResponseWriter, r *http.Request)
	GetBeneficiaryById(w http.ResponseWriter, r *http.Request)

	// Sender
	CreateSender(w http.ResponseWriter, r *http.Request)
	GetListSender(w http.ResponseWriter, r *http.Request)
	GetSenderById(w http.ResponseWriter, r *http.Request)
	UpdateSender(w http.ResponseWriter, r *http.Request)
	DeactivateSender(w http.ResponseWriter, r *http.Request)

	// Payout
	CreatePayoutSession(w http.ResponseWriter, r *http.Request)
	UploadUnderlyingDocument(w http.ResponseWriter, r *http.Request)
	ConfirmPayoutSession(w http.ResponseWriter, r *http.Request)
	GetPayoutById(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)

	// Master
	GetListMasterCountry(w http.ResponseWriter, r *http.Request)
	GetListMasterCurrency(w http.ResponseWriter, r *http.Request)
	GetListMasterCurrencyMapping(w http.ResponseWriter, r *http.Request)
	GetListMasterIdentificationType(w http.ResponseWriter, r *http.Request)
	GetListMasterAccountType(w http.ResponseWriter, r *http.Request)
	GetListMasterPurpose(w http.ResponseWriter, r *http.Request)
	GetListMasterState(w http.ResponseWriter, r *http.Request)
	GetListMasterCity(w http.ResponseWriter, r *http.Request)
	GetListMasterTransferMethod(w http.ResponseWriter, r *http.Request)
	GetListMasterSourceOfIncome(w http.ResponseWriter, r *http.Request)

	GetRfiDetails(w http.ResponseWriter, r *http.Request)
	SubmitRfiDetails(w http.ResponseWriter, r *http.Request)

	// Proxy Handler, for simplify the process
	ProxyHandler(path string, headers map[string]string) http.HandlerFunc
	// ProxyHandlerWithQueryConversion converts camelCase query params to snake_case before proxying
	ProxyHandlerWithQueryConversion(path string, headers map[string]string) http.HandlerFunc
}

type V1InternalPaymentMethodController interface {
	GetVAPaymentMethods(w http.ResponseWriter, r *http.Request)
	TopUpVAPaymentMethod(w http.ResponseWriter, r *http.Request)
}

type V1CRMXbController interface {
	GetRfiDetails(w http.ResponseWriter, r *http.Request)
	SubmitRfiDetails(w http.ResponseWriter, r *http.Request)
	GetPayoutByID(w http.ResponseWriter, r *http.Request)
	ReConfirm(w http.ResponseWriter, r *http.Request)

	// Config
	GetListConfigSpread(w http.ResponseWriter, r *http.Request)
	GetConfigSpreadDetailByID(w http.ResponseWriter, r *http.Request)
	CreateConfigSpread(w http.ResponseWriter, r *http.Request)
	UpdateConfigSpread(w http.ResponseWriter, r *http.Request)
}

type V1CRMCreditcardController interface {
	Void(w http.ResponseWriter, r *http.Request)
	GetTransactionList(w http.ResponseWriter, r *http.Request)
	CreateMID(w http.ResponseWriter, r *http.Request)
	UpdateMID(w http.ResponseWriter, r *http.Request)
	GetMIDList(w http.ResponseWriter, r *http.Request)
	GetMIDMapList(w http.ResponseWriter, r *http.Request)
	BlockCard(w http.ResponseWriter, r *http.Request)

	// Proxy Handlers
	ProxyHandlerWithQueryConversion(path string, headers map[string]string) http.HandlerFunc
	ProxyHandler(path string, headers map[string]string) http.HandlerFunc
}

type V1CRMWithdrawalController interface {
	InquiryTransaction(w http.ResponseWriter, r *http.Request)
	RetryTransaction(w http.ResponseWriter, r *http.Request)
	ChangeStatusWithdrawal(w http.ResponseWriter, r *http.Request)
}

type V1InternalTransferController interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)
	GetById(w http.ResponseWriter, r *http.Request)
}

type V1PaymentController interface {
	GetPaymentInsight(w http.ResponseWriter, r *http.Request)
	GetPaymentDashboardInsights(w http.ResponseWriter, r *http.Request)
	FilterPaymentHistory(w http.ResponseWriter, r *http.Request)
	PaymentHistory(w http.ResponseWriter, r *http.Request)
	Export(w http.ResponseWriter, r *http.Request)
	GetChannelList(w http.ResponseWriter, r *http.Request)
	GetChannelListWithPaymentToken(w http.ResponseWriter, r *http.Request)
	GetChannelDocuments(w http.ResponseWriter, r *http.Request)
	UpdatePaymentMethodStatus(w http.ResponseWriter, r *http.Request)
	GetEncryptionKey(w http.ResponseWriter, r *http.Request)
	VCCTerminalBatchCharge(w http.ResponseWriter, r *http.Request)

	// GetPaymentDetailForPaymentUI for payment UI needs
	GetPaymentDetailForPaymentUI(w http.ResponseWriter, r *http.Request)
	GetPaymentImages(w http.ResponseWriter, r *http.Request)
	GetPaymentInstructions(w http.ResponseWriter, r *http.Request)
	ConfirmPayment(w http.ResponseWriter, r *http.Request)

	// Static QRIS Dashboard methods
	FilterStaticQrisList(w http.ResponseWriter, r *http.Request)
	GetStaticQrisDetail(w http.ResponseWriter, r *http.Request)
	GetStaticQrisTransactions(w http.ResponseWriter, r *http.Request)
	GetMaxActiveStaticQRPerMerchant(w http.ResponseWriter, r *http.Request)
	DeactivateStaticQris(w http.ResponseWriter, r *http.Request)

	// Static VA Dashboard methods
	GetVARangeList(w http.ResponseWriter, r *http.Request)
	UpdateVARange(w http.ResponseWriter, r *http.Request)

	// Static VA Dashboard methods
	FilterStaticVaList(w http.ResponseWriter, r *http.Request)
	GetStaticVaDetail(w http.ResponseWriter, r *http.Request)
	GetStaticVaTransactions(w http.ResponseWriter, r *http.Request)
	DeactivateStaticVa(w http.ResponseWriter, r *http.Request)

	CreatePaymentLink(w http.ResponseWriter, r *http.Request)

	// Cases Management (Investigation History) methods
	GetInvestigationList(w http.ResponseWriter, r *http.Request)
	GetInvestigationSummary(w http.ResponseWriter, r *http.Request)
	ExportInvestigation(w http.ResponseWriter, r *http.Request)
	GetVCCTerminalList(w http.ResponseWriter, r *http.Request)
}

type V1TransferController interface {
	FilterTransferHistory(w http.ResponseWriter, r *http.Request)
	GetTransferByID(w http.ResponseWriter, r *http.Request)
}

type V1LiveFeatureController interface {
	GetList(w http.ResponseWriter, r *http.Request)
	GetAppVersion(w http.ResponseWriter, r *http.Request)
}

type V1InternalUnifiedPaymentController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	FindPaymentByReferenceId(w http.ResponseWriter, r *http.Request)
}

type V1InternalRefundController interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
}

type V1RefundController interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	GetReceipt(w http.ResponseWriter, r *http.Request)
}

type V1InternalRecurringContractController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Cancel(w http.ResponseWriter, r *http.Request)
}

type V1InternalPlatformController interface {
	GetSubMerchantBalance(w http.ResponseWriter, r *http.Request)
}

type V2InternalUnifiedPaymentController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Confirm(w http.ResponseWriter, r *http.Request)
	Cancel(w http.ResponseWriter, r *http.Request)
	GetList(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Capture(w http.ResponseWriter, r *http.Request)

	// Payment Method
	GetPaymentMethodConfig(w http.ResponseWriter, r *http.Request)

	// Payment Charge
	GetChargeList(w http.ResponseWriter, r *http.Request)
	GetChargeByID(w http.ResponseWriter, r *http.Request)

	SimulatePayment(w http.ResponseWriter, r *http.Request)

	// Card Encryption
	EncryptCard(w http.ResponseWriter, r *http.Request)
	GetEncryptedCard(w http.ResponseWriter, r *http.Request)

	// BIN Lookup
	GetBinDetailByBinNumber(w http.ResponseWriter, r *http.Request)

	// Investigation
	UploadProofOfPayment(w http.ResponseWriter, r *http.Request)
}

type V1PlatformController interface {
	TransactionList(w http.ResponseWriter, r *http.Request)
	GetSubMerchantUserList(w http.ResponseWriter, r *http.Request)
}
type V1ProcessorCallbackController interface {
	FlipDisbursementCallback(w http.ResponseWriter, r *http.Request)
	FlipInquiryAccountCallback(w http.ResponseWriter, r *http.Request)

	DanaDisbursementCallback(w http.ResponseWriter, r *http.Request)
}
type V1ReconciliationController interface {
	GetList(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
	DownloadResult(w http.ResponseWriter, r *http.Request)
	UpdateReconDetail(w http.ResponseWriter, r *http.Request)
}
type V1CRMPaymentController interface {
	GetList(w http.ResponseWriter, r *http.Request)
	GetDetailByID(w http.ResponseWriter, r *http.Request)
	InquiryByID(w http.ResponseWriter, r *http.Request)
	GetSplitRoutingByTransferID(w http.ResponseWriter, r *http.Request)
	RetryNotification(w http.ResponseWriter, r *http.Request)
	RetryStaticVANotification(w http.ResponseWriter, r *http.Request)
	GetInvestigationList(w http.ResponseWriter, r *http.Request)
	GetInvestigationProofOfPayment(http.ResponseWriter, *http.Request)
	UpdateInvestigation(w http.ResponseWriter, r *http.Request)
	GetReceipt(w http.ResponseWriter, r *http.Request)
}
type V1CRMRefundController interface {
	Create(w http.ResponseWriter, r *http.Request)
}

type V1IPWhitelistController interface {
	GetList(w http.ResponseWriter, r *http.Request)
	Detail(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type V1CRMRateLimiterController interface {
	GetList(w http.ResponseWriter, r *http.Request)
	Detail(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type V1InternalBankAccountController interface {
	GetMerchantBankAccount(w http.ResponseWriter, r *http.Request)
}

type V1InternalFeeController interface {
	CalculateWhitelabelMerchantFee(w http.ResponseWriter, r *http.Request)
	CalculateWalletTransactionFee(w http.ResponseWriter, r *http.Request)
}

type V1WalletInsightController interface {
	TotalBalance(w http.ResponseWriter, r *http.Request)
}

type V1WalletTransactionController interface {
	GetMerchantTransactionHistoryList(w http.ResponseWriter, r *http.Request)
	ExportMerchantTransactionHistoryList(w http.ResponseWriter, r *http.Request)
	GetMerchantTransactionDetail(w http.ResponseWriter, r *http.Request)
}

type V1InternalFdsController interface {
	CheckTransaction(w http.ResponseWriter, r *http.Request)
	UpdateTransaction(w http.ResponseWriter, r *http.Request)
}

type V1CRMFraudRuleController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Detail(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
}

type V1CRMAmlController interface {
	Screening(w http.ResponseWriter, r *http.Request)
	Profile(w http.ResponseWriter, r *http.Request)
	UpdateDetailStatus(w http.ResponseWriter, r *http.Request)
}

type V1ChargesController interface {
	GetChargeByID(w http.ResponseWriter, r *http.Request)
	GetChargeList(w http.ResponseWriter, r *http.Request)
	Export(w http.ResponseWriter, r *http.Request)
}

type V1RecurringContractController interface {
	GetRecurringByID(w http.ResponseWriter, r *http.Request)
}

type V1CRMCountryController interface {
	GetAll(w http.ResponseWriter, r *http.Request)
}

type V1CRMCustomerController interface {
	GetCustomerList(w http.ResponseWriter, r *http.Request)
	GetCustomerByID(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type V1CRMIndustryController interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type V1CountryController interface {
	GetAll(w http.ResponseWriter, r *http.Request)
}

type V1IndustryController interface {
	GetAll(w http.ResponseWriter, r *http.Request)
}

type V1CardFundedPayoutController interface {
	// Saved Cards Group
	GetSavedCardList(w http.ResponseWriter, r *http.Request)
	CreateSavedCard(w http.ResponseWriter, r *http.Request)
	// Payout Actions
	CreatePayout(w http.ResponseWriter, r *http.Request)
	ApprovePayout(w http.ResponseWriter, r *http.Request)
	RejectPayout(w http.ResponseWriter, r *http.Request)
	// Payout List & Detail
	GetPayoutList(w http.ResponseWriter, r *http.Request)
	GetPayoutDetail(w http.ResponseWriter, r *http.Request)
	// Receipt
	GetReceipt(w http.ResponseWriter, r *http.Request)
	// Export
	ExportPayoutList(w http.ResponseWriter, r *http.Request)
	// Payout Insights
	GetPayoutInsights(w http.ResponseWriter, r *http.Request)
	// Vendors
	GetVendorList(w http.ResponseWriter, r *http.Request)
	GetVendorDetail(w http.ResponseWriter, r *http.Request)
	// Configs
	GetTransactionConfig(w http.ResponseWriter, r *http.Request)
}

type V1CRMFdsController interface {
	UpdateTransaction(w http.ResponseWriter, r *http.Request)
}

type V1CRMDukcapilController interface {
	VerifyIdentity(w http.ResponseWriter, r *http.Request)
}

type V1InternalWithdrawalController interface {
	Withdraw(w http.ResponseWriter, r *http.Request)
	GetBankAccountList(w http.ResponseWriter, r *http.Request)
	GetWithdrawalByID(w http.ResponseWriter, r *http.Request)
}

type V1CRMInstallmentPlanController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type V1CRMCallbackController interface {
	ResendCallback(w http.ResponseWriter, r *http.Request)
}

type V1ShortLinkController interface {
	GetByCode(w http.ResponseWriter, r *http.Request)
}

type V1CRMRoleController interface {
	AddDefaultRolePermissions(w http.ResponseWriter, r *http.Request)
	DeleteDefaultRolePermissions(w http.ResponseWriter, r *http.Request)
}

type V1CRMSettlementController interface {
	CreateSettlementHold(w http.ResponseWriter, r *http.Request)
}

type V1CRMVendorController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Detail(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
}

type V1CRMPayoutManualProcessingAccountController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
}

// V1CRMTNCController exposes CRM admin endpoints for managing TNC versions.
type V1CRMTNCController interface {
	Create(w http.ResponseWriter, r *http.Request)
	Activate(w http.ResponseWriter, r *http.Request)
	Deactivate(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Detail(w http.ResponseWriter, r *http.Request)
}

// V1TNCSigningController exposes merchant-side endpoints for signing TNC and
// querying signing status / history.
type V1TNCSigningController interface {
	Sign(w http.ResponseWriter, r *http.Request)
	Status(w http.ResponseWriter, r *http.Request)
	History(w http.ResponseWriter, r *http.Request)
}

type V1CRMCardFundedPayoutController interface {
	GetPayoutTransactionList(http.ResponseWriter, *http.Request)
	PatchPayoutTransactionStatus(http.ResponseWriter, *http.Request)
}
