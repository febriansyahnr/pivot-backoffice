package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	// Config and Package
	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/conductor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	jwtCore "github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	pkgMonitor "github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt/stream"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/slackExt"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	"golang.org/x/sync/errgroup"

	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/amqp"
	chiExtMiddleware "github.com/paper-indonesia/pdk/v2/chiExt/middleware"
	pdkGoff "github.com/paper-indonesia/pdk/v2/goff"
	pdkNotifier "github.com/paper-indonesia/pdk/v2/goff/notifier"
	pdkRetriever "github.com/paper-indonesia/pdk/v2/goff/retriever"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	pdkMySql "github.com/paper-indonesia/pdk/v2/mySqlExt"
	pdkNewRelic "github.com/paper-indonesia/pdk/v2/newRelicExt"
	"github.com/paper-indonesia/pdk/v2/otelExt"
	pdkRedis "github.com/paper-indonesia/pdk/v2/redisExt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	bankAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/bankAccount"
	beneficiaryAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/beneficiaryAccount"
	callbackRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/callback"
	countryRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/country"
	creditcardCoreProcessorRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/creditcardCoreProcessor"
	customerRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/customer"
	dailyAccountTransactionRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/dailyAccountTransaction"
	disbursementRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	fraudnetrepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fdsProcessor/fraudNetRepository"
	sokratechRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fdsProcessor/sokratech"
	feeRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fee"
	fraudrulesrepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fraudRules"
	industryRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/industry"
	addrLocRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/location"
	menuRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/menu"
	merchantRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	merchantforbiddenusecaseRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchantForbiddenUsecase"
	merchantTopUpRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchantTopUp"
	outboundRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/outbound"
	paperCommRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/paperCommunication"
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
	roleRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/role"
	roleMenuPermissionRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/roleMenuPermission"
	ruleevaluationsrepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/ruleEvaluations"
	settlementHoldRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/settlementHold"
	snapCoreRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/snapCore"
	statusHistoriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/statusHistories"
	transferRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/transfer"
	userRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/user"
	userLoggedInDeviceRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/userLoggedInDevice"
	userRoleRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/userRole"
	vendorRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/vendor"
	withdrawalRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/withdrawal"
	xbCoreProcessorRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/xbCoreProcessor"

	// Service
	accountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/account"
	activityService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/activity"
	beneficiaryAccountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/beneficiaryAccount"
	callbackService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	cardFundedPayoutService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/cardFundedPayout"
	commService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/commService"
	"github.com/paper-indonesia/pivot-backoffice/internal/service/v1/country"
	creditcardService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/creditcard"
	customerService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/customer"
	disbursementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	fdsservice "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fds"
	feeService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"
	industryService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/industry"
	merchantService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	merchantForbiddenUsecaseService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchantForbiddenUsecase"
	merchantTopUpService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchantTopUp"
	notificationService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/notification"
	orchestratorService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/orchestrator"
	otpService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/otp"
	paymentService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	paymentMethodService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/paymentMethod"
	permissionService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/permission"
	platformFeeService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/platformFee"
	qrisService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/qris"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/rateLimiter"
	reconService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/reconciliation"
	recurringContractService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/recurringContract"
	refundService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/refund"
	refundProcessorService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/refundProcessor"
	reportingService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/reporting"
	roleService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/role"
	settlementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/settlement"
	settlementHoldService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/settlementHold"
	transferService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/transfer"
	userService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/user"
	userLoggedInDeviceService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/userLoggedInDevice"
	userRoleService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/userRole"
	vendorService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/vendor"
	withdrawalService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	xbPayoutService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/xbPayout"
	chargeMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/charge"
	ledgerService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/ledger"
	p2pMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/p2p"
	payInMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/payIn"
	payoutMoneyFlowService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/payOut"
	unifiedPaymentService "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	callbackPartnerService "github.com/paper-indonesia/pivot-backoffice/pkg/callback"

	// Controller
	accountConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/account"
	activityConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/activity"
	bankTransferConsumer "github.com/paper-indonesia/pivot-backoffice/port/consumer/bankTransfer"
	callbackConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/callback"
	commServiceController "github.com/paper-indonesia/pivot-backoffice/port/consumer/commService"
	creditcardConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/creditcard"
	disbursementConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/disbursement"
	merchantConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/merchant"
	notificationConsumer "github.com/paper-indonesia/pivot-backoffice/port/consumer/notification"
	paymentConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/payment"
	paymentCaptureConsumer "github.com/paper-indonesia/pivot-backoffice/port/consumer/paymentCapture"
	qrisController "github.com/paper-indonesia/pivot-backoffice/port/consumer/qris"
	reconConsumer "github.com/paper-indonesia/pivot-backoffice/port/consumer/reconciliation"
	refundConsumer "github.com/paper-indonesia/pivot-backoffice/port/consumer/refund"
	reportingConsumer "github.com/paper-indonesia/pivot-backoffice/port/consumer/reporting"
	settlementConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/settlement"
	slackConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/slack"
	withdrawalConsumer "github.com/paper-indonesia/pivot-backoffice/port/consumer/withdrawal"
	xbPayoutConsumerController "github.com/paper-indonesia/pivot-backoffice/port/consumer/xbPayout"
	callbackWorker "github.com/paper-indonesia/pivot-backoffice/port/worker/callback"
)

func init() {
	rootCmd.AddCommand(serveConsumerCmd)
}

