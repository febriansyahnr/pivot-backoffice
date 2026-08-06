package test

import (
	"context"
	"fmt"
	"time"

	pdkGoff "github.com/paper-indonesia/pdk/v2/goff"
	pdkNotifier "github.com/paper-indonesia/pdk/v2/goff/notifier"
	pdkRetriever "github.com/paper-indonesia/pdk/v2/goff/retriever"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func SetupGoff(ctx context.Context, consulAddr string, logger pdkLogger.ILogger) error {
	consulRetriever, err := pdkRetriever.NewConsulRetriever(consulAddr, "backend-portal/feature-flag", "")
	if err != nil {
		fmt.Printf("Unable to init goff - consul retriever, %v", err)
		panic(err)
	}
	logNotifier := pdkNotifier.NewLoggerNotifier(logger)

	ffconfig, err := pdkGoff.NewGoff(pdkGoff.Config{
		PollingInterval:             10 * time.Second,
		EnablePollingJitter:         false,
		Logger:                      logger,
		Context:                     ctx,
		Environment:                 "test",
		Retriever:                   consulRetriever,
		Notifiers:                   []notifier.Notifier{logNotifier},
		FileFormat:                  pdkGoff.FileFormatYAML,
		Offline:                     false,
		EvaluationContextEnrichment: nil,
		DataExporter: &logsexporter.Exporter{
			LogFormat: `goffExporter: kind={{ .Kind}}, contextKind={{ .ContextKind}}, user={{ .UserKey}}, key={{ .Key}}, variation={{ .Variation}}, value={{ .Value}}, default={{ .Default}}`,
		},
		NotifierSlackWebhookURL: "",
	})
	if err != nil {
		return err
	}

	return ffclient.Init(ffconfig)
}
