package internalPayoutController

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"google.golang.org/protobuf/proto"
)

func (c *InternalPayoutController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/payout/Create")
	defer segment.End()

	var (
		wg                           = sync.WaitGroup{}
		eg                           = new(errgroup.Group)
		chunkSize                    = constant.BulkDisbursementMaxDataRequestPerBatch
		validDataCreateDisbursements []disbursementModel.CreateSingleRequest
		bankDB                       = bankTransfer.NewBankDB()
		collectReferenceId           []string
		totalAmount                  = decimal.Zero
		traceId, _                   = ctx.Value(pdkConst.CtxTraceIdKey).(string)
		queueTTLLock                 = time.Duration(c.config.AppConfig.BulkDisbursementExpireLockMinute) * time.Minute
		err                          error
	)
	now := time.Now()

	// Merchant info from JWT
	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	merchantID := merchantCtx.MerchantId
	merchantConfigID := merchantCtx.MerchantId

	onBehalfSubMerchantId := r.Header.Get(constant.HeaderXSubMerchantID)
	if onBehalfSubMerchantId != "" {
		merchantID = onBehalfSubMerchantId

		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantCtx.MerchantId)
	}

	// Find merchant
	merchant, err := c.merchantSvc.FindMerchantByID(ctx, merchantID)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return

	} else if onBehalfSubMerchantId != "" && merchant.KYCStatus.String == constant.KYCStatusApproved {
		merchantConfigID = merchantID
	} else if onBehalfSubMerchantId == "" && merchant.ParentID.String != "" { // Sub-Merchant Direct Transaction
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)

		if merchant.KYCStatus.String == constant.KYCStatusNotRequired {
			merchantConfigID = merchant.ParentID.String
		}
	}

	var requestPayload disbursementModel.CreateDisbursementFromOpenApiRequest
	if err = json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)
			response.SendOpenApiNonSnapResponseError(ctx, w,
				pkgErrors.New(response.HttpErrValidation, fmt.Errorf("format payout is invalid")))
			return
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(requestPayload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	defer func() {
		metricData := &monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			ComponentName:        constant.ComponentNamePayout,
			MetricName:           constant.MetricNamePayout,
			MetricValue:          1,
			Attributes: map[string]any{
				"merchantId":          merchantCtx.MerchantId,
				"onBehalfSubmerchant": merchantCtx.MerchantId != merchantID,
			},
		}
		if err != nil {
			errType, errDetail := pkgErrors.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "failed to record payout custom metric", logger.Error(errMetric))
		}
	}()

	trxConfig, err := c.disbursementSvc.GetTransactionConfig(ctx, merchantConfigID)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	if len(requestPayload.Payouts) > constant.BulkDisbursementMaxDataRequest {
		err = pkgErrors.New(response.HttpErrRequest, constant.ErrMaxBulkDisbursementRequestAllowed)
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	_, isExcludeBeneficiaryLimit := c.disbursementSvc.IsMerchantAllowedExcludeBeneficiaryRules(ctx, merchantConfigID, 0)

	// validate payout items
	for i, payout := range requestPayload.Payouts {
		if err = c.validate.Struct(payout); err != nil {
			err = pkgErrors.New(response.HttpErrRequest, err)
			response.SendOpenApiNonSnapResponseError(ctx, w, err)
			return
		}

		if err = c.validatePayoutItem(ctx, &requestPayload.Payouts[i], merchantID, trxConfig, collectReferenceId); err != nil {
			if err == constant.ErrTimeout {
				err = pkgErrors.New(response.HttpErrRequestTimeout, err)
				response.SendOpenApiNonSnapResponseError(ctx, w, err)
				return
			}
			err = pkgErrors.New(response.HttpErrRequest, err)
			response.SendOpenApiNonSnapResponseError(ctx, w, err)
			return
		}

		collectReferenceId = append(collectReferenceId, payout.ReferenceID)
	}

	queueKeys, isCompleted := []string{}, false
	defer func() {
		if !isCompleted && len(queueKeys) > 0 {
			if e := c.redis.Del(ctx, queueKeys...).Err(); e != nil {
				c.logger.Error(ctx, "clears the process queue lock list", logger.Error(e))
			}
		}
	}()

	var mu sync.Mutex
	sem := semaphore.NewWeighted(int64(c.config.WorkerPoolConfig.Disbursement))

	// Build request
	for key, requestPayout := range requestPayload.Payouts {
		if err := sem.Acquire(ctx, 1); err != nil {
			c.logger.Error(ctx, "failed to acquirer semaphore", logger.Error(err))
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrInternal, constant.ErrInternalServerForUser))
			return
		}

		eg.Go(func() error {
			var beneficiaryBankCode, beneficiaryBankName string
			defer sem.Release(1)

			// Check InquiryID, if exist than query from requestAccountInquiries
			if requestPayout.InquiryID != "" {
				reqInquiryAcc, errReqInqAcc := c.inquiryAccSvc.FindLatestByInquiryID(ctx, requestPayout.InquiryID, merchantID)
				if errReqInqAcc != nil {
					return errReqInqAcc
				}

				if reqInquiryAcc.Status.String == constant.RequestAccountInquiryStatusInvalid {
					c.logger.Info(ctx, "invalid status inquiry", logger.String("inquiry_id", requestPayout.InquiryID))
					return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("invalid status inquiry"))
				}

				_, err := util.ValidateMagicNumber(http.MethodPost, reqInquiryAcc.BeneficiaryAccountNo.String)
				if c.config.Environment != constant.EnvironmentProduction && err != nil {
					if err == constant.ErrTimeout {
						return pkgErrors.New(response.HttpErrRequestTimeout, err)
					}

					return pkgErrors.New(response.HttpErrRequest, err)
				}

				bank := bankDB.FindByCode(reqInquiryAcc.BeneficiaryBankCode)
				if bank == nil {
					return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("invalid beneficiary bank code"))
				}

				// Assign to channel code & channel information
				requestPayout.ChannelCode = bank.ChannelCode
				requestPayout.ChannelInformation.AccountNumber = reqInquiryAcc.BeneficiaryAccountNo.String
				requestPayout.ChannelInformation.AccountName = reqInquiryAcc.MasterBeneficiaryAccountName
			}

			queueKey := fmt.Sprintf(constant.BulkDisbursementQueueLockFmt, merchant.UUID, requestPayout.ReferenceID)
			if ok, err := c.redis.SetNX(ctx, queueKey, true, queueTTLLock).Result(); err != nil {
				c.logger.Error(ctx, "set exclusive queue with key "+queueKey, logger.Error(err))
				return pkgErrors.New(response.HttpErrDatabase, fmt.Errorf("QUEUE: "+constant.InternalErrorFmt, traceId))
			} else if !ok {
				return pkgErrors.New(response.HttpErrDupCheck, constant.ErrPayoutsInProcess)
			}

			mu.Lock()
			queueKeys = append(queueKeys, queueKey)
			mu.Unlock()

			bank := bankDB.FindByChannelCode(requestPayout.ChannelCode)
			if bank != nil {
				beneficiaryBankCode = bank.Code
				beneficiaryBankName = bank.Name
			}

			amount, _ := decimal.NewFromString(requestPayout.Amount.Value)
			maxAmount := decimal.NewFromFloat(trxConfig.MaxAmount)

			if c.disbursementSvc.IsBankcodeOverbookingChannelAllowed(ctx, beneficiaryBankCode, merchantID) {
				maxAmount = decimal.NewFromFloat(c.config.DisbursementConfig.OverbookingBankMaxAmount)

				beneficiaryAccount, errBeneficiary := c.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(ctx, &beneficiaryModel.CheckAccountRequest{
					BeneficiaryBankCode:  beneficiaryBankCode,
					BeneficiaryAccountNo: requestPayout.ChannelInformation.AccountNumber,
					MerchantID:           merchantConfigID,
					AdditionalInfo:       map[string]any{},
				})
				if errBeneficiary != nil {
					return errBeneficiary
				}

				if max, isAllow := c.disbursementSvc.IsMerchantAllowedExcludeBeneficiaryRules(ctx, merchantConfigID, amount.InexactFloat64()); isAllow {
					maxAmount = decimal.NewFromFloat(max)
				} else if c.disbursementSvc.IsMerchantAllowedToUseBeneficiaryCustomRule(ctx, merchantConfigID, beneficiaryAccount.MetadataObj.BeneficiaryPayoutLimitRule != nil) {
					maxAmount = decimal.NewFromFloat(c.config.DisbursementConfig.OverbookingBankMaxAmountForCustomRule)
				}

				// note: when benef nil, forward ke snap to check VA

				// Check if the payout destination is a Pivot internal VA. If it matches the internal VA pattern, the transaction is declined.
				if beneficiaryAccount.MetadataObj.IsVirtualAccount {
					if !constant.IsPayoutToVirtualAccountAllowed(beneficiaryAccount.BeneficiaryBankCode, beneficiaryAccount.BeneficiaryAccountNo) {
						return pkgErrors.New(
							response.HttpErrRequest, fmt.Errorf(constant.ErrDetailMsgPayoutDstNotEligibleFmt, beneficiaryAccount.BeneficiaryBankCode, beneficiaryAccount.BeneficiaryAccountNo),
						)
					}
				}
			}

			if amount.GreaterThan(maxAmount) && !isExcludeBeneficiaryLimit {
				return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("invalid amount value greater than max amount value"))
			}

			// Append valid transaction
			createSingleDisbursementRequest := disbursementModel.CreateSingleRequest{
				ReferenceID:            requestPayout.ReferenceID,
				BeneficiaryBankCode:    beneficiaryBankCode,
				BeneficiaryBankName:    beneficiaryBankName,
				BeneficiaryAccountNo:   requestPayout.ChannelInformation.AccountNumber,
				BeneficiaryAccountName: requestPayout.ChannelInformation.AccountName,
				Amount:                 amount,
				Remark:                 requestPayout.Description,
				InquiryID:              requestPayout.InquiryID,
			}

			mu.Lock()
			totalAmount = totalAmount.Add(amount)
			validDataCreateDisbursements = append(validDataCreateDisbursements, createSingleDisbursementRequest)

			// To Assign Response
			requestPayload.Payouts[key] = requestPayout
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	payoutCutOffTime, err := c.disbursementSvc.GetCutOffTimeStatus(ctx, time.Now().UTC(), merchantID, nil)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	dailyLimit, err := c.disbursementSvc.ValidateDailyTransactionLimit(ctx, merchantID, totalAmount.InexactFloat64())
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	defer func() { dailyLimit.Close(context.Background(), isCompleted) }()

	ww := monitor.WrapResponse(w, r)
	defer func() {
		monitor.WriteAndSend(
			ctx, "internal-controller-payout-create", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", merchantID),
					fmt.Sprintf("amount:%s", totalAmount.String()),
					fmt.Sprintf("total:%d", len(validDataCreateDisbursements)),
				}
			},
		)
	}()

	// Create bulk disbursement first (bypass status to IN_PROGRESS)
	bulkDisbursement, err := c.disbursementSvc.CreateBulk(ctx, &disbursementModel.CreateBulkDisbursementRequest{
		MerchantID: merchantID,
		File:       "",
		Status:     constant.BulkDisbursementStatusInProgress,
		CreatedBy:  merchantCtx.MerchantId,
	})
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, ww, err)
		return
	}

	// create redis data for marking bulk disbursement as in progress
	queueKey := fmt.Sprintf(constant.BulkDisbursementInProgressQueueLockFmt, merchantID, bulkDisbursement.UUID)
	if _, err := c.redis.SetNX(ctx, queueKey, true, queueTTLLock).Result(); err != nil {
		c.logger.Error(ctx, "set exclusive queue with key "+queueKey, logger.Error(err))
		err = pkgErrors.New(response.HttpErrDatabase, fmt.Errorf("QUEUE: "+constant.InternalErrorFmt, traceId))
		response.SendOpenApiNonSnapResponseError(ctx, ww, err)
		return
	}

	// Create disbursements by looping batch then send to rmq
	for i := 0; i < len(validDataCreateDisbursements); i += chunkSize {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()

			end := start + chunkSize
			if end > len(validDataCreateDisbursements) {
				end = len(validDataCreateDisbursements)
			}
			batchRequest := &pb.BatchCreateDisbursementRequest{
				BulkId:       bulkDisbursement.UUID,
				MerchantId:   bulkDisbursement.MerchantID,
				MerchantName: merchant.Name,
				CreatedBy:    merchantCtx.MerchantId,
				CreatedFrom:  constant.DisbursementCreatedFromOpenApi,
				TotalTrx:     int64(len(validDataCreateDisbursements)),
				AutoApprove:  true,
				Data:         disbursementModel.TransformArrayCreateSingleRequestToProtobufType(validDataCreateDisbursements[start:end]),
			}
			payload, _ := proto.Marshal(batchRequest)

			_ = c.rabbitMqExt.Publish(ctx, rabbitMqExt.BulkDisbursementBatchCreateRoutingKey, nil, payload)
		}(i)
	}
	wg.Wait()
	isCompleted = true

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&merchantCtx.MerchantId,
		nil,
		constant.TagDisbursement,
		constant.ActivityMerchantCreateDisbursement,
		requestPayload,
	)

	// build response
	payoutResponse := disbursementModel.CreateDisbursementFromOpenApiResponse{
		UUID:       bulkDisbursement.UUID,
		MerchantID: merchantID,
		Payouts:    requestPayload.Payouts,
		Status:     bulkDisbursement.Status,
		CreatedAt:  bulkDisbursement.CreatedAt,
		UpdatedAt:  bulkDisbursement.UpdatedAt,
	}

	if payoutCutOffTime.Status == constant.DisbursementCutOffTimeStatusOngoing {
		response.SendOpenApiResponseSuccess(
			ww, http.StatusOK, c.config.DisbursementConfig.CutOffTimeWindow.TransactionInfo, payoutResponse,
		)
		return
	}
	response.SendOpenApiResponseOK(ww, payoutResponse)
}

