package creditcard

import (
	"context"

	card "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
)

func (s *CreditCardService) CreateEncryptedCardAuthenticationLink(ctx context.Context, request *card.EncryptedCardAuthenticationRequest) (*card.EncryptedCardAuthenticationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/CreateEncryptedCardAuthenticationLink")
	defer segment.End()

	authenticationResponse, err := s.creditcardCoreProcessorRepo.CreateEncryptedCardAuthenticationLink(ctx, request.ToProcessorRequestModel())
	if err != nil {
		return nil, err
	}

	return card.ToEncryptedCardAuthenticationResponse(request, authenticationResponse), nil
}

func (s *CreditCardService) Authentication(ctx context.Context, request model.AuthenticationRequest) (*model.AuthenticationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/Authentication")
	defer segment.End()

	return s.creditcardCoreProcessorRepo.Authentication(ctx, request)
}
