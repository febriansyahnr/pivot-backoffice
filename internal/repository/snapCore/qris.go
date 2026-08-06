package snapCoreRepository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validation"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

var (
	once   sync.Once
	client *http.Client

	buffPool = &sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
)

func (r *snapCoreRepository) QrUploadDocument(ctx context.Context, data *snapCoreModel.UploadDocumentReq) (*snapCoreModel.UploadDocumentResp, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/snapCore/QrUploadDocument")
	defer segment.End()

	once.Do(func() {
		client = &http.Client{}
	})

	buf := buffPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		buffPool.Put(buf)
	}()

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)

	mw := multipart.NewWriter(buf)

	dataPart, err := mw.CreateFormFile("file", filepath.Base(data.ObjectName))
	// Test hook for CreateFormFile error injection
	if r.testMultipartCreateFileHook != nil {
		if hookErr := r.testMultipartCreateFileHook(); hookErr != nil {
			err = hookErr
		}
	}
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	_, err = io.Copy(dataPart, bytes.NewReader(data.RawFile))
	// Test hook for io.Copy error injection
	if r.testMultipartCopyHook != nil {
		if hookErr := r.testMultipartCopyHook(); hookErr != nil {
			err = hookErr
		}
	}
	if err != nil {
		return nil, fmt.Errorf("copy raw file: %w", err)
	}

	dataText := map[string]string{
		"acquirer":       data.Acquirer,
		"registrationId": data.RegistrationId,
		"documentType":   data.DocumentType,
		"number":         data.DocumentNumber,
	}
	for name, value := range dataText {
		formField, err := mw.CreateFormField(name)
		// Test hook for CreateFormField error injection
		if r.testMultipartCreateFieldHook != nil {
			if hookErr := r.testMultipartCreateFieldHook(name); hookErr != nil {
				err = hookErr
			}
		}
		if err != nil {
			return nil, fmt.Errorf("create form field (text): %w", err)
		}

		_, err = strings.NewReader(value).WriteTo(formField)
		// Test hook for WriteTo error injection
		if r.testMultipartWriteToHook != nil {
			if hookErr := r.testMultipartWriteToHook(name); hookErr != nil {
				err = hookErr
			}
		}
		if err != nil {
			return nil, fmt.Errorf("write form value (text): %w", err)
		}
	}

	err = mw.Close()
	// Test hook for multipart Close error injection
	if r.testMultipartCloseHook != nil {
		if hookErr := r.testMultipartCloseHook(); hookErr != nil {
			err = hookErr
		}
	}
	if err != nil {
		return nil, fmt.Errorf("close multipart: %w", err)
	}

	snapURL := fmt.Sprintf(
		"%s/api/v1.0/internal/qr-mpm/upload", r.config.SnapCoreConfig.BaseUrl,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, snapURL, buf)
	// Test hook for NewRequestWithContext error injection
	if r.testHTTPNewRequestHook != nil {
		if hookErr := r.testHTTPNewRequestHook(); hookErr != nil {
			err = hookErr
		}
	}
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Add(constant.HeaderXRequestId, requestId)
	req.Header.Add(constant.HeaderContentType, mw.FormDataContentType())
	req.Header.Add(constant.HeaderXInternalServiceKey, r.secret.SnapCoreSecret.InternalServiceKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	rawRespBody, err := io.ReadAll(resp.Body)
	// Test hook for io.ReadAll error injection
	if r.testIOReadAllHook != nil {
		if hookErr := r.testIOReadAllHook(); hookErr != nil {
			err = hookErr
		}
	}
	if err != nil {
		return nil, fmt.Errorf("read all response body: %w", err)
	}

	r.logger.Info(
		ctx, "response from uploaded QRIS document",
		logger.String("registrationId", data.RegistrationId),
		logger.String("url", snapURL), logger.Int("statusCode", resp.StatusCode), logger.ByteString("response", rawRespBody),
	)

	respBody := &snapCoreModel.RegUploadResp{}
	if err = json.Unmarshal(rawRespBody, &respBody); err != nil {
		return nil, &validation.Fields{"snap": string(rawRespBody)}
	}

	if resp.StatusCode >= 400 {
		return nil, &validation.Fields{"snap": respBody}
	}
	return &snapCoreModel.UploadDocumentResp{
		Id:      respBody.Data.Id,
		MediaId: respBody.Data.MediaId,
	}, nil
}

func (r *snapCoreRepository) QrFinalRegistration(ctx context.Context, data *snapCoreModel.RegistrationReq) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/snapCore/QrFinalRegistration")
	defer segment.End()

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)

	snapURL := fmt.Sprintf("%s/api/v1.0/internal/qr-mpm/registration/submission", r.config.SnapCoreConfig.BaseUrl)

	data.MerchantType = strings.ToUpper(data.MerchantType)
	response, statusCode, err := r.httpRequest.POST(
		ctx, snapURL, data,
		map[string]string{
			constant.HeaderXRequestId:          requestId,
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		return err
	}

	r.logger.Info(
		ctx, "response from submitted QRIS registration",
		logger.String("registrationId", data.RegistrationId),
		logger.String("url", snapURL), logger.Int("statusCode", statusCode), logger.ByteString("response", response),
	)

	respBody := map[string]interface{}{}
	if err := json.Unmarshal([]byte(response), &respBody); err != nil {
		return &validation.Fields{"snap": string(response)}

	} else if statusCode >= 400 {
		return &validation.Fields{"snap": respBody}
	}
	return nil
}

func (r *snapCoreRepository) QrSyncRegistration(ctx context.Context, data *snapCoreModel.SyncRegistrationDataRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/snapCore/QrSyncRegistration")
	defer segment.End()

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)

	snapURL := fmt.Sprintf("%s/api/v1.0/internal/qr-mpm/registration/sync", r.config.SnapCoreConfig.BaseUrl)

	data.MerchantType = strings.ToUpper(data.MerchantType)
	response, statusCode, err := r.httpRequest.POST(
		ctx, snapURL, data,
		map[string]string{
			constant.HeaderXRequestId:          requestId,
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		return err
	}

	r.logger.Info(
		ctx, "response from sync QRIS registration",
		logger.String("registrationId", data.RegistrationID),
		logger.String("url", snapURL), logger.Int("statusCode", statusCode), logger.ByteString("response", response),
	)

	respBody := map[string]interface{}{}
	if err := json.Unmarshal([]byte(response), &respBody); err != nil {
		return &validation.Fields{"snap": string(response)}

	} else if statusCode >= 400 {
		return &validation.Fields{"snap": respBody}
	}
	return nil
}
