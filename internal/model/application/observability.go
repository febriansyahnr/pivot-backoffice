package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pdkNewRelic "github.com/paper-indonesia/pdk/v2/newRelicExt"
	"github.com/paper-indonesia/pdk/v2/otelExt"
)

func (a *Application) setupObservability() {
	var err error
	// Observability
	otelOpts := []otelExt.OptionFunc{}
	if a.cfg.OTLPConfig.Insecure {
		otelOpts = append(otelOpts, otelExt.WithInsecure())
	}
	if a.cfg.OTLPConfig.TLSClientConfig != nil {
		otelOpts = append(otelOpts, otelExt.WithTLSClientConfig(&tls.Config{
			InsecureSkipVerify: a.cfg.OTLPConfig.TLSClientConfig.InsecureSkipVerify,
		}))
	}
	// Init Open Telemetry
	a.otel, err = otelExt.New(
		otelExt.Config{
			ServiceName:  getServiceName(a.cfg.ServiceName),
			Environment:  a.cfg.Environment,
			OTLPEndpoint: a.cfg.OTLPConfig.Host,
			LicenseKey:   a.secret.NewRelicLicenseKey,
			MetricConfig: otelExt.MetricConfig{
				MetricInterval: time.Duration(a.cfg.OTLPConfig.MetricConfig.Interval) * time.Second,
				MetricTimeout:  time.Duration(a.cfg.OTLPConfig.MetricConfig.ExportTimeout) * time.Second,
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
	a.AddCloser(func() {
		if shutdownErr := a.otel.Shutdown(context.Background()); shutdownErr != nil {
			fmt.Printf("Error shutting down otel: %v\n", shutdownErr)
		}
	})
	customMetric.SetOtelExt(a.otel)

	// Init New Relic
	a.nr, err = pdkNewRelic.New(
		pdkNewRelic.Config{
			ServiceName: getServiceName(a.cfg.ServiceName) + "-" + a.cfg.Environment,
			Environment: a.cfg.Environment,
			LicenseKey:  a.secret.NewRelicLicenseKey,
		},
		pdkNewRelic.WithExcludeAttributes(strings.Split(a.cfg.AppConfig.MaskingSensitiveData, ",")),
	)
	if err != nil {
		fmt.Printf("Unable to init new relic, %v", err)
		panic(err)
	}
}

func (a *Application) GetNewRelic() pdkNewRelic.INewRelicExt {
	return a.nr
}
