package callbackService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	callbackPkgModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *CallbackService) getMerchantCallbackApiKey(ctx context.Context, merchantId string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/getMerchantCallbackApiKey")
	defer segment.End()

	callbackKey, err := s.callbackRepo.GetCallbackAPIKeyByMerchantId(ctx, merchantId)
	if err != nil {
		s.logger.Error(ctx, "failed to get callback api key", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if strings.TrimSpace(callbackKey.APIKey) == "" {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("empty callback api key"))
	}

	if callbackKey.Version > 0 {
		unwrapped, err := s.encryption.Decrypt(ctx, vault.DecryptRequest{Ciphertext: callbackKey.APIKey})
		if err != nil {
			s.logger.Error(ctx, "failed while decrypting merchant callback api key", logger.Error(err))
			return "", pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
		}
		callbackKey.APIKey = string(unwrapped.Plaintext)
	}
	return callbackKey.APIKey, nil
}

func (s *CallbackService) GetCallbackURLByMerchantId(ctx context.Context, request *callbackModel.CallbackURLSettingReq) (interface{}, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/GetCallbackURLByMerchantId")
	defer segment.End()
	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	resp, err := s.callbackRepo.GetCallbackURLByMerchantId(ctx, request.MerchantID, request.MasterName)
	if err != nil {
		s.logger.Error(ctx, "failed to get callback url by merchant id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceID))
	}

	for i := range resp {
		switch resp[i].MasterName {
		case constant.CallbackMasterPaymentSNAPAccessTokenB2b:
			resp[i].CallbackTemplate = callbackPkgModel.CallbackPayloadRequest{
				Event: constant.BuildCallbackSnapStatusTestByName(resp[i].MasterName),
				Data: map[string]interface{}{
					"grantType":      "client_credentials",
					"additionalInfo": map[string]interface{}{},
				},
			}
		case constant.CallbackMasterPaymentSNAPQRIS:
			resp[i].CallbackTemplate = callbackPkgModel.CallbackPayloadRequest{
				Event: constant.BuildCallbackSnapStatusTestByName(resp[i].MasterName),
				Data: map[string]interface{}{
					"originalReferenceNo":        "18039112312312",
					"originalPartnerReferenceNo": "18039112312313",
					"latestTransactionStatus":    "00",
					"transactionStatusDesc":      "Success",
					"amount": map[string]interface{}{
						"value":    "10000.00",
						"currency": "IDR",
					},
					"additionalInfo": map[string]interface{}{
						"RRN":             "18039112312314",
						"qrType":          "DYNAMIC",
						"qrStatus":        "INACTIVE",
						"qrExpiredDate":   "2024-10-05T10:00:00+07:00",
						"merchantName":    "HARSYA",
						"paymentStatus":   "SUCCESS",
						"transactionDate": "2024-10-04T10:00:00+07:00",
					},
				},
			}
		case constant.CallbackMasterPaymentSNAPVA:
			resp[i].CallbackTemplate = callbackPkgModel.CallbackPayloadRequest{
				Event: constant.BuildCallbackSnapStatusTestByName(resp[i].MasterName),
				Data: map[string]interface{}{
					"trxId":              "18039112312314",
					"virtualAccountNo":   "7663180391123123",
					"virtualAccountName": "HARSYA",
					"paidAmount": map[string]string{
						"value":    "10000.00",
						"currency": "IDR",
					},
					"totalAmount": map[string]string{
						"value":    "10000.00",
						"currency": "IDR",
					},
					"trxDateTime": "2024-10-04T10:00:00+07:00",
					"additionalInfo": map[string]interface{}{
						"referenceId":           "18039112312318",
						"issuer":                "PERMATA",
						"virtualAccountTrxType": "CLOSED_DYNAMIC",
						"expiredDate":           "2024-10-05T10:00:00+07:00",
						"minAmount": map[string]string{
							"value":    "0",
							"currency": "IDR",
						},
						"maxAmount": map[string]string{
							"value":    "25000000",
							"currency": "IDR",
						},
						"vaStatus":      "ACTIVE",
						"paymentStatus": "SUCCESS",
					},
				},
			}
		case constant.CallbackMasterWalletSNAPQrisMPM:
			resp[i].CallbackTemplate = callbackPkgModel.CallbackPayloadRequest{
				Event: constant.BuildCallbackSnapStatusTestByName(resp[i].MasterName),
				Data: map[string]interface{}{
					"originalReferenceNo":        "18039112312312",
					"originalPartnerReferenceNo": "18039112312313",
					"latestTransactionStatus":    "00",
					"transactionStatusDesc":      "Success",
					"amount": map[string]interface{}{
						"value":    "10000.00",
						"currency": "IDR",
					},
					"additionalInfo": map[string]interface{}{
						"qrisId":            "550e8400-e29b-41d4-a716-446655440004",
						"acquirer":          "BANKNEOCOMMERCE",
						"customerPan":       "93600000000000001234",
						"acquirerReferenceNo": "18039112312314",
						"tipAmount":         0,
						"qrisData": map[string]interface{}{
							"mode":            "DYNAMIC",
							"authorizationId": "AUTH123456",
							"dateTime":        "2024-10-04T10:00:00+07:00",
							"issuerNns":       "93600",
							"nationalMid":     "ID1234567890123",
							"terminalId":      "A01",
							"merchantData": map[string]interface{}{
								"id":          "MERCHANT001",
								"name":        "HARSYA",
								"pan":         "936000000000001",
								"criteria":    "UMI",
								"city":        "JAKARTA",
								"mcc":         "5812",
								"postalCode":  "12345",
								"countryCode": "ID",
							},
						},
					},
				},
			}
		case constant.CallbackMasterWalletSNAPDirectDebit:
			resp[i].CallbackTemplate = callbackPkgModel.CallbackPayloadRequest{
				Event: constant.BuildCallbackSnapStatusTestByName(resp[i].MasterName),
				Data: map[string]interface{}{
					"originalReferenceNo":        "18039112312312",
					"originalPartnerReferenceNo": "18039112312313",
					"latestTransactionStatus":    "00",
					"transactionStatusDesc":      "Success",
					"amount": map[string]interface{}{
						"value":    "10000.00",
						"currency": "IDR",
					},
					"additionalInfo": map[string]interface{}{
						"merchantId":      "550e8400-e29b-41d4-a716-446655440001",
						"merchantName":    "HARSYA",
						"senderId":        "550e8400-e29b-41d4-a716-446655440002",
						"senderName":      "John Doe",
						"senderBindingId": "550e8400-e29b-41d4-a716-446655440003",
					},
				},
			}
		default:
			resp[i].CallbackTemplate = callbackPkgModel.CallbackPayloadRequest{
				Event: strings.ReplaceAll(resp[i].MasterName, " ", ".") + ".TEST",
				Data:  map[string]string{"test": "OK"},
			}
		}
	}

	s.ActivityLog(ctx, &request.MerchantID, &request.UserID, request.Info, constant.ActivityUserAccessCallbackDashboard, nil)

	return resp, nil
}

func (s *CallbackService) GetCallbackAPIKeyByMerchantId(ctx context.Context, request *callbackModel.CallbackURLSettingReq) (interface{}, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/GetCallbackAPIKeyByMerchantId")
	defer segment.End()

	writeActivity, verifiedPIN := false, false
	defer func() {
		if !writeActivity {
			return
		}

		activity := constant.ActivityUserViewCallbackAPIKeyFailed
		if verifiedPIN {
			activity = constant.ActivityUserViewCallbackAPIKeySuccess
		}
		s.ActivityLog(ctx, &request.MerchantID, &request.UserID, request.Info, activity, nil)
	}()
	pin, _ := ctx.Value(constant.CtxUserPINKey).(string)

	if err := s.userSvc.CheckCurrentPin(ctx, request.UserID, pin); err != nil {
		writeActivity = errors.Is(err, constant.ErrInvalidPIN)
		return nil, err
	}

	apiKey, err := s.getMerchantCallbackApiKey(ctx, request.MerchantID)
	if err != nil {
		return nil, err
	}
	writeActivity, verifiedPIN = true, true

	return &callbackModel.CallbackAPIKeyResp{APIKey: apiKey}, nil
}

func (s *CallbackService) TestAndSaveCallbackURL(ctx context.Context, request *callbackModel.TestAndSaveCallbackURLReq) (resp *callbackModel.TestAndSaveCallbackURLResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/TestAndSaveCallbackURL")
	defer segment.End()

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	apiKey, err := s.getMerchantCallbackApiKey(ctx, request.MerchantID)
	if err != nil {
		return nil, err
	}
	callbackRequest := callbackPkgModel.CallbackRequest{
		URL:     request.URL,
		Request: request.Payload,
	}
	headers := map[string]string{
		constant.HeaderXAPIKey: apiKey,
	}

	isCompleted, now := false, time.Now().UTC()

	result, errCallback := s.callbackPartnerSvc.Callback(ctx, callbackRequest, headers)
	if errCallback != nil && strings.TrimSpace(result) == "" {
		return nil, errCallback
	}
	defer func() {
		activity := constant.ActivityUserTestCallbackURLFailed
		if errCallback == nil && !isCompleted {
			activity = constant.ActivityUserSaveCallbackURLFailed

		} else if errCallback == nil && isCompleted {
			activity = constant.ActivityUserTestAndSaveCallbackURLSuccess
		}
		buf, _ := json.Marshal(request.Payload)
		params := map[string]string{
			"callback_url":  request.URL,
			"callback_body": string(buf),
		}
		s.ActivityLog(ctx, &request.MerchantID, &request.UserID, request.Info, activity, params)
	}()

	resp = &callbackModel.TestAndSaveCallbackURLResp{
		Status:    errCallback == nil,
		RequestID: traceID,
		Duration:  time.Now().UTC().Sub(now).String(),
		Information: callbackModel.CallbackURLInfoResp{
			Product:    request.Name,
			Event:      request.Payload.Event,
			URL:        request.URL,
			Time:       now,
			CallbackID: "-", CallbackLogID: "-",
			CallbackType: "-", CallbackLength: "-",
		},
	}
	_ = json.Unmarshal([]byte(result), &resp.Body)

	resp.Information.CallbackToken = apiKey[len(apiKey)-4:]
	for i := len(apiKey) - 5; i >= 0; i-- {
		resp.Information.CallbackToken = "*" + resp.Information.CallbackToken
	}
	if errCallback != nil {
		return resp, nil
	}

	ctxTx, err := s.callbackRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to begin transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("TX: "+constant.InternalErrorFmt, traceID))
	}
	defer func() {
		if !isCompleted {
			if e := s.callbackRepo.RollbackTransaction(ctxTx); e != nil {
				s.logger.Error(ctx, "failed to rollback transaction", logger.Error(err))
				resp, err = nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("RT: "+constant.InternalErrorFmt, traceID))
			}
		}
	}()

	callbackID, err := s.callbackRepo.GetCallbackIdByMerchantAndMasterCallbackId(ctxTx, request.MerchantID, request.CallbackMasterID)
	if err != nil {
		s.logger.Error(ctx, "failed to get callback id by merchant and callback master id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("GET: "+constant.InternalErrorFmt, traceID))

	} else if callbackID != "" {
		if err := s.callbackRepo.UpdateCallbackURLById(ctxTx, callbackID, request.URL); err != nil {
			s.logger.Error(ctx, "failed to update callback url", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("UPDATE: "+constant.InternalErrorFmt, traceID))
		}

	} else if callbackID == "" {
		data := &callbackModel.Callback{
			UUID:      uuid.New(),
			URL:       request.URL,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		data.MerchantID, _ = uuid.Parse(request.MerchantID)
		data.CallbackMasterID, _ = uuid.Parse(request.CallbackMasterID)
		if err = s.callbackRepo.CreateCallback(ctxTx, data); err != nil {
			s.logger.Error(ctx, "failed to create callback url", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("CREATE: "+constant.InternalErrorFmt, traceID))
		}
		callbackID = data.UUID.String()
	}

	requestBody, _ := json.Marshal(request.Payload.Data)
	referenceId := callbackModel.GetReferenceId(request.Payload.Event, request.Payload.Data)
	callbackLog := &callbackModel.CallbackLog{
		UUID:        uuid.New(),
		Event:       &request.Payload.Event,
		Request:     string(requestBody),
		Response:    &result,
		Status:      constant.CallbackStatusDelivered,
		ReferenceId: referenceId,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	callbackLog.CallbackID, _ = uuid.Parse(callbackID)
	if err = s.callbackRepo.CreateCallbackLog(ctxTx, callbackLog); err != nil {
		s.logger.Error(ctx, "failed to create callback log", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("LOG: "+constant.InternalErrorFmt, traceID))
	}

	if err = s.callbackRepo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "failed to commit transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("CMT: "+constant.InternalErrorFmt, traceID))
	}

	isCompleted = true
	resp.Information.CallbackID = callbackID
	resp.Information.CallbackLogID = callbackLog.UUID.String()
	return
}

func (s *CallbackService) TestAndSaveB2b(ctx context.Context, request *callbackModel.TestAndSaveCallbackURLReq) (resp *callbackModel.TestAndSaveCallbackURLResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/TestAndSaveCallbackURL")
	defer segment.End()

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	isNewCallback := false

	callbackID, err := s.callbackRepo.GetCallbackIdByMerchantAndMasterCallbackId(ctx, request.MerchantID, request.CallbackMasterID)
	if err != nil {
		s.logger.Error(ctx, "failed to get callback id by merchant and callback master id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("GET: "+constant.InternalErrorFmt, traceID))
	} else if callbackID == "" {
		callbackID = uuid.New().String()
		isNewCallback = true
	}

	data := &callbackModel.Callback{
		URL:       request.URL + constant.CallbackSnapAccessTokenB2bEndpointWithoutVersion,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		BaseURL:   &request.URL,
	}
	data.UUID, _ = uuid.Parse(callbackID)
	data.MerchantID, _ = uuid.Parse(request.MerchantID)
	data.CallbackMasterID, _ = uuid.Parse(request.CallbackMasterID)

	_, accessTokenBody, err := s.getSnapAccessTokenB2bNew(ctx, data, request.Payload.Event)
	if err != nil {
		s.logger.Error(ctx, "error when get snap access token b2b", logger.Error(err))
		return nil, err
	}
	isCompleted, now := false, time.Now().UTC()

	defer func() {
		activity := constant.ActivityUserTestCallbackURLFailed
		if err == nil && !isCompleted {
			activity = constant.ActivityUserSaveCallbackURLFailed

		} else if err == nil && isCompleted {
			activity = constant.ActivityUserTestAndSaveCallbackURLSuccess
		}
		buf, _ := json.Marshal(request.Payload)
		params := map[string]string{
			"callback_url":  data.URL,
			"callback_body": string(buf),
		}
		s.ActivityLog(ctx, &request.MerchantID, &request.UserID, request.Info, activity, params)
	}()

	resp = &callbackModel.TestAndSaveCallbackURLResp{
		Status:    err == nil,
		RequestID: traceID,
		Duration:  time.Now().UTC().Sub(now).String(),
		Body:      accessTokenBody,
		Information: callbackModel.CallbackURLInfoResp{
			Product:    request.Name,
			Event:      request.Payload.Event,
			URL:        data.URL,
			Time:       now,
			CallbackID: "-", CallbackLogID: "-",
			CallbackType: "-", CallbackLength: "-",
		},
	}

	if err != nil {
		return resp, nil
	}

	ctxTx, err := s.callbackRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to begin transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("TX: "+constant.InternalErrorFmt, traceID))
	}
	defer func() {
		if !isCompleted {
			if e := s.callbackRepo.RollbackTransaction(ctxTx); e != nil {
				s.logger.Error(ctx, "failed to rollback transaction", logger.Error(err))
				resp, err = nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("RT: "+constant.InternalErrorFmt, traceID))
			}
		}
	}()

	if !isNewCallback {
		if err := s.callbackRepo.UpdateCallbackURLById(ctxTx, callbackID, data.URL); err != nil {
			s.logger.Error(ctx, "failed to update callback url", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("UPDATE: "+constant.InternalErrorFmt, traceID))
		}
		if err := s.callbackRepo.UpdateCallbackBaseURLById(ctxTx, callbackID, *data.BaseURL); err != nil {
			s.logger.Error(ctx, "failed to update callback base url", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("UPDATE: "+constant.InternalErrorFmt, traceID))
		}

	} else if isNewCallback {
		if err = s.callbackRepo.CreateCallback(ctxTx, data); err != nil {
			s.logger.Error(ctx, "failed to create callback url", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("CREATE: "+constant.InternalErrorFmt, traceID))
		}
	}

	if err = s.callbackRepo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "failed to commit transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("CMT: "+constant.InternalErrorFmt, traceID))
	}

	resp.Information.CallbackID = callbackID
	isCompleted = true
	return
}

