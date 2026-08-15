package snapCoreRepository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreQRModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qr"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *snapCoreRepository) InquiryStatusQris(ctx context.Context, request *snapCoreQRModel.InquiryStatusQrMpmRequest) (*snapCoreQRModel.QrisInquiryStatusResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/snapCore/InquiryStatusQris")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/qr-mpm/info/%s", r.config.SnapCoreConfig.BaseUrl, request.QrisUUID)
	if request.SkipPublish {
		url += "?skipPublish=true"
	}

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request to inquiry status QRIS payment", logger.Error(err), logger.String("qrisUUID", request.QrisUUID))
		return nil, err
	}

	var resp snapCoreQRModel.QrisInquiryStatusResponse
	if err = json.Unmarshal(response, &resp); err != nil {
		r.logger.Error(ctx, "error when read inquiry status QRIS payment response body", logger.Error(err), logger.String("qrisUUID", request.QrisUUID))
		return nil, err
	}

	r.logger.Info(ctx, "Response from inquiry status QRIS payment",
		logger.String("qrisUUID", request.QrisUUID),
		logger.Any("response", resp),
		logger.Int("statusCode", statusCode),
	)

	switch {
	case statusCode == 404, statusCode == 409:
		return &resp, nil
	case statusCode >= 500:
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, constant.ErrPaymentPartnerInGeneral)
	case statusCode >= 400:
		return nil, pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrPaymentPartnerInGeneral)
	}

	return &resp, nil
}
