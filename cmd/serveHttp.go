package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	netHttp "net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/cardFundedPayout"
	"github.com/paper-indonesia/pivot-backoffice/pkg/bigquery"
	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	"github.com/paper-indonesia/pivot-backoffice/pkg/dictionary"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/fds"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	jwtCore "github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	pkgMonitor "github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"

	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/encrypt"
	"github.com/paper-indonesia/pdk/v2/gcp"
	pdkGoff "github.com/paper-indonesia/pdk/v2/goff"
	pdkNotifier "github.com/paper-indonesia/pdk/v2/goff/notifier"
	pdkRetriever "github.com/paper-indonesia/pdk/v2/goff/retriever"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	pdkMySql "github.com/paper-indonesia/pdk/v2/mySqlExt"
	pdkNewRelic "github.com/paper-indonesia/pdk/v2/newRelicExt"
	"github.com/paper-indonesia/pdk/v2/otelExt"
	pdkRedis "github.com/paper-indonesia/pdk/v2/redisExt"

	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cobra"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"

	// Repository
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	accountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/account"
	accountInquriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/accountInquiries"
	accounttransactionRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/accountTransaction"
	activityRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/activity"
	adjustmentRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/adjustment"
	advanceAiRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/amlProcessor/advanceAiRepository"
	bankAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/bankAccount"
	beneficiaryAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/beneficiaryAccount"
	callbackRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/callback"
	countryRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/country"
	credSettingRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/credential"
	creditcardCoreProcessorRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/creditcardCoreProcessor"
	customerRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/customer"
	dailyAccountTransactionRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/dailyAccountTransaction"
	disbursementDatamart "github.com/paper-indonesia/pivot-backoffice/internal/repository/datamart/disbursement"
	paymentDatamart "github.com/paper-indonesia/pivot-backoffice/internal/repository/datamart/payment"
	disbursementRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	dukcapilGatewayRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/dukcapilGateway"
	fraudnetrepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fdsProcessor/fraudNetRepository"
	sokratechRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fdsProcessor/sokratech"
	feeRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fee"
	fraudrulesrepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fraudRules"
	inboundRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/inbound"
	industryRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/industry"
	installmentPlanRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/installmentPlan"
	ipWhitelistRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/ipWhitelist"
	liveFeatureRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/liveFeature"
	addrLocRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/location"
	menuRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/menu"
	merchantRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	merchantforbiddenusecaseRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchantForbiddenUsecase"
	merchantTopUpRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchantTopUp"
	outboundRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/outbound"
	passwordHistoriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/passwordHistories"
	paymentRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/payment"
	paymentCaptureRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/paymentCapture"
	paymentMethodRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/paymentMethod"
	payoutManualProcessingAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/payoutManualProcessingAccount"
	permissionRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/permission"
	productRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/product"
	qrisRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/qris"
	rateLimiterRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/rateLimiter"
	reconRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/reconciliation"
	recurringContractRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/recurringContract"
	refundRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/refund"
	reportingRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/reporting"
	requestAccountInquiryRepo "github.com/paper-indonesia/pivot-backoffice/internal/repository/requestAccountInquiry"
	roleRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/role"
	roleMenuPermissionRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/roleMenuPermission"
	ruleevaluationsrepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/ruleEvaluations"
	settlementHoldRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/settlementHold"
	shortLinkRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/shortLink"
	snapCoreRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/snapCore"
	statusHistoriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/statusHistories"
	tncRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/tnc"
	transferRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/transfer"
	userRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/user"
	userLoggedInDeviceRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/userLoggedInDevice"
	userRoleRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/userRole"
	vendorRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/vendor"
	walletTransactionRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/wallet/transaction"
	withdrawalRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/withdrawal"
	xbCoreProcessorRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/xbCoreProcessor"

	// Service
	accountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/account"
	accountinquirysvc "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/accountInquiry"
	activityService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/activity"
	adjustService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/adjustment"
	amlService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/aml"
	bankAccountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/bankAccount"
	beneficiaryAccountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/beneficiaryAccount"
	callbackService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	"github.com/paper-indonesia/pivot-backoffice/internal/service/v1/country"
	credSettingService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/credential"
	creditcardService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/creditcard"
	customerService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/customer"
	disbursementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	disbursementDashboardService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursementDashboard"
	dukcapilService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/dukcapil"
	fdsservice "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fds"
	feeService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"
	fraudruleservice "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fraudRule"
	inboundService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/inbound"
	industryService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/industry"
	installmentPlanService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/installmentPlan"
	ipwhitelistService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/ipWhitelist"
	liveFeatureService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/liveFeature"
	addrLocService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/location"
	menuService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/menu"
	merchantService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	merchantForbiddenUsecaseService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchantForbiddenUsecase"
	merchantTopUpService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchantTopUp"
	notificationService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/notification"
	orchestratorService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/orchestrator"
	otpService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/otp"
	paymentService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	paymentMethodService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/paymentMethod"
	payoutManualProcessingAccountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payoutManualProcessingAccount"
	permissionService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/permission"
	platformService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/platform"
	platformFeeService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/platformFee"
	productService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/product"
	qrisService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/qris"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/rateLimiter"
	reconService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/reconciliation"
	recurringContractService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/recurringContract"
	refundService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/refund"
	reportingService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/reporting"
	roleService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/role"
	routingprocessorService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/routingProcessor"
	settlementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/settlement"
	settlementHoldService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/settlementHold"
	shortLinkService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/shortLink"
	tncService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/tnc"
	transferService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/transfer"
	userService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/user"
	userLoggedInDeviceService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/userLoggedInDevice"
	userRoleService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/userRole"
	vendorService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/vendor"
	walletTransactionService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/wallet/transaction"
	walletInsightService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/walletInsight"
	withdrawalService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	xbPayoutService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/xbPayout"
	cancelMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/cancel"
	chargeMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/charge"
	ledgerService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/ledger"
	p2pMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/p2p"
	payInMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/payIn"
	payoutMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/payOut"
	refundMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/refundMoneyFlow"
	unifiedPaymentService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	callbackPartnerService "github.com/paper-indonesia/pivot-backoffice/pkg/callback"

	// Controller
	httpControllerUtil "github.com/paper-indonesia/pivot-backoffice/port/http/controller/util"
	v1AccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/account"
	v1ActivityController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/activity"
	bankController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/bank"
	bankAccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/bankAccount"
	beneficiaryAccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/beneficiaryAccount"
	callbackController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/callback"
	v1CardFundedPayoutController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/cardFundedPayout"
	v1ChargesController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/charges"
	countryController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/country"
	crmAccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/account"
	adjustController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/adjustment"
	crmAmlController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/aml"
	crmCallbackController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/callback"
	crmCardFundedPayoutController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/cardFundedPayout"
	crmCountryController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/country"
	crmCreditcardController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/creditcard"
	crmCustomerController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/customer"
	crmDisbursementController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/disbursement"
	crmDukcapilController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/dukcapil"
	crmfdscontroller "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/fds"
	crmFraudRuleController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/fraudRule"
	crmIndustryController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/industry"
	crmInstallmentPlanController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/installmentPlan"
	crmMerchantController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/merchant"
	crmmerchantforbiddenusecase "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/merchantForbiddenUsecase"
	v1CrmPaymentController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payment"
	crmPaymentMethodController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/paymentMethod"
	crmPayoutManualProcessingAccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payoutManualProcessingAccount"
	crmProductController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/product"
	crmRateLimiterController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/rateLimiter"
	crmRefundController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/refund"
	crmRoleController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/role"
	crmSettlementController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/settlementHold"
	crmTNCController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/tnc"
	crmUserController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/user"
	crmVendorController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/vendor"
	withdrawalCrmController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/withdrawal"
	crmXbController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/xb"
	disbursementController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/disbursement"
	industryController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/industry"
	internalAccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/account"
	internalBankAccountController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/bankAccount"
	internalFdsController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/fds"
	internalFeeController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/fee"
	internalPaymentMethodController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/paymentMethod"
	platformInternalController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/platform"
	internalTransferController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/transfer"
	internalWithdrawalsController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/withdrawals"
	ipWhitelistController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/ipWhitelist"
	liveFeatureController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/liveFeature"
	addrLocController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/location"
	menuController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/menu"
	merchantController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/merchant"
	merchantTopUpController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/merchantTopUp"
	orchestratorController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/orchestrator"
	otpController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/otp"
	paymentController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"
	paymentMethodController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/paymentMethod"
	platformController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/platform"
	v1ProcessorCallbackController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/processorCallback"
	purposeController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/purpose"
	qrisController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/qris"
	v1RecurringContractController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/recurring"
	v1RefundController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/refund"
	roleController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/role"
	apiLogController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/apiLog"
	callbackSettingController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/callback"
	credSettingController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/credential"
	depositSettingController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/setting/deposit"
	v1ShortLinkController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/shortLink"
	simulationController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/simulation"
	subMerchantController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/subMerchant"
	tncSigningController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/tnc"
	transferController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/transfer"
	userController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/user"
	walletInsightsController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/wallet/insights"
	walletTransactionController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/wallet/transaction"
	withdrawalController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/withdrawal"
	v1XbPayoutController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/xbPayout"
	ledgerController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/ledger"

	// Internal Controller
	internalAccountInquiry "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/accountInquiry"
	creditcardController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/creditcard"
	customerController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/customer"
	v1InternalMerchantController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/merchant"
	v1InternalMerchantAuthController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/merchantAuth"
	v1InternalPaymentController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/payment"
	v1InternalPayoutController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/payout"
	v1InternalRecurringContractController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/recurringContract"
	v1InternalRefundController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/refund"
	internalSubMerchant "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/subMerchant"
	v1InternalUnifiedPaymentController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/unifiedPayment"
	v1InternalXbController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/xb"
	v1ReconController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/reconciliation"
	v2InternalUnifiedPaymentController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v2/internalController/unifiedPayment"

	"github.com/paper-indonesia/pivot-backoffice/port/http"
)

func init() {
	rootCmd.AddCommand(serveHttpCmd)
}

