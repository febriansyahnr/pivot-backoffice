package application

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/bigquery"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pdk/v2/gcp"
)

func (a *Application) setupGoogleService(ctx context.Context) {

	// Init GCS
	a.gcsClient = gcs.NewGCSService(gcs.Config{
		ServiceBucketName:          a.cfg.GCSConfig.ServiceBucketName,
		ReportingBucketName:        a.cfg.GCSConfig.ReportingBucketName,
		BulkDisbursementBucketName: a.cfg.GCSConfig.BulkDisbursementBucketName,
		ProofOfTransferFolderName:  a.cfg.GCSConfig.ProofOfTransferFolderName,
	})
	a.AddCloser(func() { a.gcsClient.Close() })

	// Init Secret Manager
	gsmClient, err := gcp.NewSecretManagerClient(ctx)
	if err != nil {
		a.pdkLog.Panic(ctx, "Unable to init google secret manager: "+err.Error())
	}
	gcp.SetGlobalSecretManagerClient(gsmClient)
	a.AddCloser(func() { gsmClient.Close() })

	bqConfig := bigquery.Config{
		ProjectID:           a.cfg.BigQueryConfig.ProjectID,
		Location:            a.cfg.BigQueryConfig.Location,
		QueryTimeoutSeconds: a.cfg.BigQueryConfig.QueryTimeoutSeconds,
		MaxRetries:          a.cfg.BigQueryConfig.MaxRetries,
	}
	a.bqClient = bigquery.NewBigQueryService(bqConfig)
}
