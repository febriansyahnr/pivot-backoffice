package disbursementService

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) checkMerchantWhitelist(ctx context.Context, merchantId string) bool {
	trxMerchantWhitelist := ffcontext.NewEvaluationContext(merchantId)
	trxMerchantWhitelist.AddCustomAttribute("environment", s.config.Environment)
	trxMerchantWhitelist.AddCustomAttribute("merchant_id", merchantId)

	if merchantWhitelist, err := ffclient.BoolVariation("trx-merchant-whitelist", trxMerchantWhitelist, false); err != nil {
		s.logger.Error(ctx, "getting merchant whitelist", logger.Error(err))
		return false
	} else {
		return merchantWhitelist
	}
}

func (s *DisbursementService) randomTrxWhitelist() bool {
	randomPercentageFlag := ffcontext.NewEvaluationContext(s.config.Environment)
	randomPercentageFlag.AddCustomAttribute("environment", s.config.Environment)

	var (
		randomPercentage int
		err              error
	)
	randomPercentage, err = ffclient.IntVariation("trx-random-whitelist", randomPercentageFlag, 10)
	if err != nil {
		randomPercentage = 10
	}

	// Generate a random number between 0 and 99 using crypto/rand
	randomBig, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		// Fall back to default behavior if random number generation fails
		return false
	}
	randomNumber := int(randomBig.Int64())

	return randomNumber < randomPercentage
}

func (s *DisbursementService) checkAndSetTrxMerchantWhitelist(ctx context.Context, merchantId, transactionId string) error {
	isMerchantWhitelist := s.checkMerchantWhitelist(ctx, merchantId)
	isTrxWhitelist := s.randomTrxWhitelist()

	// if s.checkMerchantWhitelist(ctx, merchantId) && s.randomTrxWhitelist() {
	if isMerchantWhitelist && isTrxWhitelist {
		err := s.redisExt.Set(ctx, fmt.Sprintf("backend-portal:trx-merchant-whitelist:%s", transactionId), true, time.Hour).Err()
		if err != nil {
			s.logger.Error(ctx, "setting trx merchant whitelist", logger.Error(err))
			return err
		}
		s.logger.Info(
			ctx,
			"trx merchant whitelisted",
			logger.String("merchant_id", merchantId),
			logger.String("transaction_id", transactionId),
			logger.Bool("isMerchantWhitelist", isMerchantWhitelist),
			logger.Bool("isTrxWhitelist", isTrxWhitelist))
		return nil
	}
	s.logger.Info(
		ctx,
		"trx merchant not whitelisted",
		logger.String("merchant_id", merchantId),
		logger.String("transaction_id", transactionId),
		logger.Bool("isMerchantWhitelist", isMerchantWhitelist),
		logger.Bool("isTrxWhitelist", isTrxWhitelist))
	return nil
}

