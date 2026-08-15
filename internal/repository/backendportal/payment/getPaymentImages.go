package paymentRepository

import (
	"context"
	"encoding/json"
	"fmt"

	pdkRetriever "github.com/paper-indonesia/pdk/v2/goff/retriever"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
)

func (r *PaymentRepository) RetrieveImages(ctx context.Context) (paymentModel.ImageResponse, error) {
	var imageRetriever ConsulRetriever
	var err error

	// Use injected retriever if available (for testing), otherwise create a new one
	if r.consulRetriever != nil {
		imageRetriever = r.consulRetriever
	} else {
		imageRetriever, err = pdkRetriever.NewConsulRetriever(
			r.config.FeatureFlagConfig.ConsulAddr,
			r.config.FeatureFlagConfig.ConsulPaymentImages,
			r.secret.ConsulSecret.Token,
		)
		if err != nil {
			return paymentModel.ImageResponse{}, err
		}
	}

	data, err := imageRetriever.Retrieve(ctx)
	if err != nil {
		return paymentModel.ImageResponse{}, err
	}

	var response paymentModel.ImageResponse
	if err = json.Unmarshal(data, &response); err != nil {
		return paymentModel.ImageResponse{}, fmt.Errorf("failed to unmarshal data into ImageResponse struct: %w", err)
	}

	return response, nil
}
