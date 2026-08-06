package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) CreateMerchantAuth(ctx context.Context, merchantID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/CreateMerchantAuth")
	defer segment.End()

	// generate public and private key
	privKey, err := s.cryptoExt.GenerateRandomPKCS8Key()
	if err != nil {
		s.logger.Error(ctx, "error when generate public and private key", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, constant.ErrGenerateMerchantAuth)
	}

	merchantAuthUUID := uuid.New()

	secretKey := s.cryptoExt.SecretKeyFromUUID(merchantAuthUUID)
	encPriv, err := s.cryptoExt.Encrypt(string(privKey), secretKey)
	if err != nil {
		s.logger.Error(ctx, "error when encrypt key", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, constant.ErrGenerateMerchantAuth)
	}

	merchantSecret, _ := util.GenerateRandomString(40)

	wrappedSecret, err := s.encryption.Encrypt(ctx, vault.EncryptRequest{Plaintext: []byte(merchantSecret)})
	if err != nil {
		s.logger.Error(ctx, "failed while encrypting merchant secret", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	merchantAuth := merchant.NewMerchantAuth(&merchant.NewMerchantAuthRequest{
		ID:                  merchantAuthUUID.String(),
		MerchantID:          merchantID,
		Secret:              wrappedSecret.Ciphertext,
		SecretVersion:       wrappedSecret.KeyVersion,
		SnapPrivateKey:      encPriv,
		SnapPrivateKeyValid: true,
	})

	// skip on error
	err = s.repo.CreateMerchantAuth(ctx, merchantAuth)
	if err != nil {
		s.logger.Error(ctx, "error when create merchant auth", logger.Error(err), logger.Any("request", merchantAuth))
	}

	return nil
}
