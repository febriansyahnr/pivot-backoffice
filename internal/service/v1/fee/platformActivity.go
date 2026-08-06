package feeService

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *FeeService) PlatformActivitiesFee(ctx context.Context, date time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/PlatformActivitiesFee")
	defer segment.End()

	merchants, err := s.merchantRepo.GetListOfMerchantsWhoHaveSubMerchant(ctx)
	if err != nil {
		return err

	} else if len(merchants) == 0 {
		s.logger.Info(ctx, "List of merchants who have sub merchant is not found")
		return nil
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, tz)
	endOfMonth := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, tz)

	for _, merchant := range merchants {

		_, feeDetail, err := s.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
			MerchantID: merchant.ID,
			Reference:  constant.ReferencePlatformActivity,
		})
		if err != nil {
			return err
		}

		deductionDate := time.Date(date.Year(), date.Month(), int(*feeDetail.DeductionDay), 0, 0, 0, 0, tz)
		if date.Month() != deductionDate.Month() {
			deductionDate = endOfMonth
		}

		if date.Day() != deductionDate.Day() {
			s.logger.Info(ctx,
				"There is no schedule for calculating platform activity fees", logger.String("merchantId", merchant.ID),
			)
			continue
		}

		if feeDetail.DeductionLastDate == nil {
			feeDetail.DeductionLastDate = &merchant.CreatedAt
		}
		startDate := feeDetail.DeductionLastDate.Add(time.Second)
		endDate := date.Add(-time.Second).UTC()
		period := date.Format("2006-01")

		// Check transactions for sub-merchants. If a transaction is found, the transaction activity fee will be charged.
		errTrx := func() (err error) {

			ctxTrx, err := s.feeRepo.BeginTransaction(ctx)
			if err != nil {
				return err
			}

			isComplated := false
			defer func() {
				if isComplated {
					err = s.feeRepo.CommitTransaction(ctxTrx)
					return
				}

				if e := s.feeRepo.RollbackTransaction(ctxTrx); e != nil {
					s.logger.Error(ctx, "Failed to rollback transaction", logger.Error(err))
				}
			}()

			activeSubMerchants, err := s.accountTransactionRepo.GetPlatformTransactionActivities(ctxTrx, merchant.SubMerchants, startDate, endDate)
			if err != nil {
				return err
			}

			if feeDetail.IsDefaultConfig {
				// Save the default fee config to the database and update the last calculation date.
				merchantFee := &merchantModel.MerchantFee{
					UUID:          uuid.NewString(),
					MerchantID:    merchant.ID,
					AmountType:    feeDetail.AmountType,
					Amount:        feeDetail.Amount,
					Reference:     feeDetail.Type,
					DeductionType: feeDetail.DeductionType,
					DeductionDay:  feeDetail.DeductionDay,
					TaxType:       feeDetail.TaxType,
					TaxPercentage: feeDetail.TaxPercentage,
					CreatedAt:     time.Now().UTC(), UpdatedAt: time.Now().UTC(),
				}
				if err = s.merchantRepo.CreateMerchantFee(ctxTrx, merchantFee); err != nil {
					return err
				}
			}

			// Calculation of fee and charge on main merchants
			feeAmount, taxAmount := s.CalculateFee(ctxTrx, &feeModel.GetFeeRequest{}, &feeModel.FeeMetadataObject{
				AmountType:    constant.MerchantFeeAmountType,
				Amount:        float64(len(activeSubMerchants)) * feeDetail.Amount,
				TaxType:       feeDetail.TaxType,
				TaxPercentage: feeDetail.TaxPercentage,
			})
			feeDetail.TaxAmount = taxAmount

			s.logger.Info(ctx,
				"Platform activity fee calculation for merchant id "+merchant.ID,
				logger.String("period", period), logger.Any("details", activeSubMerchants), logger.Float64("feeAmount", feeAmount), logger.Float64("taxAmount", taxAmount),
			)

			// When a sub merchant on the platform has no transaction activity, it will be updated directly to the last calculation date.
			if len(activeSubMerchants) == 0 {

				err = s.merchantRepo.UpdateMerchantFeeLastDeductionDate(ctxTrx, merchant.ID, feeDetail.Type, endDate)
				isComplated = (err == nil)
				return
			}

			trxStatus := constant.StatusPending
			if feeDetail.DeductionType == constant.MerchantFeeDeductionTypeAutomated {
				trxStatus = constant.StatusSuccess
			}

			merchantUUID, _ := uuid.Parse(merchant.ID)
			feeTrxRequest := &orchestrator_model.CreateAccountTransactionRequest{
				UUID:                 util.GenerateUUID(),
				ReferenceID:          "",
				Type:                 orchestrator_model.TypeFee,
				MerchantID:           merchantUUID,
				Currency:             constant.CurrencyIDR,
				Debit:                feeAmount,
				Channel:              "",
				Remarks:              fmt.Sprintf("Platform activity fee period %s", period),
				Status:               trxStatus,
				TransactionTimestamp: time.Now().UTC(),
				Usecase:              constant.TypeDisbursement,
			}
			feeTrxRequest.AdditionalInfo.Valid = true
			feeTrxRequest.AdditionalInfo.JSONText, _ = json.Marshal(feeDetail)

			if err = s.orchestratorSvc.PostAccountTransaction(ctxTrx, feeTrxRequest); err != nil {
				return err
			}

			err = s.merchantRepo.UpdateMerchantFeeLastDeductionDate(ctxTrx, merchant.ID, feeDetail.Type, endDate)
			isComplated = (err == nil)
			return
		}()

		if errTrx != nil {
			return errTrx
		}
	}

	return nil
}
