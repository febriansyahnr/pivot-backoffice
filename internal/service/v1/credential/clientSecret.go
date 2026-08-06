package credential

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	credModel "github.com/paper-indonesia/pivot-backoffice/internal/model/credential"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) ClientSecretById(ctx context.Context, request *credModel.ClientSecretReq) (resp *credModel.ClientSecretResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/credential/GetClientSecretById")
	defer segment.End()

	verifiedPIN, writeActivityLog := false, true
	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	defer func() {
		if !writeActivityLog {
			return
		}

		var activity string
		if request.Action == constant.ActionGet {
			activity = constant.ActivityUserViewClientSecretFailed

		} else if request.Action == constant.ActionPost {
			activity = constant.ActivityUserRegenerateClientSecretFailed
		}
		if verifiedPIN {
			if request.Action == constant.ActionGet {
				activity = constant.ActivityUserViewClientSecretSuccess

			} else if request.Action == constant.ActionPost {
				activity = constant.ActivityUserRegenerateClientSecretSuccess
			}
		}
		s.ActivityLog(ctx, &request.MerchantID, &request.UserID, request.Info, activity, map[string]string{"secret_id": request.SecretID})
	}()
	if err := s.userSvc.CheckCurrentPin(ctx, request.UserID, request.PIN); errors.Is(err, constant.ErrInvalidPIN) {
		return nil, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrInvalidPIN)

	} else if err != nil {
		writeActivityLog = false
		return nil, err
	}

	clientSecret := &credModel.ClientSecret{}
	verifiedPIN, writeActivityLog = true, false

	if request.Action == constant.ActionGet {
		if clientSecret, err = s.repo.GetClientSecretById(ctx, request.MerchantID, request.SecretID); err != nil {
			s.log.Error(ctx, "failed to get client secret by id", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceID))

		} else if clientSecret == nil {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
		}
		// Backward compatibility while the migration process is in progress
		if clientSecret.SecretVersion > 0 {
			unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: clientSecret.Secret})
			if err != nil {
				s.log.Error(ctx, "failed while decrypting merchant secret", logger.Error(err))
				return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
			}
			clientSecret.Secret = string(unwrapped.Plaintext)
		}

	} else if request.Action == constant.ActionPost {
		merchantSecret, _ := util.GenerateRandomString(40)
		wrappedSecret, err := s.encryption.Encrypt(ctx, vault.EncryptRequest{Plaintext: []byte(merchantSecret)})
		if err != nil {
			s.log.Error(ctx, "failed while encrypting merchant secret", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
		}

		clientSecret.Secret = wrappedSecret.Ciphertext
		clientSecret.SecretVersion = wrappedSecret.KeyVersion
		clientSecret.UpdatedAt = time.Now().UTC()
		if ok, err := s.repo.UpdateClientSecretById(ctx, request.MerchantID, request.SecretID, clientSecret); err != nil {
			s.log.Error(ctx, "failed to update client secret by id", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceID))

		} else if !ok {
			return nil, pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
		}
		clientSecret.Secret = merchantSecret
	}

	writeActivityLog = true

	return clientSecret.ToResponse(), nil
}
