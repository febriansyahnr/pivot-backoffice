package snapCoreRepository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreVAModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *snapCoreRepository) InquiryStatusVirtualAccount(ctx context.Context, request *snapCoreVAModel.InquiryStatusVARequest) (*snapCoreVAModel.InquiryStatusVAResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/InquiryStatusVAPayment")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/inquiry-status", r.config.SnapCoreConfig.BaseUrl)

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request to inquiry status VA payment", logger.Error(err))
		return nil, err
	}

	var resp snapCoreVAModel.InquiryStatusVAResponse

	if err = json.Unmarshal(response, &resp); err != nil {
		r.logger.Error(ctx, "error when read inquiry status VA payment response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "Response from inquiry status VA payment", logger.Any("request", request), logger.Any("response", resp), logger.Int("statusCode", statusCode))

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
