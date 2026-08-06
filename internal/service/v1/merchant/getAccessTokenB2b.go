package merchant

import (
	"context"
	errL "errors"
	"fmt"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/go/snap"
	snapSignature "github.com/paper-indonesia/pdk/go/snap/signature"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) GetAccessTokenB2b(ctx context.Context, clientID, clientSecret string) (*string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetAccessTokenB2b")
	defer segment.End()

	merchantAuth, err := s.repo.GetMerchantAuthByMerchantId(ctx, clientID)
	if err != nil {
		return nil, errors.New(responseHttp.HttpErrDatabase, err)
	}

	if merchantAuth != nil && merchantAuth.SecretVersion > 0 {
		unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: merchantAuth.Secret})
		if err != nil {
			s.logger.Error(ctx, "failed while decrypting merchant authentication", logger.Error(err))
			return nil, errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
		}
		merchantAuth.Secret = string(unwrapped.Plaintext)
	}

	if merchantAuth == nil || merchantAuth.Secret != clientSecret {
		return nil, errors.New(responseHttp.HttpErrUnauthorized, fmt.Errorf("client id %s not found", clientID))
	}

	merchant, err := s.FindMerchantByID(ctx, merchantAuth.MerchantID)
	if err != nil {
		return nil, errors.New(responseHttp.HttpErrDatabase, err)
	} else if merchant == nil {
		return nil, errors.New(responseHttp.HttpErrUnauthorized, constant.ErrMerchantNotFound)
	}

	if err = s.validateMerchantStatus(merchant.Status); err != nil {
		return nil, err
	}

	// generate access token
	accessToken, err := s.JWT.GenerateMerchantToken(ctx, clientID, merchantAuth.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to generate access token b2b", logger.Error(err))
		return nil, errors.New(responseHttp.HttpErrInternal, err)
	}

	return &accessToken, nil
}

func (s *MerchantService) GetSNAPAccessTokenB2B(ctx context.Context, request *merchantModel.SNAPAccessTokenB2BReq) (*merchantModel.SNAPAccessTokenB2BResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetSNAPAccessTokenB2B")
	defer segment.End()

	if err := s.ValidateSNAPAccessTokenRequestSignature(ctx, &merchantModel.SNAPValidateB2b2cTokenSignatureRequest{
		ClientId:  request.ClientId,
		Timestamp: request.Timestamp,
		Signature: request.Signature,
	}); err != nil {
		return nil, err
	}

	merchant, err := s.FindMerchantByID(ctx, request.ClientId)
	if err != nil {
		return nil, err
	} else if merchant == nil {
		return nil, errors.New(responseHttp.HttpErrUnauthorized, constant.ErrMerchantNotFound)
	}

	if err = s.validateMerchantStatus(merchant.Status); err != nil {
		return nil, err
	}

	accessToken, err := s.JWT.GenerateMerchantToken(ctx, request.ClientId, request.ClientId)
	if err != nil {
		s.logger.Error(ctx, "Open Api Snap | Generate access token B2B", logger.Error(err))
		return nil, err
	}

	code, msg := snap.GenerateResponseCode(snap.SNAP_SUCCESS, snap.SNAP_SERVICE_B2B)
	return &merchantModel.SNAPAccessTokenB2BResp{
		ResponseCode:    code,
		ResponseMessage: msg,
		AccessToken:     accessToken,
		TokenType:       "Bearer",
		ExpiresIn:       strconv.FormatFloat(constant.MerchantAuthExpirationDuration.Seconds(), 'f', -1, 64),
	}, nil
}

func (s *MerchantService) ValidateSNAPAccessTokenRequestSignature(ctx context.Context, request *merchantModel.SNAPValidateB2b2cTokenSignatureRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/ValidateSNAPAccessTokenRequestSignature")
	defer segment.End()

	merchantAuth, err := s.repo.GetMerchantSnapPKCS8KeyByMerchantID(ctx, request.ClientId)

	if err != nil || merchantAuth.SnapPrivateKey.String == "" || merchantAuth.MerchantPublicKey.String == "" {
		if err == nil {
			s.logger.Warn(
				ctx, "Open Api Snap | Empty private key or public key", logger.Any("empty_secret", map[string]bool{
					"private_key": merchantAuth.SnapPrivateKey.String == "",
					"public_key":  merchantAuth.MerchantPublicKey.String == "",
				}))
		}
		return errors.New(responseHttp.HttpErrUnauthorized, fmt.Errorf(constant.UnauthorizedSnapFmt, constant.HeaderXClientKey))
	}

	merchantUUID, _ := uuid.Parse(merchantAuth.MerchantID)
	secretKey := s.cryptoExt.SecretKeyFromUUID(merchantUUID)

	rawPublicKey, err := s.cryptoExt.Decrypt(merchantAuth.MerchantPublicKey.String, secretKey)
	if err != nil {
		s.logger.Error(ctx, "Open Api Snap | Decrypt public key with secret", logger.Error(err))
		return err
	}

	publicKey, err := s.cryptoExt.ToPublicKey(*rawPublicKey)
	if err != nil {
		s.logger.Error(ctx, "Open Api Snap | Convert public key", logger.Error(err))
		return err
	}

	b2bSign := snapSignature.NewB2bTokenSignature(
		snapSignature.B2bTokenSignature{
			PublicKey: publicKey,
			ClientID:  merchantAuth.MerchantID,
			Timestamp: request.Timestamp,
			Signature: request.Signature,
		},
	)

	if !b2bSign.Verify(request.Signature) {
		s.logger.Warn(ctx, "Open Api Snap | Invalid signature", logger.String("signature", request.Signature))
		return errors.New(responseHttp.HttpErrUnauthorized, fmt.Errorf(constant.UnauthorizedSnapFmt, constant.HeaderXSignature))
	}

	return nil
}

func (s *MerchantService) validateMerchantStatus(merchantStatus string) error {
	switch merchantStatus {
	case constant.MerchantStatusClosed:
		return errors.New(responseHttp.HttpErrUnauthorized, errL.New("merchant status is closed"))
	case constant.MerchantStatusBlocked:
		return errors.New(responseHttp.HttpErrUnauthorized, errL.New("merchant is blocked"))
	case constant.MerchantStatusDeactivated:
		return errors.New(responseHttp.HttpErrUnauthorized, errL.New("merchant is deactivated"))
	case constant.MerchantStatusCreated:
		// Will be removed when we have merchant onboarding process feature
		return errors.New(responseHttp.HttpErrUnauthorized, errL.New("merchant is not activated yet"))
	}

	return nil
}
