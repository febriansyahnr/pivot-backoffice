package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	_ "github.com/paper-indonesia/pivot-backoffice/docs"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	httpControllerUtil "github.com/paper-indonesia/pivot-backoffice/port/http/controller/util"
	customMiddleware "github.com/paper-indonesia/pivot-backoffice/port/http/middleware"
	"github.com/paper-indonesia/pivot-backoffice/port/http/middleware/openApi"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/chiExt"
	inboundPdk "github.com/paper-indonesia/pdk/v2/chiExt/inbound"
	chiExtMiddleware "github.com/paper-indonesia/pdk/v2/chiExt/middleware"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	pdkNewRelic "github.com/paper-indonesia/pdk/v2/newRelicExt"
	"github.com/paper-indonesia/pdk/v2/otelExt"
	httpSwagger "github.com/swaggo/http-swagger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

type RouterModule struct {
	Cfg             *config.Config
	Sct             *config.Secret
	Monitor         *monitoring.Monitor
	PdkLog          pdkLogger.ILogger
	Logger          logger.ILogger
	Otel            otelExt.IOtelExt
	Nr              pdkNewRelic.INewRelicExt
	Redis           redisExt.IRedisExt
	RabbitMQ        rabbitMqExt.IRabbitMQExt
	Secret          *config.Secret
	IJwt            jwt.IJwt
	MySQLDB         mySqlExt.IMySqlExt
	InboundRecorder inboundPdk.Recorder
	Workflow        conductor.IWorkflow

	// Service Modules
	V1UserService                     service.IUserService
	V1MerchantForbiddenUseCaseService service.IMerchantForbiddenUseCaseService
	V1MerchantService                 service.IMerchantService
	V1PermissionService               service.IPermissionService
	V1ProductService                  service.IProductService
	V1IPWhitelistService              service.IIPWhitelistService
	V1IndustryService                 service.IIndustryService
	V1RateLimitService                service.IRateLimiter

	// API Route Modules
	V1ActivityController                 controller.V1ActivityController
	V1UserController                     controller.V1UserController
	V1MerchantController                 controller.V1MerchantController
	V1MerchantTopUpController            controller.V1MerchantTopUpController
	V1SubMerchantController              controller.V1SubMerchantController
	V1RoleController                     controller.V1RoleController
	V1CallbackController                 controller.V1CallbackController
	V1DisbursementController             controller.V1DisbursementController
	V1BeneficiaryAccountController       controller.V1BeneficiaryAccountController
	V1BankController                     controller.V1BankController
	V1MasterPurposeController            controller.V1MasterPurposeController
	V1PaymentMethodController            controller.V1PaymentMethodController
	V1OrchestratorController             controller.V1OrchestratorController
	V1AccountController                  controller.V1AccountController
	V1OTPController                      controller.V1OTPController
	V1MenuController                     controller.V1MenuController
	V1CredentialSettingController        controller.V1CredentialSettingController
	V1CallbackSettingController          controller.V1CallbackSettingController
	V1DepositSettingController           controller.V1DepositSettingController
	V1AddrLocationController             controller.V1AddrLocationController
	V1SimulationController               controller.V1SimulationController
	V1XbPayoutController                 controller.V1XbPayoutController
	V1WithdrawalController               controller.V1WithdrawalController
	V1PaymentController                  controller.V1PaymentController
	V1TransferController                 controller.V1TransferController
	V1LiveFeatureController              controller.V1LiveFeatureController
	V1PlatformController                 controller.V1PlatformController
	V1ProcessorCallbackController        controller.V1ProcessorCallbackController
	V1IPWhitelistConfigurationController controller.V1IPWhitelistController
	V1CRMRateLimiterController           controller.V1CRMRateLimiterController
	V1WalletInsightController            controller.V1WalletInsightController
	V1WalletTransactionController        controller.V1WalletTransactionController
	V1ApiLogsSettingController           controller.V1ApiLogsSettingController
	V1ChargesController                  controller.V1ChargesController
	V1RecurringContractController        controller.V1RecurringContractController
	V1RefundController                   controller.V1RefundController
	V1CountryController                  controller.V1CountryController
	V1IndustryController                 controller.V1IndustryController
	V1ShortLinkController                controller.V1ShortLinkController
	V1CardFundedPayoutController         controller.V1CardFundedPayoutController

	// Internal Route Module
	V1InternalPaymentController           controller.V1InternalPaymentController
	V1InternalMerchantAuthController      controller.V1InternalMerchantAuthController
	V1InternalPayoutController            controller.V1InternalPayoutController
	V1InternalAccountInquiryController    controller.V1InternalAccountInquiryController
	V1InternalSubMerchantController       controller.V1InternalSubMerchantController
	V1InternalCreditCardController        controller.V1CreditCardController
	V1InternalMerchantController          controller.V1InternalMerchantController
	V1InternalAccountController           controller.V1InternalAccountController
	V1InternalCustomerController          controller.V1InternalCustomerController
	V1InternalXbController                controller.V1InternalXbController
	V1InternalPaymentMethodController     controller.V1InternalPaymentMethodController
	V1InternalTransferController          controller.V1InternalTransferController
	V1InternalUnifiedPaymentController    controller.V1InternalUnifiedPaymentController
	V1InternalBankAccountController       controller.V1InternalBankAccountController
	V1InternalFeeController               controller.V1InternalFeeController
	V1InternalMerchantRcnController       controller.V1InternalMerchantRcnController
	V1InternalFdsController               controller.V1InternalFdsController
	V1InternalRefundController            controller.V1InternalRefundController
	V1InternalWithdrawalController        controller.V1InternalWithdrawalController
	V1InternalPlatformController          controller.V1InternalPlatformController
	V1InternalRecurringContractController controller.V1InternalRecurringContractController

	// V2 Internal Route Module
	V2InternalLedgerController         controller.V2InternalLedgerController
	V2InternalUnifiedPaymentController controller.V2InternalUnifiedPaymentController

	// CRM Routes Module
	V1CRMAdjustmentController                    controller.V1CRMAdjustment
	V1CRMDisbursementController                  controller.V1CRMDisbursementController
	V1CRMUserController                          controller.V1CRMUserController
	V1CRMMerchantController                      controller.V1CRMMerchantController
	V1CRMMerchantForbiddenUseCaseController      controller.V1CRMMerchantForbiddenUseCaseController
	V1CRMPaymentMethodController                 controller.V1CRMPaymentMethodController
	V1QrisController                             controller.V1QrisController
	V1CRMXbController                            controller.V1CRMXbController
	V1CRMCreditcardController                    controller.V1CRMCreditcardController
	V1CRMBankAccountController                   controller.V1BankAccountController
	V1CRMWithdrawalController                    controller.V1CRMWithdrawalController
	V1CRMProductController                       controller.V1CRMProductController
	V1CRMAccountController                       controller.V1CRMAccountController
	V1CRMReconController                         controller.V1ReconciliationController
	V1CRMPaymentController                       controller.V1CRMPaymentController
	V1CRMFraudRuleController                     controller.V1CRMFraudRuleController
	V1CRMAmlController                           controller.V1CRMAmlController
	V1CRMCallbackController                      controller.V1CRMCallbackController
	V1CRMCountryController                       controller.V1CRMCountryController
	V1CRMCustomerController                      controller.V1CRMCustomerController
	V1CRMIndustryController                      controller.V1CRMIndustryController
	V1CRMFdsController                           controller.V1CRMFdsController
	V1CRMRefundController                        controller.V1CRMRefundController
	V1CRMDukcapilController                      controller.V1CRMDukcapilController
	V1CRMInstallmentPlanController               controller.V1CRMInstallmentPlanController
	V1CRMRoleController                          controller.V1CRMRoleController
	V1CRMSettlementController                    controller.V1CRMSettlementController
	V1CRMVendorController                        controller.V1CRMVendorController
	V1CRMPayoutManualProcessingAccountController controller.V1CRMPayoutManualProcessingAccountController
	V1CRMTNCController                           controller.V1CRMTNCController
	V1TNCSigningController                       controller.V1TNCSigningController
	V1CRMCardFundedPayoutController              controller.V1CRMCardFundedPayoutController

	InternalWalletRequestSetup *httpControllerUtil.InternalWalletRequestSetup
}

func (module *RouterModule) Validate() error {
	if module.Logger == nil {
		return errors.New("module.Logger must not be nil")
	}
	if module.V1UserController == nil {
		return errors.New("module.V1UserController must not be nil")
	}
	if module.V1MerchantController == nil {
		return errors.New("module.V1MerchantController must not be nil")
	}
	if module.V1RoleController == nil {
		return errors.New("module.V1RoleController must not be nil")
	}
	if module.V1MenuController == nil {
		return errors.New("module.V1MenuController must not be nil")
	}
	if module.V1ActivityController == nil {
		return errors.New("module.V1ActivityController must not be nil")
	}
	if module.V1CallbackController == nil {
		return errors.New("module.V1CallbackController must not be nil")
	}
	if module.V1DisbursementController == nil {
		return errors.New("module.V1DisbursementController must not be nil")
	}
	if module.V1BeneficiaryAccountController == nil {
		return errors.New("module.V1BeneficiaryAccountController must not be nil")
	}
	if module.V1BankController == nil {
		return errors.New("module.V1BankController must not be nil")
	}
	if module.V1MasterPurposeController == nil {
		return errors.New("module.V1MasterPurposeController must not be nil")
	}
	if module.V1PaymentMethodController == nil {
		return errors.New("module.V1PaymentMethodController must not be nil")
	}
	if module.V1LiveFeatureController == nil {
		return errors.New("module.V1LiveFeatureController must not be nil")
	}
	if module.V1OrchestratorController == nil {
		return errors.New("module.V1OrchestratorController must not be nil")
	}
	if module.V1SimulationController == nil {
		return errors.New("module.V1SimulationController must not be nil")
	}
	if module.V1XbPayoutController == nil {
		return errors.New("module.V1XbPayoutController must not be nil")
	}
	if module.IJwt == nil {
		return errors.New("module.IJwt must not be nil")
	}
	if module.Secret == nil {
		return errors.New("module.Secret must not be nil")
	}
	if module.V1TransferController == nil {
		return errors.New("module.V1TransferController must not be nil")
	}
	if module.V1ApiLogsSettingController == nil {
		return errors.New("module.V1ApiLogsSettingController must not be nil")
	}
	if module.V1ChargesController == nil {
		return errors.New("module.V1ChargesController must not be nil")
	}

	// Internal Route Modules
	if module.V1InternalPaymentController == nil {
		return errors.New("module.V1InternalPaymentController must not be nil")
	}
	if module.V1InternalMerchantAuthController == nil {
		return errors.New("module.V1InternalMerchantAuthController must not be nil")
	}
	if module.V1InternalPayoutController == nil {
		return errors.New("module.V1InternalPayoutController must not be nil")
	}
	if module.V1InternalAccountInquiryController == nil {
		return errors.New("module.V1InternalAccountInquiryController must not be nil")
	}
	if module.V1InternalCreditCardController == nil {
		return errors.New("module.V1InternalCreditCardController must not be nil")
	}
	if module.V1InternalMerchantController == nil {
		return errors.New("module.V1InternalMerchantController must not be nil")
	}
	if module.V1InternalAccountController == nil {
		return errors.New("module.V1InternalAccountController must not be nil")
	}
	if module.V1InternalMerchantController == nil {
		return errors.New("module.V1InternalCustomerController must not be nil")
	}
	if module.V1InternalXbController == nil {
		return errors.New("module.V1InternalXbController must not be nil")
	}
	if module.V1ProcessorCallbackController == nil {
		return errors.New("module.V1ProcessorCallbackController must not be nil")
	}
	if module.V1InternalFdsController == nil {
		return errors.New("module.V1InternalFdsController must not be nil")
	}

	// CRM Route Modules
	if module.V1CRMDisbursementController == nil {
		return errors.New("module.V1CRMDisbursementController must not be nil")
	}
	if module.V1CRMPaymentMethodController == nil {
		return errors.New("module.V1CRMPaymentMethodController must not be nil")
	}
	if module.V1CRMXbController == nil {
		return errors.New("module.V1CRMXbController must not be nil")
	}

	// Service module
	if module.V1UserService == nil {
		return errors.New("module.V1UserService must not be nil")
	}

	if module.V1RateLimitService == nil {
		return errors.New("module.V1RateLimitService should not be nil")
	}

	if module.V1CRMBankAccountController == nil {
		return errors.New("module.V1CRMBankAccountController must not be nil")
	}

	if module.V1CRMAccountController == nil {
		return errors.New("module.V1CRMAccountController must not be nil")
	}

	if module.V1CRMReconController == nil {
		return errors.New("module.V1CRMReconController must not be nil")
	}
	if module.V1CRMRefundController == nil {
		return errors.New("module.V1CRMRefundController must not be nil")
	}
	if module.V1CRMRoleController == nil {
		return errors.New("module.V1CRMRoleController must not be nil")
	}

	if module.V1InternalMerchantRcnController == nil {
		return errors.New("module.V1MerchantRcnController must not be nil")
	}
	return nil
}

