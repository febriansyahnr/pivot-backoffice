package vccsettlement

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cimbProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cimbProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
)

func (s *VccSettlementService) RcnTransactionInquiry(ctx context.Context, request *vccSettlement.VccTransactionInquiryRequest) (*vccSettlement.VccTransactionInquiryResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/vccSettlement/RcnTransactionInquiry")
	defer span.End()

	if err := request.Validate(); err != nil {
		return nil, errPkg.New(response.HttpErrRequest, err)
	}
	request.PartnerReferenceNo = util.GenerateVccRandomPartnerReferenceNo()

	merchantRcn, err := s.merchantRcnSvc.GetRcnDetail(ctx, request.RcnId, request.MerchantId)
	if err != nil {
		s.logger.Error(ctx, "error retrieve merchant rcn detail", logger.Error(err))
		return nil, err
	}
	defer merchantRcn.EraseSensitiveData()

	err = s.rabbitMq.Publish(ctx, rabbitMqExt.VccSettlementInquiryRoutingKey, nil, request)
	if err != nil {
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrPublishSettlementTransactionProcess)
	}

	return &vccSettlement.VccTransactionInquiryResponse{
		PartnerReferenceNo: request.PartnerReferenceNo,
	}, nil
}

func (s *VccSettlementService) ProcessRcnTransactionInquiry(ctx context.Context, request *vccSettlement.VccTransactionInquiryRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/vccSettlement/ProcessRcnTransactionInquiry")
	defer span.End()

	// Acquire key from cache to force 1 process per request
	lockKey := fmt.Sprintf(constant.ProcessRcnTransactionInquiryCacheLockKey, s.config.ServiceName, request.MerchantId, request.RcnId, request.PostingDate)
	islockAcquired, err := s.cache.SetNX(ctx, lockKey, "1", constant.ProcessRcnTransactionInquiryCacheLockDuration).Result()
	if err != nil && err != redis.Nil {
		s.logger.Error(ctx, "error acquire vcc settlement transaction inquiry process lock", logger.Error(err), logger.String("key", lockKey))
		return constant.ErrAcquireTransactionInquiryLock
	}
	if !islockAcquired {
		s.logger.Warn(ctx, "failed acquire vcc settlement transaction inquiry process lock. there is ongoing process", logger.String("key", lockKey))
		return nil
	}
	defer func() {
		_, err = s.cache.Del(ctx, lockKey).Result()
		if err != nil {
			s.logger.Warn(ctx, "error delete lock key", logger.Error(err), logger.Any("key", lockKey))
		}
	}()

	merchantRcn, err := s.merchantRcnSvc.GetRcnDetail(ctx, request.RcnId, request.MerchantId)
	if err != nil {
		s.logger.Error(ctx, "error retrieve merchant rcn detail", logger.Error(err))
		return err
	}
	defer merchantRcn.EraseSensitiveData()

	stateKey := fmt.Sprintf(constant.ProcessRcnTransactionInquiryCacheStateKey, s.config.ServiceName, request.MerchantId, request.RcnId, request.PostingDate)
	lastProcessedPage, err := s.cache.Get(ctx, stateKey).Result()
	if err != nil && err != redis.Nil {
		s.logger.Warn(ctx, "error when retrieve settlement transaction inquiry last state.", logger.String("key", stateKey), logger.Error(err))
	}
	var page = constant.DefaultPage
	if lastProcessedPage != "" {
		parsedPage, err := strconv.ParseInt(lastProcessedPage, 10, 32)
		if err != nil {
			s.logger.Warn(ctx, "invalid state page value", logger.Error(err), logger.String("lastProcessedPage", lastProcessedPage))
		}
		if parsedPage > 0 {
			page = int(parsedPage)
		}
	}
	if page == constant.DefaultPage {
		/// Notes: soft delete previous posting date to avoid data duplicates without valid identifier
		postingDatetime, _ := time.Parse(constant.VccDateFormat, request.PostingDate)
		err = s.vccSettlementRepo.Delete(ctx, request.RcnId, postingDatetime)
		if err != nil {
			s.logger.Error(ctx, "error when soft delete existing settlement transaction data", logger.String("rcnId", request.RcnId), logger.String("postingDate", request.PostingDate))
			return constant.ErrProcessSettlementTransactionInquiry
		}
		s.logger.Info(ctx, "soft delete settlement data for same posting date & rcnId")
	}

	for {
		s.logger.Info(ctx, "retrieve merchant transaction.", logger.String("partnerReferenceNo", request.PartnerReferenceNo), logger.Int("page", page))
		processorTrxList, err := s.cimbProcessor.InquiryTransactionCorporateCreditCard(ctx, &cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest{
			PartnerReferenceNo: request.PartnerReferenceNo,
			RecordType:         request.RecordType,
			BillingCycle:       request.BillingCycle,
			PostingDate:        request.PostingDate,
			BankCardNo:         merchantRcn.CardNumber,
			Page:               page,
		})
		if err != nil {
			s.logger.Error(ctx, "error when inquiry merchant rcn transaction from bank partner", logger.Error(err))
			// TODO: Add Send slack alert
			_ = s.notifyRcnTransactionInquiryUpdates(ctx, &vccSettlement.VccTransactionInquiryAlert{
				Title:       "Error Inquiry Merchant RCN Transaction to Bank Partner",
				Recipient:   s.config.VccSlackRecipient.DefaultRecipient,
				Description: fmt.Sprintf(`Error when inquiry merchant rcn transaction from bank partner. Error detail: %s`, err.Error()),
				RcnId:       request.RcnId,
				PostingDate: request.PostingDate,
			})
			return err
		}

		vccTrxList := processorTrxList.ToVccSettlementModel(request.RcnId)
		if len(vccTrxList) == 0 {
			s.logger.Info(ctx, "no data found")
			break
		}
		err = s.vccSettlementRepo.BulkInsert(ctx, vccTrxList)
		if err != nil {
			s.logger.Error(ctx, "error when bulk insert vcc trx list to database")
			_ = s.notifyRcnTransactionInquiryUpdates(ctx, &vccSettlement.VccTransactionInquiryAlert{
				Title:       "Error Insert Merchant RCN Transaction to Database",
				Recipient:   s.config.VccSlackRecipient.DefaultRecipient,
				Description: fmt.Sprintf(`Error when insert merchant rcn transaction to database. Error detail: %s`, err.Error()),
				RcnId:       request.RcnId,
				PostingDate: request.PostingDate,
			})
			return constant.ErrProcessSettlementTransactionInquiry
		}

		// store state to redis
		_, err = s.cache.Set(ctx, stateKey, fmt.Sprintf("%d", page), constant.ProcessRcnTransactionInquiryCacheStateDuration).Result()
		if err != nil {
			s.logger.Warn(ctx, "error when store page state", logger.Error(err), logger.String("stateKey", stateKey), logger.Int("page", page))
		}

		if !processorTrxList.HasNextPage {
			break
		}

		page++
	}

	// Delete state cache key
	_, err = s.cache.Del(ctx, stateKey).Result()
	if err != nil {
		s.logger.Warn(ctx, "error when store page state", logger.Error(err), logger.String("stateKey", stateKey), logger.Int("page", page))
	}

	s.logger.Info(ctx, "finish process settlement transaction inquiry")

	return nil
}

func (s *VccSettlementService) notifyRcnTransactionInquiryUpdates(ctx context.Context, request *vccSettlement.VccTransactionInquiryAlert) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/vccSettlement/notifyRcnTransactionInquiryUpdates")
	defer span.End()

	s.logger.Info(ctx, "send alert notification")
	errNotify := s.notificationSvc.SendVccSettlementTransactionAlert(ctx, request)
	if errNotify != nil {
		s.logger.Warn(ctx, "failed send vcc transaction alert to slack", logger.Error(errNotify), logger.Any("request", request))
		return errNotify
	}
	return nil
}
