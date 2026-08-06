package callbackService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	pdkSnapHeader "github.com/paper-indonesia/pdk/go/snap/header"
	pdkSnapSign "github.com/paper-indonesia/pdk/go/snap/signature"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *CallbackService) buildSnapRequestHeaders(ctx context.Context, event string, callback *callbackModel.Callback, request any) (headers *pdkSnapHeader.Header, err error) {

	urlPath, err := getUrlPath(callback.URL)
	if err != nil {
		s.logger.Error(ctx, "error when get url path", logger.Error(err))
		return nil, err
	}

	callbackDetails, err := s.callbackRepo.FindCallbackByNameAndMerchantID(ctx, constant.CallbackMasterPaymentSNAPAccessTokenB2b, callback.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error when find callback token", logger.Error(err))
		return nil, err

	} else if callbackDetails == nil {
		return nil, errors.New("callback token not found")
	}

	val, _, err := s.getSnapAccessTokenB2bNew(ctx, callbackDetails, event)
	if err != nil {
		s.logger.Error(ctx, "error when get snap access token b2b", logger.Error(err))
		return nil, err

	} else if val == nil || *val == "" {
		return nil, errors.New("invalid snap access token")
	}

	merchantAuth, err := s.merchantRepo.GetMerchantSnapPKCS8KeyByMerchantID(ctx, callback.MerchantID.String())
	if err != nil {
		s.logger.Error(ctx, "error when try to get merchant auth", logger.Error(err))
		return nil, err

	} else if merchantAuth == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, errors.New("merchant auth is not found"))
	}

	// Decrypt merchant secret
	if merchantAuth.SecretVersion > 0 {
		unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: merchantAuth.Secret})
		if err != nil {
			s.logger.Error(ctx, "failed while decrypting merchant secret", logger.Error(err))
			return nil, pkgErr.New(response.HttpErrInternal, err)
		}
		merchantAuth.Secret = string(unwrapped.Plaintext)
	}

	// Headers Request
	headers = &pdkSnapHeader.Header{
		ContentType:   constant.ContentTypeApplicationJson,
		Authorization: fmt.Sprintf("Bearer %s", *val),
		XTimestamp:    util.SnapCompatible(time.Now()),
		XPartnerID:    callback.MerchantID.String(),
		ChannelID:     constant.BuildSnapChannelIdByEvent(event),
		XCLientKey:    callback.MerchantID.String(),
	}
	headers.XExternalID = headers.GenerateExternalID()

	var requestBytes []byte

	switch v := request.(type) {
	default:
		requestBytes, _ = json.Marshal(v)

	case []byte:
		requestBytes = v

	case string:
		requestBytes = []byte(v)
	}

	signature, _ := pdkSnapSign.NewTrxSignature(
		pdkSnapSign.TrxSignature{
			HttpMethod:   http.MethodPost,
			URL:          urlPath,
			AccessToken:  fmt.Sprintf("Bearer %s", *val),
			ClientSecret: merchantAuth.Secret,
			Timestamp:    headers.XTimestamp,
			BodyPayload:  requestBytes,
		},
	)
	snapSignatureValue, _ := signature.Create()

	headers.XSignature = *snapSignatureValue

	return headers, nil
}