func initHttpPkg(ctx context.Context) (
	cfg *config.Config,
	secret *config.Secret,
	pdkLog pdkLogger.ILogger,
	log logger.ILogger,
	otel otelExt.IOtelExt,
	nr pdkNewRelic.INewRelicExt,
	monitor *monitoring.Monitor,
	mySqlDB mySqlExt.IMySqlExt,
	redisClient redisExt.IRedisExt,
	rabbitMq rabbitMqExt.IRabbitMQExt,
	gcsClient gcs.IGCSService,
	httpRequestClient httpRequestExt.IHTTPRequest,
	jwtConfig jwtCore.IJwt,
	validate *validatorExt.Validate,
	encryptExt encryption.ICrypto,
	encryptGcs encryption.GCSClient,
	vaultClient *vault.Client,
	bqClient bigquery.IBigQueryService,
) {
	var (
		err error

		closes = make([]func(), 0)
	)

	defer func() {
		if r := recover(); r != nil {
			for _, close := range closes {
				close()
			}
			fmt.Printf("Panic occurred: %v\n", r)
			panic(r)
		}
	}()

	// Init config
	cfg, secret, err = config.LoadConfig(cfgFile, scrtFile)
	if err != nil {
		fmt.Printf("Unable to load configuration and secret: %v", err)
		panic(err)
	}

	// Set isDevelopment
	isDevelopment := true
	if cfg.Environment == constant.EnvironmentProduction {
		isDevelopment = false
	}

	// Init Logger
	if cfg.AppConfig.PdkLoggerUsed == constant.PdkLoggerSloggerName {
		pdkLog = pdkLogger.NewSlogger(
			pdkLogger.Config{
				IsDevelopment: isDevelopment,
				Environment:   cfg.Environment,
				ServiceName:   getServiceName(cfg.ServiceName),
			},
		)
	} else {
		pdkLog, err = pdkLogger.NewZapLogger(
			pdkLogger.Config{
				IsDevelopment: isDevelopment,
				Environment:   cfg.Environment,
				ServiceName:   getServiceName(cfg.ServiceName),
			},
			pdkLogger.WithZapMaskingSensitiveData(strings.Split(cfg.AppConfig.MaskingSensitiveData, ",")),
		)
		if err != nil {
			fmt.Printf("Unable to init pdk logger, %v", err)
			panic(err)
		}
	}
	closes = append(closes, func() {
		if syncErr := pdkLog.Sync(); syncErr != nil {
			fmt.Printf("Error syncing pdk logger: %v\n", syncErr)
		}
	})

	// Deprecated: Use pdkLogger
	log, err = logger.New(
		logger.Config{
			Environment: cfg.Environment,
			ServiceName: cfg.ServiceName,
		},
		logger.WithMaskingSensitiveData(strings.Split(cfg.AppConfig.MaskingSensitiveData, ",")),
	)
	if err != nil {
		fmt.Printf("Unable to init logger, %v", err)
		panic(err)
	}
	closes = append(closes, func() {
		if syncErr := log.Sync(); syncErr != nil {
			fmt.Printf("Error syncing logger: %v\n", syncErr)
		}
	})

	// Init Feature Flag
	consulRetriever, err := pdkRetriever.NewConsulRetriever(
		cfg.FeatureFlagConfig.ConsulAddr,
		cfg.FeatureFlagConfig.ConsulConfigPath,
		secret.ConsulSecret.Token,
	)
	if err != nil {
		fmt.Printf("Unable to init goff - consul retriever, %v", err)
		panic(err)
	}
	logNotifier := pdkNotifier.NewLoggerNotifier(pdkLog)

	ffconfig, err := pdkGoff.NewGoff(pdkGoff.Config{
		PollingInterval:             time.Duration(cfg.FeatureFlagConfig.PollingInterval) * time.Second,
		EnablePollingJitter:         false,
		Logger:                      pdkLog,
		Context:                     context.Background(),
		Environment:                 cfg.Environment,
		Retriever:                   consulRetriever,
		Notifiers:                   []notifier.Notifier{logNotifier},
		FileFormat:                  pdkGoff.FileFormatYAML,
		Offline:                     cfg.FeatureFlagConfig.Offline,
		EvaluationContextEnrichment: nil,
		DataExporter: &logsexporter.Exporter{
			LogFormat: `goffExporter: kind={{ .Kind}}, contextKind={{ .ContextKind}}, user={{ .UserKey}}, key={{ .Key}}, variation={{ .Variation}}, value={{ .Value}}, default={{ .Default}}`,
		},
		NotifierSlackWebhookURL: cfg.FeatureFlagConfig.ExporterSlackWebhookURL,
	})
	if err != nil {
		fmt.Printf("Unable to init feature flag config, %v", err)
		panic(err)
	}
	if err := ffclient.Init(ffconfig); err != nil {
		fmt.Printf("Unable to init feature flag client, %v", err)
		panic(err)
	}
	closes = append(closes, func() { ffclient.Close() })

	// Observability
	otelOpts := []otelExt.OptionFunc{}
	if cfg.OTLPConfig.Insecure {
		otelOpts = append(otelOpts, otelExt.WithInsecure())
	}
	if cfg.OTLPConfig.TLSClientConfig != nil {
		otelOpts = append(otelOpts, otelExt.WithTLSClientConfig(&tls.Config{
			InsecureSkipVerify: cfg.OTLPConfig.TLSClientConfig.InsecureSkipVerify,
		}))
	}
	// Init Open Telemetry
	otel, err = otelExt.New(
		otelExt.Config{
			ServiceName:  getServiceName(cfg.ServiceName),
			Environment:  cfg.Environment,
			OTLPEndpoint: cfg.OTLPConfig.Host,
			LicenseKey:   secret.NewRelicLicenseKey,
			MetricConfig: otelExt.MetricConfig{
				MetricInterval: time.Duration(cfg.OTLPConfig.MetricConfig.Interval) * time.Second,
				MetricTimeout:  time.Duration(cfg.OTLPConfig.MetricConfig.ExportTimeout) * time.Second,
				DropMetricConfigs: []otelExt.MetricViewConfig{
					{
						InstrumentMetricName: otelExt.OtelMetricPrefixHttp,
					},
					{
						InstrumentMetricName: otelExt.OtelMetricPrefixMysql,
					},
					{
						InstrumentMetricName: otelExt.OtelMetricPrefixRedis,
					},
				},
				MetricTemporality: otelExt.MetricTemporalityDelta,
			},
		}, otelOpts...,
	)
	if err != nil {
		fmt.Printf("Unable to init opentelemetry, %v", err)
		panic(err)
	}
	closes = append(closes, func() {
		if shutdownErr := otel.Shutdown(ctx); shutdownErr != nil {
			fmt.Printf("Error shutting down otel: %v\n", shutdownErr)
		}
	})
	customMetric.SetOtelExt(otel)

	// Init New Relic
	nr, err = pdkNewRelic.New(
		pdkNewRelic.Config{
			ServiceName: getServiceName(cfg.ServiceName) + "-" + cfg.Environment,
			Environment: cfg.Environment,
			LicenseKey:  secret.NewRelicLicenseKey,
		},
		pdkNewRelic.WithExcludeAttributes(strings.Split(cfg.AppConfig.MaskingSensitiveData, ",")),
	)
	if err != nil {
		fmt.Printf("Unable to init new relic, %v", err)
		panic(err)
	}

	// Deprecated: Use otelExt metric provider
	// Init PDK Monitoring
	if cfg.MonitoringConfig.IsEnabled {
		monitor, err = monitoring.New(
			getServiceName(cfg.ServiceName)+"-"+cfg.Environment,
			secret.StatsdHost,
			secret.StatsdPort,
			monitoring.WithStatsdOptions([]statsd.Option{
				statsd.WithMaxBytesPerPayload(cfg.MonitoringConfig.MaxBytesPerPayload),
				statsd.WithMaxMessagesPerPayload(cfg.MonitoringConfig.MaxMessagesPerPayload),
				statsd.WithWriteTimeout(time.Duration(cfg.MonitoringConfig.WriteTimeout) * time.Second),
			}),
		)
		if err != nil {
			fmt.Printf("Unable to init monitoring, %v", err)
			panic(err)
		}

		pkgMonitor.SetGlobalMonitoring(monitor)
	}

	// Init MySql Database
	mySqlDB, err = mySqlExt.New(
		pdkMySql.Config{
			Host:         cfg.MySQLConfig.Host,
			Port:         cfg.MySQLConfig.Port,
			Username:     secret.MySQLSecret.Username,
			Password:     secret.MySQLSecret.Password,
			DBName:       secret.MySQLSecret.Database,
			MaxIdleConns: cfg.MySQLConfig.MaxIdleConns,
			MaxIdleTime:  cfg.MySQLConfig.MaxOpenConns,
			MaxLifeTime:  cfg.MySQLConfig.MaxLifeTime,
			MaxOpenConns: cfg.MySQLConfig.MaxOpenConns,
			SlaveHost:    cfg.MySQLConfig.SlaveHost,
			SlavePort:    cfg.MySQLConfig.SlavePort,
		},
		pdkMySql.WithLogger(pdkLog),
		pdkMySql.WithTracerProvider(otel.TracerProvider()),
		pdkMySql.WithMetricProvider(otel.MeterProvider()),
	)
	if err != nil {
		fmt.Printf("Unable to init mysql, %v", err)
		panic(err)
	}
	closes = append(closes, func() { mySqlDB.Close() })

	// Init Redis
	redisClient, err = redisExt.New(
		pdkRedis.Config{
			Addr:             cfg.RedisConfig.Host + ":" + cfg.RedisConfig.Port,
			Password:         secret.RedisSecret.Password,
			DB:               cfg.RedisConfig.CacheDB,
			IsRedsyncEnabled: true,
		},
		pdkRedis.WithTracerProvider(otel.TracerProvider()),
		pdkRedis.WithMetricProvider(otel.MeterProvider()),
		pdkRedis.WithMaxRetries(cfg.RedisConfig.MaxRetries),
		pdkRedis.WithMinRetryBackoff(time.Duration(cfg.RedisConfig.MinRetryBackoff)*time.Second),
		pdkRedis.WithMaxRetryBackoff(time.Duration(cfg.RedisConfig.MaxRetryBackoff)*time.Second),
		pdkRedis.WithDialTimeout(time.Duration(cfg.RedisConfig.DialTimeout)*time.Second),
		pdkRedis.WithReadTimeout(time.Duration(cfg.RedisConfig.ReadTimeout)*time.Second),
		pdkRedis.WithWriteTimeout(time.Duration(cfg.RedisConfig.WriteTimeout)*time.Second),
		pdkRedis.WithPoolSize(cfg.RedisConfig.PoolSize),
		pdkRedis.WithPoolTimeout(time.Duration(cfg.RedisConfig.PoolTimeout)*time.Second),
	)
	if err != nil {
		fmt.Printf("Unable to init redis cache, %v", err)
		panic(err)
	}
	closes = append(closes, func() { redisClient.Close() })

	// Init RabbitMQ
	// TODO: Add to PDK V2
	cfg.RabbitMQConfig.ServiceName = getServiceName(cfg.ServiceName)
	//
	rabbitMq, err = rabbitMqExt.New(
		cfg.RabbitMQConfig,
		secret.RabbitMQSecret,
		pdkLog,
		nr,
	)
	if err != nil {
		fmt.Printf("Unable to init rabbitmq, %v", err)
		panic(err)
	}
	closes = append(closes, func() { rabbitMq.Close() })

	// Init GCS
	// TODO: Add to PDK V2
	gcsClient = gcs.NewGCSService(gcs.Config{
		ServiceBucketName:          cfg.GCSConfig.ServiceBucketName,
		ReportingBucketName:        cfg.GCSConfig.ReportingBucketName,
		BulkDisbursementBucketName: cfg.GCSConfig.BulkDisbursementBucketName,
		ProofOfTransferFolderName:  cfg.GCSConfig.ProofOfTransferFolderName,
	})
	closes = append(closes, func() { gcsClient.Close() })

	// Init Secret Manager
	gsmClient, err := gcp.NewSecretManagerClient(ctx)
	if err != nil {
		log.Panic(ctx, "Unable to init google secret manager: "+err.Error())
	}
	gcp.SetGlobalSecretManagerClient(gsmClient)
	closes = append(closes, func() { gsmClient.Close() })

	// Init Dictionary
	// TODO: Add to PDK V2
	dictionary.Dict, err = dictionary.New(cfg.DictionaryConfig)
	if err != nil {
		fmt.Printf("Unable to init dictionary, %v", err)
		panic(err)
	}

	// Init HTTP Request Client
	httpRequestClient = httpRequestExt.New(
		httpRequestExt.WithLogger(pdkLog),
		httpRequestExt.WithOutbound(outboundRepository.New(mySqlDB)),
		httpRequestExt.WithMaskingSensitiveData(strings.Split(cfg.AppConfig.MaskingSensitiveData, ",")),
	)

	// Init JWT Config
	if jwtConfig, err = jwtCore.New(cfg, secret, redisClient); err != nil {
		panic("Unable to init JWT package, " + err.Error())
	}

	// Init Vault Client
	vaultClient, err = vault.New(vault.Config{
		Address: cfg.Vault.Address,
		Token:   secret.Vault.Token,
	})
	if err != nil {
		panic("Unable to init Vault client, " + err.Error())
	}

	// Init BigQuery Client
	bqConfig := bigquery.Config{
		ProjectID:           cfg.BigQueryConfig.ProjectID,
		Location:            cfg.BigQueryConfig.Location,
		QueryTimeoutSeconds: cfg.BigQueryConfig.QueryTimeoutSeconds,
		MaxRetries:          cfg.BigQueryConfig.MaxRetries,
	}
	bqClient = bigquery.NewBigQueryService(bqConfig)

	// Init Validator
	validate = validatorExt.New()

	// Init Encryption
	encryptExt = encryption.New()

	// Init Encryption Gcs
	encryptGcs = encryption.NewGCS(secret)

	return cfg, secret, pdkLog, log, otel, nr, monitor, mySqlDB, redisClient, rabbitMq, gcsClient, httpRequestClient, jwtConfig, validate, encryptExt, encryptGcs, vaultClient, bqClient
}

