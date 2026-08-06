package merchantTopUp

import (
	"context"
	stdErrors "errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *merchantTopUpService) FindOrCreate(ctx context.Context, merchantId, accountName, paymentMethodId string) (*model.MerchantTopUp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantTopUp/FindOrCreate")
	defer segment.End()

	var parentMerchant *merchant.Merchant

	merchant, err := s.merchantService.FindMerchantByID(ctx, merchantId)
	if err != nil {
		return nil, errors.New(responseHttp.HttpErrDatabase, constant.ErrFindMerchant)

	} else if merchant == nil {
		return nil, errors.New(responseHttp.HttpErrRequest, constant.ErrMerchantNotFound)

	} else if merchant.ParentID.String != "" {
		if parentMerchant, err = s.merchantService.FindMerchantByID(ctx, merchant.ParentID.String); err != nil {
			return nil, errors.New(responseHttp.HttpErrForbidden, constant.ErrFindParentMerchant)
		}
	}

	reference, err := s.merchantTopUpRepo.GetByMerchantAccountNameAndPaymentMethodId(ctx, merchant.UUID, accountName, paymentMethodId)
	if err != nil {
		return nil, errors.New(responseHttp.HttpErrNotFound, err)
	}

	paymentMethod, err := s.paymentMethodRepo.GetPaymentMethodById(ctx, paymentMethodId)
	if err != nil {
		s.logger.Error(ctx, "failed to get payment_method by payment_method_id", logger.Error(err))
		if stdErrors.Is(err, constant.ErrPaymentMethodNotFound) {
			return nil, errors.New(responseHttp.HttpErrNotFound, err)
		}
		return nil, errors.New(responseHttp.HttpErrDatabase, err)

	} else if paymentMethod.Type != paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT {
		return nil, errors.New(responseHttp.HttpErrRequest, constant.ErrIncorrectPaymentMethod)
	}

	if reference != nil {
		return reference, nil
	}

	// if reference not found, generate a va to snap-core and then create a new merchant top up
	referenceUUID := uuid.NewString()

	snapBody := snapCoreModel.CreateVirtualAccountRequest{
		MID:          merchant.MID.String,
		CustomerNo:   merchant.UUID,
		AccountName:  merchant.Name,
		AccountEmail: merchant.MerchantEmail,
		AccountPhone: merchant.MerchantPhone,
		TotalAmount: snapCoreModel.Amount{
			Currency: "IDR", Value: "0",
		},
		Acquirer:      paymentMethod.Acquirer,
		IsCloseAmount: snapCoreModel.VaTrxType(paymentConst.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC).IsCloseAmount,
		IsSingleUse:   snapCoreModel.VaTrxType(paymentConst.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC).IsSingleUsed,
		Purpose:       paymentConst.VIRTUAL_ACCOUNT_TRX_PURPOSE_TOPUP_BALANCE,
		MerchantID:    merchantId,
	}
	if parentMerchant != nil {
		snapBody.MID = parentMerchant.MID.String
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    referenceUUID,
		ReferenceId: merchant.UUID,
		From:        "Merchant-Topup-Service",
	})

	// generate va to snap-core
	va, err := s.snapCore.CreateVirtualAccount(ctx, snapBody)
	if err != nil {
		s.logger.Error(ctx, "failed to create virtual_account to snap_core", logger.Error(err))
		// Propagate the already-typed downstream error from snap-core repository
		// (mapPartnerHTTPStatusToErrorType wraps partner 5xx/timeout/4xx correctly).
		return nil, err
	}

	// create a new merchant top up
	reference = &model.MerchantTopUp{
		ID:              referenceUUID,
		MerchantID:      merchant.UUID,
		AccountName:     accountName,
		PaymentMethodID: paymentMethodId,
		ReferenceNumber: va.VirtualAccountNo,
		CreatedAt:       time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err = s.merchantTopUpRepo.Create(ctx, reference); err != nil {
		s.logger.Error(ctx, "failed to create merchant top up reference", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrDatabase, err)
	}
	return reference, nil
}
