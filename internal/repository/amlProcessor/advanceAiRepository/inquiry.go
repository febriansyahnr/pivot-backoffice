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

const inquiryURL = "/openapi/onestop/inquiry/info"

func (r *AdvanceAiRepository) Inquiry(ctx context.Context, transactionID string) (*amlcommon.InquiryResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/amlProcessor/advanceAiRepository/Inquiry")
	defer span.End()

	baseURL := r.baseURL() + inquiryURL
	params := url.Values{}
	params.Add("transactionId", transactionID)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	var (
		response *amlcommon.InquiryResponse
		err      error
	)

	header := map[string]string{
		constant.HeaderXAdvAIKey: r.secret.AdvanceAISecret.ApiKey,
	}

	respBytes, _, err := r.httpRequest.GET(ctx, fullURL, header)
	if err != nil {
		response, err := ValidateHttpResponse[amlcommon.InquiryResponse](bytes.NewReader(respBytes))
		if err != nil {
			r.logger.Error(ctx,
				"AdvanceAiRepository | Failed to validate inquiry response",
				logger.Any("error", err),
			)
			return nil, err
		}
		return response, err
	}

	response, err = ValidateHttpResponse[amlcommon.InquiryResponse](bytes.NewReader(respBytes))
	if err != nil {
		r.logger.Error(ctx,
			"AdvanceAiRepository | Failed to validate inquiry response",
			logger.Any("error", err),
		)
		return nil, err
	}

	return response, nil
}
