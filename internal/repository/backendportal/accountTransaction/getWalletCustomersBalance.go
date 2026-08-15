package accounttransaction_repository

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pdk/v2/logger"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"golang.org/x/sync/errgroup"
)

func (r *AccountTransactionRepository) GetWalletCustomersTotalBalance(ctx context.Context, request *orchestratorModel.GetWalletTotalBalanceRequest) (float64, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/accountTransaction/GetWalletCustomersTotalBalance")
	defer segment.End()

	var (
		totalBalance     float64
		totalCreditDebit float64
		whereClause      = []string{}
		whereArgs        = []interface{}{}
		errG             = new(errgroup.Group)
	)
	totalBalanceQuery := `
		SELECT COALESCE(SUM(a.eod_balance) ,0)
		FROM customers c 
		JOIN accounts a ON c.uuid = a.reference_id AND a.name = 'WALLET'
		WHERE c.merchant_id = ?
	`
	errG.Go(func() error {
		if err := r.db.GetContext(ctx, &totalBalance, totalBalanceQuery, request.MerchantID); err != nil {
			r.logger.Error(ctx, "error when calculate account total balance", logger.Error(err), logger.Any("request", request))
			return err
		}
		return nil
	})

	query := `
		SELECT COALESCE(SUM(at2.credit),0) - COALESCE(SUM(at2.debit),0)
		FROM customers c 
		JOIN accounts a ON c.uuid = a.reference_id AND a.name = 'WALLET'
		JOIN account_transactions at2 ON a.uuid = at2.account_id	
	`

	if request.MerchantID != "" {
		whereClause = append(whereClause, "c.merchant_id = ?")
		whereArgs = append(whereArgs, request.MerchantID)
	}
	if len(request.Status) > 0 {
		statusQuery, statusArgs, _ := sqlx.In("at2.status IN (?)", request.Status)
		whereClause = append(whereClause, statusQuery)
		whereArgs = append(whereArgs, statusArgs...)
	}
	if !request.IncludeIndirectFee {
		whereClause = append(whereClause,
			`(
				at2.type != 'FEE'
				OR (
					at2.additional_info->>'$.deductionType' != 'MANUAL' 
					AND (
						(at2.additional_info->>'$.deductionType' = 'DIRECT' OR at2.additional_info->>'$.deductionType' = '')
						OR ( at2.additional_info->>'$.deductionType' = 'AUTOMATED' AND at2.status = 'SUCCESS' )
					)
				)
			)`,
		)
	}
	whereClause = append(whereClause, "at2.updated_at >= a.last_update_balance_at")
	whereQuery := " WHERE " + strings.Join(whereClause, " AND ")
	query = query + whereQuery

	errG.Go(func() error {
		err := r.db.GetContext(ctx, &totalCreditDebit, query, whereArgs...)
		if err != nil {
			r.logger.Error(ctx, "error when calculate wallet customers total credit debit", logger.Error(err), logger.Any("request", request), logger.String("query", query))
			return err
		}
		return nil
	})

	if err := errG.Wait(); err != nil {
		return 0, err
	}

	return totalBalance + totalCreditDebit, nil
}
