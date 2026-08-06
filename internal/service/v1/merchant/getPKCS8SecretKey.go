package merchant

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) GetPKCS8SecretKey(ctx context.Context, merchantID string) (*merchant.PKCS8SecretKeyResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetPKCS8SecretKey")
	defer segment.End()

	merchantAuth, err := s.repo.GetMerchantSnapPKCS8KeyByMerchantID(ctx, merchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to find secret key", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrNotFound, err)
	}

	if merchantAuth.SecretVersion > 0 {
		unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: merchantAuth.Secret})
		if err != nil {
			s.logger.Error(ctx, "failed while decrypting merchant secret", logger.Error(err))
			return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
		}
		merchantAuth.Secret = string(unwrapped.Plaintext)
	}

	dataB, err := json.Marshal(
		merchant.PKCS8SecretKeyResponse{
			MerchantID:        merchantID,
			MerchantSecret:    merchantAuth.Secret,
			SnapPrivateKey:    merchantAuth.SnapPrivateKey.String,
			MerchantPublicKey: merchantAuth.MerchantPublicKey.String,
		})
	if err != nil {
		s.logger.Error(ctx, "failed to marshal data", logger.Error(err))
		return nil, err
	}

	response := &merchant.PKCS8SecretKeyResponse{
		MerchantID: merchantID,
		Data:       base64.StdEncoding.EncodeToString(dataB),
	}

	return response, nil
}
