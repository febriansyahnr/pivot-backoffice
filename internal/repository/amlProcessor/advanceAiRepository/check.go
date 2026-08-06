package advanceairepository

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const checkURL = "/intl/openapi/journey/v2/submit"

func (r *AdvanceAiRepository) Check(ctx context.Context, request *amlcommon.CheckRequest) (*amlcommon.CheckResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/amlProcessor/advanceAiRepository/Check")
	defer span.End()

	baseURL := r.baseURL() + checkURL
	params := url.Values{}
	params.Add("journeyId", r.journeyID())
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	var (
		response *amlcommon.CheckResponse
		err      error
	)

	header := map[string]string{
		constant.HeaderXAdvAIKey: r.secret.AdvanceAISecret.ApiKey,
	}

	respBytes, _, err := r.httpRequest.POST(ctx, fullURL, request, header)
	if err != nil {
		response, err := ValidateHttpResponse[amlcommon.CheckResponse](bytes.NewReader(respBytes))
		if err != nil {
			r.logger.Error(ctx,
				"AdvanceAiRepository | Failed to validate check response",
				logger.Any("error", err),
			)
			return nil, err
		}
		return response, err
	}

	response, err = ValidateHttpResponse[amlcommon.CheckResponse](bytes.NewReader(respBytes))
	if err != nil {
		r.logger.Error(ctx,
			"AdvanceAiRepository | Failed to validate check response",
			logger.Any("error", err),
		)
		return nil, err
	}

	return response, nil
}
