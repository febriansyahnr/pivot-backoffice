package cron

import "github.com/paper-indonesia/pivot-backoffice/port/cron/handler"

type cron struct {
	BalanceHandler           handler.IOrchestratorBalanceHandler
	RoleAndPermissionHandler handler.IRoleAndPermissionHandler
	FeeHandler               handler.IFeeHandler
	MerchantHandler          handler.IMerchantHandler
	TablePartitionHandler    handler.ITablePartitionHandler
	WithdrawalHandler        handler.IWithdrawalHandler
	PaymentHandler           handler.IPaymentHandler
	PayoutHandler            handler.IPayoutHandler
}

func NewCron(
	balanceHandler handler.IOrchestratorBalanceHandler,
	roleAndPermissionHandler handler.IRoleAndPermissionHandler,
	feeHandler handler.IFeeHandler,
	merchantHandler handler.IMerchantHandler,
	tablePartitionHandler handler.ITablePartitionHandler,
	withdrawalHandler handler.IWithdrawalHandler,
	paymentHandler handler.IPaymentHandler,
	payoutHandler handler.IPayoutHandler,
) *cron {
	return &cron{
		BalanceHandler:           balanceHandler,
		RoleAndPermissionHandler: roleAndPermissionHandler,
		FeeHandler:               feeHandler,
		MerchantHandler:          merchantHandler,
		TablePartitionHandler:    tablePartitionHandler,
		WithdrawalHandler:        withdrawalHandler,
		PaymentHandler:           paymentHandler,
		PayoutHandler:            payoutHandler,
	}
}
