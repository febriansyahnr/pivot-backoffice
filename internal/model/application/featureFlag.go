package application

import (
	"context"
	"fmt"
	"time"

	pdkGoff "github.com/paper-indonesia/pdk/v2/goff"
	pdkNotifier "github.com/paper-indonesia/pdk/v2/goff/notifier"
	pdkRetriever "github.com/paper-indonesia/pdk/v2/goff/retriever"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func (a *Application) setupFeatureFlag() {

	// Init Feature Flag
	consulRetriever, err := pdkRetriever.NewConsulRetriever(
		a.cfg.FeatureFlagConfig.ConsulAddr,
		a.cfg.FeatureFlagConfig.ConsulConfigPath,
		a.secret.ConsulSecret.Token,
	)
	if err != nil {
		fmt.Printf("Unable to init goff - consul retriever, %v", err)
		panic(err)
	}
	logNotifier := pdkNotifier.NewLoggerNotifier(a.pdkLog)
	ffconfig, err := pdkGoff.NewGoff(pdkGoff.Config{
		PollingInterval:             time.Duration(a.cfg.FeatureFlagConfig.PollingInterval) * time.Second,
		EnablePollingJitter:         false,
		Logger:                      a.pdkLog,
		Context:                     context.Background(),
		Environment:                 a.cfg.Environment,
		Retriever:                   consulRetriever,
		Notifiers:                   []notifier.Notifier{logNotifier},
		FileFormat:                  pdkGoff.FileFormatYAML,
		Offline:                     a.cfg.FeatureFlagConfig.Offline,
		EvaluationContextEnrichment: nil,
		DataExporter: &logsexporter.Exporter{
			LogFormat: `goffExporter: kind={{ .Kind}}, contextKind={{ .ContextKind}}, user={{ .UserKey}}, key={{ .Key}}, variation={{ .Variation}}, value={{ .Value}}, default={{ .Default}}`,
		},
		NotifierSlackWebhookURL: a.cfg.FeatureFlagConfig.ExporterSlackWebhookURL,
	})
	if err != nil {
		fmt.Printf("Unable to init feature flag config, %v", err)
		panic(err)
	}
	if err := ffclient.Init(ffconfig); err != nil {
		fmt.Printf("Unable to init feature flag client, %v", err)
		panic(err)
	}
	a.AddCloser(func() { ffclient.Close() })
}
