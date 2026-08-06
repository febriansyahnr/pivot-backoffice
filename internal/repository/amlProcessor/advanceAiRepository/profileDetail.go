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

const profileDetailURL = "/openapi/onestop/inquiry/aml-profile-detail"

func (r *AdvanceAiRepository) ProfileDetail(ctx context.Context, inquiryID string, profileID string) (*amlcommon.ProfileDetailResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/amlProcessor/advanceAiRepository/ProfileDetail")
	defer span.End()

	baseURL := r.baseURL() + profileDetailURL
	params := url.Values{}
	params.Add("inquiryId", inquiryID)
	params.Add("profileId", profileID)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	var (
		response *amlcommon.ProfileDetailResponse
		err      error
	)

	header := map[string]string{
		constant.HeaderXAdvAIKey: r.secret.AdvanceAISecret.ApiKey,
	}

	respBytes, _, err := r.httpRequest.GET(ctx, fullURL, header)
	if err != nil {
		response, err := ValidateHttpResponse[amlcommon.ProfileDetailResponse](bytes.NewReader(respBytes))
		if err != nil {
			r.logger.Error(ctx,
				"AdvanceAiRepository | Failed to validate profile detail response",
				logger.Any("error", err),
			)
			return nil, err
		}
		return response, err
	}

	response, err = ValidateHttpResponse[amlcommon.ProfileDetailResponse](bytes.NewReader(respBytes))
	if err != nil {
		r.logger.Error(ctx,
			"AdvanceAiRepository | Failed to validate profile detail response",
			logger.Any("error", err),
		)
		return nil, err
	}

	return response, nil
}
