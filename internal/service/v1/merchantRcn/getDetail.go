package merchantRcn

import (
	"context"
	"encoding/base64"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantRcn"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (m *MerchantRcnService) GetRcnDetail(ctx context.Context, id string, merchantId string) (*merchantRcn.MerchantRcnDetail, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchantRCN/GetRcnDetail")
	defer segment.End()

	merchantRcnResp, err := m.repo.FindByIDAndMerchantID(ctx, id, merchantId)
	if err != nil {
		m.logger.Error(ctx, "error when finding merchant by id", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrInternal, err)
	}
	if merchantRcnResp == nil {
		m.logger.Error(ctx, "merchant rcn not found")
		return nil, errors.New(responseHttp.HttpErrNotFound, constant.ErrMerchantRcnNotFound)
	}

	decodedRealCardNumber, err := base64.StdEncoding.DecodeString(merchantRcnResp.RealCardNumber)
	if err != nil {
		m.logger.Error(ctx, "failed to decode real card number", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrDecodeMerchantRcn)
	}
	cardNumber, err := m.gcsEncryption.DecryptSymmetric(ctx, decodedRealCardNumber)
	if err != nil {
		m.logger.Error(ctx, "error decrypt merchant real card number", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrDecryptMerchantRcn)
	}

	return &merchantRcn.MerchantRcnDetail{
		ID:              merchantRcnResp.ID,
		MerchantID:      merchantRcnResp.MerchantID,
		PrincipalIssuer: merchantRcnResp.PrincipalIssuer,
		CardNumber:      cardNumber,
		CreatedAt:       merchantRcnResp.CreatedAt,
		UpdatedAt:       merchantRcnResp.UpdatedAt,
	}, nil
}