func NewRouter(module RouterModule) http.Handler {
	isDevelopment := module.Cfg.Environment != constant.EnvironmentProduction
	loggerWorkers := ffcontext.NewEvaluationContext(module.Cfg.Environment)
	loggerWorkers.AddCustomAttribute("environment", module.Cfg.Environment)

	requestTimeoutSeconds, err := ffclient.IntVariation(constant.FeatureFlagKeyCustomContextTimeout, loggerWorkers, 300)
	if err != nil {
		module.PdkLog.Warn(context.Background(), "failed to get feature flag for custom context timeout", pdkLogger.Error(err))
		requestTimeoutSeconds = 300 // set context.Timeout globally to 300 seconds if feature flag is not set
	}

	loggerWorkersCount, err := ffclient.IntVariation("backend-portal-logger-total-worker-pool", loggerWorkers, 25)
	if err != nil {
		module.PdkLog.Warn(context.Background(), "failed to get feature flag for logger worker count", pdkLogger.Error(err))
		loggerWorkersCount = 25
	}

	r := chiExt.New(
		&chiExt.Config{
			LoggerWorkerCount:          loggerWorkersCount,
			LoggerWorkerExpiryDuration: 3 * time.Minute,
			Recorder:                   module.InboundRecorder,
			ContextOverrideConfig: &chiExtMiddleware.ContextOverrideConfig{
				Methods: []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodGet},
			},
		},
		chiExt.WithLogger(module.PdkLog),
		chiExt.WithNewRelic(module.Nr),
	).(*chi.Mux) // chiExt.New return http.Handler, but we need chi.Mux
	r.Use(customMiddleware.CacheHTTPRequestMiddleware(module.Cfg, module.PdkLog, module.Redis))
	r.Use(customMiddleware.DynamicTimeout(module.PdkLog, int(requestTimeoutSeconds)))
	r.Mount("/debug", middleware.Profiler())

	if isDevelopment {
		// swagger
		r.Route("/swagger", func(r chi.Router) {
			r.Use(cors.Handler(cors.Options{
				AllowedOrigins:   []string{"https://*", "http://*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
				ExposedHeaders:   []string{"Link"},
				AllowCredentials: false,
				MaxAge:           300, // Maximum value not ignored by any of major browsers
			}))

			r.Mount("/", httpSwagger.Handler())
		})
	}

	// Health check
	healthCheckTargets := []chiExtMiddleware.HealthCheckTarget{
		{Name: "rabbitMQAvailable", Health: module.RabbitMQ.HealthCheck},
		{Name: "redisAvailable", Health: func(ctx context.Context) error {
			return customMiddleware.HealthCheckContextMiddleware(module.PdkLog, ctx, constant.HealthCheckDependentServiceCache, func(ctx context.Context) error {
				return module.Redis.Ping(ctx).Err()
			})
		}},
		{Name: "mysqlAvailable", Health: func(context.Context) error { return module.MySQLDB.Ping() }},
		{Name: "serviceAvailable", Health: func(context.Context) error { return nil }},
	}
	r.Get("/health-check", chiExtMiddleware.HealthCheckHandler(module.PdkLog, healthCheckTargets...))
	r.Get("/app-version", module.V1LiveFeatureController.GetAppVersion)

	// ShortLink
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.InternalServiceMiddleware(module.Secret))
		r.Get("/s/{code}", module.V1ShortLinkController.GetByCode)
	})

	r.Route("/api/v1", func(r chi.Router) {
		// TODO: Move before api/v1 so it's not exposed to newrelic and public
		r.Get("/health-check", chiExtMiddleware.HealthCheckHandler(module.PdkLog, healthCheckTargets...))

		// Live services / features
		r.Get("/services", module.V1LiveFeatureController.GetList)

		// master such as province, city and district
		r.Route("/master", func(r chi.Router) {
			r.Route("/address/locations", func(r chi.Router) {
				r.Get("/{name}", module.V1AddrLocationController.Get)
			})
		})

		// auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", module.V1UserController.Login)
			r.Post("/refresh", module.V1UserController.Refresh)
			r.Post("/register", module.V1UserController.Register)
			r.Post("/forgot-password", module.V1UserController.ForgotPassword)
			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.SpecialCaseRequireAuthForSendOTP(module.IJwt))
				r.Post("/otp/send", module.V1OTPController.Send)
			})

			r.Route("/otp", func(r chi.Router) {
				r.Use(customMiddleware.AuthTokenFromOTP(module.IJwt, module.Redis))
				r.Post("/verify", module.V1OTPController.Verify)
			})
		})

		// validate user invitation
		r.Post("/users/validate-invitation", module.V1UserController.ValidateInvitationToken)

		// One-time token authentication (From OTP verification)
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.AuthTokenFromFeature2FA(module.IJwt, module.Redis))
			r.Patch("/auth/reset-password", module.V1UserController.ResetPassword)
			r.Patch("/auth/reset-pin", module.V1UserController.ResetPIN)
			r.Patch("/users/activate", module.V1UserController.Activate)
			r.Get("/users/2fa/token", module.V1UserController.SessionFromLogin2FA)
		})

		// Authenticated Route
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.AuthMiddleware(module.IJwt, module.Redis))
			r.Use(customMiddleware.MerchantStatusMiddleware(module.V1MerchantService, module.Cfg, response.SendApiResponseError))
			r.Use(customMiddleware.CheckMerchantMiddleware(module.V1MerchantService))

			r.Post("/logout", module.V1UserController.Logout)

			// tnc (terms and conditions) signing for the authenticated merchant
			r.Route("/tnc", func(r chi.Router) {
				r.Get("/status", module.V1TNCSigningController.Status)
				r.Get("/history", module.V1TNCSigningController.History)
				r.Post("/sign", module.V1TNCSigningController.Sign)
			})

			// callback
			r.Route("/callbacks", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingView))
					r.Get("/", module.V1CallbackController.GetCallbackList)
					r.Get("/events", module.V1CallbackController.GetCallbackEvents)
					r.Get("/histories", module.V1CallbackController.GetCallbackLogList)
					r.Get("/histories/{id}", module.V1CallbackController.GetCallbackLogDetail)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingCreate))
					r.Post("/", module.V1CallbackController.RegisterCallback)
					r.Post("/histories/{id}/resend", module.V1CallbackController.ResendCallbackByID)
					r.Post("/snap/histories/{id}/resend", module.V1CallbackController.ResendSNAPCallbackByID)
				})
			})

			// user
			r.Route("/users", func(r chi.Router) {
				r.Get("/profile", module.V1UserController.UserProfile)

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugTeamMemberView))

					r.Get("/", module.V1UserController.ListByMerchantID)
					r.Get("/{user_id}", module.V1UserController.FindByID)
					r.Get("/user-detail", module.V1UserController.UserDetail)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugTeamMemberCreate))

					r.Post("/", module.V1UserController.Create)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugTeamMemberEdit))

					r.Put("/{user_id}", module.V1UserController.Update)
					r.Put("/{user_id}/activate", module.V1UserController.ActivateUser)
					r.Put("/{user_id}/deactivate", module.V1UserController.DeactivateUser)
					r.Post("/{user_id}/resend-invitation", module.V1UserController.ResendInvitation)
				})

				r.Route("/pin", func(r chi.Router) {
					r.Post("/", module.V1UserController.CreatePin)
					r.Post("/check", module.V1UserController.CheckCurrentPin)
					r.Put("/", module.V1UserController.ChangePin)
				})

				r.Post("/check-password", module.V1UserController.CheckCurrentPassword)
				r.Post("/activities", module.V1ActivityController.Create)

				r.Route("/mfa/totp", func(r chi.Router) {
					r.Post("/enroll", module.V1UserController.EnrollTOTP)
					r.Post("/confirm", module.V1UserController.ConfirmTOTP)
				})

				r.Patch("/mfa/preferred-method", module.V1UserController.SetPreferred2FAMethod)
			})

			// Insights Usecase
			r.Route("/insight", func(r chi.Router) {
				r.Get("/payout", module.V1DisbursementController.GetDisbursementInsight)
				r.Get("/payment", module.V1PaymentController.GetPaymentDashboardInsights)
				r.Get("/xb-payout", module.V1XbPayoutController.GetXbPayoutDashboardInsights)
			})

			// Disbursement Usecase
			r.Route("/disbursements", func(r chi.Router) {
				r.Get("/configs", module.V1DisbursementController.GetTransactionConfig) // Min and max transaction amount and configuration fee
				r.Get("/limits", module.V1DisbursementController.GetTransactionLimit)   // Min and max transaction amount and daily transactions limit
				r.Get("/limits/sub-merchants/{id}", module.V1DisbursementController.GetTransactionLimitSubMerchant)
				r.Get("/daily-limits/{type}", module.V1DisbursementController.GetDailyTransactionLimit) // Daily transactions limit
				r.Get("/cut-off-time-status", module.V1DisbursementController.GetCutOffTimeStatus)
				r.Get("/dashboard", module.V1DisbursementController.GetDisbursementDashboard)

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDisbursementView, constant.PermissionSlugDisbursementHistoryView))
					r.Get("/", module.V1DisbursementController.GetList)
					r.Post("/export", module.V1DisbursementController.ExportToExcel)
					r.Get("/{id}", module.V1DisbursementController.FindByID)
					r.Get("/{id}/receipt", module.V1DisbursementController.GetReceiptByID)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDisbursementApprovalView))
					r.Get("/approval-dashboard", module.V1DisbursementController.GetDisbursementApprovalDashboard)
				})

				// Merchant Top Up
				r.Group(func(r chi.Router) {
					// Flagging From Disbursement Usecase
					r.Use(func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), constant.CtxAccountName, constant.TypeDisbursement)))
						})
					})
					// Routes
					r.With(
						customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDisbursementTopUpCreate),
					).Post("/top-up", module.V1MerchantTopUpController.Topup)
					r.Post("/top-up-simulation-va", module.V1MerchantTopUpController.TopUpSimulation)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.CheckUserMerchantForbiddenUsecase(module.V1MerchantForbiddenUseCaseService, constant.ReferenceDisbursement))

					// Approval action for single & bulk
					r.Route("/approval-actions", func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDisbursementApprovalCreate))
						r.Use(customMiddleware.CheckPINMiddleware(module.V1UserService, module.RabbitMQ))
						r.Use(customMiddleware.ApprovalActionsIdempotencyMiddleware(module.Redis, module.Cfg.ServiceName, module.Logger))
						r.Post("/", module.V1DisbursementController.ApprovalActions)
					})

					// Single disbursement action
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCreateTransactionCreate))
						r.Post("/single/create", module.V1DisbursementController.CreateSingle)
					})

					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDisbursementApprovalCreate))
						// r.Use(customMiddleware.IdempotencyApiMiddleware(module.Redis, module.Cfg.ServiceName, "retry", constant.HeaderXIdempotencyKey)) // enable when the FE was ready
						r.Post("/single/retry", module.V1DisbursementController.RetrySingle)
						r.Post("/bulk/retry", module.V1DisbursementController.RetryBulk)
						r.Post("/cancel", module.V1DisbursementController.Cancel)
					})

					// Bulk disbursement action
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDisbursementView))
						r.Get("/bulk/list", module.V1DisbursementController.GetListBulkDisbursement)
						r.Get("/bulk/{id}", module.V1DisbursementController.GetBulkDisbursementDetail)
					})
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCreateTransactionCreate))
						r.Post("/bulk/preview", module.V1DisbursementController.BulkPreview)
						r.Post("/bulk/validate", module.V1DisbursementController.BulkValidate)
						r.Post("/bulk/upload", module.V1DisbursementController.CreateBulk)
					})
				})
			})

			// Withdrawal
			r.Route("/withdrawals", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(
						customMiddleware.PlatformPermissionOption(
							customMiddleware.PermittedPermissions(module.V1PermissionService,
								constant.PermissionSlugPaymentWithdrawalView,
								constant.PermissionSlugDepositSettingView,
								constant.PermissionSlugVccTerminalBalanceView),
							module.V1PermissionService, constant.PermissionSlugPlatformView,
						),
					)
					// For backward compatibility
					r.Get("/", func(w http.ResponseWriter, r *http.Request) {
						r.SetPathValue("account", "payments")
						module.V1WithdrawalController.GetList(w, r)
					})
					r.Post("/{account}/export", module.V1WithdrawalController.Export)
					r.Get("/insights", module.V1WithdrawalController.GetInsights)
				})

				r.Get("/preparation", module.V1WithdrawalController.Preparation)
				r.Get("/{account}", module.V1WithdrawalController.GetList)
				r.Get("/{account}/{id}", module.V1WithdrawalController.GetById)

				r.Group(func(r chi.Router) {
					r.Use(
						customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentWithdrawalCreate, constant.PermissionSlugVccTerminalBalanceWithdraw),
						customMiddleware.CheckUserMerchantForbiddenUsecase(module.V1MerchantForbiddenUseCaseService, constant.ReferenceWithdrawal),
						customMiddleware.IdempotencyApiMiddleware(module.Redis, module.Cfg.ServiceName, "withdrawals", constant.HeaderXIdempotencyKey),
					)

					r.Post("/", module.V1WithdrawalController.Create)

					r.With(customMiddleware.CheckUserMerchantForbiddenUsecase(module.V1MerchantForbiddenUseCaseService, constant.ReferenceDisbursement)).
						Post("/balance", module.V1WithdrawalController.TransferBalance)
				})
			})

			// payment transfers
			r.Route("/transfers", func(r chi.Router) {
				r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentTransferView))
				r.Get("/", module.V1TransferController.FilterTransferHistory)
				r.Get("/{id}", module.V1TransferController.GetTransferByID)
			})

			// beneficiary-accounts
			r.Route("/beneficiary-accounts", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDisbursementBeneficiaryListView))
					r.Get("/", module.V1BeneficiaryAccountController.GetList)
				})

				r.Group(func(r chi.Router) {
					// create sub-merchant has ability to validate a bank account using the endpoint
					r.Use(
						customMiddleware.PermittedPermissions(
							module.V1PermissionService,
							constant.PermissionSlugDisbursementBeneficiaryListView,
							constant.PermissionSlugCreateTransactionCreate,
							constant.PermissionSlugPlatformMerchantCreate,
							constant.PermissionSlugPaymentRefundCreate,
						),
					)
					r.Post("/inquiry", module.V1BeneficiaryAccountController.CheckBeneficiary)
				})
			})

			// Wallets Usecase
			r.Route("/wallets", func(r chi.Router) {
				walletReverseProxy := httputil.ReverseProxy(&httputil.ReverseProxyConfiguration{
					PrepareFunc: []func(*http.Request) error{
						module.InternalWalletRequestSetup.PrepareInternalWalletRequest,
					},
					Logger: module.PdkLog,
				})

				// Merchant Routes
				r.Route("/merchants", func(r chi.Router) {
					// Flagging From Wallet Usecase
					r.Use(func(h http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), constant.CtxAccountName, constant.TypeWallet)))
						})
					})
					// Merchant Wallet Balance Top Up
					r.With(
						customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugWalletMerchantBalanceView),
					).Post("/top-up", module.V1MerchantTopUpController.Topup)
					r.Post("/top-up-simulation-va", module.V1MerchantTopUpController.TopUpSimulation) // Staging Only

					r.Route("/transaction-histories", func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugWalletMerchantTransactionHistoriesView))

						r.Get("/", module.V1WalletTransactionController.GetMerchantTransactionHistoryList)
						r.With(
							customMiddleware.GetTimeLocationFromHeader(constant.HeaderTimezone),
						).Get("/export", module.V1WalletTransactionController.ExportMerchantTransactionHistoryList)
						r.Get("/{id}", module.V1WalletTransactionController.GetMerchantTransactionDetail)
					})
				})

				// Customer Routes
				r.Route("/customers", func(r chi.Router) {
					r.Route("/insights", func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugWalletCustomersInsightsView))
						r.Get("/balance", walletReverseProxy)
						r.Get("/total-users", walletReverseProxy)
					})

					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugWalletCustomersView))
						r.Get("/", walletReverseProxy)
					})

					r.Route("/{customerId}", func(r chi.Router) {
						r.Group(func(r chi.Router) {
							r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugWalletCustomersView))
							r.Get("/", walletReverseProxy)
							r.Get("/balance", walletReverseProxy)
						})

						r.Route("/verification", func(r chi.Router) {
							r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugWalletCustomersVerificationView))
							r.Get("/", walletReverseProxy)
							r.Get("/selfie", walletReverseProxy)
						})

						r.Route("/transactions", func(r chi.Router) {
							r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugWalletCustomersTransactionsView))
							r.Get("/", walletReverseProxy)
							r.Post("/export", walletReverseProxy)
							r.Get("/{transactionId}", walletReverseProxy)
						})
					})
				})
			})

			// activities
			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugActivityLogView))
				r.Get("/activities", module.V1ActivityController.GetList)
			})

			// transactions history
			r.Route("/transaction-histories", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugTransactionHistoryView))
					r.Get("/", module.V1OrchestratorController.GetList)
					r.Get("/details/{transaction_id}", module.V1OrchestratorController.GetDetailById)
					r.Get("/export", module.V1OrchestratorController.ExportToExcelTransactionHistory)
				})
			})

			// User Role
			r.Route("/roles", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugRolesAndPermissionsCreate))
					r.Post("/", module.V1RoleController.Create)
					r.Post("/assign", module.V1UserController.AddUserRole)
				})
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugRolesAndPermissionsView))
					r.Get("/", module.V1RoleController.GetList)
					r.Get("/{role_id}/permissions", module.V1RoleController.FindPermissionsByRoleId) // TO BE DEPRECATED
				})
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugRolesAndPermissionsEdit))

					r.Put("/{role_id}", module.V1RoleController.Update)
				})
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugRolesAndPermissionsDelete))
					r.Delete("/{role_id}", module.V1RoleController.Delete)
				})
			})

			r.Route("/settings", func(r chi.Router) {
				r.Use(cors.Handler(cors.Options{
					ExposedHeaders: []string{constant.HeaderXResponseSignature, constant.HeaderXRequestId},
				}))
				r.Use(customMiddleware.AuthMiddleware(module.IJwt, module.Redis))
				r.Route("/credentials", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingView))
						r.Get("/", module.V1CredentialSettingController.Get)
						r.Get("/client-secrets/{id}", module.V1CredentialSettingController.GetClientSecretById)
					})

					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingCreate))
						r.Post("/client-secrets/{id}", module.V1CredentialSettingController.GenerateClientSecretById)
					})
				})
				r.Route("/callbacks", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingView))
						r.Get("/", module.V1CallbackSettingController.Get)
						r.Get("/api-key", module.V1CallbackSettingController.GetApiKey)
					})

					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingEdit))

						r.Post("/urls/{master_id}", module.V1CallbackSettingController.TestAndSaveCallbackURL)
					})

					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingEdit))
						r.Post("/snap/urls/access-token/b2b/{master_id}", module.V1CallbackSettingController.TestAndSaveSnapB2b)
						r.Post("/snap/urls/{master_id}", module.V1CallbackSettingController.TestAndSaveSnapPayment)
					})
				})
				r.Route("/deposit", func(r chi.Router) {
					r.Use(
						customMiddleware.PlatformPermissionOption(
							customMiddleware.PermittedPermissions(
								module.V1PermissionService,
								constant.PermissionSlugDepositSettingView,
							),
							module.V1PermissionService,
							constant.PermissionSlugPlatformView,
						),
					)
					r.Get("/", module.V1DepositSettingController.Get)

					r.Group(func(r chi.Router) {
						r.Use(
							customMiddleware.PlatformPermissionOption(
								customMiddleware.PermittedPermissions(
									module.V1PermissionService,
									constant.PermissionSlugDepositSettingEdit,
								),
								module.V1PermissionService,
								constant.PermissionSlugPlatformEdit,
							),
						)
						r.Patch("/auto-withdrawal", module.V1DepositSettingController.SetAutoWithdrawal)
					})
				})

				r.Route("/ip-whitelist", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Get("/", module.V1IPWhitelistConfigurationController.GetList)
						r.Get("/{id}", module.V1IPWhitelistConfigurationController.Detail)

						r.Group(func(r chi.Router) {
							r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingEdit))
							r.Put("/{id}", module.V1IPWhitelistConfigurationController.Update)
						})

						r.Group(func(r chi.Router) {
							r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingCreate))
							r.Post("/", module.V1IPWhitelistConfigurationController.Create)
						})

						r.Group(func(r chi.Router) {
							r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingDelete))
							r.Delete("/{id}", module.V1IPWhitelistConfigurationController.Delete)
						})
					})
				})

				r.Route("/api-logs", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDeveloperSettingView))
						r.Get("/", module.V1ApiLogsSettingController.GetList)
						r.Get("/{id}", module.V1ApiLogsSettingController.GetByID)
						r.Get("/{id}/snap", module.V1ApiLogsSettingController.GetSnapVersionByID)
					})
				})
			})

			// XB Route
			r.Route("/xb", func(r chi.Router) {
				r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugInternationalPayoutView))
				{
					r.Get("/fx-rate", module.V1XbPayoutController.GetFxRate)
					r.Get("/configs", module.V1XbPayoutController.GetFeeConfig)

					r.Route("/payout", func(r chi.Router) {
						r.Group(func(r chi.Router) {
							r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugInternationalPayoutCreate))
							r.Post("/", module.V1XbPayoutController.CreateSession)
							r.Post("/{id}/confirm", module.V1XbPayoutController.Confirm)
						})
						r.Get("/", module.V1XbPayoutController.GetList)
						r.Get("/{id}", module.V1XbPayoutController.GetDetail)
						r.Post("/{id}/upload", module.V1XbPayoutController.UploadUnderlyingDocument)
						r.Post("/export", module.V1XbPayoutController.ExportToExcel)
						r.Route("/upload/limit", func(r chi.Router) {
							pathDocLimit := "/api/v1/payout/upload/limit"
							r.Get("/", func(w http.ResponseWriter, r *http.Request) {
								module.V1InternalXbController.ProxyHandler(pathDocLimit, nil)(w, r)
							})
						})
					})

					r.Route("/master", func(r chi.Router) {
						r.Route("/country", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterCountry)
						})

						r.Route("/state", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterState)
						})

						r.Route("/city", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterCity)
						})

						r.Route("/currency", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterCurrency)
							r.Route("/map", func(r chi.Router) {
								r.Get("/list", module.V1XbPayoutController.GetListMasterCurrencyMapping)
							})
						})

						r.Route("/identification-type", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterIdentificationType)
						})

						r.Route("/account-type", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterAccountType)
						})

						r.Route("/purpose", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterPurpose)
						})

						r.Route("/transfer-method", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterTransferMethod)
						})

						r.Route("/source-of-income", func(r chi.Router) {
							r.Get("/list", module.V1XbPayoutController.GetListMasterSourceOfIncome)
						})
					})
				}
			})

			// All permissions
			r.Route("/menus", func(r chi.Router) {
				r.Get("/", module.V1MenuController.GetAll)
				r.Get("/role", module.V1MenuController.GetByRole)
				r.Get("/role/{role_id}", module.V1MenuController.GetByRoleId)
			})

			r.Route("/banks", func(r chi.Router) {
				r.Get("/", module.V1BankController.List)
			})

			r.Route("/purposes", func(r chi.Router) {
				r.Get("/", module.V1MasterPurposeController.List)
			})

			r.Route("/payment-methods", func(r chi.Router) {
				r.Get("/{category}", module.V1PaymentMethodController.FindPaymentMethodByCategory)
			})

			r.Route("/accounts", func(r chi.Router) {
				r.Get("/{id}", module.V1AccountController.GetByUUID)
			})
			r.Get("/balances", module.V1AccountController.GetBalance)

			r.Patch("/change-password", module.V1UserController.ChangePassword)
			r.Post("/generate-random-password", module.V1UserController.GenerateRandomPassword)

			// merchant
			r.Route("/merchants", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedRoles(constant.RoleAdmin))
					r.Post("/create", module.V1MerchantController.Create)
				})

				r.Get("/fee", module.V1MerchantController.FindMerchantFeeByMerchantIDAndType)
				r.Get("/{id}", module.V1MerchantController.FindByMerchantID)
				r.Get("/actived-products", module.V1MerchantController.GetActiveProducts)
				r.Post("/set-public-key", module.V1MerchantController.SetPKCS8MerchantPublicKey)
				r.Route("/utilities", func(r chi.Router) {
					r.Post("/generate-signature", module.V1MerchantController.GenOpenAPISignature)
				})

				r.With(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugNotificationSettingsView)).
					Get("/notification-config", module.V1MerchantController.GetNotificationConfig)
				r.With(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugNotificationSettingsEdit)).
					Put("/notification-config", module.V1MerchantController.UpdateNotificationConfig)
			})

			// subMerchant
			r.Route("/sub-merchants", func(r chi.Router) {
				r.Use(customMiddleware.MerchantUserProductValidationMiddleware(module.V1ProductService, constant.ProductPlatform))

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPlatformView))
					r.Get("/", module.V1SubMerchantController.ListSubMerchantByParentID)
					r.Get("/{id}", module.V1SubMerchantController.DetailSubMerchantByID)
					r.Get("/balance/{id}", module.V1SubMerchantController.GetSubMerchantBalance)
					r.Get("/transactions", module.V1PlatformController.TransactionList)
					r.Get("/users", module.V1PlatformController.GetSubMerchantUserList)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentHistoriesView))
					r.Post("/{subMerchantId}/payments/export", module.V1SubMerchantController.ExportPaymentHistory)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugDisbursementView, constant.PermissionSlugDisbursementHistoryView))
					r.Post("/{subMerchantId}/disbursements/export", module.V1SubMerchantController.ExportDisbursementHistory)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPlatformPayoutDailyLimitView))
					r.Get("/{id}/daily-limits/{type}", module.V1SubMerchantController.GetSubMerchantDailyLimit)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPlatformMerchantEdit))
					r.Put("/{id}", module.V1SubMerchantController.Update)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPlatformMerchantCreate))
					r.Post("/create", module.V1SubMerchantController.Create)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPlatformMerchantDeactivate))
					r.Post("/block", module.V1SubMerchantController.Block)
					r.Post("/unblock", module.V1SubMerchantController.Unblock)
				})
			})

			// payments
			r.Route("/payments", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentHistoriesView))

					r.Get("/", module.V1PaymentController.FilterPaymentHistory)
					r.Route("/payment-methods", func(r chi.Router) {
						r.Get("/", module.V1PaymentController.GetChannelList)
						r.Get("/{paymentMethodId}/documents", module.V1PaymentController.GetChannelDocuments)
						r.Patch("/{paymentMethodId}", module.V1PaymentController.UpdatePaymentMethodStatus)
					})

					r.Post("/export", module.V1PaymentController.Export)
					r.Get("/{payment_id}", module.V1PaymentController.PaymentHistory)

					r.Route("/static-qris", func(r chi.Router) {
						r.Get("/", module.V1PaymentController.FilterStaticQrisList)
						r.Get("/max-active-limit", module.V1PaymentController.GetMaxActiveStaticQRPerMerchant)
						r.Get("/{paymentId}", module.V1PaymentController.GetStaticQrisDetail)
						r.Get("/{paymentId}/transactions", module.V1PaymentController.GetStaticQrisTransactions)
						r.Put("/{paymentId}/deactivate", module.V1PaymentController.DeactivateStaticQris)
					})

					r.Route("/static-va", func(r chi.Router) {
						r.Get("/", module.V1PaymentController.FilterStaticVaList)
						r.Get("/{paymentId}", module.V1PaymentController.GetStaticVaDetail)
						r.Get("/{paymentId}/transactions", module.V1PaymentController.GetStaticVaTransactions)
						r.Put("/{paymentId}/deactivate", module.V1PaymentController.DeactivateStaticVa)

						r.Route("/ranges", func(r chi.Router) {
							r.Get("/", module.V1PaymentController.GetVARangeList)
							r.Put("/{id}", module.V1PaymentController.UpdateVARange)
						})
					})

					r.Route("/link", func(r chi.Router) {
						r.Group(func(r chi.Router) {
							r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentLinkCreate))
							r.Post("/", module.V1PaymentController.CreatePaymentLink)
						})
					})

					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugVccTerminalCreate))
						r.Get("/encryption-key", module.V1PaymentController.GetEncryptionKey)
						r.Route("/vcc-terminal", func(r chi.Router) {
							r.Route("/charges", func(r chi.Router) {
								r.With(
									customMiddleware.CheckPINMiddleware(module.V1UserService, module.RabbitMQ),
									customMiddleware.IdempotencyApiMiddlewareWithInvalidateOnError(
										module.Redis, module.PdkLog, module.Cfg.ServiceName, "vcc-terminal", constant.HeaderXIdempotencyKey,
									),
								).Post("/batch", module.V1PaymentController.VCCTerminalBatchCharge)
							})
						})
					})
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugVccTerminalChargeHistoryView))
						r.Route("/vcc-terminals", func(r chi.Router) {
							r.Get("/", module.V1PaymentController.GetVCCTerminalList)
						})
					})
				})
			})

			r.Route("/cases", func(r chi.Router) {
				r.Use(customMiddleware.MerchantUserProductValidationMiddleware(module.V1ProductService, constant.ProductPaymentInvestigation))
				r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCasesManagementView))
				r.Get("/", module.V1PaymentController.GetInvestigationList)
				r.Get("/summary", module.V1PaymentController.GetInvestigationSummary)
				r.Post("/export", module.V1PaymentController.ExportInvestigation)
				r.Get("/{payment_id}", module.V1PaymentController.PaymentHistory)
			})

			// refunds
			r.Route("/refunds", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentHistoriesView))
					r.Get("/{uuid}", module.V1RefundController.GetByID)
					r.Get("/{uuid}/receipt", module.V1RefundController.GetReceipt)
				})
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentRefundCreate))
					r.Post("/", module.V1RefundController.Create)
				})
			})

			// charges
			r.Route("/charges", func(r chi.Router) {
				r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentHistoriesView, constant.PermissionSlugCardFundedPayoutView))
				r.Get("/", module.V1ChargesController.GetChargeList)
				r.Get("/{uuid}", module.V1ChargesController.GetChargeByID)
				r.With(
					customMiddleware.GetTimeLocationFromHeader(constant.HeaderTimezone),
				).Post("/export", module.V1ChargesController.Export)
			})

			// recurrings
			r.Route("/recurrings", func(r chi.Router) {
				r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugPaymentHistoriesView))
				r.Get("/{uuid}", module.V1RecurringContractController.GetRecurringByID)
			})

			r.Route("/countries", func(r chi.Router) {
				r.Get("/", module.V1CountryController.GetAll)
			})

			r.Route("/industries", func(r chi.Router) {
				r.Get("/", module.V1IndustryController.GetAll)
			})

			r.Route("/card-funded-payouts", func(r chi.Router) {
				r.Use(customMiddleware.MerchantUserProductValidationMiddleware(module.V1ProductService, constant.ProductCardFundedPayout))

				// Register static routes first to prevent dynamic route conflicts (e.g. /{payoutId})
				r.Group(func(r chi.Router) {
					r.Get("/configs", module.V1CardFundedPayoutController.GetTransactionConfig)
				})

				r.Route("/saved-cards", func(r chi.Router) {
					r.With(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCardFundedPayoutSavedCardsView)).
						Get("/", module.V1CardFundedPayoutController.GetSavedCardList)
					r.With(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCardFundedPayoutSavedCardsCreate)).
						Post("/", module.V1CardFundedPayoutController.CreateSavedCard)
				})

				r.Route("/vendors", func(r chi.Router) {
					r.With(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCardFundedPayoutVendorView)).
						Get("/", module.V1CardFundedPayoutController.GetVendorList)
					r.With(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCardFundedPayoutVendorView)).
						Get("/{id}", module.V1CardFundedPayoutController.GetVendorDetail)
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCardFundedPayoutView))
					r.Get("/", module.V1CardFundedPayoutController.GetPayoutList)
					r.Get("/insights", module.V1CardFundedPayoutController.GetPayoutInsights)
					r.Get("/{payoutId}", module.V1CardFundedPayoutController.GetPayoutDetail)
					r.Get("/{payoutId}/receipt", module.V1CardFundedPayoutController.GetReceipt)
					r.Post("/export", module.V1CardFundedPayoutController.ExportPayoutList)
				})

				r.Group(func(r chi.Router) {
					r.Use(
						customMiddleware.IdempotencyApiMiddlewareWithInvalidateOnError(
							module.Redis, module.PdkLog, module.Cfg.ServiceName, "card-funded-payout", constant.HeaderXIdempotencyKey,
						),
					)
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.CheckPINMiddleware(module.V1UserService, module.RabbitMQ))
						r.With(
							customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCardFundedPayoutCreate),
						).Post("/", module.V1CardFundedPayoutController.CreatePayout)
						r.With(
							customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCardFundedPayoutNeedActionCreate),
						).Post("/{payoutId}/reject", module.V1CardFundedPayoutController.RejectPayout)
					})
					r.With(
						customMiddleware.PermittedPermissions(module.V1PermissionService, constant.PermissionSlugCardFundedPayoutNeedActionCreate),
					).Post("/{payoutId}/approve", module.V1CardFundedPayoutController.ApprovePayout)
				})
			})
		})

		// Supported Simulations Endpoint
		r.Route("/simulations", func(r chi.Router) {
			r.Use(customMiddleware.EnvironmentCheck(module.Cfg.Environment, constant.EnvironmentStaging))
			r.Use(customMiddleware.KeyAuth(constant.HeaderXSimulationKey, module.Secret.InternalApiKeySecret.Simulation))

			r.Get("/payment-methods/payment", module.V1SimulationController.GetPaymentMethodForPayment)
			r.Route("/payments", func(r chi.Router) {
				r.Get("/{id}", module.V1SimulationController.GetPaymentByID)
				r.Post("/{id}/pay", module.V1SimulationController.ProcessPayment)
			})
		})

		// Payment UI with payment token verification
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.PaymentTokenMiddleware(module.IJwt, module.Redis, module.Logger))
			r.Get("/payments/channels", module.V1PaymentController.GetChannelListWithPaymentToken)
			r.Get("/payments/detail", module.V1PaymentController.GetPaymentDetailForPaymentUI)
			r.Get("/payments/detail/image", module.V1PaymentController.GetPaymentImages)
			r.Get("/payments/detail/instruction", module.V1PaymentController.GetPaymentInstructions)
			r.Post("/payments/confirm", module.V1PaymentController.ConfirmPayment)
		})
	})

	// Internal Routes
	r.Route("/internal/v1", func(r chi.Router) {
		// Authenticated internal routes with merchant access token
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.InternalServiceMiddleware(module.Secret))
			r.Use(customMiddleware.MerchantAuthMiddleware(module.IJwt))
			r.Use(customMiddleware.MerchantStatusMiddleware(
				module.V1MerchantService, module.Cfg, response.SendOpenApiNonSnapResponseError,
			))
			r.Use(customMiddleware.CheckSubMerchantMiddleware(module.V1MerchantService, module.V1ProductService))

			r.Get("/me", module.V1InternalMerchantAuthController.GetAuthInfo)

			r.Route("/payments", func(r chi.Router) {
				r.Patch("/{id}", module.V1InternalPaymentController.Update)
				r.Get("/{id}", module.V1InternalPaymentController.FindPaymentById)
				r.Post("/", module.V1InternalPaymentController.Create)
				r.Get("/query/qr-mpm-dynamic", module.V1InternalPaymentController.QueryQrMpmDynamic)
				r.Post("/query/qr-mpm-static", module.V1InternalPaymentController.QueryQrMpmStatic)
			})

			r.Route("/payouts", func(r chi.Router) {
				r.Use(customMiddleware.CheckMerchantForbiddenUsecase(module.V1MerchantForbiddenUseCaseService, constant.ReferenceDisbursement))
				{
					r.Post("/", module.V1InternalPayoutController.Create)
				}
				r.Get("/{id}", module.V1InternalPayoutController.FindByBulkId)
				r.Post("/{id}/retry", module.V1InternalPayoutController.RetryBulk)
			})

			r.Route("/inquiry-accounts", func(r chi.Router) {
				r.Post("/", module.V1InternalAccountInquiryController.RequestAccountInquiry)
				r.Get("/{inquiryId}", module.V1InternalAccountInquiryController.CheckStatusRequestInquiry)
			})

			r.Post("/secret-key", module.V1InternalMerchantAuthController.CreatePKCS8SecretKey)

			r.Route("/sub-merchants", func(r chi.Router) {
				r.Use(customMiddleware.MerchantProductValidationMiddleware(module.V1ProductService, constant.ProductPlatform))
				r.Post("/create", module.V1InternalSubMerchantController.Create)
				r.Put("/{id}", module.V1InternalSubMerchantController.Update)
				r.Post("/admin", module.V1InternalSubMerchantController.AssignAdmin)
			})

			// Creditcard
			r.Route("/cards", func(r chi.Router) {
				r.Post("/payment-session", module.V1InternalCreditCardController.CreatePayment)
				r.Get("/payment-session/{uuid}", module.V1InternalCreditCardController.GetPaymentById)
			})

			r.Route("/merchants/forbidden-usecase", func(r chi.Router) {
				r.Post("/block", module.V1InternalMerchantController.Block)
				r.Post("/unblock", module.V1InternalMerchantController.Unblock)
			})

			r.Get("/balances", module.V1InternalAccountController.GetBalance)

			r.Route("/transfers", func(r chi.Router) {
				r.Use(customMiddleware.SetErrorSourceMiddleware())
				r.Use(customMiddleware.MerchantProductValidationMiddleware(module.V1ProductService, constant.ProductPlatform))
				r.Post("/", module.V1InternalTransferController.Create)
				r.Get("/", module.V1InternalTransferController.GetList)
				r.Get("/{id}", module.V1InternalTransferController.GetById)
			})

			r.Route("/merchants/rcns", func(r chi.Router) {
				r.Get("/{rcn_id}", module.V1InternalMerchantRcnController.FindByIDAndMerchantID)
				r.Post("/transactions/inquiry", module.V1InternalMerchantRcnController.InquiryTransactions)
			})
		})

		// Authenticated internal routes without merchant access token
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.InternalServiceMiddleware(module.Secret))

			// TODO:  create new middleware function which give header reference of the merchantId
			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.IPWhitelistMiddleware(module.V1IPWhitelistService, constant.ClientIdKey))
				r.With(customMiddleware.InboundFeatureMiddleware(constant.InboundFeatureAuth)).
					Post("/access-token/b2b", module.V1InternalMerchantAuthController.GetAccessTokenB2b)
				r.Get("/secret-key", module.V1InternalMerchantAuthController.GetPKCS8SecretKey)
			})

			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.IPWhitelistMiddleware(module.V1IPWhitelistService, constant.HeaderXMerchantId))
				r.Post("/access-token/b2b/validate", module.V1InternalMerchantAuthController.ValidateAccessTokenB2b)
			})

			r.Group(func(r chi.Router) {
				r.Route("/snap", func(r chi.Router) {
					r.Use(customMiddleware.IPWhitelistMiddleware(module.V1IPWhitelistService, constant.HeaderXMerchantId))
					r.Route("/access-token/b2b", func(r chi.Router) {
						r.Post("/", module.V1InternalMerchantAuthController.GetSNAPAccessTokenB2B)
						r.Post("/validate", module.V1InternalMerchantAuthController.ValidateSNAPAccessTokenB2b)
					})
					r.Route("/signature", func(r chi.Router) {
						r.Post("/validate", module.V1InternalMerchantAuthController.ValidateSNAPSignature)
						r.Post("/b2b2c/validate", module.V1InternalMerchantAuthController.ValidateB2B2CTokenSNAPSignature)
						r.Post("/generate", module.V1InternalMerchantAuthController.GenerateSNAPSignature)
						r.Post("/generate-b2b-token", module.V1InternalMerchantAuthController.GenerateB2BTokenSNAPSignature)
					})
				})
			})

			r.Post("/util/encrypting-key", module.V1InternalMerchantAuthController.UtilEncryptingKey)

			r.Group(func(r chi.Router) {
				r.Get("/merchants/{id}", module.V1InternalMerchantController.Detail)

				r.Route("/wallet", func(r chi.Router) {
					r.Get("/merchants/{merchantId}/jit-api-key", module.V1InternalMerchantController.GetJITApiKey)

					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.CheckMerchantMiddleware(module.V1MerchantService))
						r.Route("/merchants", func(r chi.Router) {
							r.Get("/{merchantId}/account", module.V1InternalAccountController.GetWalletMerchantAccount)
						})

						r.Route("/customers", func(r chi.Router) {
							r.Post("/", module.V1InternalCustomerController.CreateWalletCustomer)
							r.Get("/{phoneNumber}", module.V1InternalCustomerController.GetByPhoneNumber)
							r.Get("/{customerId}/account", module.V1InternalAccountController.GetWalletCustomerAccount)
							r.Post("/account", module.V1InternalAccountController.CreateWalletCustomerAccount)
						})

						r.Post("/whitelabel/transaction-fee", module.V1InternalFeeController.CalculateWhitelabelMerchantFee)
						r.Post("/transaction-fee", module.V1InternalFeeController.CalculateWalletTransactionFee)
					})
				})

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.CheckMerchantMiddleware(module.V1MerchantService))
					r.Get("/bank-accounts/{merchantId}", module.V1InternalBankAccountController.GetMerchantBankAccount)
				})
			})

			r.Route("/fds", func(r chi.Router) {
				r.Post("/check/{id}", module.V1InternalFdsController.CheckTransaction)
				r.Patch("/update/{id}", module.V1InternalFdsController.UpdateTransaction)
			})

			r.Route("/customers", func(r chi.Router) {
				r.Post("/", module.V1InternalCustomerController.Create)
				r.Get("/", module.V1InternalCustomerController.GetList)
				r.Put("/{id}", module.V1InternalCustomerController.Update)
				r.Get("/{id}", module.V1InternalCustomerController.GetById)
				r.Delete("/{id}", module.V1InternalCustomerController.Delete)
			})

			r.Route("/cards/stored-card", func(r chi.Router) {
				r.Get("/{merchantId}/{customerId}", module.V1InternalCreditCardController.GetStoredCardByCustomerID)
				r.With(
					customMiddleware.PaymentTokenMiddleware(module.IJwt, module.Redis, module.Logger),
				).Delete("/{merchantId}/{customerId}/{tokenId}", module.V1InternalCreditCardController.RemoveCardByCustomerIDAndTokenID)
			})
			r.Post("/extend-payment-token", module.V1InternalCreditCardController.GeneratePaymentToken)
		})
	})

	r.Route("/internal/v2", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.InternalServiceMiddleware(module.Secret))

			r.Route("/ledger", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.CheckMerchantMiddleware(module.V1MerchantService))
					r.Post("/", module.V2InternalLedgerController.Create)
				})

				r.Put("/{referenceId}", module.V2InternalLedgerController.Update)
				r.Get("/{referenceId}", module.V2InternalLedgerController.GetLedgerDetail)
				r.Get("/", module.V2InternalLedgerController.GetTransactions)
				r.Route("/balance", func(r chi.Router) {
					r.Get("/{accountId}", module.V2InternalLedgerController.GetLedgerBalance)
					r.Post("/query", module.V2InternalLedgerController.CalculateBulkLedgerBalance)
				})
			})
		})
	})

	// CRM Routes
	r.Route("/crm/v1", func(r chi.Router) {
		r.Use(customMiddleware.KeyAuth(constant.HeaderXCRMKey, module.Secret.CrmSecret.ApiKey))

		r.Route("/balances", func(r chi.Router) {
			r.Get("/flip-balance", module.V1CRMDisbursementController.GetFlipEscrowBalance)
			r.Get("/dana-balance", module.V1CRMDisbursementController.GetDanaEscrowBalance)

			r.Route("/topup", func(r chi.Router) {
				r.Post("/manual", module.V1CRMAdjustmentController.CreateManualTopup)
				r.Post("/manual/{id}/adjustment", module.V1CRMAdjustmentController.CreateAdjustmentFromManualTopup)
			})
			r.Route("/hold", func(r chi.Router) {
				r.Post("/", module.V1CRMAdjustmentController.HoldMerchantBalance)
				r.Get("/", module.V1CRMAdjustmentController.GetHoldedMerchantBalance)
			})
		})
		r.Post("/unblock", module.V1UserController.UnblockUser)

		r.Route("/adjustments", func(r chi.Router) {
			r.Post("/", module.V1CRMAdjustmentController.CreateBalanceAdjustment)
		})

		r.Route("/bank-accounts", func(r chi.Router) {
			r.Post("/", module.V1CRMBankAccountController.Create)
			r.Put("/{id}", module.V1CRMBankAccountController.Update)
		})

		r.Post("/withdrawals/{id}/inquiry-transaction", module.V1CRMWithdrawalController.InquiryTransaction)
		r.Post("/withdrawals/{id}/retry-transaction", module.V1CRMWithdrawalController.RetryTransaction)
		r.Post("/withdrawals/{id}/change-transaction-status", module.V1CRMWithdrawalController.ChangeStatusWithdrawal)

		// payment methods
		r.Route("/payment-methods", func(r chi.Router) {
			r.Post("/", module.V1CRMPaymentMethodController.Create)
		})

		// merchant
		r.Route("/merchants", func(r chi.Router) {
			r.Post("/", module.V1CRMMerchantController.Create)
			r.Get("/{merchantId}", module.V1CRMMerchantController.Get)
			r.Put("/{id}", module.V1CRMMerchantController.Update)
			r.Get("/{merchantId}/balances", module.V1CRMAccountController.GetBalance)
			r.Post("/assign", module.V1CRMMerchantController.Assign)
			r.Post("/{id}/documents", module.V1MerchantController.UploadDocument)
			r.Get("/{id}/documents", module.V1MerchantController.GetDocuments)
			r.Post("/{id}/bods", module.V1MerchantController.CreateMerchantBOD)
			r.Get("/{id}/bods", module.V1MerchantController.GetListMerchantBOD)
			r.Put("/{id}/bods/{bod_id}", module.V1MerchantController.UpdateMerchantBOD)
			r.Patch("/{id}/transaction-configs", module.V1CRMMerchantController.TransactionConfig)
			r.Get("/{id}/transaction-configs", module.V1CRMMerchantController.GetTransactionConfig)
			r.Put("/{id}/fds-configs", module.V1CRMMerchantController.FDSConfig)
			r.Get("/{id}/fds-configs", module.V1CRMMerchantController.GetFDSConfig)
			r.Patch("/{id}/payment-investigation-configs", module.V1CRMMerchantController.PaymentInvestigationConfig)
			r.Put("/{id}/kyc", module.V1CRMMerchantController.UpdateKYC)
			r.Patch("/{merchantId}/beneficiary-payout-limit", module.V1CRMMerchantController.SetCustomLimitConfig)
			r.Post("/reserve-shortname", module.V1CRMMerchantController.UploadReservedShortName)
			r.Post("/{id}/block", module.V1CRMMerchantController.BlockMerchant)
			r.Post("/{id}/unblock", module.V1CRMMerchantController.UnblockMerchant)
			r.Get("/{id}/tncs", module.V1CRMMerchantController.GetMerchantTNCHistory)

			// merchant fee
			r.Route("/fee", func(r chi.Router) {
				r.Post("/", module.V1MerchantController.CreateMerchantFee)
				r.Put("/{id}", module.V1MerchantController.UpdateMerchantFee)
				r.Patch("/{id}/settlement-configs", module.V1CRMMerchantController.UpdateSettlementConfig)
				r.Patch("/{id}/tiers", module.V1CRMMerchantController.UpdateFeeTieringConfig)
			})

			// Fee On Behalf Of Sub-Merchants
			r.Route("/fee-on-behalf", func(r chi.Router) {
				r.Post("/", module.V1CRMMerchantController.CreateFeeConfigOnBehalf)
				r.Get("/details", module.V1CRMMerchantController.GetFeeConfigOnBehalf)
				r.Patch("/{id}", module.V1CRMMerchantController.UpdateFeeConfigOnBehalf)
			})

			// Manual Invoicing
			r.Route("/{merchantId}/billing", func(r chi.Router) {
				r.Get("/fees", module.V1CRMMerchantController.GetBillingFees)
				r.Post("/fees/pay", module.V1CRMMerchantController.PayBillingFees)
			})

			// merchant forbidden use case
			r.Route("/forbidden-usecase", func(r chi.Router) {
				r.Post("/block", module.V1CRMMerchantForbiddenUseCaseController.Block)
				r.Post("/unblock", module.V1CRMMerchantForbiddenUseCaseController.Unblock)
			})

			// merchant payment method
			r.Route("/{id}/payment-methods", func(r chi.Router) {
				r.Get("/", module.V1CRMPaymentMethodController.GetByMerchant)
				r.Patch("/activate", module.V1CRMPaymentMethodController.ActivateAllPaymentMethod) // staging only
				r.Patch("/{paymentMethodId}/activate", module.V1CRMPaymentMethodController.ActivatePaymentMethodMerchant)
				r.Patch("/{paymentMethodId}/deactivate", module.V1CRMPaymentMethodController.DeactivatePaymentMethodMerchant)
				r.Patch("/{paymentMethodId}/config", module.V1CRMPaymentMethodController.SetupConfig)
				r.Patch("/{paymentMethodId}/activation-status", module.V1CRMPaymentMethodController.ChangeActivationStatus)
				r.Get("/{paymentMethodId}/documents", module.V1CRMPaymentMethodController.GetRequiredMerchantDocumentList)
				r.Get("/static-va", module.V1CRMPaymentMethodController.GetStaticVAByMerchant)
				r.Get("/static-qris", module.V1CRMPaymentMethodController.GetStaticQRByMerchant)
			})

			r.Patch("/{id}/close", module.V1MerchantController.CloseMerchant)

			r.Route("/{merchantId}/rate-limiter", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Get("/", module.V1CRMRateLimiterController.GetList)
					r.Get("/{id}", module.V1CRMRateLimiterController.Detail)
					r.Post("/", module.V1CRMRateLimiterController.Create)
					r.Put("/{id}", module.V1CRMRateLimiterController.Update)
				})
			})
		})

		// Installment plan
		r.Route("/installment-plans", func(r chi.Router) {
			r.Post("/", module.V1CRMInstallmentPlanController.Create)
			r.Put("/{id}", module.V1CRMInstallmentPlanController.Update)
		})

		r.Route("/sub-merchants", func(r chi.Router) {
			r.Post("/bulk-create", module.V1CRMMerchantController.BulkCreateSubmerchant)
			r.Get("/{merchantId}/bulk-create/{sessionId}/summary", module.V1CRMMerchantController.GetBulkCreateSubmerchantSummary)
		})

		// customers
		r.Route("/customers", func(r chi.Router) {
			r.Get("/", module.V1CRMCustomerController.GetCustomerList)
			r.Get("/{id}", module.V1CRMCustomerController.GetCustomerByID)
			r.Post("/", module.V1CRMCustomerController.Create)
			r.Put("/{id}", module.V1CRMCustomerController.Update)
		})

		// users
		r.Route("/users", func(r chi.Router) {
			r.Post("/", module.V1CRMUserController.Create)
			r.Put("/{user_id}", module.V1CRMUserController.Update)
			r.Post("/resend-invitation", module.V1CRMUserController.ResendInvitation)
		})

		// disbursement case
		r.Route("/disbursements", func(r chi.Router) {
			r.Post("/{id}/retry-transaction", module.V1CRMDisbursementController.RetryTransaction)
			r.Post("/{id}/inquiry-transaction", module.V1CRMDisbursementController.InquiryTransaction)
			r.Post("/{id}/reversal", module.V1CRMDisbursementController.Reversal)

			// for internal tool purpose
			r.Post("/change-transaction-status", module.V1CRMDisbursementController.ChangeStatus)
			r.Post("/check-transactions", module.V1CRMDisbursementController.CheckTransactionStatus)

			r.Post("/payout-status", module.V1CRMDisbursementController.GetPayoutStatusAndRouting)
			r.Post("/batch-payout-status", module.V1CRMDisbursementController.GetBatchPayoutStatusAndRouting)

			r.Post("/receipt", module.V1CRMDisbursementController.GetReceipt)

			// manual processing accounts
			r.Route("/manual-processing-accounts", func(r chi.Router) {
				r.Get("/", module.V1CRMPayoutManualProcessingAccountController.List)
				r.Post("/", module.V1CRMPayoutManualProcessingAccountController.Create)
				r.Put("/{uuid}", module.V1CRMPayoutManualProcessingAccountController.Update)
			})
		})

		// Address data such as province, city and district
		r.Route("/address/locations", func(r chi.Router) {
			r.Get("/{name}", module.V1AddrLocationController.Get)
		})

		r.Route("/qr", func(r chi.Router) {
			r.Post("/registrations", module.V1QrisController.Registration)
			r.Get("/registrations/merchants/{id}", module.V1QrisController.RegistrationList)
			r.Put("/registrations/documents", module.V1QrisController.ReuploadDocument)
			r.Post("/registrations/duplicate", module.V1QrisController.DuplicateRegistration)
		})

		// xb
		r.Route("/xb", func(r chi.Router) {
			r.Route("/payout", func(r chi.Router) {
				r.Get("/{id}", module.V1CRMXbController.GetPayoutByID)
				r.Post("/{id}/get-rfi", module.V1CRMXbController.GetRfiDetails)
				r.Post("/{id}/submit-rfi", module.V1CRMXbController.SubmitRfiDetails)
				r.Post("/{id}/reconfirm", module.V1CRMXbController.ReConfirm)
			})

			r.Route("/config", func(r chi.Router) {
				r.Route("/spread", func(r chi.Router) {
					r.Post("/", module.V1CRMXbController.CreateConfigSpread)
					r.Get("/{id}", module.V1CRMXbController.GetConfigSpreadDetailByID)
					r.Get("/list", module.V1CRMXbController.GetListConfigSpread)
					r.Put("/{id}", module.V1CRMXbController.UpdateConfigSpread)
				})
			})
		})

		// creditcard
		r.Route("/creditcard", func(r chi.Router) {
			r.Post("/void", module.V1CRMCreditcardController.Void)
			r.Route("/transaction", func(r chi.Router) {
				r.With(
					customMiddleware.GetTimeLocationFromHeader(constant.HeaderTimezone),
				).Get("/list", module.V1CRMCreditcardController.GetTransactionList)

				r.Get("/detail", module.V1CRMCreditcardController.ProxyHandlerWithQueryConversion("/api/v1/transaction/detail", nil))
			})
			r.Route("/mid", func(r chi.Router) {
				r.Post("/", module.V1CRMCreditcardController.CreateMID)
				r.Put("/{id}", module.V1CRMCreditcardController.UpdateMID)
				r.Get("/list", module.V1CRMCreditcardController.GetMIDList)

				r.Get("/map/list", module.V1CRMCreditcardController.GetMIDMapList)
			})
			r.Route("/bin-mappings", func(r chi.Router) {
				binMappingsPath := "/api/v1/bin-mappings"
				r.Get("/", module.V1CRMCreditcardController.ProxyHandlerWithQueryConversion(binMappingsPath, nil))
				r.Post("/", module.V1CRMCreditcardController.ProxyHandlerWithQueryConversion(binMappingsPath, nil))
				r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
					id := chi.URLParam(req, "id")
					module.V1CRMCreditcardController.ProxyHandlerWithQueryConversion(binMappingsPath+"/"+id, nil)(w, req)
				})
				r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
					id := chi.URLParam(req, "id")
					module.V1CRMCreditcardController.ProxyHandler(binMappingsPath+"/"+id, nil)(w, req)
				})
			})

			r.Route("/bin", func(r chi.Router) {
				binPath := "/api/v1/bin"
				r.Get("/list", module.V1CRMCreditcardController.ProxyHandlerWithQueryConversion(binPath+"/list", nil))
				r.Put("/do-not-update", module.V1CRMCreditcardController.ProxyHandler(binPath+"/do-not-update", nil))
				r.Post("/bulk-upload", module.V1CRMCreditcardController.ProxyHandler(binPath+"/bulk-upload", nil))
			})
		})

		// card
		r.Route("/card", func(r chi.Router) {
			r.Put("/block", module.V1CRMCreditcardController.BlockCard)
		})

		r.Route("/products", func(r chi.Router) {
			r.Get("/", module.V1CRMProductController.GetProductList)
			r.Put("/", module.V1CRMProductController.UpdateProductAvailability)
			r.Get("/merchants/{merchantId}", module.V1CRMProductController.GetMerchantSelectedProducts)
			r.Post("/merchants", module.V1CRMProductController.AddMerchantSelectedProduct)
			r.Put("/merchants/{merchantId}", module.V1CRMProductController.UpdateMerchantProductAvailability)
		})

		// reconciliation
		r.Route("/recon", func(r chi.Router) {
			r.Get("/list", module.V1CRMReconController.GetList)
			r.Post("/upload", module.V1CRMReconController.UploadFile)
			r.Post("/download-result", module.V1CRMReconController.DownloadResult)
			r.Put("/detail", module.V1CRMReconController.UpdateReconDetail)
		})

		r.Route("/payments", func(r chi.Router) {
			r.Get("/", module.V1CRMPaymentController.GetList)
			r.Route("/investigations", func(r chi.Router) {
				r.Get("/", module.V1CRMPaymentController.GetInvestigationList)
				r.Get("/{paymentId}/proof-of-payment", module.V1CRMPaymentController.GetInvestigationProofOfPayment)
			})
			r.Get("/{id}", module.V1CRMPaymentController.GetDetailByID)
			r.Post("/{id}/inquiry", module.V1CRMPaymentController.InquiryByID)
			r.Post("/{id}/investigation", module.V1CRMPaymentController.UpdateInvestigation)
			r.Get("/{paymentId}/split-routing-by-transfer-id/{transferId}", module.V1CRMPaymentController.GetSplitRoutingByTransferID)
			r.Post("/{id}/retry-notification", module.V1CRMPaymentController.RetryNotification)
			r.Post("/static-va/retry-payment-notif", module.V1CRMPaymentController.RetryStaticVANotification)
			r.Post("/receipt", module.V1CRMPaymentController.GetReceipt)
		})

		r.Route("/settlements", func(r chi.Router) {
			r.Post("/hold-release", module.V1CRMSettlementController.CreateSettlementHold)
		})

		r.Route("/refunds", func(r chi.Router) {
			r.Post("/", module.V1CRMRefundController.Create)
		})

		// fraud rule
		r.Route("/fraud-rule", func(r chi.Router) {
			r.Post("/", module.V1CRMFraudRuleController.Create)
			r.Put("/{id}", module.V1CRMFraudRuleController.Update)
			r.Delete("/{id}", module.V1CRMFraudRuleController.Delete)
			r.Get("/{id}", module.V1CRMFraudRuleController.Detail)
			r.Get("/", module.V1CRMFraudRuleController.List)
		})

		// card-funded-payout vendors
		r.Route("/card-funded-payout", func(r chi.Router) {
			r.Route("/vendors", func(r chi.Router) {
				r.Post("/", module.V1CRMVendorController.Create)
				r.Put("/{id}", module.V1CRMVendorController.Update)
				r.Delete("/{id}", module.V1CRMVendorController.Delete)
				r.Get("/{id}", module.V1CRMVendorController.Detail)
				r.Get("/", module.V1CRMVendorController.List)
			})
		})

		r.Route("/references", func(r chi.Router) {
			r.Get("/banks", module.V1BankController.List)
		})

		// aml
		r.Route("/aml", func(r chi.Router) {
			r.Post("/screening", module.V1CRMAmlController.Screening)
			r.Post("/profile", module.V1CRMAmlController.Profile)
			r.Put("/screening/{profileId}/status", module.V1CRMAmlController.UpdateDetailStatus)
		})

		// dukcapil
		r.Route("/dukcapil", func(r chi.Router) {
			r.Post("/verify", module.V1CRMDukcapilController.VerifyIdentity)
		})

		r.Route("/country", func(r chi.Router) {
			r.Get("/", module.V1CRMCountryController.GetAll)
		})

		r.Route("/industry", func(r chi.Router) {
			r.Get("/", module.V1CRMIndustryController.GetAll)
			r.Post("/", module.V1CRMIndustryController.Create)
			r.Put("/{id}", module.V1CRMIndustryController.Update)
			r.Delete("/{id}", module.V1CRMIndustryController.Delete)
		})

		// crm fds
		r.Route("/fds", func(r chi.Router) {
			r.Patch("/update/{id}", module.V1CRMFdsController.UpdateTransaction)
		})

		// roles
		r.Route("/roles", func(r chi.Router) {
			r.Put("/default-permissions", module.V1CRMRoleController.AddDefaultRolePermissions)
			r.Delete("/default-permissions", module.V1CRMRoleController.DeleteDefaultRolePermissions)
		})

		// callback
		r.Route("/callback", func(r chi.Router) {
			r.Post("/resend", module.V1CRMCallbackController.ResendCallback)
		})

		// Card-funded Payout Endpoints
		r.Route("/card-funded-payouts", func(r chi.Router) {
			r.Route("/transactions", func(r chi.Router) {
				r.Route("/payouts", func(r chi.Router) {
					r.Get("/", module.V1CRMCardFundedPayoutController.GetPayoutTransactionList)
					r.Patch("/{payoutId}/status", module.V1CRMCardFundedPayoutController.PatchPayoutTransactionStatus)
				})
			})
		})

		// tnc (terms and conditions) version management
		r.Route("/tncs", func(r chi.Router) {
			r.Get("/", module.V1CRMTNCController.List)
			r.Post("/", module.V1CRMTNCController.Create)
			r.Get("/{id}", module.V1CRMTNCController.Detail)
			r.Patch("/{id}/activate", module.V1CRMTNCController.Activate)
			r.Patch("/{id}/deactivate", module.V1CRMTNCController.Deactivate)
		})
	})

	// Supported Automated Endpoint
	r.Route("/api/v1/tests", func(r chi.Router) {
		r.Use(customMiddleware.KeyAuth(constant.HeaderXAutomatedTestKey, module.Secret.InternalApiKeySecret.AutomatedTest))
		r.Route("/users", func(r chi.Router) {
			r.Get("/invitations/{encoded_email}", module.V1UserController.GetInvitationURL)
		})
	})

	// Open API
	snapAuthMiddleware := openApi.NewSnapAuthMiddleware(module.Logger, module.IJwt, module.V1MerchantService)
	snapIdempotencyAuthMiddleware := openApi.NewIdempotentMiddleware(module.Logger, *module.Cfg, *module.Secret, module.Redis)
	r.Route("/open-api", func(r chi.Router) {
		r.Use(customMiddleware.InternalServiceMiddleware(module.Secret))
		r.Route("/snap/v1", func(r chi.Router) {
			r.Use(openApi.IdentitySnapServiceCode)

			// Placeholder for Open-API endpoints with SNAP standard
			r.Group(func(r chi.Router) {
				r.Use(customMiddleware.IPWhitelistMiddleware(module.V1IPWhitelistService, constant.HeaderXClientKey))
				r.Post("/access-token/b2b", module.V1InternalMerchantAuthController.GetSNAPAccessTokenB2B)
			})
			r.Group(func(r chi.Router) {
				r.Use(snapAuthMiddleware.Authorize)
				r.Use(customMiddleware.IPWhitelistMiddleware(module.V1IPWhitelistService, constant.HeaderAuthorization))
				r.Use(customMiddleware.MerchantRateLimiterMiddleware(module.V1RateLimitService, module.Cfg))
				r.Use(snapIdempotencyAuthMiddleware.SnapIdempotentMiddleware("qr"))
				r.Post("/payments/query/qr-mpm-static", module.V1InternalPaymentController.SNAPQueryQrMpmStatic)
				r.Post("/payments/query/qr-mpm-dynamic", module.V1InternalPaymentController.SNAPQueryQrMpmDynamic)
				r.Post("/payments/qr/qr-mpm-generate", module.V1InternalPaymentController.SNAPGenerateQRMpm)
				r.Post("/payments/transfer-va/create-va", module.V1InternalPaymentController.SNAPCreateVA)
				r.Put("/payments/transfer-va/update-va", module.V1InternalPaymentController.SNAPUpdateVA)
				r.Post("/payments/transfer-va/get-va", module.V1InternalPaymentController.SNAPGetVA)
			})
		})

		r.Route("/v1", func(r chi.Router) {
			r.Use(customMiddleware.MerchantAuthMiddleware(module.IJwt))
			r.Use(customMiddleware.IPWhitelistMiddleware(module.V1IPWhitelistService, constant.HeaderAuthorization))
			r.Use(customMiddleware.MerchantRateLimiterMiddleware(module.V1RateLimitService, module.Cfg))
			r.Use(customMiddleware.CheckSubMerchantMiddleware(module.V1MerchantService, module.V1ProductService))
			r.Use(customMiddleware.MerchantStatusMiddleware(
				module.V1MerchantService, module.Cfg, func(ctx context.Context, w http.ResponseWriter, err error) {
					if errors.Is(err, constant.ErrMerchantNotFound) {
						if id, _ := ctx.Value(constant.CtxSubMerchantIDKey).(string); id != "" {
							ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("sub-account", id)) // Standardize error response
						}
					}
					response.SendOpenApiNonSnapResponseError(ctx, w, err)
				},
			))
			r.Use(customMiddleware.CheckMerchantMiddleware(module.V1MerchantService))

			r.Route("/inquiry-account", func(r chi.Router) {
				r.Use(customMiddleware.SetErrorSourceMiddleware())
				r.Post("/", module.V1InternalAccountInquiryController.RequestAccountInquiry)
			})
			r.Route("/payouts", func(r chi.Router) {
				r.Use(customMiddleware.SetErrorSourceMiddleware())
				r.Use(customMiddleware.InboundFeatureMiddleware(constant.InboundFeaturePayout))

				r.Group(func(r chi.Router) {
					r.Use(customMiddleware.CheckMerchantForbiddenUsecase(module.V1MerchantForbiddenUseCaseService, constant.ReferenceDisbursement))
					r.Use(snapIdempotencyAuthMiddleware.InternalIdempotencyMiddleware("payouts", constant.HeaderXRequestId))
					r.Post("/", module.V1InternalPayoutController.Create)
				})

				r.Get("/{id}", module.V1InternalPayoutController.FindByBulkId)
				r.Get("/balance", module.V1InternalAccountController.GetBalance)
				r.Post("/{id}/retry", module.V1InternalPayoutController.RetryBulk)
			})

			r.Route("/payment-methods", func(r chi.Router) {
				r.Use(customMiddleware.SetErrorSourceMiddleware())
				r.Get("/virtual-accounts", module.V1InternalPaymentMethodController.GetVAPaymentMethods)
				r.Get("/virtual-accounts/{paymentMethodId}/top-up", module.V1InternalPaymentMethodController.TopUpVAPaymentMethod) // Payout Only
			})

			r.Route("/balances", func(r chi.Router) {
				r.Use(customMiddleware.SetErrorSourceMiddleware())
				r.Get("/", module.V1InternalAccountController.GetBalance)
				r.Post("/sub-merchants", module.V1InternalPlatformController.GetSubMerchantBalance)
			})

			r.Route("/xb", func(r chi.Router) {
				r.Use(customMiddleware.SetV2ErrorCodeMiddleware())

				r.Get("/fx-rate", module.V1InternalXbController.GetFxRate)

				r.Route("/payouts", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(customMiddleware.CheckMerchantForbiddenUsecase(module.V1MerchantForbiddenUseCaseService, constant.ReferenceDisbursement))
						r.Use(snapIdempotencyAuthMiddleware.InternalIdempotencyMiddleware("xb", constant.HeaderXRequestId))
						r.Post("/", module.V1InternalXbController.CreatePayoutSession)
					})

					r.Post("/{id}/upload", module.V1InternalXbController.UploadUnderlyingDocument)
					r.Post("/{id}/confirm", module.V1InternalXbController.ConfirmPayoutSession)
					r.Get("/{id}", module.V1InternalXbController.GetPayoutById)
					r.Get("/list", module.V1InternalXbController.GetList)
				})

				r.Route("/master", func(r chi.Router) {
					r.Route("/country", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterCountry)
					})

					r.Route("/state", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterState)
					})

					r.Route("/city", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterCity)
					})

					r.Route("/currency", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterCurrency)
						r.Route("/map", func(r chi.Router) {
							r.Get("/list", module.V1InternalXbController.GetListMasterCurrencyMapping)
						})
					})

					r.Route("/identification-type", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterIdentificationType)
					})

					r.Route("/account-type", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterAccountType)
					})

					r.Route("/purpose", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterPurpose)
					})

					r.Route("/transfer-method", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterTransferMethod)
					})

					r.Route("/source-of-income", func(r chi.Router) {
						r.Get("/list", module.V1InternalXbController.GetListMasterSourceOfIncome)
					})

					r.Route("/swift", func(r chi.Router) {
						pathSwift := "/api/v1/master/swift"
						r.Get("/", func(w http.ResponseWriter, r *http.Request) {
							code := r.URL.Query().Get("code")
							if code != "" {
								pathSwift = fmt.Sprintf("%s?code=%s", pathSwift, code)
							}

							module.V1InternalXbController.ProxyHandler(pathSwift, nil)(w, r)
						})
						r.Post("/", func(w http.ResponseWriter, r *http.Request) {
							module.V1InternalXbController.ProxyHandler(pathSwift, nil)(w, r)
						})
					})

					r.Route("/bank", func(r chi.Router) {
						pathBank := "/api/v1/master/bank"
						r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
							module.V1InternalXbController.ProxyHandlerWithQueryConversion(pathBank+"/list", nil)(w, r)
						})
					})
				})

				r.Route("/beneficiary", func(r chi.Router) {
					r.Post("/", module.V1InternalXbController.CreateBeneficiary)
					r.Get("/list", module.V1InternalXbController.GetListBeneficiary)
					r.Get("/{id}", module.V1InternalXbController.GetBeneficiaryById)
					r.Put("/{id}", module.V1InternalXbController.UpdateBeneficiary)
					r.Patch("/{id}/deactivate", module.V1InternalXbController.DeactivateBeneficiary)
				})

				r.Route("/sender", func(r chi.Router) {
					r.Post("/", module.V1InternalXbController.CreateSender)
					r.Get("/list", module.V1InternalXbController.GetListSender)
					r.Get("/{id}", module.V1InternalXbController.GetSenderById)
					r.Put("/{id}", module.V1InternalXbController.UpdateSender)
					r.Patch("/{id}/deactivate", module.V1InternalXbController.DeactivateSender)
				})

				r.Route("/payout/rfi", func(r chi.Router) {
					r.Post("/{id}", module.V1InternalXbController.SubmitRfiDetails)
					r.Get("/{id}", module.V1InternalXbController.GetRfiDetails)
				})
			})
			r.Route("/sub-merchants", func(r chi.Router) {
				r.Use(customMiddleware.SetErrorSourceMiddleware())
				r.Use(customMiddleware.InboundFeatureMiddleware(constant.InboundFeaturePlatform))
				r.Use(customMiddleware.MerchantProductValidationMiddleware(module.V1ProductService, constant.ProductPlatform))
				r.Get("/", module.V1InternalSubMerchantController.ListSubMerchantByParentID)
				r.Get("/{id}", module.V1InternalSubMerchantController.DetailSubMerchantByID)
				r.Get("/balance/{id}", module.V1InternalSubMerchantController.GetSubMerchantBalance)
				r.Post("/", module.V1InternalSubMerchantController.Create)
				r.Put("/{id}", module.V1InternalSubMerchantController.Update)
				r.Post("/admin", module.V1InternalSubMerchantController.AssignAdmin)
				r.Post("/users/resend-invitation", module.V1InternalSubMerchantController.ResendInvitation)
			})

			r.Route("/countries", func(r chi.Router) {
				r.Get("/", module.V1CountryController.GetAll)
			})

			r.Route("/industries", func(r chi.Router) {
				r.Get("/", module.V1IndustryController.GetAll)
			})

			r.Route("/merchants", func(r chi.Router) {
				r.Route("/forbidden-usecase", func(r chi.Router) {
					r.Post("/block", module.V1InternalMerchantController.Block)
					r.Post("/unblock", module.V1InternalMerchantController.Unblock)
				})
			})

			r.Route("/cards", func(r chi.Router) {
				r.Post("/payment-session", module.V1InternalCreditCardController.CreatePayment)
				r.Get("/payment-session/{uuid}", module.V1InternalCreditCardController.GetPaymentById)
			})

			r.Route("/payments", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(snapIdempotencyAuthMiddleware.InternalIdempotencyMiddleware("payments", constant.HeaderXRequestId))
					r.Post("/", module.V1InternalUnifiedPaymentController.Create)
				})
				r.Patch("/", module.V1InternalUnifiedPaymentController.Update)
				r.Get("/{referenceId}", module.V1InternalUnifiedPaymentController.FindPaymentByReferenceId)
			})

			r.Route("/refunds", func(r chi.Router) {
				r.Use(customMiddleware.InboundFeatureMiddleware(constant.InboundFeaturePayment))
				r.Post("/", module.V1InternalRefundController.Create)
				r.Get("/", module.V1InternalRefundController.GetList)
				r.Get("/{uuid}", module.V1InternalRefundController.GetByID)
			})

			r.Route("/recurring", func(r chi.Router) {
				r.Use(customMiddleware.InboundFeatureMiddleware(constant.InboundFeatureRecurringPayment))
				r.Post("/", module.V1InternalRecurringContractController.Create)
				r.Post("/{uuid}/cancel", module.V1InternalRecurringContractController.Cancel)
			})

			r.Get("/balance-histories", module.V1OrchestratorController.GetOpenApiBalanceHistories)

			r.Route("/encrypt-card", func(r chi.Router) {
				r.Post("/", module.V2InternalUnifiedPaymentController.EncryptCard)
				r.Get("/{uuid}", module.V2InternalUnifiedPaymentController.GetEncryptedCard)
			})

			r.Route("/customers", func(r chi.Router) {
				r.Get("/{id}", module.V1InternalCustomerController.GetByIDForUnifiedPayment)
			})

			r.Route("/withdrawals", func(r chi.Router) {
				r.Use(customMiddleware.SetErrorSourceMiddleware())
				r.Use(customMiddleware.CheckMerchantForbiddenUsecase(module.V1MerchantForbiddenUseCaseService, constant.ReferenceWithdrawal))
				r.Post("/", module.V1InternalWithdrawalController.Withdraw)
				r.Get("/{id}", module.V1InternalWithdrawalController.GetWithdrawalByID)
			})
		})

		r.Route("/v2", func(r chi.Router) {
			r.Use(customMiddleware.SetV2ErrorCodeMiddleware())
			r.Use(customMiddleware.MerchantAuthMiddleware(module.IJwt))
			r.Use(customMiddleware.IPWhitelistMiddleware(module.V1IPWhitelistService, constant.HeaderAuthorization))
			r.Use(customMiddleware.MerchantRateLimiterMiddleware(module.V1RateLimitService, module.Cfg))
			r.Use(customMiddleware.MerchantStatusMiddleware(
				module.V1MerchantService, module.Cfg, response.SendOpenApiNonSnapResponseError,
			))
			r.Use(customMiddleware.CheckSubMerchantMiddleware(module.V1MerchantService, module.V1ProductService))

			r.Route("/payments", func(r chi.Router) {
				r.Use(customMiddleware.InboundFeatureMiddleware(constant.InboundFeaturePayment))
				r.Group(func(r chi.Router) {
					r.Use(snapIdempotencyAuthMiddleware.InternalIdempotencyMiddleware("payments", constant.HeaderXRequestId))
					r.Post("/", module.V2InternalUnifiedPaymentController.Create)
				})
				r.Group(func(r chi.Router) {
					r.Use(snapIdempotencyAuthMiddleware.InternalIdempotencyMiddleware("payments-confirm", constant.HeaderXRequestId))
					r.Post("/{uuid}/confirm", module.V2InternalUnifiedPaymentController.Confirm)
				})
				r.Post("/{uuid}/cancel", module.V2InternalUnifiedPaymentController.Cancel)
				r.Get("/{uuid}", module.V2InternalUnifiedPaymentController.GetByID)
				r.Get("/", module.V2InternalUnifiedPaymentController.GetList)
				r.Post("/{uuid}/capture", module.V2InternalUnifiedPaymentController.Capture)
				r.Get("/bin/{binNumber}", module.V2InternalUnifiedPaymentController.GetBinDetailByBinNumber)
				r.Post("/{uuid}/investigation/proof-of-payment", module.V2InternalUnifiedPaymentController.UploadProofOfPayment)
			})

			r.Route("/payments/simulations", func(r chi.Router) {
				r.Post("/", module.V2InternalUnifiedPaymentController.SimulatePayment)
			})

			r.Route("/charges", func(r chi.Router) {
				r.Use(customMiddleware.InboundFeatureMiddleware(constant.InboundFeaturePayment))
				r.Get("/", module.V2InternalUnifiedPaymentController.GetChargeList)
				r.Get("/{uuid}", module.V2InternalUnifiedPaymentController.GetChargeByID)
			})
			r.Route("/payment-method-configs", func(r chi.Router) {
				r.Use(customMiddleware.InboundFeatureMiddleware(constant.InboundFeaturePayment))
				r.Get("/", module.V2InternalUnifiedPaymentController.GetPaymentMethodConfig)
			})
		})
	})

	return r
}
