package merchant

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/google/uuid"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) SetMerchantPublicKey(ctx context.Context, merchantId string, publicKey string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/SetMerchantPublicKey")
	defer segment.End()

	// get merchant auth by id
	merchantAuth, err := s.repo.GetMerchantSnapPKCS8KeyByMerchantID(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "failed to find secret key", logger.Error(err))
		return err
	}

	// remove hyphens
	merchantUUID, err := uuid.Parse(merchantId)
	if err != nil {
		s.logger.Error(ctx, "failed to parse merchant id to uuid", logger.Error(err))
		return err
	}

	secretKey := s.cryptoExt.SecretKeyFromUUID(merchantUUID)
	// decrypt the secret key
	decryptedKey, err := s.cryptoExt.Decrypt(publicKey, secretKey)
	if err != nil {
		s.logger.Error(ctx, "failed to decrypt key", logger.Error(err))
		return err
	}

	if !s.isValidPublicKey(*decryptedKey) {
		return errors.New(responseHttp.HttpErrForbidden, fmt.Errorf("invalid public key"))
	}

	merchantAuth.MerchantPublicKey = sql.NullString{
		String: publicKey,
		Valid:  true,
	}

	merchantAuth.UpdatedAt = time.Now()

	if err = s.repo.UpdateMerchantAuth(ctx, merchantAuth); err != nil {
		s.logger.Error(ctx, "failed to update merchant auth", logger.Error(err))
		return err
	}

	return nil
}

func (s *MerchantService) isValidPublicKey(publicKey []byte) bool {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		s.logger.Error(context.Background(), "failed to decode pem block")
		return false
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		s.logger.Error(context.Background(), "failed to parse public key", logger.Error(err))
		return false
	}

	length := pub.(*rsa.PublicKey).N.BitLen()
	return length == 2048
}

// simple utility function to encrypting data(will be deprecated soon :D )
func (s *MerchantService) UtilEncryptingKey(ctx context.Context, key string, data string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UtilEncryptingKey")
	defer segment.End()

	var keyParam string

	keyUUID, err := uuid.Parse(key)
	if err == nil {
		keyParam = s.cryptoExt.SecretKeyFromUUID(keyUUID)

	} else if len(key) == 32 {
		keyParam = key

	} else {
		err := errors.New(responseHttp.HttpErrForbidden, fmt.Errorf("invalid key"))
		s.logger.Error(ctx, "invalid key")
		return "", err
	}

	encrypted, err := s.cryptoExt.Encrypt(data, keyParam)
	if err != nil {
		s.logger.Error(ctx, "failed to encrypt data", logger.Error(err))
		return "", err
	}

	return encrypted, nil
}
