package fraudnetrepository

import (
	"bytes"
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	fraudnetmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fraudNet"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const updateURL = "/v2/risk/transaction/banking_marketplace"

func (r *FraudNetRepository) Update(ctx context.Context, request *fdscommon.UpdateRequest) (*fdscommon.UpdateResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/fdsProcessor/fraudNetRepository/Update")
	defer span.End()

	var (
		response *fdscommon.UpdateResponse
		err      error
		url      = r.baseURL() + updateURL
	)

	basicAuth := CreateBasicAuth(
		r.secret.FraudNetSecret.AccessKey,
		r.secret.FraudNetSecret.AccessSecret,
	)

	header := map[string]string{
		constant.HeaderAuthorization: basicAuth,
		constant.HeaderContentType:   constant.MIMEApplicationJSON,
	}

	// For now we can use directly, since the mapping is directly one to one
	respBytes, _, err := r.httpRequest.PATCH(ctx, url, request, header)
	if err != nil {
		resp, err := ValidateHttpResponse[fraudnetmodel.MarketplaceUpdateResponse](bytes.NewReader(respBytes))
		if err != nil {
			r.logger.Error(ctx,
				"FraudNetRepository | Failed to validate fraud net update response",
				logger.Any("error", err),
			)
			return nil, err
		}
		response = UpdateRequestMapToCommonResponse(resp)
		return response, err
	}

	resp, err := ValidateHttpResponse[fraudnetmodel.MarketplaceUpdateResponse](bytes.NewReader(respBytes))
	if err != nil {
		r.logger.Error(ctx,
			"FraudNetRepository | Failed to validate fraud net update response",
			logger.Any("error", err),
		)
		return nil, err
	}

	response = UpdateRequestMapToCommonResponse(resp)
	return response, nil
}