func (s *CallbackService) TestAndSaveSnapPayment(ctx context.Context, request *callbackModel.TestAndSaveCallbackURLReq) (resp *callbackModel.TestAndSaveCallbackURLResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/callback/TestAndSaveCallbackURL")
	defer segment.End()

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	isNewCallback := false

	callbackMaster, err := s.callbackRepo.FindCallbackMasterByName(
		ctx,
		request.Name)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase)
	}

	if request == nil || request.URL == "" {
		return nil, errors.New("invalid request URL")
	}

	if callbackMaster == nil || callbackMaster.Name == "" {
		return nil, errors.New("invalid callback master")
	}

	endpoint := constant.BuildCallbackSnapEndpointByName(callbackMaster.Name)
	if endpoint == "" {
		return nil, errors.New("invalid callback endpoint")
	}

	callbackID, err := s.callbackRepo.GetCallbackIdByMerchantAndMasterCallbackId(ctx, request.MerchantID, request.CallbackMasterID)
	if err != nil {
		s.logger.Error(ctx, "failed to get callback id by merchant and callback master id", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("GET: "+constant.InternalErrorFmt, traceID))
	} else if callbackID == "" {
		callbackID = uuid.New().String()
		isNewCallback = true
	}

	data := &callbackModel.Callback{
		URL:       request.URL + endpoint,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		BaseURL:   &request.URL,
	}
	data.UUID, _ = uuid.Parse(callbackID)
	data.MerchantID, _ = uuid.Parse(request.MerchantID)
	data.CallbackMasterID, _ = uuid.Parse(request.CallbackMasterID)

	snapResp, err := s.processSnapNew(ctx, request.Payload.Data, data, request.Payload.Event)
	if err != nil {
		s.logger.Error(ctx, "error when process snap", logger.Error(err))
		return nil, err
	}
	isCompleted, now := false, time.Now().UTC()
	defer func() {
		activity := constant.ActivityUserTestCallbackURLFailed
		if err == nil && !isCompleted {
			activity = constant.ActivityUserSaveCallbackURLFailed

		} else if err == nil && isCompleted {
			activity = constant.ActivityUserTestAndSaveCallbackURLSuccess
		}
		buf, _ := json.Marshal(request.Payload)
		params := map[string]string{
			"callback_url":  data.URL,
			"callback_body": string(buf),
		}
		s.ActivityLog(ctx, &request.MerchantID, &request.UserID, request.Info, activity, params)
	}()

	resp = &callbackModel.TestAndSaveCallbackURLResp{
		Status:    err == nil,
		RequestID: traceID,
		Duration:  time.Now().UTC().Sub(now).String(),
		Information: callbackModel.CallbackURLInfoResp{
			Product:    request.Name,
			Event:      request.Payload.Event,
			URL:        data.URL,
			Time:       now,
			CallbackID: "-", CallbackLogID: "-",
			CallbackType: "-", CallbackLength: "-",
		},
	}
	_ = json.Unmarshal([]byte(snapResp), &resp.Body)

	if err != nil {
		return resp, nil
	}

	ctxTx, err := s.callbackRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to begin transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("TX: "+constant.InternalErrorFmt, traceID))
	}
	defer func() {
		if !isCompleted {
			if e := s.callbackRepo.RollbackTransaction(ctxTx); e != nil {
				s.logger.Error(ctx, "failed to rollback transaction", logger.Error(err))
				resp, err = nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("RT: "+constant.InternalErrorFmt, traceID))
			}
		}
	}()

	if !isNewCallback {
		if err := s.callbackRepo.UpdateCallbackURLById(ctxTx, callbackID, data.URL); err != nil {
			s.logger.Error(ctx, "failed to update callback url", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("UPDATE: "+constant.InternalErrorFmt, traceID))
		}
		if err := s.callbackRepo.UpdateCallbackBaseURLById(ctxTx, callbackID, *data.BaseURL); err != nil {
			s.logger.Error(ctx, "failed to update callback base url", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("UPDATE: "+constant.InternalErrorFmt, traceID))
		}

	} else if isNewCallback {
		if err = s.callbackRepo.CreateCallback(ctxTx, data); err != nil {
			s.logger.Error(ctx, "failed to create callback url", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("CREATE: "+constant.InternalErrorFmt, traceID))
		}
	}

	if err = s.callbackRepo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "failed to commit transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("CMT: "+constant.InternalErrorFmt, traceID))
	}

	resp.Information.CallbackID = callbackID
	isCompleted = true
	return
}
