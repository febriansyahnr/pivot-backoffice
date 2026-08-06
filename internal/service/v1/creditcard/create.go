package creditcard

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentMethodConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (c *CreditCardService) CreatePayment(ctx context.Context, request creditcardModel.CreateCardPaymentRequest) (*creditcardModel.CreateCardPaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/CreatePayment")
	defer segment.End()

	var (
		now             = time.Now().UTC().Truncate(time.Millisecond)
		expired         = now.Add(constant.CreditCardPaymentExpired)
		paymentUUID     = uuid.New().String()
		discount        = decimal.NewFromFloat(0) // ignore rn, will update soon
		processorStatus = constant.CreditCardStatusWaitingForPayment
	)

	if request.PaymentUUID != uuid.Nil {
		paymentUUID = request.PaymentUUID.String()
	}
	changePaymentMethod, _ := ctx.Value(constant.CtxChangePaymentMethod).(bool)

	paymentURL := c.generatePaymentLink(paymentUUID, request.MerchantID.String())

	merchant, errFind := c.merchantRepo.FindMerchantByID(ctx, request.MerchantID.String())
	if errFind != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, errFind)
	} else if merchant == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}
	// When sub-merchant create payment directly
	if _, ok := ctx.Value(constant.CtxParentMerchantId).(string); merchant.ParentID.String != "" && !ok {
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)
	}

	derivedMerchantId := request.MerchantID.String()
	if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		derivedMerchantId = merchant.ParentID.String
	}

	payment, err := c.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.MerchantID.String(), request.ReferenceID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if payment != nil && !changePaymentMethod {
		err = constant.ErrCreditcardReferenceIdAlreadyExist
		c.logger.Error(ctx, err.Error(), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDupCheck, err)
	}

	paymentMethod, err := c.paymentMethodSvc.GetActivePaymentMethodDetailForPaymentRequest(ctx, paymentModel.GetActivePaymentMethodRequest{
		MerchantID: derivedMerchantId,
		Category:   paymentMethodConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
		Type:       paymentMethodConstant.PAYMENT_METHOD_CREDIT_CARD,
		Acquirer:   constant.ACQUIRER_CC_HARSYA,
	})
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)

	} else if paymentMethod == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodNotFound)

	} else if paymentMethod.ChannelType == constant.PaymentMethodChannelTypeDirect && len(util.ValueOfPtr(request.SplitRoutingConfigurations)) > 0 {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrDoNotApplySplitRouteInFacilitatorModel)
	}

	amount := request.Amount // client input amount
	feeInDecimal := decimal.NewFromFloat(0)
	totalAmount := amount.Add(feeInDecimal).Sub(discount).Round(2)

	creditCardMetadata := creditcardModel.CreditcardMetadata{
		AuthenticationMethod: request.AuthenticationMethod,
		BankMerchantID:       request.BankMerchantID,
		ProcessorStatus:      processorStatus,
		RedirectUrl:          request.RedirectUrl,
		ClientRedirectUrl:    request.UnifiedPaymentRedirectUrl,
		IsUnifiedPayment:     request.IsUnifiedPayment,
	}
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		creditCardMetadata.OnBehalf = &merchantModel.OnBehalfObject{
			ParentMerchantId: parentMerchantId,
		}
	}
	if request.SplitRoutingConfigurations != nil && len(*request.SplitRoutingConfigurations) > 0 {
		creditCardMetadata.SplitRoutingConfigurations = request.SplitRoutingConfigurations
	}
	// Build credit card metadata
	creditCardMetadataByte, err := json.Marshal(creditCardMetadata)
	if err != nil {
		c.logger.Error(ctx, constant.ErrWhenMarshalCreditcardMetadata, logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, err)
	}

	metaDataString := string(creditCardMetadataByte)

	paymentDto := &paymentModel.PaymentDTO{
		UUID:            paymentUUID,
		ReferenceID:     &request.ReferenceID,
		MerchantID:      request.MerchantID.String(),
		PaymentMethodID: paymentMethod.UUID,
		Amount:          amount,
		Fee:             &feeInDecimal,
		Discount:        &discount,
		TotalAmount:     totalAmount,
		Currency:        request.Currency,
		Status:          constant.StatusPending,
		PaymentURL:      paymentURL,
		Metadata:        &metaDataString,
		ExpiredAt:       &expired,
		CreatedBy:       &request.CreatedBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	accountTrxMetadata := orchestratorModel.MetadataPayment[orchestratorModel.MetadataPaymentMethodCC]{
		ReconReferenceNo: "",
		ExpiredAt:        expired,
		MethodDetail:     orchestratorModel.MetadataPaymentMethodCC{}, // Should be filled once the is payment notification
	}
	rawAccountTrxMetadata, _ := json.Marshal(accountTrxMetadata)

	// Begin Tx
	ctx, err = c.paymentRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	isCompleted := false
	defer func() {
		if !isCompleted {
			if err = c.paymentRepo.RollbackTransaction(ctx); err != nil {
				return
			}
		}
	}()

	if changePaymentMethod {
		currentPaymentMethod, _ := ctx.Value(constant.CtxCurrentPaymentMethod).(string)
		// Update new payment method
		if err = c.paymentRepo.ChangePaymentMethod(ctx, paymentDto); err == nil {
			trxRequest := orchestratorModel.UpdateTransactionWithPendingStatus{
				Channel:         constant.ChannelCreditCard,
				Metadata:        rawAccountTrxMetadata,
				UpdatedAt:       time.Now().UTC(),
				Processor:       constant.CreditCardCoreProcessor,
				ProcessorID:     "",
				SettlementModel: paymentMethod.ChannelType,
			}
			err = c.accountTransactionRepo.UpdateTransactionWithPendingStatusByReferenceIdTypeAndChannel(
				ctx, paymentUUID, constant.TypePayment, currentPaymentMethod, trxRequest,
			)
		}
	} else {
		// Create new payment
		if err = c.paymentRepo.CreatePayment(ctx, paymentDto); err == nil {
			// Create PENDING transaction with QRIS detail
			trxRequest := &orchestratorModel.CreateAccountTransactionRequest{
				UUID:                 uuid.New(),
				ReferenceID:          paymentUUID,
				Type:                 constant.TypePayment,
				MerchantID:           request.MerchantID,
				Currency:             request.Currency,
				Credit:               totalAmount.InexactFloat64(),
				Channel:              constant.ChannelCreditCard,
				Status:               constant.StatusPending,
				SettlementStatus:     util.ValueToPtr(constant.StatusPending),
				TransactionTimestamp: paymentDto.CreatedAt,
				Usecase:              constant.TypePayment,
				Processor:            constant.CreditCardCoreProcessor,
				ProcessorID:          "",
				AdditionalInfo: types.NullJSONText{
					Valid: true, JSONText: rawAccountTrxMetadata,
				},
				SettlementModel: util.ValueToPtr(paymentMethod.ChannelType),
			}
			// Action For Post Transaction
			err = c.orchestratorSvc.PostAccountTransaction(ctx, trxRequest)
		}
	}
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Commit Tx
	if err = c.paymentRepo.CommitTransaction(ctx); err != nil {
		return nil, err
	}
	isCompleted = true

	return &creditcardModel.CreateCardPaymentResponse{
		UUID:           paymentDto.UUID,
		MerchantID:     request.MerchantID,
		BankMerchantID: request.BankMerchantID,
		ReferenceID:    request.ReferenceID,
		Amount:         paymentDto.Amount,
		Currency:       paymentDto.Currency,
		Created:        paymentDto.CreatedAt.Format(time.RFC3339),
		Expired:        paymentDto.ExpiredAt.Format(time.RFC3339),
		PaymentURL:     paymentDto.PaymentURL,
		Status:         processorStatus,
	}, nil
}

func (c *CreditCardService) generatePaymentLink(id, merchantId string) string {
	id = c.hashURL(id)
	merchantId = c.hashURL(merchantId)
	uniqueLink := fmt.Sprintf("%s|%s|%d", id, merchantId, time.Now().UnixNano())
	return fmt.Sprintf("%s/pay/%s", c.config.CreditcardConfig.WebviewURL, c.hashURL(uniqueLink))
}

func (c *CreditCardService) hashURL(data string) string {
	return base64.URLEncoding.EncodeToString([]byte(data))
}