func initConsumerPkg(ctx context.Context) (
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
	slack slackExt.SlackNotifier,
	httpRequestClient httpRequestExt.IHTTPRequest,
	jwtConfig jwtCore.IJwt,
	encryptExt encryption.ICrypto,
	vaultClient *vault.Client,
	gcsEncryption encryption.GCSClient,
	rabbitMqStream *stream.Client,
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
	pdkLog, err = pdkLogger.NewZapLogger(
		pdkLogger.Config{
			IsDevelopment: isDevelopment,
			Environment:   cfg.Environment,
			ServiceName:   cfg.ServiceName + "-consumer",
		},
		pdkLogger.WithZapMaskingSensitiveData(strings.Split(cfg.AppConfig.MaskingSensitiveData, ",")),
	)
	if err != nil {
		fmt.Printf("Unable to init pdk logger, %v", err)
		panic(err)
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
			ServiceName:  cfg.ServiceName + "-consumer",
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
			ServiceName: cfg.ServiceName + "-consumer-" + cfg.Environment,
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
			cfg.ServiceName+"-consumer-"+cfg.Environment,
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

	// Init Slack
	slack = slackExt.NewSlackExt("Payment Gateway", "Backend Portal")

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

	// Init Encryption
	encryptExt = encryption.New()

	gcsEncryption = encryption.NewGCS(secret)

	// Init Vault Client
	vaultClient, err = vault.New(vault.Config{
		Address: cfg.Vault.Address,
		Token:   secret.Vault.Token,
	})
	if err != nil {
		panic("Unable to init Vault client, " + err.Error())
	}

	// RabbitMQ Stream
	streamConfig := stream.Config{
		Host:             cfg.RabbitMQStream.Host,
		Port:             cfg.RabbitMQStream.Port,
		VHost:            cfg.RabbitMQStream.VHost,
		Username:         secret.RabbitMQStream.Username,
		Password:         secret.RabbitMQStream.Password,
		HeartbeatSeconds: cfg.RabbitMQStream.HeartbeatInSeconds,
		NR:               nr,
	}
	if rabbitMqStream, err = stream.New(streamConfig); err != nil {
		panic("Unable to init RabbitMQ Stream, " + err.Error())
	}
	closes = append(closes, func() {
		if closeErr := rabbitMqStream.Close(); closeErr != nil {
			fmt.Printf("Error close RabbitMQ Stream: %s\n", closeErr)
		}
	})

	return cfg, secret, pdkLog, log, otel, nr, monitor, mySqlDB, redisClient, rabbitMq, gcsClient, slack, httpRequestClient, jwtConfig, encryptExt, vaultClient, gcsEncryption, rabbitMqStream
}

// TODO: Tech Debt
// 1. Logger need to move to pdkLogger, currently only to support PDK V2
// 2. Redis need to refactor to not have extra layer in pkg, right now focusing on tracer and metric centralize in PDK

var serveConsumerCmd = &cobra.Command{
	Use:   "serveConsumer",
	Short: "Start Consumer",
	Long:  `Start Backend Portal Consumer`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// Init Consumer PKG
		config, secret, pdkLog, logger, otelExt, newRelicExt, _, dbClient, cacheClient, rmqExt, gcs, slack, httpExt, jwtConfig, encryptExt, vaultClient, _, rabbitMqStream := initConsumerPkg(ctx)
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
			ctxShutdown, stop := context.WithTimeout(context.Background(), 60*time.Second)
			defer stop()

			groups := new(errgroup.Group)
			groups.Go(func() error { return otelExt.MeterProvider().ForceFlush(ctxShutdown) })
			groups.Go(func() error { return otelExt.TracerProvider().ForceFlush(ctxShutdown) })
			if errOtel := groups.Wait(); errOtel != nil {
				fmt.Printf("Error force flush otel provider: %v\n", errOtel)
			}

			if errOtel := otelExt.Shutdown(ctxShutdown); errOtel != nil {
				fmt.Printf("Error shutting down otel: %v\n", errOtel)
			}
		}()
		defer newRelicExt.GetApp().Shutdown(10 * time.Second)
		defer dbClient.Close()
		defer cacheClient.Close()
		defer rmqExt.Close()
		defer gcs.Close()
		defer rabbitMqStream.Close()

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
			Logger:         conductor.NewZapLogger(logger.GetLogger()),
		})
		if err != nil {
			panic("Unable init conductor client, error: " + err.Error())
		}

		// Vault
		merchantCredsEncryption := vaultClient.NewTransit(config.Vault.Secrets.MerchantCredentials.SecretPath, config.Vault.Secrets.MerchantCredentials.SecretKey)

		// Disbursement Worker Pools
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

		accounttransactionRepo := accounttransactionRepository.New(
			dbClient,
			pdkLog)

		accountRepo := accountRepository.New(
			dbClient,
			pdkLog)

		callbackRepo := callbackRepository.New(
			dbClient,
			pdkLog)

		activityRepositoryFactory := &activityRepository.ActivityRepository{
			Mongo:  nil,
			Mysql:  dbClient,
			Logger: pdkLog,
		}
		activityRepository := activityRepositoryFactory.CreateRepository(activityRepository.MySQLType)
		paymentRepository := paymentRepository.New(dbClient, pdkLog)
		paymentMethodRepository := paymentMethodRepository.New(dbClient, pdkLog)
		customerRepo := customerRepository.New(dbClient, pdkLog)
		merchantRepo := merchantRepository.New(dbClient, pdkLog, merchantRepository.WithServiceConfig(config))
		userRepository := userRepository.New(dbClient, pdkLog)
		passHistory := passwordHistoriesRepository.New(dbClient, pdkLog)
		disbursementRepo := disbursementRepository.New(dbClient, pdkLog)
		statusHistoriesRepo := statusHistoriesRepository.New(dbClient)
		paymentCaptureRepo := paymentCaptureRepository.New(dbClient, pdkLog)
		accountInquiriesRepo := accountInquriesRepository.New(dbClient, pdkLog)
		beneficiaryAccountRepo := beneficiaryAccountRepository.New(dbClient, pdkLog)
		forbiddenUsecaseRepo := merchantforbiddenusecaseRepository.New(dbClient, pdkLog)
		qrisRepository := qrisRepository.New(dbClient)
		feeRepo := feeRepository.New(dbClient, pdkLog)
		xbCoreProcessorRepo := xbCoreProcessorRepository.New(config, secret, pdkLog, httpExt)
		creditcardCoreProcessorRepo := creditcardCoreProcessorRepository.New(config, secret, pdkLog, httpExt)
		paperCommRepo := paperCommRepository.New(&config.PaperCommunication, pdkLog, httpExt)
		bankAccountRepo := bankAccountRepository.New(dbClient, pdkLog)
		withdrawalRepo := withdrawalRepository.New(dbClient)
		addrLocRepo := addrLocRepository.New(dbClient)
		permissionRepo := permissionRepository.New(dbClient, pdkLog)
		menuRepo := menuRepository.New(dbClient, pdkLog)
		userLoggedInDeviceRepo := userLoggedInDeviceRepository.New(dbClient, pdkLog)
		roleMenuPermRepo := roleMenuPermissionRepository.New(dbClient, pdkLog)
		roleRepo := roleRepository.New(dbClient, pdkLog)
		userRoleRepo := userRoleRepository.New(dbClient, pdkLog)
		rateLimiterRepository := rateLimiterRepository.New(dbClient, pdkLog)
		productRepo := productRepository.New(dbClient, pdkLog)
		industryRepo := industryRepository.New(dbClient, pdkLog)
		countryRepo := countryRepository.New(dbClient, pdkLog)
		vendorRepo := vendorRepository.New(pdkLog, dbClient)

		reconRepo := reconRepository.New(dbClient, pdkLog)
		transferRepo := transferRepository.New(dbClient, pdkLog)
		dailyAccountTrxRepo := dailyAccountTransactionRepository.New(dbClient, pdkLog)
		merchantTopUpRepo := merchantTopUpRepository.New(dbClient, pdkLog)
		refundRepo := refundRepository.New(dbClient, pdkLog)
		fraudRulesRepo := fraudrulesrepository.New(pdkLog, dbClient)
		ruleEvaluationsRepo := ruleevaluationsrepository.New(pdkLog, dbClient)
		fraudNetRepo := fraudnetrepository.New(config, secret, pdkLog, httpExt)
		recurringContractRepo := recurringContractRepository.New(pdkLog, dbClient)
		sokratechRepo := sokratechRepository.New(config.Sokratech, secret.Sokratech, httpExt, pdkLog)
		settlementHoldRepo := settlementHoldRepository.New(dbClient, pdkLog)
		reportingRepository := reportingRepository.New(dbClient, pdkLog, config.AppConfig)
		payoutManualProcessingAccountRepo := payoutManualProcessingAccountRepository.New(pdkLog, dbClient)

		// Init 3rd Party service
		// e.g. Callback Partner Service, snap core service, etc
		merchantCallbackHTTPClient := httputil.NewHTTPClient(
			httputil.ServiceConfig(config.HTTPClients.MerchantCallback), httputil.WithLogger(pdkLog),
		)
		defer merchantCallbackHTTPClient.CloseIdleConnections()
		callbackPartnerSvc := callbackPartnerService.New(logger, merchantCallbackHTTPClient)

		// Init service
		// e.g. business logic, etc.
		notificationSvc := notificationService.New(config, pdkLog, rmqExt)
		otpSvc := otpService.New(config, pdkLog, cacheClient, jwtConfig, rmqExt, userRepository, redisExt.NewLimiter(cacheClient.Client()))
		userLoggedInDeviceSvc := userLoggedInDeviceService.New(config, secret, pdkLog, userLoggedInDeviceService.Repositories{
			UserRepo:               userRepository,
			UserLoggedInDeviceRepo: userLoggedInDeviceRepo,
		})
		permissionSvc := permissionService.New(permissionRepo, pdkLog, permissionService.WithRedisClient(cacheClient))
		roleSvc := roleService.New(
			roleRepo, pdkLog,
			roleService.WithMenuRepository(menuRepo), roleService.WithRoleMenuPermissionRepository(roleMenuPermRepo), roleService.WithUserRoleRepository(userRoleRepo),
			roleService.WithRedisClient(cacheClient),
		)
		rateLimiterService := ratelimiter.New(pdkLog, cacheClient, rateLimiterRepository, ratelimiter.WithRedisLimiter(redisExt.NewLimiter(cacheClient.Client())), ratelimiter.WithConfig(config))
		userRoleSvc := userRoleService.New(userRoleRepo, pdkLog)

		userSvc := userService.New(
			config, secret, pdkLog, userRepository, passHistory,
			userService.WithJWT(jwtConfig),
			userService.WithRedisClient(cacheClient),
			userService.WithRateLimiter(rateLimiterService),
			userService.WithRabbitMQClient(rmqExt),
			userService.WithOTPService(otpSvc),
			userService.WithUserLoggedInDeviceService(userLoggedInDeviceSvc),
			userService.WithPermissionService(permissionSvc),
			userService.WithLimiter(redisExt.NewLimiter(cacheClient.Client())),
			userService.WithRoleService(roleSvc),
			userService.WithUserRoleService(userRoleSvc),
			userService.WithUserLoggedInDeviceRepo(userLoggedInDeviceRepo),
		)

		beneficiaryAccountSvc := beneficiaryAccountService.New(pdkLog, beneficiaryAccountRepo, accountInquiriesRepo, snapCoreRepository,
			beneficiaryAccountService.WithConfig(config),
		)

		orchestratorSvc := orchestratorService.New(
			pdkLog,
			gcs,
			accounttransactionRepo,
			accountRepo,
			orchestratorService.WithRedisClient(cacheClient),
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
		vendorSvc := vendorService.New(vendorRepo, pdkLog)

		accountSvc := accountService.New(pdkLog, accounttransactionRepo, accountRepo, dailyAccountTrxRepo)
		merchantSvc := merchantService.New(merchantRepo, pdkLog, userRepository, jwtConfig, rmqExt, encryptExt,
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
			merchantService.WithPaymentMethodRepository(paymentMethodRepository),
			merchantService.WithQrisService(qrisService),
			merchantService.WithIndustryService(industrySvc),
			merchantService.WithCountryService(countrySvc),
			merchantService.WithExcelLibrary(xlsx.New()),
			merchantService.WithVaultTransit(merchantCredsEncryption),
		)
		paymentMethodService := paymentMethodService.New(pdkLog, paymentMethodRepository, snapCoreRepository, creditcardCoreProcessorRepo,
			paymentMethodService.WithQrisService(qrisService),
			paymentMethodService.WithMerchantService(merchantSvc),
			paymentMethodService.WithConfig(config),
			paymentMethodService.WithRedisClient(cacheClient),
			paymentMethodService.WithMerchantRepository(merchantRepo),
		)
		feeSvc := feeService.New(pdkLog, feeRepo, merchantRepo, feeService.WithPaymentMethodService(paymentMethodService), feeService.WithRedisClient(cacheClient), feeService.WithConfig(config), feeService.WithAccountTransactionRepository(accounttransactionRepo))
		merchantTopUpSvc := merchantTopUpService.New(
			config, pdkLog, paymentMethodRepository, merchantTopUpRepo, snapCoreRepository,
			merchantTopUpService.WithRabbitMQClient(rmqExt),
			merchantTopUpService.WithMerchantService(merchantSvc),
			merchantTopUpService.WithOrchestratorService(orchestratorSvc),
			merchantTopUpService.WithFeeService(feeSvc),
		)

		callbackSvc := callbackService.New(
			pdkLog, cacheClient, callbackRepo, callbackPartnerSvc, merchantSvc,
			callbackService.WithMerchantRepository(merchantRepo),
			callbackService.WithVaultTransit(merchantCredsEncryption),
		)
		activitySvc := activityService.New(activityRepository)
		merchantForbiddenUsecaseSvc := merchantForbiddenUsecaseService.New(pdkLog, forbiddenUsecaseRepo, rmqExt, merchantSvc)
		// Platform Service Preparation
		customerService := customerService.New(customerRepo, accountSvc, pdkLog)
		ledgerSvc := ledgerService.New(pdkLog, accounttransactionRepo, accountRepo, merchantSvc, customerService, accountSvc)
		payInMoneyFlowSvc := payInMoneyFlowService.New(pdkLog, accounttransactionRepo, accountSvc, merchantSvc)
		p2pMoneyFlowSvc := p2pMoneyFlowService.New(pdkLog, accounttransactionRepo, accountSvc, ledgerSvc, merchantSvc)
		payOutMoneyFLowSvc := payoutMoneyFlowService.New(pdkLog, accounttransactionRepo, accountSvc, ledgerSvc, merchantSvc)
		chargeMoneyFLowSvc := chargeMoneyFlowService.New(pdkLog, accounttransactionRepo, accountSvc, ledgerSvc, merchantSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypePayIn, payInMoneyFlowSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypeP2P, p2pMoneyFlowSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypePayOut, payOutMoneyFLowSvc)
		ledgerService.WithMoneyFlowService(ledgerSvc, constant.TransferTypeCharge, chargeMoneyFLowSvc)
		platformFeeSvc := platformFeeService.New(pdkLog, ledgerSvc, feeSvc, accountSvc)
		transferSvc := transferService.New(pdkLog, ledgerSvc, accountSvc, platformFeeSvc, merchantSvc, transferRepo)

		disbursementSvc := disbursementService.New(
			config, pdkLog, merchantRepo, disbursementRepo, snapCoreRepository, bankAccountRepo,
			disbursementService.WithOrchestratorService(orchestratorSvc),
			disbursementService.WithBeneficiaryAccService(beneficiaryAccountSvc),
			disbursementService.WithRabbitMQClient(rmqExt),
			disbursementService.WithGoogleCloudStorage(gcs),
			disbursementService.WithRedisClient(cacheClient),
			disbursementService.WithMerchantForbiddenUseCaseService(merchantForbiddenUsecaseSvc),
			disbursementService.WithFeeService(feeSvc),
			disbursementService.WithAccountTransactionRepository(accounttransactionRepo),
			disbursementService.WithDisbursementWorkerPool(disbursementWorkerPools),
			disbursementService.WithTransferService(transferSvc),
			disbursementService.WithLedgerService(ledgerSvc),
			disbursementService.WithStatusHistoriesRepository(statusHistoriesRepo),
			disbursementService.WithWorkflowFDSRepository(sokratechRepo),
			disbursementService.WithPayoutManualProcessingAccountRepository(payoutManualProcessingAccountRepo),
		)
		defer disbursementSvc.WPRelease()

		paymentLedgerSvc := paymentService.NewPaymentLedgerService(config, pdkLog, paymentRepository, merchantRepo, accounttransactionRepo, orchestratorSvc, feeSvc, transferSvc, ledgerSvc, rmqExt)
		creditcardSvc := creditcardService.New(config, pdkLog, rmqExt, paymentRepository, paymentMethodRepository, creditcardCoreProcessorRepo,
			creditcardService.WithFeeService(feeSvc),
			creditcardService.WithOrchestratorService(orchestratorSvc),
			creditcardService.WithPaymentLedgerService(paymentLedgerSvc),
			creditcardService.WithCustomerRepo(customerRepo),
			creditcardService.WithMerchantRepo(merchantRepo),
			creditcardService.WithAccountTransactionRepo(accounttransactionRepo),
			creditcardService.WithPaymentMethodService(paymentMethodService),
			creditcardService.WithRedis(cacheClient),
		)
		paymentSvc := paymentService.New(paymentRepository, pdkLog, snapCoreRepository, customerRepo, merchantRepo, paymentMethodRepository, accountRepo,
			paymentService.WithOrchestratorService(orchestratorSvc),
			paymentService.WithAccountTransactionRepository(accounttransactionRepo),
			paymentService.WithRabbitMQClient(rmqExt),
			paymentService.WithQrisService(qrisService),
			paymentService.WithConfig(config),
			paymentService.WithFeeService(feeSvc),
			paymentService.WithTransferService(transferSvc),
			paymentService.WithCreditCardService(creditcardSvc),
			paymentService.WithPaymentMethodService(paymentMethodService),
			paymentService.WithLedgerService(ledgerSvc),
			paymentService.WithStatusHistoriesRepository(statusHistoriesRepo),
			paymentService.WithRedisClient(cacheClient),
			paymentService.WithDisbursementRepository(disbursementRepo),
		)
		xbPayoutSvc := xbPayoutService.New(pdkLog, disbursementRepo, beneficiaryAccountRepo,
			xbCoreProcessorRepo,
			xbPayoutService.WithFeeService(feeSvc),
			xbPayoutService.WithOrchestratorService(orchestratorSvc),
			xbPayoutService.WithRabbitMQClient(rmqExt),
			xbPayoutService.WithConfig(config),
			xbPayoutService.WithStatusHistories(statusHistoriesRepo),
		)
		settlementSvc := settlementService.New(pdkLog, accounttransactionRepo,
			settlementService.WithPaymentSvc(paymentSvc),
			settlementService.WithMerchantSvc(merchantSvc),
		)
		withdrawalService := withdrawalService.New(
			pdkLog, withdrawalRepo, orchestratorSvc, bankAccountRepo, nil,
			withdrawalService.WithRedisClient(cacheClient),
			withdrawalService.WithSnapCoreRepository(snapCoreRepository),
			withdrawalService.WithAccountRepository(accountRepo),
			withdrawalService.WithWithdrawalConfig(&config.WithdrawalConfig),
			withdrawalService.WithBankTransferConfig(disbursementSvc),
			withdrawalService.WithRabbitMQClient(rmqExt),
			withdrawalService.WithMerchantRepository(merchantRepo),
			withdrawalService.WithNotificationService(notificationSvc),
		)
		reconSvc := reconService.New(
			config,
			pdkLog,
			reconRepo,
			reconService.WithGCSService(gcs),
			reconService.WithExcelService(xlsx.New()),
			reconService.WithRabbitMQClient(rmqExt),
			reconService.WithAccountTransactionRepository(accounttransactionRepo),
			reconService.WithSnapCoreRepository(snapCoreRepository),
			reconService.WithCreditCardCoreProcessorRepository(creditcardCoreProcessorRepo),
		)

		fdsProcessors := map[string]repository.IFdsProcessorRepository{
			constant.PROVIDER_FRAUD_NET: fraudNetRepo,
			constant.PROVIDER_SOKRATECH: sokratechRepo.NewFDSProcessor(),
		}

		fdsProcessorService := fdsservice.New(
			config,
			pdkLog,
			fraudRulesRepo,
			ruleEvaluationsRepo,
			accounttransactionRepo,
			paymentRepository,
			merchantRepo,
			fdsProcessors,
			fdsservice.WithCustomerRepository(customerRepo),
			fdsservice.WithRabbitMqExt(rmqExt),
		)
		recurringContractSvc := recurringContractService.New(pdkLog, recurringContractRepo, customerService)

		unifiedPaymentSvc := unifiedPaymentService.New(config, pdkLog, paymentRepository, paymentMethodRepository, accounttransactionRepo,
			unifiedPaymentService.WithMerchantRepo(merchantRepo),
			unifiedPaymentService.WithSnapCoreRepo(snapCoreRepository),
			unifiedPaymentService.WithJWTExt(jwtConfig),
			unifiedPaymentService.WithRabbitMQClient(rmqExt),
			unifiedPaymentService.WithRedisClient(cacheClient),
			unifiedPaymentService.WithFeeService(feeSvc),
			unifiedPaymentService.WithOrchestratorService(orchestratorSvc),
			unifiedPaymentService.WithQRISService(qrisService),
			unifiedPaymentService.WithPaymentService(paymentSvc),
			unifiedPaymentService.WithCustomerRepo(customerRepo),
			unifiedPaymentService.WithFdsService(fdsProcessorService),
			unifiedPaymentService.WithStatusHistoriesRepository(statusHistoriesRepo),
			unifiedPaymentService.WithPaymentCaptureRepository(paymentCaptureRepo),
			unifiedPaymentService.WithCreditCardCoreProcessorRepo(creditcardCoreProcessorRepo),
			unifiedPaymentService.WithRecurringContractService(recurringContractSvc),
			unifiedPaymentService.WithPaymentMethodService(paymentMethodService),
			unifiedPaymentService.WithCreditCardService(creditcardSvc),
			unifiedPaymentService.WithCryptoProvider(encryption.NewCryptoProvider()),
			unifiedPaymentService.WithCache(cacheClient),
		)
		paymentService.WithUnifiedPaymentService(paymentSvc, unifiedPaymentSvc)
		unifiedPaymentService.WithMerchantService(unifiedPaymentSvc, merchantSvc)

		refundSvc := refundService.New(config, pdkLog, refundRepo, paymentRepository, accounttransactionRepo, merchantRepo, snapCoreRepository, callbackRepo,
			refundService.WithOrchestratorService(orchestratorSvc),
			refundService.WithRabbitMQClient(rmqExt),
			refundService.WithFeeService(feeSvc),
			refundService.WithRedisClient(cacheClient),
			refundService.WithPaymentMethodRepository(paymentMethodRepository),
			refundService.WithStatusHistoriesRepository(statusHistoriesRepo),
		)
		refundProcessorSvc := refundProcessorService.New(pdkLog, refundRepo, snapCoreRepository, creditcardCoreProcessorRepo, beneficiaryAccountSvc, orchestratorSvc, cacheClient,
			refundProcessorService.WithRefundService(refundSvc),
			refundProcessorService.WithFeeService(feeSvc),
			refundProcessorService.WithTransferService(transferSvc),
			refundProcessorService.WithSettlementService(settlementSvc),
			refundProcessorService.WithMerchantService(merchantSvc),
		)
		reportingService := reportingService.New(pdkLog, reportingRepository, accountRepo)

		settlementHoldSvc := settlementHoldService.New(pdkLog, settlementHoldRepo, paymentSvc, settlementSvc, accounttransactionRepo)
		paymentService.WithSettlementHoldService(paymentSvc, settlementHoldSvc)

		cardFundedPayoutSvc := cardFundedPayoutService.New(config, pdkLog,
			cardFundedPayoutService.WithFeeService(feeSvc),
			cardFundedPayoutService.WithVendorService(vendorSvc),
			cardFundedPayoutService.WithCustomerService(customerService),
			cardFundedPayoutService.WithUnifiedPaymentService(unifiedPaymentSvc),
			cardFundedPayoutService.WithDisbursementRepository(disbursementRepo),
			cardFundedPayoutService.WithStatusHistoriesRepository(statusHistoriesRepo),
			cardFundedPayoutService.WithCacheClient(cacheClient),
			cardFundedPayoutService.WithPaymentRepository(paymentRepository),
			cardFundedPayoutService.WithCreditCardService(creditcardSvc),
			cardFundedPayoutService.WithCryptoProvider(encryption.NewCryptoProvider()),
			cardFundedPayoutService.WithAccountTransactionRepository(accounttransactionRepo),
			cardFundedPayoutService.WithOrchestratorService(orchestratorSvc),
			cardFundedPayoutService.WithSnapCoreRepository(snapCoreRepository),
		)
		settlementService.WithCardFundedPayoutSvc(settlementSvc, cardFundedPayoutSvc)
		unifiedPaymentService.WithCardFundedPayoutService(unifiedPaymentSvc, cardFundedPayoutSvc)

		// Init consumer
		// e.g. consumers job, etc.
		callbackConsumer := callbackConsumerController.New(pdkLog, callbackSvc, rmqExt)
		activityConsumer := activityConsumerController.New(pdkLog, activitySvc)
		paymentConsumer := paymentConsumerController.New(config, pdkLog, paymentSvc, merchantTopUpSvc, orchestratorSvc, rmqExt, merchantSvc, unifiedPaymentSvc)
		disbursementConsumer := disbursementConsumerController.New(pdkLog, disbursementSvc, refundProcessorSvc)
		slackConsumer := slackConsumerController.New(slack, pdkLog)
		creditcardConsumer := creditcardConsumerController.New(pdkLog, creditcardSvc, orchestratorSvc, paymentSvc, unifiedPaymentSvc, refundSvc)
		accountConsumer := accountConsumerController.New(pdkLog, accountSvc)
		qrisController := qrisController.New(pdkLog, qrisService)
		xbPayoutConsumer := xbPayoutConsumerController.New(pdkLog, xbPayoutSvc)
		settlementConsumer := settlementConsumerController.New(pdkLog, settlementSvc)
		commServiceConsumer := commServiceController.New(pdkLog, commService.New(paperCommRepo))
		withdrawalConsumer := withdrawalConsumer.New(pdkLog, withdrawalService)
		reconConsumer := reconConsumer.New(pdkLog, reconSvc)
		notificationConsumer := notificationConsumer.New(pdkLog, rmqExt)
		refundConsumer := refundConsumer.New(pdkLog, refundSvc, refundProcessorSvc, paymentSvc, orchestratorSvc)
		paymentCaptureConsumer := paymentCaptureConsumer.New(pdkLog, unifiedPaymentSvc)
		bankTransferConsumer := bankTransferConsumer.New(pdkLog, &bankTransferConsumer.Service{
			LedgerSvc:           orchestratorSvc,
			DisbursementSvc:     disbursementSvc,
			RefundProcSvc:       refundProcessorSvc,
			WithdrawalSvc:       withdrawalService,
			CardFundedPayoutSvc: cardFundedPayoutSvc,
			RedisExt:            cacheClient,
		})
		merchantConsumer := merchantConsumerController.New(config, pdkLog, merchantSvc)
		callbackWorker := callbackWorker.New(pdkLog, callbackSvc)
		reportingConsumer := reportingConsumer.New(pdkLog, reportingService)

		wg := new(sync.WaitGroup)
		signal, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		defer stop()

		// Consumer config
		directExchangeType := amqp.ExchangeDirect

		// Listening Topic: snap.va.payment
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.SnapVAPaymentRoutingKey, nil,
				paymentConsumer.ProcessPaymentNotification)
			if err != nil {
				fmt.Printf("Unable to consume snap va payment: %v", err)
				panic(err)
			}
		}()

		// Listening Topic: snap.qris.payment
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.SnapQrisPaymentRoutingKey, nil,
				paymentConsumer.ProcessPaymentNotification)
			if err != nil {
				fmt.Printf("Unable to consume snap qris payment: %v", err)
				panic(err)
			}
		}()

		// Listening Topic: snap.ewallet.payment
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.SnapEwalletPaymentRoutingKey, nil,
				paymentConsumer.ProcessPaymentNotification)
			if err != nil {
				fmt.Printf("Unable to consume snap ewallet payment: %v", err)
				panic(err)
			}
		}()

		// Listening Topic: backend-portal.callback.process
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.CallbackRoutingKey, nil,
				callbackConsumer.ProcessCallback)
			if err != nil {
				fmt.Printf("Unable to consume callback: %v", err)
				panic(err)
			}
		}()

		// Listening Topic: backend-portal.activity.insert
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.ActivityInsertRoutingKey, &directExchangeType,
				activityConsumer.Insert)
			if err != nil {
				fmt.Printf("Unable to consume activity: %v", err)
				panic(err)
			}
		}()

		// Listening Topic: backend-portal.bulk-disbursement.batch-create
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.BulkDisbursementBatchCreateRoutingKey, nil,
				disbursementConsumer.BatchCreateDisbursement)
			if err != nil {
				fmt.Printf("Unable to consume backend-portal.bulk-disbursement.batch-create: %v", err)
				panic(err)
			}
		}()

		// Listening Topic: backend-portal.bulk-disbursement.batch-process
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.BulkDisbursementBatchProcessRoutingKey, nil,
				disbursementConsumer.BatchProcessDisbursement)
			if err != nil {
				fmt.Printf("Unable to consume backend-portal.bulk-disbursement.batch-process: %v", err)
				panic(err)
			}
		}()

		// Listening Topic: backend-portal.slack.post-webhook
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.SlackPostWebhookRoutingKey, nil,
				slackConsumer.ProcessSlackPostWebhook)
			if err != nil {
				fmt.Printf("Unable to consume slack post webhook: %v", err)
			}
		}()

		// Listening Topic: creditcard.payment.notification
		go func() {
			err = rmqExt.Consume(signal, rabbitMqExt.CreditcardPaymentNotificationRoutingKey, nil,
				creditcardConsumer.PaymentNotification)
			if err != nil {
				fmt.Printf("Unable to consume creditcard payment notification: %v", err)
			}
		}()

		// Listening Topic: backend-portal.account.bulk-create
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.BulkCreateAccountRoutingKey, nil,
				accountConsumer.BulkCreateAccount)
			if err != nil {
				fmt.Printf("Unable to consume bulk create account: %v", err)
			}
		}()

		// Consume Queue (Topic): q.snap.qris.registration-callback
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.QrisRegistrationCallbackRoutingKey, nil, qrisController.Process)
			if err != nil {
				fmt.Printf("Unable to consume qris registration callback: %v", err)
			}
		}()

		// Consume Queue (Topic): q.xb.payout.status-change
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.XbPayoutStatusChangeRoutingKey, nil, xbPayoutConsumer.UpdateStatus)
			if err != nil {
				fmt.Printf("Unable to consume xb payout status change: %v", err)
			}
		}()

		// Consume Queue (Topic): q.backend-portal.scheduling.settlement.process
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.SettlementProcessingRoutingKey, &directExchangeType, settlementConsumer.ProcessPaymentSettlement)
			if err != nil {
				fmt.Printf("Unable to consume settlement processing: %v", err)
			}
		}()

		// Consume Queue (Topic): backend-portal.payout.alert.processing
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.PayoutAlertProcessingRoutingKey, &directExchangeType, disbursementConsumer.PayoutTransactionAlertProcess)
			if err != nil {
				fmt.Printf("Unable to consume payout alert processing: %v", err)
			}
		}()

		// Consume Queue : q.backend-portal.comm-service.email
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(
				signal, rabbitMqExt.CommServiceEmailRoutingKey, nil, commServiceConsumer.PostEmailHandler,
			)
			if err != nil {
				fmt.Printf("Unable to consume post email processing: %v", err)
			}
		}()

		// Listening Topic: snap-core.transfer.status
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.SnapTransferStatusRoutingKey, nil, bankTransferConsumer.UpdateTransferStatus)
			if err != nil {
				fmt.Printf("Unable to consume snap-core.transfer.status: %v", err)
				panic(err)
			}
		}()

		// Listening Topic: snap-core.transfer.cutoff-report
		wg.Add(1)
		go func() {
			defer wg.Done()

			delayedExchangeType := rabbitMqExt.DelayedExchangeType
			err = rmqExt.Consume(signal, rabbitMqExt.SnapTransferCutOffReportRoutingKey, &delayedExchangeType, bankTransferConsumer.CutOffReportTrigger)
			if err != nil {
				fmt.Printf("Unable to consume snap-core.transfer.cutoff-report: %v", err)
				panic(err)
			}
		}()

		// Listening Queue: q.backend-portal.recon.process
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.ReconProcessRoutingKey, nil,
				reconConsumer.ReconciliationProcess)
			if err != nil {
				fmt.Printf("Unable to consume snap-core.transfer.status: %v", err)
				panic(err)
			}
		}()

		if config.WithdrawalConfig.ActiveConsumer <= 0 {
			config.WithdrawalConfig.ActiveConsumer = 1
		}
		for range config.WithdrawalConfig.ActiveConsumer {
			// Listening Queue (Topic): q.backend-portal.withdrawal.process
			wg.Add(1)
			go func() {
				defer wg.Done()

				err = rmqExt.Consume(
					signal, rabbitMqExt.WithdrawalProcessRoutingKey, nil, withdrawalConsumer.WithdrawalProcess,
				)
				if err != nil {
					fmt.Printf("Unable to consume q.backend-portal.withdrawal.process: %v", err)
					panic(err)
				}
			}()
		}

		if config.VccTerminal.ConsumerCount <= 0 {
			config.VccTerminal.ConsumerCount = 1
		}
		for range config.VccTerminal.ConsumerCount {
			// Listening Queue (Topic): q.backend-portal.vcc-terminal.charges
			wg.Add(1)
			go func() {
				defer wg.Done()

				err = rmqExt.Consume(signal, rabbitMqExt.VccTerminalChargeRoutingKey, nil, paymentConsumer.VCCTerminalSubmitCharge)
				if err != nil {
					fmt.Printf("Unable to consume q.backend-portal.vcc-terminal.charges: %v", err)
					panic(err)
				}
			}()
		}

		// Listening Queue: q.backend-portal.payment.expiration.process
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.PaymentExpirationRoutingKey, &directExchangeType,
				paymentConsumer.ProcessPaymentExpiration)
			if err != nil {
				fmt.Printf("Unable to consume q.backend-portal.payment.expiration.process: %v", err)
				panic(err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			config := &rabbitMqExt.ConsumeOptsConfig{
				QueueName: rabbitMqExt.NotificationDLQueueName,
			}
			err := rmqExt.ConsumeWithOpts(signal, config, notificationConsumer.RetryNotification, rabbitMqExt.ConsumerSetupNotification)
			if err != nil {
				fmt.Printf("Unable to consume %s: %v", config.QueueName, err)
				panic(err)
			}
		}()

		// Listening Queue: q.backend-portal.sub-merchants.bulk-create
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.SubMerchantBulkCreateRoutingKey, nil,
				merchantConsumer.ProcessBulkCreateSubMerchant)
			if err != nil {
				fmt.Printf("Unable to consume %s: %v", rabbitMqExt.SubMerchantBulkCreateRoutingKey, err)
				panic(err)
			}
		}()

		// Listening Topic: backend-portal.refund.process
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.RefundProcessRoutingKey, nil,
				refundConsumer.RefundProcess)
			if err != nil {
				fmt.Printf("Unable to consume refund process: %v", err)
			}
		}()

		// Listening Topic: backend-portal.payment-capture.process
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.PaymentCaptureProcessRoutingKey, nil, paymentCaptureConsumer.PaymentCaptureProcess)
			if err != nil {
				fmt.Printf("Unable to consume payment capture process: %v", err)
			}
		}()

		// Listening Topic: snap-core.transfer.reconcile
		wg.Add(1)
		go func() {
			defer wg.Done()

			err = rmqExt.Consume(signal, rabbitMqExt.SnapTransferReconcileRoutingKey, nil,
				reconConsumer.SnapCoreTransferReconcile)
			if err != nil {
				fmt.Printf("Unable to consume snap-core.transfer.reconcile: %v", err)
				panic(err)
			}
		}()

		logger.Info(ctx, "rabbitmq listening")

		// Currently, the worker responsible for executing tasks runs alongside the consumer, as both share similar
		// behavior — the consumer listens for messages from the message broker, while the task runner (worker) retrieves messages for processing.
		// This approach is also chosen to speed up the development process. When scalability becomes a concern, separate instances (pods) will be created.
		taskRunner := conductorClient.Workers()

		// Run the task runner on a separate thread because the process is blocking.
		go func() {
			// Register Worker
			workerTaskList := []conductor.WorkerDefinition{
				{
					TaskName:     constant.WorkflowTaskMerchantCallbackPreparation,
					Handler:      callbackWorker.Preparation,
					BatchSize:    config.MerchantCallbackTask.Preparation.BatchSize,
					PollInterval: config.MerchantCallbackTask.Preparation.PollingInterval,
					PollTimeout:  config.MerchantCallbackTask.Preparation.PollingTimeout,
				},
				{
					TaskName:     constant.WorkflowTaskMerchantCallbackDelivery,
					Handler:      callbackWorker.SendCallback,
					BatchSize:    config.MerchantCallbackTask.SendCallback.BatchSize,
					PollInterval: config.MerchantCallbackTask.SendCallback.PollingInterval,
					PollTimeout:  config.MerchantCallbackTask.SendCallback.PollingTimeout,
				},
				{
					TaskName:     constant.WorkflowTaskMerchantCallbackDeliveryWithoutRetry,
					Handler:      callbackWorker.SendCallback,
					BatchSize:    config.MerchantCallbackTask.SendCallback.BatchSize,
					PollInterval: config.MerchantCallbackTask.SendCallback.PollingInterval,
					PollTimeout:  config.MerchantCallbackTask.SendCallback.PollingTimeout,
				},
				{
					TaskName:     constant.WorkflowTaskMerchantCallbackWriteLog,
					Handler:      callbackWorker.WriteCallbackLog,
					BatchSize:    config.MerchantCallbackTask.WriteCallbackLog.BatchSize,
					PollInterval: config.MerchantCallbackTask.WriteCallbackLog.PollingInterval,
					PollTimeout:  config.MerchantCallbackTask.WriteCallbackLog.PollingTimeout,
				},
				{
					TaskName:     constant.WorkflowTaskMerchantCallbackWriteMetric,
					Handler:      callbackWorker.WriteCallbackMetric,
					BatchSize:    config.MerchantCallbackTask.WriteCallbackMetric.BatchSize,
					PollInterval: config.MerchantCallbackTask.WriteCallbackMetric.PollingInterval,
					PollTimeout:  config.MerchantCallbackTask.WriteCallbackMetric.PollingTimeout,
				},
			}
			// Run Task Runner (blocking)
			if runError := taskRunner.RunWorkers(context.Background(), workerTaskList); runError != nil {
				log.Println("Running Task Failed:", runError.Error())
			}
		}()

		// Reporting Service
		// Report: Balance History
		if conf := config.ReportingConsumers.ReportBalanceHistory; conf.Enabled {
			wg.Add(1)
			go func() {
				defer wg.Done()

				config := stream.ReadMessageConfig{
					StreamQueueName: conf.StreamQueueName,
					ConsumerName:    conf.ConsumerName,
					RetryCount:      conf.RetryCount,
					RetryDelay:      conf.RetryDelay,
					CommitSize:      conf.CommitSize,
					CommitInterval:  conf.CommitInterval,
					Handler:         reportingConsumer.BalanceHistory,
					ReconnectDelay:  conf.ReconnectDelay,
				}
				if err := rabbitMqStream.ReadMessage(signal, config); err != nil && !errors.Is(err, context.Canceled) {
					log.Println("Running Report Balance History Failed:", err.Error())
				}
			}()
		}

		// Service Monitoring
		r := chi.NewRouter()
		r.Mount("/debug", middleware.Profiler())

		healthCheckTargets := []chiExtMiddleware.HealthCheckTarget{
			{Name: "rabbitMQAvailable", Health: rmqExt.HealthCheck},
			{Name: "redisAvailable", Health: func(ctx context.Context) error { return cacheClient.Ping(ctx).Err() }},
			{Name: "mysqlAvailable", Health: func(context.Context) error { return dbClient.Ping() }},
			{Name: "serviceAvailable", Health: func(context.Context) error { return nil }},
		}
		r.Get("/health-check", chiExtMiddleware.HealthCheckHandler(pdkLog, healthCheckTargets...))

		port := os.Getenv("PORT")
		if port == "" {
			port = "3001"
		}

		server := &http.Server{
			Addr: ":" + port, Handler: r,
		}
		go func() {
			if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalln("Failed listen and server http server:", err)
			}
		}()
		logger.Info(ctx, "Monitoring service is running at :3001")

		<-signal.Done()

		stop()
		logger.Info(ctx, "Receive a signal to shutdown the service")

		// Gracefully Shutdown Task Runner
		func() {
			shutdownCtx, stopShutdown := context.WithTimeout(context.Background(), 30*time.Second)
			defer stopShutdown()

			if err := taskRunner.Close(shutdownCtx); err != nil {
				log.Println("Shutdown Task Runner Failed:", err.Error())
			}
		}()

		// Gracefully Shutdown Consumer
		close := make(chan struct{}, 1)
		timeout := time.After(time.Duration(config.GracefulWaitTime) * time.Second)
		go func() {
			wg.Wait()

			close <- struct{}{}
		}()
		select {
		case <-timeout:
			logger.Info(signal, "Consumer termination timeout exceeded")

		case <-close:
			logger.Info(signal, "Consumer successfully shutdown")
		}

		ctx, cancel := context.WithTimeout(ctx, time.Duration(config.GracefulWaitTime)*time.Second)
		defer cancel()

		// Gracefully Shutdown HTTP Server
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalln("Failed shutting down http server:", err)
		}
	},
}
