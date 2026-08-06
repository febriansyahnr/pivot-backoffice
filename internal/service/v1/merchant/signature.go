package merchant

import (
	"context"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	snapSignature "github.com/paper-indonesia/pdk/go/snap/signature"
)

func (s *MerchantService) GenOpenAPISignature(ctx context.Context, req *merchant.GenSignatureReq) (string, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GenOpenAPISignature")
	defer segment.End()

	privateKey, err := util.RSAPrivateKey(req.PrivateKey)
	if err != nil {
		return "", pkgErrs.New(responseHttp.HttpErrUnprocessableContent, err)
	}

	signe := snapSignature.NewB2bTokenSignature(snapSignature.B2bTokenSignature{
		ClientID:   req.MerchantId,
		Timestamp:  req.Timestamp,
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	})

	signature, err := signe.Create()
	if err != nil {
		return "", pkgErrs.New(responseHttp.HttpErrUnprocessableContent, err)
	}
	return *signature, nil
}

func (s *MerchantService) ValidateSnapRequestSignature(ctx context.Context, req *merchant.ValidateSnapSignatureRequest) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/ValidateSnapRequestSignature")
	defer segment.End()

	merchantAuth, err := s.repo.GetMerchantSnapPKCS8KeyByMerchantID(ctx, req.ClientID)
	if err != nil {
		return pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrValidateRequestSignature)
	}

	if merchantAuth.SecretVersion > 0 {
		unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: merchantAuth.Secret})
		if err != nil {
			return pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
		}
		merchantAuth.Secret = string(unwrapped.Plaintext)
	}

	body, err := json.Marshal(req.Body)
	if err != nil {
		return pkgErrs.New(responseHttp.HttpErrRequest, constant.ErrMalformedRequestBodyPayload)
	}

	trxSignatureRequest := snapSignature.TrxSignature{
		HttpMethod:   req.Method,
		URL:          req.Url,
		BodyPayload:  body,
		AccessToken:  req.AccessToken,
		Timestamp:    req.Timestamp,
		ClientSecret: merchantAuth.Secret,
	}
	_, err = trxSignatureRequest.Verify(req.Signature)
	if err != nil {
		return pkgErrs.New(responseHttp.HttpErrUnprocessableContent, err)
	}

	return nil
}

func (s *MerchantService) GenerateSnapRequestSignature(ctx context.Context, req *merchant.GenerateSnapSignatureRequest) (string, error) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GenerateSnapRequestSignature")
	defer segment.End()

	merchantAuth, err := s.repo.GetMerchantSnapPKCS8KeyByMerchantID(ctx, req.ClientID)
	if err != nil {
		return "", pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrGenerateRequestSignature)
	}

	if merchantAuth.SecretVersion > 0 {
		unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: merchantAuth.Secret})
		if err != nil {
			return "", pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
		}
		merchantAuth.Secret = string(unwrapped.Plaintext)
	}

	body, err := json.Marshal(req.Body)
	if err != nil {
		return "", pkgErrs.New(responseHttp.HttpErrRequest, constant.ErrMalformedRequestBodyPayload)
	}

	trxSignatureRequest := snapSignature.TrxSignature{
		HttpMethod:   req.Method,
		URL:          req.Url,
		BodyPayload:  body,
		AccessToken:  req.AccessToken,
		Timestamp:    req.Timestamp,
		ClientSecret: merchantAuth.Secret,
	}
	signature, err := trxSignatureRequest.Create()
	if err != nil {
		return "", pkgErrs.New(responseHttp.HttpErrUnprocessableContent, err)
	}

	return *signature, nil
}
