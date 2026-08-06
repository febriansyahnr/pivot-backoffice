package merchant

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) CreatePKCS8SecretKey(ctx context.Context, merchantID string) (*merchant.PKCS8SecretKeyResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/CreatePKCS8SecretKey")
	defer segment.End()

	// get merchant auth
	merchantAuth, err := s.repo.GetMerchantSnapPKCS8KeyByMerchantID(ctx, merchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to find secret key", logger.Error(err))
		return nil, pkgErrors.New(responseHttp.HttpErrNotFound, err)
	}

	// generate pkcs8 secret key
	privKey, err := s.cryptoExt.GenerateRandomPKCS8Key()
	if err != nil {
		s.logger.Error(ctx, "failed to generate pkcs8 secret key", logger.Error(err))
		return nil, pkgErrors.New(responseHttp.HttpErrInternal, err)
	}

	// encrypt the key
	merchantUUID, err := uuid.Parse(merchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to parse merchant id to uuid", logger.Error(err))
		return nil, pkgErrors.New(responseHttp.HttpErrInternal, err)
	}

	encPriv, err := s.cryptoExt.Encrypt(string(privKey), s.cryptoExt.SecretKeyFromUUID(merchantUUID))
	if err != nil {
		s.logger.Error(ctx, "failed to encrypt key", logger.Error(err))
		return nil, pkgErrors.New(responseHttp.HttpErrInternal, err)
	}

	merchantAuth.SnapPrivateKey = sql.NullString{
		String: encPriv,
		Valid:  true,
	}

	// update database
	if err := s.repo.UpdateMerchantAuth(ctx, merchantAuth); err != nil {
		s.logger.Error(ctx, "failed to update secret key", logger.Error(err))
		return nil, pkgErrors.New(responseHttp.HttpErrInternal, err)
	}

	resp := &merchant.PKCS8SecretKeyResponse{
		MerchantID: merchantID,
	}

	return resp, nil
}
