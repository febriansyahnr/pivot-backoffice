package paymentRepository

import (
	"context"
	"encoding/json"
	"fmt"

	pdkRetriever "github.com/paper-indonesia/pdk/v2/goff/retriever"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
)

func (r *PaymentRepository) RetrieveInstructions(ctx context.Context) ([]paymentModel.InstructionResponse, error) {
	var instructionRetriever ConsulRetriever
	var err error

	// Use injected retriever if available (for testing), otherwise create a new one
	if r.consulRetriever != nil {
		instructionRetriever = r.consulRetriever
	} else {
		// Create a new retriever specifically for instructions
		instructionRetriever, err = pdkRetriever.NewConsulRetriever(
			r.config.FeatureFlagConfig.ConsulAddr,
			r.config.FeatureFlagConfig.ConsulPaymentInstructions,
			r.secret.ConsulSecret.Token,
		)
		if err != nil {
			return nil, err
		}
	}

	data, err := instructionRetriever.Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	// Now expecting a slice of InstructionResponse
	var instructions []paymentModel.InstructionResponse
	if err = json.Unmarshal(data, &instructions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data into InstructionResponse struct: %w", err)
	}

	return instructions, nil
}