func (c *InternalPayoutController) validatePayoutItem(
	ctx context.Context, payout *disbursementModel.PayoutObjectForCreate, merchantID string, trxConfig *disbursementModel.TransactionConfig, collectReferenceId []string,
) error {
	// Validate Channel Information
	if payout.InquiryID == "" {
		if payout.ChannelInformation.AccountNumber == "" {
			return fmt.Errorf("accountNumber is required")
		}

		if payout.ChannelInformation.AccountName == "" {
			return fmt.Errorf("accountName is required")
		}
	}

	bankDB := bankTransfer.NewBankDB()

	_, err := util.ValidateMagicNumber(http.MethodPost, payout.ChannelInformation.AccountNumber)
	if c.config.Environment != constant.EnvironmentProduction && err != nil {
		return err
	}

	if len(payout.ChannelInformation.AccountName) > constant.DisbursementMaxLengthBeneficiaryName {
		return constant.ErrBeneficiaryNameLengthExceeded
	}

	if err = c.validateMinAmount(payout.Amount.Value, trxConfig.MinAmount); err != nil {
		return err
	}

	if utf8.RuneCountInString(payout.Description) > constant.DisbursementMaxLengthRemark {
		// if payout description is more than 20 characters, truncate it
		payout.Description = payout.Description[:constant.DisbursementMaxLengthRemark]
	}

	bank := bankDB.FindByChannelCode(payout.ChannelCode)
	if bank == nil {
		return fmt.Errorf("channel code not found")
	}

	// validate reference ID
	for _, refID := range collectReferenceId {
		if refID == payout.ReferenceID {
			return constant.ErrDuplicateDisbursementReferenceId
		}
	}

	if c.disbursementSvc.IsExistReferenceID(ctx, merchantID, payout.ReferenceID) {
		return constant.ErrDisbursementReferenceIdAlreadyExist
	}
	return nil
}

func (c *InternalPayoutController) validateMinAmount(amount string, minAmount float64) error {
	val, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return constant.ErrInvalidDisbursementAmount
	}

	if val < int64(minAmount) {
		return fmt.Errorf("min amount %s", util.ConvertFloatToCurrency(minAmount))
	}

	return nil
}
