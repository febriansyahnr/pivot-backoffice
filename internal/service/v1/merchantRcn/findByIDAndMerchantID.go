package merchantRcn

import (
	"context"
	"encoding/base64"
	"fmt"

	cimbProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cimbProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantRcn"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (m *MerchantRcnService) FindByIDAndMerchantID(ctx context.Context, id string, merchantId string) (*merchantRcn.MerchantRcnResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantRCN/FindByIDAndMerchantID")
	defer segment.End()

	merchantRcnResp, err := m.repo.FindByIDAndMerchantID(ctx, id, merchantId)
	if err != nil {
		m.logger.Error(ctx, "error when finding merchant by id", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrInternal, err)
	}

	if merchantRcnResp == nil {
		m.logger.Error(ctx, "merchant rcn not found")
		return nil, errors.New(responseHttp.HttpErrNotFound, fmt.Errorf("merchant rcn not found"))
	}

	decodedRealCardNumber, err := base64.StdEncoding.DecodeString(merchantRcnResp.RealCardNumber)
	if err != nil {
		m.logger.Info(ctx, "merchant rcn failed to decode real card number", logger.Error(err))
		return nil, err
	}

	cardNumber, err := m.gcsEncryption.DecryptSymmetric(ctx, decodedRealCardNumber)
	if err != nil {
		m.logger.Info(ctx, "merchant rcn decrypt real card card number error", logger.Error(err))
		return nil, err
	}

	card, err := m.cimbProcessor.InquiryCorporateCreditCard(ctx, &cimbProcessorModel.InquiryCorporateCreditCardRequest{
		BankCardNo: cardNumber,
	})

	if err != nil {
		m.logger.Info(ctx, "inquiry card error", logger.Error(err))
		return nil, err
	}

	resp := merchantRcn.BuildMerchantRcnResponse(card)

	return &resp, nil
}