// TODO: Tech Debt
// 1. Logger need to move to pdkLogger, currently only to support PDK V2
// 2. Redis need to refactor to not have extra layer in pkg, right now focusing on tracer and metric centralize in PDK

var serveHttpCmd = &cobra.Command{
	Use:   "serveHttp",
	Short: "Start HTTP server",
	Long:  `Start Boilerplate HTTP server`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// Init HTTP PKG
		config, secret, pdkLog, logger, otelExt, newRelicExt, monitor, dbClient, cacheClient, rabbitMqExt, gcs, httpExt, jwtConfig, validate, encryptExt, _, vaultClient, bqClient := initHttpPkg(ctx)
		defer func() {
			if syncErr := pdkLog.Sync(); syncErr != nil {
				fmt.Printf("Error syncing pdk logger: %v\n", syncErr)
			}
		}()
		defer pdkLog.Info(ctx, "Service successfully stopped")
		defer func() {
			if syncErr := logger.Sync(); syncErr != nil {
				fmt.Printf("Error syncing logger: %v\n", syncErr)
			}
		}()
		defer ffclient.Close()
		defer func() {
			if shutdownErr := otelExt.Shutdown(ctx); shutdownErr != nil {
				fmt.Printf("Error shutting down otel: %v\n", shutdownErr)
			}
		}()
		defer newRelicExt.GetApp().Shutdown(10 * time.Second)
		defer dbClient.Close()
		defer cacheClient.Close()
		defer rabbitMqExt.Close()
		defer gcs.Close()

		// Vault KV
		userEncryptionKey := vaultClient.NewKeyValue(
			config.Vault.Secrets.UserEncryptionKey.MountPath,
			config.Vault.Secrets.UserEncryptionKey.SecretPath,
		)
		paymentEncryptionKey := vaultClient.NewKeyValue(
			config.Vault.Secrets.PaymentEncryptionKey.MountPath,
			config.Vault.Secrets.PaymentEncryptionKey.SecretPath,
		)
		paymentCryptoAesGcm, err := encrypt.NewAesGcmCipher(secret.Payment.KeyEncryptionKey, encrypt.KeyFormat(-1))
		if err != nil {
			panic("Unable to init payment encryption tool, error " + err.Error())
		}
		merchantCredsEncryption := vaultClient.NewTransit(config.Vault.Secrets.MerchantCredentials.SecretPath, config.Vault.Secrets.MerchantCredentials.SecretKey)

		// Conductor Workflow
		var conductorAuthentication conductor.Authentication
		if secret.Conductor.BasicAuth != nil {
			conductorAuthentication = &conductor.BasicAuthentication{
				Username: secret.Conductor.BasicAuth.Username,
				Password: secret.Conductor.BasicAuth.Password,
			}
		}
		conductorClient, err := conductor.NewClient(conductor.Config{
			BaseURL:        config.Conductor.Address,
			Authentication: conductorAuthentication,
		})
		if err != nil {
			panic("Unable init conductor client, error: " + err.Error())
		}

		// Init Disbursement Worker Pools
		disbursementWorkers := ffcontext.NewEvaluationContext(config.Environment)
		disbursementWorkers.AddCustomAttribute("environment", config.Environment)

		disbursementWorkerPools, err := ffclient.IntVariation("backend-portal-disbursement-total-worker-pool", disbursementWorkers, config.WorkerPoolConfig.Disbursement)
		if err != nil {
			logger.Warn(context.Background(), "failed to get total worker pool", zap.Error(err))
			disbursementWorkerPools = config.WorkerPoolConfig.Disbursement
		}

		// Init repository
		// e.g. database, external/internal services repository, etc.
		snapCoreRepository := snapCoreRepository.New(config, secret, pdkLog, httpExt)

		activityRepositoryFactory := &activityRepository.ActivityRepository{
			Mongo:  nil,
			Mysql:  dbClient,
			Logger: pdkLog,
		}
		activityRepository := activityRepositoryFactory.CreateRepository(activityRepository.MySQLType)
		userRepository := userRepository.New(dbClient, pdkLog)
		passHistory := passwordHistoriesRepository.New(dbClient, pdkLog)
		merchantRepo := merchantRepository.New(dbClient, pdkLog, merchantRepository.WithServiceConfig(config))
		roleRepo := roleRepository.New(dbClient, pdkLog)
		userRoleRepo := userRoleRepository.New(dbClient, pdkLog)
		paymentRepository := paymentRepository.New(dbClient, pdkLog,
			paymentRepository.WithAppConfig(&config.AppConfig),
		)
		paymentMethodRepo := paymentMethodRepository.New(dbClient, pdkLog)
		callbackRepo := callbackRepository.New(dbClient, pdkLog)
		customerRepo := customerRepository.New(dbClient, pdkLog)
		beneficiaryAccountRepo := beneficiaryAccountRepository.New(dbClient, pdkLog)
		accountTransactionRepo := accounttransactionRepository.New(dbClient, pdkLog,
			accounttransactionRepository.WithAppConfig(&config.AppConfig),
		)
		accountRepo := accountRepository.New(dbClient, pdkLog)
		disbursementRepo := disbursementRepository.New(dbClient, pdkLog,
			disbursementRepository.WithConfig(&config.DisbursementConfig),
			disbursementRepository.WithAppConfig(&config.AppConfig),
		)
		statusHistoriesRepo := statusHistoriesRepository.New(dbClient)
		paymentCaptureRepo := paymentCaptureRepository.New(dbClient, pdkLog)
		accountInquiriesRepo := accountInquriesRepository.New(dbClient, pdkLog)
		permissionRepo := permissionRepository.New(dbClient, pdkLog)
		menuRepo := menuRepository.New(dbClient, pdkLog)
		roleMenuPermRepo := roleMenuPermissionRepository.New(dbClient, pdkLog)
		credSettingRepo := credSettingRepository.New(dbClient)
		reqAccountInquiryRepo := requestAccountInquiryRepo.New(dbClient, pdkLog)
		forbiddenUsecaseRepo := merchantforbiddenusecaseRepository.New(dbClient, pdkLog)
		userLoggedInDeviceRepo := userLoggedInDeviceRepository.New(dbClient, pdkLog)
		addrLocRepo := addrLocRepository.New(dbClient)
		qrisRepository := qrisRepository.New(dbClient)
		feeRepo := feeRepository.New(dbClient, pdkLog)
		xbCoreProcessorRepo := xbCoreProcessorRepository.New(config, secret, pdkLog, httpExt)
		creidtcardCoreProcessorRepo := creditcardCoreProcessorRepository.New(config, secret, pdkLog, httpExt)
		bankAccountRepo := bankAccountRepository.New(dbClient, pdkLog)
		withdrawalRepo := withdrawalRepository.New(dbClient)
		transferRepo := transferRepository.New(dbClient, pdkLog)
		productRepo := productRepository.New(dbClient, pdkLog)
		liveFeatureRepo := liveFeatureRepository.New(dbClient, pdkLog)
		reconRepo := reconRepository.New(dbClient, pdkLog)
		ipWhitelistRepo := ipWhitelistRepository.New(pdkLog, dbClient)
		industryRepo := industryRepository.New(dbClient, pdkLog)
		rateLimiterRepository := rateLimiterRepository.New(dbClient, pdkLog)
		outboundRepo := outboundRepository.New(dbClient)
		dailyAccountTrxRepo := dailyAccountTransactionRepository.New(dbClient, pdkLog)
		merchantTopUpRepo := merchantTopUpRepository.New(dbClient, pdkLog,
			merchantTopUpRepository.WithAppConfig(&config.AppConfig),
		)
		walletTransactionRepo := walletTransactionRepository.New(dbClient)
		inboundRepo := inboundRepository.New(dbClient)
		fraudRulesRepo := fraudrulesrepository.New(pdkLog, dbClient)
		vendorRepo := vendorRepository.New(pdkLog, dbClient)
		payoutManualProcessingAccountRepo := payoutManualProcessingAccountRepository.New(pdkLog, dbClient)
		tncRepo := tncRepository.New(pdkLog, dbClient)
		ruleEvaluationsRepo := ruleevaluationsrepository.New(pdkLog, dbClient)
		fraudNetRepo := fraudnetrepository.New(config, secret, pdkLog, httpExt)
		refundRepo := refundRepository.New(dbClient, pdkLog)
		recurringContractRepo := recurringContractRepository.New(pdkLog, dbClient)
		advanceAiRepo := advanceAiRepository.New(config, secret, pdkLog, httpExt)
		dukcapilGatewayRepo := dukcapilGatewayRepository.New(config, secret, pdkLog, httpExt)
		countryRepo := countryRepository.New(dbClient, pdkLog)
		installmentPlanRepo := installmentPlanRepository.New(dbClient, pdkLog)
		shortLinkRepo := shortLinkRepository.New(dbClient, pdkLog)
		paymentDatamartRepo := paymentDatamart.New(config.BigQueryConfig, bqClient, pdkLog)
		disbursementDatamartRepo := disbursementDatamart.New(config.BigQueryConfig, bqClient, pdkLog)
		sokratechRepo := sokratechRepository.New(config.Sokratech, secret.Sokratech, httpExt, pdkLog)
		settlementHoldRepo := settlementHoldRepository.New(dbClient, pdkLog)
		reportingRepo := reportingRepository.New(dbClient, pdkLog, config.AppConfig)

		liveFeatureRepo.WithConfig(config)
		liveFeatureRepo.WithSecret(secret)

		paymentRepository.WithConfig(config)
		paymentRepository.WithSecret(secret)
		// Init External Services
		merchantCallbackHTTPClient := httputil.NewHTTPClient(
			httputil.ServiceConfig(config.HTTPClients.MerchantCallback), httputil.WithLogger(pdkLog),
		)
		defer merchantCallbackHTTPClient.CloseIdleConnections()
		callbackPartnerService := callbackPartnerService.New(logger, merchantCallbackHTTPClient)

		// Init services
		notificationSvc := notificationService.New(config, pdkLog, rabbitMqExt)
		accountSvc := accountService.New(pdkLog, accountTransactionRepo, accountRepo, dailyAccountTrxRepo)
		otpSvc := otpService.New(
			config, pdkLog, cacheClient, jwtConfig, rabbitMqExt, userRepository, redisExt.NewLimiter(cacheClient.Client()),
		)
		otpSvc.WithUserEncryptionKey(userEncryptionKey)

		productSvc := productService.New(pdkLog, productRepo, productService.WithMerchantRepo(merchantRepo))
		userLoggedInDeviceSvc := userLoggedInDeviceService.New(config, secret, pdkLog, userLoggedInDeviceService.Repositories{
			UserRepo:               userRepository,
			UserLoggedInDeviceRepo: userLoggedInDeviceRepo,
		})
		permissionSvc := permissionService.New(permissionRepo, pdkLog, permissionService.WithRedisClient(cacheClient))
		menuSvc := menuService.New(menuRepo, pdkLog, menuService.WithProductService(productSvc))
		roleSvc := roleService.New(
			roleRepo, pdkLog,
			roleService.WithMenuRepository(menuRepo), roleService.WithRoleMenuPermissionRepository(roleMenuPermRepo), roleService.WithUserRoleRepository(userRoleRepo),
			roleService.WithRedisClient(cacheClient),
		)
		rateLimiterService := ratelimiter.New(pdkLog, cacheClient, rateLimiterRepository, ratelimiter.WithRedisLimiter(redisExt.NewLimiter(cacheClient.Client())), ratelimiter.WithConfig(config))
		userRoleSvc := userRoleService.New(userRoleRepo, pdkLog)
		activityService := activityService.New(activityRepository)
		tncSvc := tncService.New(
			tncRepo, merchantRepo, pdkLog,
			tncService.WithGCSService(gcs),
			tncService.WithActivityService(activityService),
			tncService.WithConfig(config),
		)
		userSvc := userService.New(
			config, secret, pdkLog, userRepository, passHistory,
			userService.WithJWT(jwtConfig),
			userService.WithRedisClient(cacheClient),
			userService.WithRateLimiter(rateLimiterService),
			userService.WithRabbitMQClient(rabbitMqExt),
			userService.WithOTPService(otpSvc),
			userService.WithUserLoggedInDeviceService(userLoggedInDeviceSvc),
			userService.WithPermissionService(permissionSvc),
			userService.WithLimiter(redisExt.NewLimiter(cacheClient.Client())),
			userService.WithRoleService(roleSvc),
			userService.WithUserRoleService(userRoleSvc),
			userService.WithUserLoggedInDeviceRepo(userLoggedInDeviceRepo),
			userService.WithEncryptionKey(userEncryptionKey),
			userService.WithTNCService(tncSvc),
		)
		orchestratorSvc := orchestratorService.New(pdkLog, gcs, accountTransactionRepo, accountRepo,
			orchestratorService.WithRedisClient(cacheClient),
		)
		orchestratorService.WithAccountService(orchestratorSvc, accountSvc)
		orchestratorService.WithReportingRepository(orchestratorSvc, reportingRepo)

		// Routing Processor Service
		routingProcessors := map[string]repository.IRoutingProcessorRepository{
			constant.SnapCoreProcessor: snapCoreRepository,
		}
		routingProcessorSvc := routingprocessorService.New(config, pdkLog, routingProcessors,
			routingprocessorService.WithOutboundRepository(outboundRepo),
			routingprocessorService.WithRabbitMqExt(rabbitMqExt),
			routingprocessorService.WithRequestAccountInquiryRepository(reqAccountInquiryRepo),
		)

		beneficiaryAccountSvc := beneficiaryAccountService.New(pdkLog, beneficiaryAccountRepo, accountInquiriesRepo, snapCoreRepository,
			beneficiaryAccountService.WithRoutingProcessorService(routingProcessorSvc),
			beneficiaryAccountService.WithConfig(config),
		)

		qrisService := qrisService.New(
			pdkLog, qrisRepository, merchantRepo, snapCoreRepository,
			qrisService.WithGCSService(gcs),
			qrisService.WithServiceConfig(config),
			qrisService.WithPDFGenerator(pdf.NewPDFGenerator(
				pdf.WithGCSService(gcs),
			)),
		)

		industrySvc := industryService.NewIndustryService(industryRepo, pdkLog)
		countrySvc := country.New(countryRepo, pdkLog)

		merchantSvc := merchantService.New(
			merchantRepo, pdkLog, userRepository, jwtConfig, rabbitMqExt, encryptExt,
			merchantService.WithGCSService(gcs),
			merchantService.WithServiceConfig(config),
			merchantService.WithAccountService(accountSvc),
			merchantService.WithAccountRepository(accountRepo),
			merchantService.WithRedisClient(cacheClient),
			merchantService.WithLocationRepository(addrLocRepo),
			merchantService.WithUserService(userSvc),
			merchantService.WithFeeCalculation(&feeService.FeeService{}),
			merchantService.WithOrchestratorService(orchestratorSvc),
			merchantService.WithBankAccountRepository(bankAccountRepo),
			merchantService.WithProductRepository(productRepo),
			merchantService.WithBeneficiaryAccountRepo(beneficiaryAccountRepo),
			merchantService.WithBeneficiaryAccountService(beneficiaryAccountSvc),
			merchantService.WithPaymentMethodRepository(paymentMethodRepo),
			merchantService.WithQrisService(qrisService),
			merchantService.WithIndustryService(industrySvc),
			merchantService.WithCountryService(countrySvc),
			merchantService.WithValidator(validate),
			merchantService.WithExcelLibrary(xlsx.New()),
			merchantService.WithVaultTransit(merchantCredsEncryption),
			merchantService.WithSnapCoreRepo(snapCoreRepository),
			merchantService.WithMerchantTopUpRepo(merchantTopUpRepo),
		)
		accountService.WithMerchantService(accountSvc, merchantSvc)

		callbackSvc := callbackService.New(
			pdkLog, cacheClient, callbackRepo, callbackPartnerService, merchantSvc,
			callbackService.WithUserService(userSvc),
			callbackService.WithRabbitMQExt(rabbitMqExt),
			callbackService.WithMerchantRepository(merchantRepo),
			callbackService.WithVaultTransit(merchantCredsEncryption),
		)

		liveFeatureSvc := liveFeatureService.New(pdkLog, liveFeatureRepo, rabbitMqExt)

		paymentMethodSvc := paymentMethodService.New(pdkLog, paymentMethodRepo, snapCoreRepository, creidtcardCoreProcessorRepo,
			paymentMethodService.WithQrisService(qrisService),
			paymentMethodService.WithMerchantService(merchantSvc),
			paymentMethodService.WithConfig(config),
			paymentMethodService.WithRedisClient(cacheClient),
			paymentMethodService.WithMerchantRepository(merchantRepo),
			paymentMethodService.WithPaymentRepository(paymentRepository),
		)
		feeSvc := feeService.New(pdkLog, feeRepo, merchantRepo, feeService.WithPaymentMethodService(paymentMethodSvc), feeService.WithRedisClient(cacheClient), feeService.WithConfig(config))

		merchantForbiddenUsecaseSvc := merchantForbiddenUsecaseService.New(pdkLog, forbiddenUsecaseRepo, rabbitMqExt, merchantSvc)
		disbursementDashboardSvc := disbursementDashboardService.New(pdkLog, disbursementRepo, accountTransactionRepo, accountRepo, orchestratorSvc)

		customerService := customerService.New(customerRepo, accountSvc, pdkLog)

		// Bind account service with customer service
		accountService.WithCustomerService(accountSvc, customerService)

		adjustSvc := adjustService.New(config.SlackConfig, adjustmentRepository.New(dbClient), merchantRepo)
		adjustService.WithGCSService(adjustSvc, gcs)
		adjustService.WithRabbitMQ(adjustSvc, rabbitMqExt)
		adjustService.WithOrchestratorService(adjustSvc, orchestratorSvc)
		adjustService.WithLogger(adjustSvc, pdkLog)
		adjustService.WithAccountRepository(adjustSvc, accountRepo)
		credSettingSvc := credSettingService.New(
			pdkLog, credSettingRepo, rabbitMqExt, credSettingService.WithUserService(userSvc), credSettingService.WithVaultTransit(merchantCredsEncryption),
		)
		xbPayoutSvc := xbPayoutService.New(pdkLog, disbursementRepo, beneficiaryAccountRepo,
			xbCoreProcessorRepo,
			xbPayoutService.WithFeeService(feeSvc),
			xbPayoutService.WithOrchestratorService(orchestratorSvc),
			xbPayoutService.WithRabbitMQClient(rabbitMqExt),
			xbPayoutService.WithGCS(gcs),
			xbPayoutService.WithConfig(config),
			xbPayoutService.WithStatusHistories(statusHistoriesRepo),
		)

		ledgerSvc := ledgerService.New(pdkLog, accountTransactionRepo, accountRepo, merchantSvc, customerService, accountSvc)
		payInMoneyFlowSvc := payInMoneyFlowService.New(pdkLog, accountTransactionRepo, accountSvc, merchantSvc)
		p2pMoneyFlowSvc := p2pMoneyFlowService.New(pdkLog, accountTransactionRepo, accountSvc, ledgerSvc, merchantSvc, p2pMoneyFlowService.WithRabbitMQClient(rabbitMqExt))
		payOutMoneyFLowSvc := payoutMoneyFlowService.New(pdkLog, accountTransactionRepo, accountSvc, ledgerSvc, merchantSvc)
		chargeMoneyFLowSvc := chargeMoneyFlowService.New(pdkLog, accountTransactionRepo, accountSvc, ledgerSvc, merchantSvc, chargeMoneyFlowService.WithRabbitMQClient(rabbitMqExt))
		cancelMoneyFlowSvc := cancelMoneyFlowService.New(pdkLog, accountTransactionRepo, accountSvc, ledgerSvc, merchantSvc, cancelMoneyFlowService.WithRabbitMQClient(rabbitMqExt))
		refundMoneyFlowSvc := refundMoneyFlowService.New(pdkLog, accountTransactionRepo, accountSvc, ledgerSvc, merchantSvc, refundMoneyFlowService.WithRabbitMQClient(rabbitMqExt))
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypePayIn, payInMoneyFlowSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypeP2P, p2pMoneyFlowSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypePayOut, payOutMoneyFLowSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypeCharge, chargeMoneyFLowSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypeCancel, cancelMoneyFlowSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypeRefund, refundMoneyFlowSvc)
		addrLocService := addrLocService.New(pdkLog, addrLocRepo)
		platformFeeSvc := platformFeeService.New(pdkLog, ledgerSvc, feeSvc, accountSvc)
		transferSvc := transferService.New(pdkLog, ledgerSvc, accountSvc, platformFeeSvc, merchantSvc, transferRepo, transferService.WithPaymentRepository(paymentRepository))
		bankaccountSvc := bankAccountService.New(bankAccountRepo, pdkLog)

		fdsProcessors := map[string]repository.IFdsProcessorRepository{
			constant.PROVIDER_FRAUD_NET: fraudNetRepo,
			constant.PROVIDER_SOKRATECH: sokratechRepo.NewFDSProcessor(),
		}

		fdsProcessorService := fdsservice.New(
			config,
			pdkLog,
			fraudRulesRepo,
			ruleEvaluationsRepo,
			accountTransactionRepo,
			paymentRepository,
			merchantRepo,
			fdsProcessors,
			fdsservice.WithCustomerRepository(customerRepo),
			fdsservice.WithRabbitMqExt(rabbitMqExt),
			fdsservice.WithPaymentMethodRepository(paymentMethodRepo),
		)

		disbursementSvc := disbursementService.New(
			config, pdkLog, merchantRepo, disbursementRepo, snapCoreRepository, bankAccountRepo,
			disbursementService.WithOrchestratorService(orchestratorSvc),
			disbursementService.WithRabbitMQClient(rabbitMqExt),
			disbursementService.WithBeneficiaryAccService(beneficiaryAccountSvc),
			disbursementService.WithGoogleCloudStorage(gcs),
			disbursementService.WithRedisClient(cacheClient),
			disbursementService.WithMerchantForbiddenUseCaseService(merchantForbiddenUsecaseSvc),
			disbursementService.WithFeeService(feeSvc),
			disbursementService.WithAccountTransactionRepository(accountTransactionRepo),
			disbursementService.WithExcelLibrary(xlsx.New()),
			disbursementService.WithDisbursementWorkerPool(disbursementWorkerPools),
			disbursementService.WithRoutingProcessorService(routingProcessorSvc),
			disbursementService.WithTransferService(transferSvc),
			disbursementService.WithLedgerService(ledgerSvc),
			disbursementService.WithStatusHistoriesRepository(statusHistoriesRepo),
			disbursementService.WithMerchantService(merchantSvc),
			disbursementService.WithAccountRepository(accountRepo),
			disbursementService.WithDisbursementMetricsRepository(disbursementDatamartRepo),
			disbursementService.WithWorkflowFDSRepository(sokratechRepo),
			disbursementService.WithPayoutManualProcessingAccountRepository(payoutManualProcessingAccountRepo),
		)
		defer disbursementSvc.WPRelease()

		withdrawalService := withdrawalService.New(
			pdkLog, withdrawalRepo, orchestratorSvc, bankAccountRepo, userSvc,
			withdrawalService.WithRedisClient(cacheClient),
			withdrawalService.WithSnapCoreRepository(snapCoreRepository),
			withdrawalService.WithAccountRepository(accountRepo),
			withdrawalService.WithWithdrawalConfig(&config.WithdrawalConfig),
			withdrawalService.WithRabbitMQClient(rabbitMqExt),
			withdrawalService.WithGCSService(gcs),
			withdrawalService.WithBankTransferConfig(disbursementSvc),
			withdrawalService.WithMerchantRepository(merchantRepo),
			withdrawalService.WithNotificationService(notificationSvc),
			withdrawalService.WithAccountTransactionRepository(accountTransactionRepo),
		)

		paymentLedgerSvc := paymentService.NewPaymentLedgerService(config, pdkLog, paymentRepository, merchantRepo, accountTransactionRepo, orchestratorSvc, feeSvc, transferSvc, ledgerSvc, rabbitMqExt)
		creditcardSvc := creditcardService.New(config, pdkLog, rabbitMqExt, paymentRepository, paymentMethodRepo, creidtcardCoreProcessorRepo,
			creditcardService.WithFeeService(feeSvc),
			creditcardService.WithPaymentLedgerService(paymentLedgerSvc),
			creditcardService.WithOrchestratorService(orchestratorSvc),
			creditcardService.WithCustomerRepo(customerRepo),
			creditcardService.WithMerchantRepo(merchantRepo),
			creditcardService.WithAccountTransactionRepo(accountTransactionRepo),
			creditcardService.WithPaymentMethodService(paymentMethodSvc),
			creditcardService.WithRedis(cacheClient),
		)
		paymentSvc := paymentService.New(paymentRepository, pdkLog, snapCoreRepository, customerRepo, merchantRepo, paymentMethodRepo, accountRepo,
			paymentService.WithOrchestratorService(orchestratorSvc),
			paymentService.WithRabbitMQClient(rabbitMqExt),
			paymentService.WithQrisService(qrisService),
			paymentService.WithConfig(config),
			paymentService.WithFeeService(feeSvc),
			paymentService.WithAccountTransactionRepository(accountTransactionRepo),
			paymentService.WithGCSService(gcs),
			paymentService.WithRedisClient(cacheClient),
			paymentService.WithJWTExt(jwtConfig),
			paymentService.WithCreditCardService(creditcardSvc),
			paymentService.WithTransferService(transferSvc),
			paymentService.WithPaymentMethodService(paymentMethodSvc),
			paymentService.WithLedgerService(ledgerSvc),
			paymentService.WithStatusHistoriesRepository(statusHistoriesRepo),
			paymentService.WithPaymentMetricsRepository(paymentDatamartRepo),
			paymentService.WithSecretManager(paymentEncryptionKey),
			paymentService.WithCryptoAesGcm(paymentCryptoAesGcm),
			paymentService.WithCryptoProvider(encryption.NewCryptoProvider()),
			paymentService.WithValidator(validate),
			paymentService.WithMerchantService(merchantSvc),
			paymentService.WithDisbursementRepository(disbursementRepo),
		)
		unifiedPaymentSvc := unifiedPaymentService.New(config, pdkLog, paymentRepository, paymentMethodRepo, accountTransactionRepo,
			unifiedPaymentService.WithMerchantRepo(merchantRepo),
			unifiedPaymentService.WithSnapCoreRepo(snapCoreRepository),
			unifiedPaymentService.WithJWTExt(jwtConfig),
			unifiedPaymentService.WithRabbitMQClient(rabbitMqExt),
			unifiedPaymentService.WithRedisClient(cacheClient),
			unifiedPaymentService.WithFeeService(feeSvc),
			unifiedPaymentService.WithOrchestratorService(orchestratorSvc),
			unifiedPaymentService.WithQRISService(qrisService),
			unifiedPaymentService.WithPaymentService(paymentSvc),
			unifiedPaymentService.WithCustomerRepo(customerRepo),
			unifiedPaymentService.WithPaymentMethodService(paymentMethodSvc),
			unifiedPaymentService.WithCreditCardService(creditcardSvc),
			unifiedPaymentService.WithCreditCardCoreProcessorRepo(creidtcardCoreProcessorRepo),
			unifiedPaymentService.WithSecret(secret),
			unifiedPaymentService.WithCache(cacheClient),
			unifiedPaymentService.WithStorage(gcs),
			unifiedPaymentService.WithFdsService(fdsProcessorService),
			unifiedPaymentService.WithStatusHistoriesRepository(statusHistoriesRepo),
			unifiedPaymentService.WithPaymentCaptureRepository(paymentCaptureRepo),
			unifiedPaymentService.WithRecurringContractRepository(recurringContractRepo),
			unifiedPaymentService.WithFDSVelocityCheck(fds.NewVelocityCheck(cacheClient.Client())),
		)
		unifiedPaymentService.WithMerchantService(unifiedPaymentSvc, merchantSvc)
		unifiedPaymentService.WithCustomerService(unifiedPaymentSvc, customerService)
		paymentService.WithUnifiedPaymentService(paymentSvc, unifiedPaymentSvc)

		refundSvc := refundService.New(config, pdkLog, refundRepo, paymentRepository, accountTransactionRepo, merchantRepo, snapCoreRepository, callbackRepo,
			refundService.WithOrchestratorService(orchestratorSvc),
			refundService.WithRabbitMQClient(rabbitMqExt),
			refundService.WithFeeService(feeSvc),
			refundService.WithRedisClient(cacheClient),
			refundService.WithPaymentMethodRepository(paymentMethodRepo),
			refundService.WithStatusHistoriesRepository(statusHistoriesRepo),
			refundService.WithGCS(gcs),
		)
		recurringContractSvc := recurringContractService.New(pdkLog, recurringContractRepo, customerService)

		paymentService.WithRefundService(refundSvc)(paymentSvc)

		requestAccountInquirySvc := accountinquirysvc.New(pdkLog, snapCoreRepository, reqAccountInquiryRepo, accountInquiriesRepo, orchestratorSvc, merchantSvc, feeSvc,
			accountinquirysvc.WithBeneficiaryAccountRepository(beneficiaryAccountRepo),
			accountinquirysvc.WithRoutingProcessorService(routingProcessorSvc),
			accountinquirysvc.WithConfig(config),
			accountinquirysvc.WithTransferService(transferSvc),
			accountinquirysvc.WithOutboundRepository(outboundRepo),
			accountinquirysvc.WithRabbitMqExt(rabbitMqExt),
		)
		merchantTopUpSvc := merchantTopUpService.New(
			config, pdkLog, paymentMethodRepo, merchantTopUpRepo, snapCoreRepository,
			merchantTopUpService.WithRabbitMQClient(rabbitMqExt),
			merchantTopUpService.WithMerchantService(merchantSvc),
			merchantTopUpService.WithOrchestratorService(orchestratorSvc),
			merchantTopUpService.WithFeeService(feeSvc),
		)
		adjustService.WithMerchantTopUpCallbackService(adjustSvc, merchantTopUpSvc)
		walletTransactionSvc := walletTransactionService.New(pdkLog, walletTransactionRepo, cacheClient, gcs)
		platformSvc := platformService.New(pdkLog, disbursementSvc, paymentSvc, merchantSvc, transferSvc, withdrawalService, merchantTopUpSvc, platformService.WithUserService(userSvc), platformService.WithUnifiedPaymentService(unifiedPaymentSvc), platformService.WithOrchestratorService(orchestratorSvc))
		reconSvc := reconService.New(
			config,
			pdkLog,
			reconRepo,
			reconService.WithGCSService(gcs),
			reconService.WithExcelService(xlsx.New()),
			reconService.WithRabbitMQClient(rabbitMqExt),
			reconService.WithAccountTransactionRepository(accountTransactionRepo),
		)
		ipWhitelistService := ipwhitelistService.New(pdkLog, ipWhitelistRepo, cacheClient, ipwhitelistService.WithConfig(config))
		walletInsightService := walletInsightService.New(orchestratorSvc, pdkLog, cacheClient)

		inboundSvc := inboundService.New(config, pdkLog, inboundRepo,
			inboundService.WithMaskingSensitiveData(strings.Split(config.AppConfig.MaskingSensitiveData, ",")),
		)

		amlProcessors := map[string]repository.IAmlProcessorRepository{
			constant.PROVIDER_ADVANCE_AI: advanceAiRepo,
		}
		amlService := amlService.New(
			config,
			pdkLog,
			merchantRepo,
			amlProcessors,
		)

		dukcapilService := dukcapilService.New(
			config,
			secret,
			pdkLog,
			dukcapilGatewayRepo,
			merchantRepo,
		)

		fraudRuleSvc := fraudruleservice.New(pdkLog, fraudRulesRepo)
		vendorSvc := vendorService.New(vendorRepo, pdkLog)
		payoutManualProcessingAccountSvc := payoutManualProcessingAccountService.New(payoutManualProcessingAccountRepo, pdkLog)

		installmentPlanSvc := installmentPlanService.NewInstallmentPlanService(pdkLog, installmentPlanRepo, creditcardSvc, merchantSvc)
		installmentPlanService.WithPaymentMethodService(installmentPlanSvc, paymentMethodSvc)
		paymentMethodService.WithInstallmentPlanService(paymentMethodSvc, installmentPlanSvc)
		unifiedPaymentService.WithInstallmentPlanService(unifiedPaymentSvc, installmentPlanSvc)

		shortLinkSvc := shortLinkService.NewShortLinkService(pdkLog, shortLinkRepo)
		shortLinkService.WithConfig(shortLinkSvc, config)
		unifiedPaymentService.WithShortLinkService(unifiedPaymentSvc, shortLinkSvc)
		settlementSvc := settlementService.New(pdkLog, accountTransactionRepo,
			settlementService.WithPaymentSvc(paymentSvc),
			settlementService.WithMerchantSvc(merchantSvc),
		)

		settlementHoldSvc := settlementHoldService.New(pdkLog, settlementHoldRepo, paymentSvc, settlementSvc, accountTransactionRepo)
		paymentService.WithSettlementHoldService(paymentSvc, settlementHoldSvc)
		reportingSvc := reportingService.New(pdkLog, reportingRepo, accountRepo)
		cardFundedPayoutSvc := cardFundedPayoutService.New(config, pdkLog,
			cardFundedPayoutService.WithFeeService(feeSvc),
			cardFundedPayoutService.WithVendorService(vendorSvc),
			cardFundedPayoutService.WithCustomerService(customerService),
			cardFundedPayoutService.WithUnifiedPaymentService(unifiedPaymentSvc),
			cardFundedPayoutService.WithDisbursementRepository(disbursementRepo),
			cardFundedPayoutService.WithPaymentMethodRepository(paymentMethodRepo),
			cardFundedPayoutService.WithStatusHistoriesRepository(statusHistoriesRepo),
			cardFundedPayoutService.WithCacheClient(cacheClient),
			cardFundedPayoutService.WithPaymentRepository(paymentRepository),
			cardFundedPayoutService.WithGCS(gcs),
			cardFundedPayoutService.WithOrchestratorService(orchestratorSvc),
		)
		settlementService.WithCardFundedPayoutSvc(settlementSvc, cardFundedPayoutSvc)

		// This is for preventive measure to avoid memory leak
		defer ants.Release()

		// Init controller
		activityController := v1ActivityController.New(config, validate, activityService)
		userController := userController.New(validate, userSvc, roleSvc, userRoleSvc, merchantSvc, jwtConfig, config, secret, rabbitMqExt, pdkLog)
		merchantController := merchantController.New(merchantSvc, validate, rabbitMqExt, merchantController.WithFeeService(feeSvc), merchantController.WithProductService(productSvc), merchantController.WithConfig(config))
		subMerchantController := subMerchantController.New(merchantSvc, accountSvc, orchestratorSvc, merchantForbiddenUsecaseSvc, validate, rabbitMqExt, subMerchantController.WithFeeService(feeSvc), subMerchantController.WithDisbursementService(disbursementSvc), subMerchantController.WithPaymentService(paymentSvc))
		roleController := roleController.New(validate, roleSvc, permissionSvc)
		callbackController := callbackController.New(config, validate, callbackSvc, rabbitMqExt)
		beneficiaryAccountController := beneficiaryAccountController.New(config, validate, rabbitMqExt, beneficiaryAccountSvc, disbursementSvc, merchantSvc, feeSvc)
		disbursementController := disbursementController.New(
			config,
			validate,
			monitor,
			disbursementController.Services{
				MerchantSvc:              merchantSvc,
				DisbursementDashboardSvc: disbursementDashboardSvc,
				DisbursementSvc:          disbursementSvc,
				BeneficiaryAccountSvc:    beneficiaryAccountSvc,
				FeeSvc:                   feeSvc,
			},
			rabbitMqExt, gcs, disbursementController.WithRedisClient(cacheClient), disbursementController.WithLogger(pdkLog),
		)
		paymentMethodController := paymentMethodController.New(paymentMethodSvc)
		bankController := bankController.New(config)
		purposeController := purposeController.New(config)
		orchestratorController := orchestratorController.New(config, orchestratorSvc, merchantSvc, validate, reportingSvc)
		v1AccountController := v1AccountController.New(config, accountSvc, orchestratorSvc)
		otpHandler := otpController.New(otpSvc, userSvc)
		v1MenuController := menuController.New(config, validate, menuSvc, userRoleSvc, roleSvc)
		adjustController := adjustController.New(adjustSvc)
		credSettingController := credSettingController.New(validate, &secret.SecuritySecret, credSettingSvc)
		callbackSettingController := callbackSettingController.New(validate, &secret.SecuritySecret, callbackSvc)
		depositSettingController := depositSettingController.New(validate, pdkLog, merchantSvc)
		v1AddrLocController := addrLocController.New(validate, addrLocService)
		v1SimulationController := simulationController.New(validate,
			simulationController.WithPaymentMethodService(paymentMethodSvc),
			simulationController.WithPaymentService(paymentSvc),
		)
		v1XbPayoutController := v1XbPayoutController.New(config,
			v1XbPayoutController.WithXbPayoutService(xbPayoutSvc),
			v1XbPayoutController.WithMerchantService(merchantSvc),
			v1XbPayoutController.WithDisbursementService(disbursementSvc),
		)
		PaymentController := paymentController.New(config, validate, monitor, paymentController.WithPaymentService(paymentSvc), paymentController.WithMerchantService(merchantSvc), paymentController.WithPaymentMethodService(paymentMethodSvc), paymentController.WithUserService(userSvc), paymentController.WithUnifiedPaymentService(unifiedPaymentSvc))
		v1LiveFeatureController := liveFeatureController.New(liveFeatureSvc)
		v1PlatformController := platformController.New(pdkLog, validate, platformSvc)
		v1IpWhitelistController := ipWhitelistController.New(ipWhitelistService, validate)
		v1CRMRateLimiterController := crmRateLimiterController.New(rateLimiterService, validate)
		v1CRMPaymentController := v1CrmPaymentController.New(paymentSvc)
		v1TransferController := transferController.New(config, validate, monitor, transferSvc, transferController.WithMerchantService(merchantSvc))
		v1WalletInsightController := walletInsightsController.New(walletInsightService)
		internalWalletRequestSetup := httpControllerUtil.NewInternalWalletRequestSetup(config, secret, pdkLog)
		v1MerchantTopUpController := merchantTopUpController.New(validate, merchantTopUpSvc)
		v1WalletTransactionController := walletTransactionController.New(validate, walletTransactionSvc)
		v1ApiLogSetttingController := apiLogController.New(inboundSvc)
		v1CountryController := countryController.New(countrySvc, validate)
		v1IndustryController := industryController.NewController(industrySvc, validate)

		// Internal Controller
		internalPaymentController := v1InternalPaymentController.New(
			validate, paymentSvc, merchantSvc, rabbitMqExt,
			v1InternalPaymentController.WithLogger(pdkLog),
			v1InternalPaymentController.WithUnifiedPaymentService(unifiedPaymentSvc),
			v1InternalPaymentController.WithConfig(config),
		)
		internalMerchantAuthController := v1InternalMerchantAuthController.New(validate, merchantSvc, v1InternalMerchantAuthController.WithLogger(pdkLog))
		internalPayoutController := v1InternalPayoutController.New(
			config, validate, disbursementSvc, merchantSvc, requestAccountInquirySvc, rabbitMqExt,
			v1InternalPayoutController.WithLogger(pdkLog),
			v1InternalPayoutController.WithRedisClient(cacheClient),
			v1InternalPayoutController.WithBeneficiaryAccountService(beneficiaryAccountSvc),
		)
		internalAccountInquiryController := internalAccountInquiry.New(requestAccountInquirySvc)
		v1InternalAccountController := internalAccountController.New(accountSvc, orchestratorSvc)
		internalAccountController.WithLogger(v1InternalAccountController, pdkLog)
		internalSubMerchantController := internalSubMerchant.New(merchantSvc, accountSvc, orchestratorSvc, validate)
		internalCreditcardController := creditcardController.New(
			config,
			validate,
			pdkLog,
			monitor,
			creditcardController.Services{
				MerchantSvc:      merchantSvc,
				CreditcardSvc:    creditcardSvc,
				OrchestratorSvc:  orchestratorSvc,
				CustomerSvc:      customerService,
				PaymentMethodSvc: paymentMethodSvc,
				PaymentSvc:       paymentSvc,
			},
		)
		internalV1UnifiedPaymentController := v1InternalUnifiedPaymentController.New(config, monitor,
			v1InternalUnifiedPaymentController.WithLogger(pdkLog),
			v1InternalUnifiedPaymentController.WithPaymentService(paymentSvc),
			v1InternalUnifiedPaymentController.WithUnifiedPaymentService(unifiedPaymentSvc),
			v1InternalUnifiedPaymentController.WithCustomerService(customerService),
		)

		internalV2UnifiedPaymentController := v2InternalUnifiedPaymentController.New(config, monitor,
			v2InternalUnifiedPaymentController.WithLogger(pdkLog),
			v2InternalUnifiedPaymentController.WithUnifiedPaymentService(unifiedPaymentSvc),
			v2InternalUnifiedPaymentController.WithCustomerService(customerService),
		)
		internalV1RefundController := v1InternalRefundController.New(config,
			v1InternalRefundController.WithLogger(pdkLog),
			v1InternalRefundController.WithRefundService(refundSvc),
		)
		internalV1RecurringContractController := v1InternalRecurringContractController.New(validate, recurringContractSvc)

		internalV2LedgerController := ledgerController.New(ledgerSvc)

		internalCustomerController := customerController.New(customerService, validate)
		v1InternalXbController := v1InternalXbController.New(config,
			v1InternalXbController.WithXbPayoutService(xbPayoutSvc),
			v1InternalXbController.WithMerchantService(merchantSvc),
			v1InternalXbController.WithDisbursementService(disbursementSvc),
			v1InternalXbController.WithSecret(secret),
			v1InternalXbController.WithLogger(pdkLog),
		)
		v1InternalPaymentMethodController := internalPaymentMethodController.New(merchantTopUpSvc, paymentMethodSvc)
		v1InternalTransferController := internalTransferController.New(transferSvc, validate)

		internalWithdrawalController := internalWithdrawalsController.New(config, validate,
			internalWithdrawalsController.WithWithdrawalService(withdrawalService),
		)
		internalPlatformController := platformInternalController.New(config, platformSvc)

		// CRM Controller
		v1CrmDisbursementController := crmDisbursementController.New(disbursementSvc, routingProcessorSvc)
		v1CrmUserController := crmUserController.New(
			config, secret, userSvc, roleSvc, userRoleSvc, merchantSvc,
			crmUserController.WithJWT(jwtConfig),
			crmUserController.WithValidator(validate),
			crmUserController.WithRabbitMQClient(rabbitMqExt),
		)
		v1CrmMerchantController := crmMerchantController.New(
			merchantSvc,
			userSvc,
			validate,
			rabbitMqExt,
			crmMerchantController.WithConfig(config),
			crmMerchantController.WithLogger(pdkLog),
			crmMerchantController.WithTNCService(tncSvc),
		)
		v1CrmMerchantForbiddenUsecase := crmmerchantforbiddenusecase.New(merchantForbiddenUsecaseSvc, validate)
		v1CrmPaymentMethodController := crmPaymentMethodController.New(paymentMethodSvc, crmPaymentMethodController.WithMerchantService(merchantSvc), crmPaymentMethodController.WithConfig(config))
		v1QrisController := qrisController.New(validate, qrisService, config)
		v1CrmXbController := crmXbController.New(xbPayoutSvc,
			crmXbController.WithLogger(pdkLog),
		)
		v1CrmCreditcardController := crmCreditcardController.New(config, secret, creditcardSvc)
		v1CrmProductController := crmProductController.New(productSvc, validate)
		v1CrmAccountController := crmAccountController.New(accountSvc, orchestratorSvc)
		v1CrmReconController := v1ReconController.New(pdkLog, validate, reconSvc)
		v1CrmFraudRuleController := crmFraudRuleController.New(fraudRuleSvc, validate)
		v1CrmVendorController := crmVendorController.New(vendorSvc, validate)
		v1CrmPayoutManualProcessingAccountController := crmPayoutManualProcessingAccountController.New(payoutManualProcessingAccountSvc, validate)
		v1CrmTNCController := crmTNCController.New(tncSvc, validate)
		v1TNCSigningController := tncSigningController.New(tncSvc, validate)
		v1CrmAmlController := crmAmlController.New(amlService, validate)
		v1CrmDukcapilController := crmDukcapilController.New(dukcapilService, validate)
		v1CrmCountryController := crmCountryController.New(countrySvc, validate)
		v1CrmIndustryController := crmIndustryController.NewController(industrySvc, validate)
		v1CrmFdsController := crmfdscontroller.New(config, pdkLog, validate, fdsProcessorService)
		v1CrmRefundController := crmRefundController.New(config,
			crmRefundController.WithLogger(pdkLog),
			crmRefundController.WithRefundService(refundSvc),
		)
		v1CrmRoleController := crmRoleController.New(roleSvc, validate, crmRoleController.WithLogger(pdkLog))
		v1CRMCustomerController := crmCustomerController.New(customerService, validate)
		v1CRMInstallmentPlanController := crmInstallmentPlanController.NewController(installmentPlanSvc, validate)
		v1CRMCallbackController := crmCallbackController.New(pdkLog, unifiedPaymentSvc, disbursementSvc)
		v1CRMSettlementController := crmSettlementController.New(settlementHoldSvc, validate)
		v1CRMCardFundedPayoutController := crmCardFundedPayoutController.New(validate, cardFundedPayoutSvc)

		v1BankAccountController := bankAccountController.New(validate, bankaccountSvc)
		v1WithdrawalController := withdrawalController.New(validate, pdkLog, withdrawalService, userSvc)
		v1WithdrawalCrmController := withdrawalCrmController.New(validate, withdrawalService)

		internalMerchantController := v1InternalMerchantController.New(merchantForbiddenUsecaseSvc, merchantSvc, validate)
		processorCallbackController := v1ProcessorCallbackController.New(pdkLog, disbursementSvc, routingProcessorSvc, requestAccountInquirySvc, validate)
		internalBankAccountController := internalBankAccountController.New(bankaccountSvc)
		internalFeeController := internalFeeController.New(feeSvc)
		internalFdsController := internalFdsController.New(config, pdkLog, validate, fdsProcessorService)
		v1ChargesController := v1ChargesController.New(config, validate, monitor,
			v1ChargesController.WithLogger(pdkLog),
			v1ChargesController.WithUnifiedPaymentService(unifiedPaymentSvc),
			v1ChargesController.WithMerchantService(merchantSvc),
		)
		v1RecurringContractController := v1RecurringContractController.New(config, validate, monitor,
			v1RecurringContractController.WithLogger(pdkLog),
			v1RecurringContractController.WithRecurringContractService(recurringContractSvc),
			v1RecurringContractController.WithMerchantService(merchantSvc),
		)
		v1RefundController := v1RefundController.New(
			v1RefundController.WithLogger(pdkLog),
			v1RefundController.WithRefundService(refundSvc),
		)
		v1ShortLinkController := v1ShortLinkController.New(config, shortLinkSvc)
		v1CardFundedPayoutController := v1CardFundedPayoutController.New(config, validate, cardFundedPayoutSvc,
			v1CardFundedPayoutController.WithLogger(pdkLog),
			v1CardFundedPayoutController.WithFeeService(feeSvc),
			v1CardFundedPayoutController.WithMerchantService(merchantSvc),
			v1CardFundedPayoutController.WithVendorService(vendorSvc),
		)

		// Init router
		routerModule := http.RouterModule{
			Cfg:             config,
			Sct:             secret,
			Monitor:         monitor,
			PdkLog:          pdkLog,
			Logger:          logger,
			Otel:            otelExt,
			Nr:              newRelicExt,
			MySQLDB:         dbClient,
			InboundRecorder: inboundSvc,
			Workflow:        conductorClient,

			Redis:                                cacheClient,
			RabbitMQ:                             rabbitMqExt,
			V1UserController:                     userController,
			V1MerchantController:                 merchantController,
			V1SubMerchantController:              subMerchantController,
			V1RoleController:                     roleController,
			V1ActivityController:                 activityController,
			V1CallbackController:                 callbackController,
			IJwt:                                 jwtConfig,
			Secret:                               secret,
			V1DisbursementController:             disbursementController,
			V1BeneficiaryAccountController:       beneficiaryAccountController,
			V1BankController:                     bankController,
			V1MasterPurposeController:            purposeController,
			V1PaymentMethodController:            paymentMethodController,
			V1OrchestratorController:             orchestratorController,
			V1AccountController:                  v1AccountController,
			V1OTPController:                      otpHandler,
			V1MenuController:                     v1MenuController,
			V1CredentialSettingController:        credSettingController,
			V1CallbackSettingController:          callbackSettingController,
			V1DepositSettingController:           depositSettingController,
			V1SimulationController:               v1SimulationController,
			V1XbPayoutController:                 v1XbPayoutController,
			V1WithdrawalController:               v1WithdrawalController,
			V1PaymentController:                  PaymentController,
			V1LiveFeatureController:              v1LiveFeatureController,
			V1PlatformController:                 v1PlatformController,
			V1ProcessorCallbackController:        processorCallbackController,
			V1IPWhitelistConfigurationController: v1IpWhitelistController,
			V1CRMRateLimiterController:           v1CRMRateLimiterController,
			V1TransferController:                 v1TransferController,
			V1WalletInsightController:            v1WalletInsightController,
			InternalWalletRequestSetup:           internalWalletRequestSetup,
			V1MerchantTopUpController:            v1MerchantTopUpController,
			V1WalletTransactionController:        v1WalletTransactionController,
			V1ApiLogsSettingController:           v1ApiLogSetttingController,
			V1ChargesController:                  v1ChargesController,
			V1RecurringContractController:        v1RecurringContractController,
			V1RefundController:                   v1RefundController,
			V1CountryController:                  v1CountryController,
			V1IndustryController:                 v1IndustryController,
			V1CardFundedPayoutController:         v1CardFundedPayoutController,

			// Internal Controller
			V1InternalPaymentController:           internalPaymentController,
			V1InternalMerchantAuthController:      internalMerchantAuthController,
			V1InternalPayoutController:            internalPayoutController,
			V1InternalAccountInquiryController:    internalAccountInquiryController,
			V1InternalSubMerchantController:       internalSubMerchantController,
			V1InternalCreditCardController:        internalCreditcardController,
			V1InternalMerchantController:          internalMerchantController,
			V1InternalAccountController:           v1InternalAccountController,
			V1InternalCustomerController:          internalCustomerController,
			V1InternalXbController:                v1InternalXbController,
			V1InternalPaymentMethodController:     v1InternalPaymentMethodController,
			V1InternalTransferController:          v1InternalTransferController,
			V1InternalUnifiedPaymentController:    internalV1UnifiedPaymentController,
			V1InternalBankAccountController:       internalBankAccountController,
			V1InternalFeeController:               internalFeeController,
			V1InternalFdsController:               internalFdsController,
			V1InternalRefundController:            internalV1RefundController,
			V1InternalWithdrawalController:        internalWithdrawalController,
			V1InternalPlatformController:          internalPlatformController,
			V1InternalRecurringContractController: internalV1RecurringContractController,

			// Internal V2 Controller
			V2InternalLedgerController:         internalV2LedgerController,
			V2InternalUnifiedPaymentController: internalV2UnifiedPaymentController,

			// Service module
			V1UserService:                     userSvc,
			V1MerchantForbiddenUseCaseService: merchantForbiddenUsecaseSvc,
			V1MerchantService:                 merchantSvc,
			V1PermissionService:               permissionSvc,
			V1ProductService:                  productSvc,
			V1IPWhitelistService:              ipWhitelistService,
			V1IndustryService:                 industrySvc,
			V1RateLimitService:                rateLimiterService,

			// CRM Module
			V1CRMAdjustmentController:                    adjustController,
			V1CRMDisbursementController:                  v1CrmDisbursementController,
			V1CRMUserController:                          v1CrmUserController,
			V1CRMMerchantController:                      v1CrmMerchantController,
			V1CRMMerchantForbiddenUseCaseController:      v1CrmMerchantForbiddenUsecase,
			V1AddrLocationController:                     v1AddrLocController,
			V1CRMPaymentMethodController:                 v1CrmPaymentMethodController,
			V1QrisController:                             v1QrisController,
			V1CRMXbController:                            v1CrmXbController,
			V1CRMCreditcardController:                    v1CrmCreditcardController,
			V1CRMBankAccountController:                   v1BankAccountController,
			V1CRMWithdrawalController:                    v1WithdrawalCrmController,
			V1CRMProductController:                       v1CrmProductController,
			V1CRMAccountController:                       v1CrmAccountController,
			V1CRMReconController:                         v1CrmReconController,
			V1CRMPaymentController:                       v1CRMPaymentController,
			V1CRMFraudRuleController:                     v1CrmFraudRuleController,
			V1CRMVendorController:                        v1CrmVendorController,
			V1CRMPayoutManualProcessingAccountController: v1CrmPayoutManualProcessingAccountController,
			V1CRMTNCController:                           v1CrmTNCController,
			V1TNCSigningController:                       v1TNCSigningController,
			V1CRMAmlController:                           v1CrmAmlController,
			V1CRMDukcapilController:                      v1CrmDukcapilController,
			V1CRMCountryController:                       v1CrmCountryController,
			V1CRMIndustryController:                      v1CrmIndustryController,
			V1CRMFdsController:                           v1CrmFdsController,
			V1CRMRefundController:                        v1CrmRefundController,
			V1CRMRoleController:                          v1CrmRoleController,
			V1CRMCustomerController:                      v1CRMCustomerController,
			V1CRMInstallmentPlanController:               v1CRMInstallmentPlanController,
			V1CRMCallbackController:                      v1CRMCallbackController,
			V1ShortLinkController:                        v1ShortLinkController,
			V1CRMSettlementController:                    v1CRMSettlementController,
			V1CRMCardFundedPayoutController:              v1CRMCardFundedPayoutController,
		}

		go func() {
			liveFeatureSvc.PollForChanges(ctx, 10*time.Second, config) // Poll every 10 seconds
		}()

		if err := routerModule.Validate(); err != nil {
			log.Fatalf("error new http router: %v", err)
		}

		r := http.NewRouter(routerModule)

		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}

		// Use config values for server host and port if available
		serverHost := config.ServerHost
		if serverHost == "" {
			serverHost = "0.0.0.0" // Default to IPv4 only
		}

		serverPort := config.ServerPort
		if serverPort == "" {
			serverPort = port // Fall back to PORT env var
		}

		serverAddr := fmt.Sprintf("%s:%s", serverHost, serverPort)
		server := &netHttp.Server{Addr: serverAddr, Handler: r}
		// Start the server in a goroutine
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, netHttp.ErrServerClosed) {
				log.Fatalf("listen: %s\n", err)
			}
		}()

		// Set up signal capturing
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Block until we receive our signal.
		<-quit

		// Doesn't block if no connections, but will otherwise wait
		// until the timeout deadline or until all connections have returned.
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server Shutdown Failed:%+v", err)
		}

		log.Println("Server gracefully stopped")
	},
}

func getServiceName(serviceName string) string {
	// First check environment variable
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		serviceName = name
	}

	// Get mode from command line args first (since this is how we run it)
	mode := os.Getenv("MODE")
	if len(os.Args) > 1 {
		mode = os.Args[1] // The command (serveHttp/serveRmqConsumer) is always the first argument
	}

	// Append suffix based on mode
	switch mode {
	case "serveHttp":
		return serviceName + "-http"
	case "serveConsumer":
		return serviceName + "-consumer"
	default:
		return serviceName
	}
}
