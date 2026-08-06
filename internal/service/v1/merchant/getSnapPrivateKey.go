package merchant

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const (
	PRIVPKCS1 = "RSA PRIVATE KEY"
	PRIVPKCS8 = "PRIVATE KEY"

	PUBPKCS1 = "RSA PUBLIC KEY"
	PUBPKCS8 = "PUBLIC KEY"
)

func (s *MerchantService) GetSnapPrivateKey(ctx context.Context, merchantId string) (*rsa.PrivateKey, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetSnapPrivateKey")
	defer segment.End()

	// get merchant auth by id
	merchantAuth, err := s.repo.GetMerchantSnapPKCS8KeyByMerchantID(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "failed to find secret key", logger.Error(err))
		return nil, err
	}

	// remove hyphens
	merchantUUID, err := uuid.Parse(merchantId)
	if err != nil {
		s.logger.Error(ctx, "failed to parse merchant id to uuid", logger.Error(err))
		return nil, err
	}

	secretKey := s.cryptoExt.SecretKeyFromUUID(merchantUUID)
	// decrypt the secret key
	decryptedKey, err := s.cryptoExt.Decrypt(merchantAuth.SnapPrivateKey.String, secretKey)
	if err != nil {
		s.logger.Error(ctx, "failed to decrypt key", logger.Error(err))
		return nil, err
	}
	if decryptedKey == nil {
		return nil, fmt.Errorf("decrypted key is empty")
	}

	privateKey, err := parsePrivateKey(*decryptedKey)
	if err != nil {
		s.logger.Error(ctx, "failed to parse private key", logger.Error(err))
		return nil, err
	}
	return privateKey, nil
}

func parsePrivateKey(decryptedPrivateKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(decryptedPrivateKey)
	switch block.Type {
	case PRIVPKCS8:
		p, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return p.(*rsa.PrivateKey), err
	case PRIVPKCS1:
		p, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, fmt.Errorf("private key not found")
	}
}
