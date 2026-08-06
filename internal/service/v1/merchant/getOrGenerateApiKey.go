package merchant

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) GetOrGenerateCallbackApiKey(ctx context.Context, id string) (*string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetOrGenerateCallbackApiKey")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, id)
	if err != nil {
		return nil, errors.New(responseHttp.HttpErrInternal, err)

	} else if merchant == nil {
		return nil, errors.New(responseHttp.HttpErrNotFound, fmt.Errorf("merchant with id %s not found", id))
	}

	if merchant.CallbackApiKey != nil {
		if merchant.CallbackApiKeyVersion > 0 {
			unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: *merchant.CallbackApiKey})
			if err != nil {
				s.logger.Error(ctx, "failed while decrypting merchant callback api key", logger.Error(err))
				return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
			}
			merchant.CallbackApiKey = util.ValueToPtr(string(unwrapped.Plaintext))
		}
		return merchant.CallbackApiKey, nil
	}

	callbackApiKey, _ := util.GenerateRandomString(32)

	wrapped, err := s.encryption.Encrypt(ctx, vault.EncryptRequest{Plaintext: []byte(callbackApiKey)})
	if err != nil {
		s.logger.Error(ctx, "failed when encrypting merchant callback api key", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	// Update merchant callback api key
	err = s.repo.UpdateCallbackApiKey(ctx, id, wrapped.Ciphertext, wrapped.KeyVersion)
	if err != nil {
		return nil, errors.New(responseHttp.HttpErrInternal, err)
	}
	return &callbackApiKey, nil
}

func (s *MerchantService) GetOrGenerateJITApiKey(ctx context.Context, id string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetOrGenerateJITApiKey")
	defer segment.End()

	merchant, err := s.repo.FindMerchantByID(ctx, id)
	if err != nil {
		return "", errors.New(responseHttp.HttpErrInternal, constant.ErrFindMerchant)

	} else if merchant == nil {
		return "", errors.New(responseHttp.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	if merchant.JITApiKey != "" {
		if merchant.JITApiKeyVersion > 0 {
			unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: merchant.JITApiKey})
			if err != nil {
				s.logger.Error(ctx, "failed while decrypting merchant jit api key", logger.Error(err))
				return "", errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
			}
			merchant.JITApiKey = string(unwrapped.Plaintext)
		}
		return merchant.JITApiKey, nil
	}

	jitApiKey, _ := util.GenerateRandomString(32)

	wrapped, err := s.encryption.Encrypt(ctx, vault.EncryptRequest{Plaintext: []byte(jitApiKey)})
	if err != nil {
		s.logger.Error(ctx, "failed when encrypting merchant jit api key", logger.Error(err))
		return "", errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	merchant.JITApiKey = wrapped.Ciphertext
	merchant.JITApiKeyVersion = wrapped.KeyVersion

	if err = s.repo.Update(ctx, merchant); err != nil {
		return "", errors.New(responseHttp.HttpErrInternal, constant.ErrUpdateMerchant)
	}
	return jitApiKey, nil
}
