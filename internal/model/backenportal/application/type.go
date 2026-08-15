package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pdk/go/monitoring"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	pdkNewRelic "github.com/paper-indonesia/pdk/v2/newRelicExt"
	"github.com/paper-indonesia/pdk/v2/otelExt"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/bigquery"
	"github.com/paper-indonesia/pivot-backoffice/pkg/dictionary"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	jwtCore "github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
)

type Application struct {
	cfg               *config.Config
	secret            *config.Secret
	pdkLog            pdkLogger.ILogger
	otel              otelExt.IOtelExt
	nr                pdkNewRelic.INewRelicExt
	monitor           *monitoring.Monitor
	mySqlDB           mySqlExt.IMySqlExt
	redisClient       redisExt.IRedisExt
	rabbitMq          rabbitMqExt.IRabbitMQExt
	gcsClient         gcs.IGCSService
	httpRequestClient httpRequestExt.IHTTPRequest
	jwtConfig         jwtCore.IJwt
	validate          *validatorExt.Validate
	encryptExt        encryption.ICrypto
	encryptGcs        encryption.GCSClient
	vaultClient       *vault.Client
	bqClient          bigquery.IBigQueryService

	repo    AppRepository
	service AppService

	closes []func()
}

func NewApplication(cfg *config.Config, secret *config.Secret, ctx context.Context) *Application {
	var err error
	app := Application{
		cfg:    cfg,
		secret: secret,
		closes: make([]func(), 0),
	}

	isDevelopment := true
	if cfg.Environment == constant.EnvironmentProduction {
		isDevelopment = false
	}
	// Init Logger
	if cfg.AppConfig.PdkLoggerUsed == constant.PdkLoggerSloggerName {
		app.pdkLog = pdkLogger.NewSlogger(
			pdkLogger.Config{
				IsDevelopment: isDevelopment,
				Environment:   cfg.Environment,
				ServiceName:   getServiceName(cfg.ServiceName),
			},
		)
	} else {
		app.pdkLog, err = pdkLogger.NewZapLogger(
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
	app.AddCloser(func() {
		if syncErr := app.pdkLog.Sync(); syncErr != nil {
			fmt.Printf("Error syncing pdk logger: %v\n", syncErr)
		}
	})

	app.setupFeatureFlag()
	app.setupObservability()
	app.setupDatabase()

	// Init RabbitMQ
	cfg.RabbitMQConfig.ServiceName = getServiceName(cfg.ServiceName)
	app.rabbitMq, err = rabbitMqExt.New(
		cfg.RabbitMQConfig,
		secret.RabbitMQSecret,
		app.pdkLog,
		app.nr,
	)
	if err != nil {
		fmt.Printf("Unable to init rabbitmq, %v", err)
		panic(err)
	}
	app.AddCloser(func() { app.rabbitMq.Close() })

	app.setupGoogleService(ctx)

	dictionary.Dict, err = dictionary.New(cfg.DictionaryConfig)
	if err != nil {
		fmt.Printf("Unable to init dictionary, %v", err)
		panic(err)
	}

	// Init HTTP Request Client
	app.httpRequestClient = httpRequestExt.New(
		httpRequestExt.WithLogger(app.pdkLog),
		httpRequestExt.WithMaskingSensitiveData(strings.Split(cfg.AppConfig.MaskingSensitiveData, ",")),
	)

	// Init JWT Config
	if app.jwtConfig, err = jwtCore.New(cfg, secret, app.redisClient); err != nil {
		panic("Unable to init JWT package, " + err.Error())
	}

	// Init Vault Client
	app.vaultClient, err = vault.New(vault.Config{
		Address: cfg.Vault.Address,
		Token:   secret.Vault.Token,
	})
	if err != nil {
		panic("Unable to init Vault client, " + err.Error())
	}

	app.validate = validatorExt.New()

	// Init Encryption
	app.encryptExt = encryption.New()

	// Init Encryption Gcs
	app.encryptGcs = encryption.NewGCS(secret)

	return &app
}

func (a *Application) AddCloser(closer func()) {
	a.closes = append(a.closes, closer)
}

func (a *Application) Recover() {
	if r := recover(); r != nil {
		for _, close := range a.closes {
			close()
		}
		fmt.Printf("Panic occurred: %v\n", r)
		panic(r)
	}
}

func (a *Application) Closer() {
	for _, close := range a.closes {
		close()
	}
}

func (a *Application) Setup() {
	a.SetupRepositories()
	a.SetupServices()
}