func (s *DisbursementService) BatchCreateDisbursement(ctx context.Context, request *disbursementModel.BatchCreateDisbursementRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/BatchCreateDisbursement")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error(ctx, "Panic recovery from BatchCreateDisbursement", logger.Error(fmt.Errorf("%v", r)))
			// TODO: send to slack
		}
	}()

	var (
		now    = time.Now()
		amount decimal.Decimal
	)

	// TODO: Need to highly monitor this
	defer monitor.WriteAndSend(
		ctx, "disbursement-batch-create", now, nil, err, func() []string {
			return []string{
				fmt.Sprintf("merchant_id:%s", request.MerchantID),
				fmt.Sprintf("bulk_id:%s", request.BulkID),
				fmt.Sprintf("total:%d", len(request.Data)),
				"proc_identifier:batch-create",
			}
		},
	)

	if s.batchCreateWP == nil {
		s.newBatchCreateWP()
	}

	var (
		wg    sync.WaitGroup
		start = time.Now().UTC()
	)

	// Get Merchant By merchantID
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return err

	} else if merchant == nil {
		return constant.ErrMerchantNotFound
	}
	ctx = context.WithValue(ctx, constant.CtxMerchantData, merchant)

	// delete redis data for marking bulk disbursement as in progress
	if err := s.redisExt.Del(ctx, fmt.Sprintf(constant.BulkDisbursementInProgressQueueLockFmt, request.MerchantID, request.BulkID)).Err(); err != nil {
		s.logger.Error(ctx, "deleting redis data for marking bulk disbursement as in progress", logger.Error(err))
	}

	// TODO: Need to highly monitor this, because there will be so much span in go routine
	for _, createRequest := range request.Data {
		wg.Add(1)

		createRequest.BulkID = &request.BulkID
		createRequest.MerchantID = request.MerchantID
		createRequest.MerchantName = request.MerchantName
		createRequest.CreatedBy = &request.CreatedBy
		createRequest.CreatedFrom = request.CreatedFrom

		amount = amount.Add(createRequest.Amount)

		if err := s.checkAndSetTrxMerchantWhitelist(ctx, request.MerchantID, createRequest.ReferenceID); err != nil {
			s.logger.Error(ctx, "failed to check and set trx merchant whitelist", logger.Error(err))
			// Continue processing despite this error as it's not critical to the main flow
		}

		// Invoke create disbursement
		_ = s.batchCreateWP.Invoke(batchCreateWPData{
			ctx:           ctx,
			wg:            &wg,
			bulkID:        request.BulkID,
			createRequest: createRequest,
		})
	}

	wg.Wait()
	s.logger.Info(
		ctx, "Batch create disbursement", logger.Any("details", map[string]any{
			"total":    len(request.Data),
			"duration": time.Now().UTC().Sub(start).String(),
		}),
	)
	// Trigger check length disbursements, then update bulkDisbursement status to WAITING
	countByBulkID := s.disbursementRepo.CountByBulkID(ctx, request.BulkID)
	if countByBulkID != request.TotalTrx {
		return nil
	}

	err = s.disbursementRepo.UpdateBulkDisbursementStatusByID(ctx, request.BulkID, constant.BulkDisbursementStatusWaiting)
	if err != nil {
		return err
	}

	err = s.rabbitMqExt.PushNotification(ctx, &notification.PushNotification{
		RoutingKey: fmt.Sprintf(constant.NotificationRoutingKeyFmt, request.CreatedBy),
		Payload: notification.PushNotificationPayload{
			ID:        uuid.NewString(),
			Subject:   "Upload Successful!",
			Type:      constant.CreateBulkDisbursementNotifType,
			Message:   fmt.Sprintf("Your batch transaction <b>%s</b> has been successfully uploaded.", request.BulkID[len(request.BulkID)-12:]),
			CreatedAt: time.Now().UTC(),
		},
	})
	if err != nil {
		s.logger.Error(ctx, "push notification for batch transaction "+request.BulkID, logger.Error(err))
	}

	if request.AutoApprove {
		disbursements, err := s.disbursementRepo.GetAllDisbursementByBulkID(ctx, request.BulkID)
		if err != nil {
			return err
		}

		var totalAmount = 0.0
		var approveActions []disbursementModel.ApproveActionObject
		for _, disbursement := range disbursements {
			totalAmount += disbursement.Amount.InexactFloat64()
			approveActions = append(approveActions, disbursementModel.ApproveActionObject{DisbursementID: disbursement.UUID})
		}

		requestApprove := &disbursementModel.ApproveRequest{
			ApproveAction: approveActions,
			MerchantID:    request.MerchantID,
			ApprovedBy:    request.CreatedBy,
			BulkID:        request.BulkID,
			CreatedFrom:   request.CreatedFrom,
			TotalAmount:   totalAmount,
		}
		if err = s.self.Approve(ctx, requestApprove); err != nil {
			return err
		}
	}

	return nil
}
