package paymentService

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) GetEncryptionKey(ctx context.Context) (*model.GetEncryptionKeyResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetEncryptionKey")
	defer segment.End()

	encryptionKey, err := s.redis.Get(ctx, constant.PaymentEncryptionKeyCacheKey).Result()
	if err != nil && !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "Failed to get encryption key from cache", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	if encryptionKey != "" {
		return &model.GetEncryptionKeyResponse{EncryptionKey: encryptionKey}, nil
	}

	publicKeyPem, err := s.secretManager.GetSecretKeyString(ctx, "public_key")
	if err != nil {
		s.logger.Error(ctx, "Failed to get public key from secret manager", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	encryptionKeyBytes, err := s.cryptoAesGcm.Encrypt(publicKeyPem.Value)
	if err != nil {
		s.logger.Error(ctx, "Failed to encrypt public key with crypto aes-gcm", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	encryptionKey = base64.StdEncoding.EncodeToString(encryptionKeyBytes)

	if err = s.redis.Set(ctx, constant.PaymentEncryptionKeyCacheKey, encryptionKey, constant.PaymentEncryptionKeyTTL).Err(); err != nil {
		s.logger.Warn(ctx, "Failed to store encryption key to cache", logger.Error(err))
	}
	return &model.GetEncryptionKeyResponse{EncryptionKey: encryptionKey}, nil
}

func (s *PaymentService) DecryptRequest(ctx context.Context, data *encryption.DataEncryption, dst any) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/DecryptRequest")
	defer segment.End()

	privateKeyPem, err := s.secretManager.GetSecretKeyString(ctx, "private_key")
	if err != nil {
		s.logger.Error(ctx, "Failed to get private key from secret manager", logger.Error(err))
		return pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
	}
	data.PrivateKeyPEM = privateKeyPem.Value
	defer func() { privateKeyPem, data.PrivateKeyPEM = nil, "" }()

	plaintext, err := s.cryptoProvider.DecryptHybrid(data)
	if err != nil {
		s.logger.Warn(ctx, "Failed to decrypt request payload", logger.Error(err))
		return pkgErrs.New(response.HttpErrRequest, constant.ErrMalformedRequestBodyPayload)
	}

	if err := json.Unmarshal(plaintext, dst); err != nil {
		s.logger.Warn(ctx, "Failed to unmarshal plaintext to destination", logger.Error(err))
		return pkgErrs.New(response.HttpErrRequest, constant.ErrMalformedRequestBodyPayload)
	}
	return nil
}
