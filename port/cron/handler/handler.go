package handler

import (
	"context"
	"time"

	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/partition"
)

type IOrchestratorBalanceHandler interface {
	CalculateAllMerchantEODBalance(ctx context.Context)
	CalculateAllMerchantDailyAccountTransaction(ctx context.Context, location *time.Location)
}

type IRoleAndPermissionHandler interface {
	SetupPredefinedRoleMenuPermissions(ctx context.Context)
}

type IFeeHandler interface {
	PlatformActivitiesFee(ctx context.Context, dateStr string)
	DeductBalanceForIndirectFeeType(ctx context.Context, dateStr string)
	DetermineFeeTierLvlFromMonthlyTPV(ctx context.Context, dateStr string)
}

type IMerchantHandler interface {
	DormantMerchant(ctx context.Context, dateStr string)
}

type ITablePartitionHandler interface {
	CreateAccountTransactionPartition(ctx context.Context)
	CreateCallbackLogPartition(ctx context.Context)
	CreatePaymentPartition(ctx context.Context)
	CreateDisbursementPartition(ctx context.Context)
	CreateActivityLogPartition(ctx context.Context)
	ReorganizeMonthlyRangePartition(ctx context.Context, request partitionModel.ReorganizeRangePartitionRequest) error
}

type IWithdrawalHandler interface {
	TriggeringAutoWithdrawalProcess(ctx context.Context)
	ForceAutoWithdrawalProcess(ctx context.Context, dateStr string)
}

type IPaymentHandler interface {
	PublishPendingPaymentExpirationEvent(ctx context.Context)
	ProcessInvestigationMonthlyReconciliation(ctx context.Context, dateStr string)
}

type IPayoutHandler interface {
	ReportAfterPayoutCutOffTime(ctx context.Context)
	InquirePendingPayout(ctx context.Context)
}
