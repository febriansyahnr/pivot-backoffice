package console

import "context"

type IMerchantCommand interface {
	MigrateMerchantSecretsToEncryption(ctx context.Context) error
	MigrateBalanceHistoryToDataReporting(ctx context.Context, startDate, endDate string) error
}
