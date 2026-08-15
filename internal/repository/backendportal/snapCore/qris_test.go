package snapCoreRepository_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qris"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/snapCore"
	httpReqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newServer(t *testing.T, handler http.Handler) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Log("Test server run at", server.URL)

	return server
}

func TestQrUploadDocument(t *testing.T) {
	var (
		badReqResp  = `{"code":"40","message":"bad request","error":"registerId is required field"}`
		successResp = `{"code":"00","message":"OK","data":{"uuid":"123","mediaId":"mediaId-123"}}`
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(c.HeaderContentType, c.MIMEApplicationJSON)

		if r.PostFormValue("acquirer") == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`acquirer not found`))

		} else if r.PostFormValue("registrationId") == "" || r.PostFormValue("documentType") == "" || r.PostFormValue("number") == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(badReqResp))

		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(successResp))
		}
	})

	server := newServer(t, handler)
	defer server.Close()

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	tests := []struct {
		name           string
		baseURL        string
		data           *snapCoreModel.UploadDocumentReq
		wantErr        string
		wantResult     *snapCoreModel.UploadDocumentResp
		setupTestHooks func(repo interface {
			SetTestMultipartCreateFileHook(func() error)
			SetTestMultipartCopyHook(func() error)
			SetTestMultipartCreateFieldHook(func(string) error)
			SetTestMultipartWriteToHook(func(string) error)
			SetTestMultipartCloseHook(func() error)
			SetTestHTTPNewRequestHook(func() error)
			SetTestIOReadAllHook(func() error)
		})
	}{
		{
			name:    "ERROR:Unsupported protocol scheme",
			baseURL: "http",
			data:    &snapCoreModel.UploadDocumentReq{},
			wantErr: `execute request: Post "http/api/v1.0/internal/qr-mpm/upload": unsupported protocol scheme ""`,
		},
		{
			name:    "ERROR:Decode response",
			baseURL: server.URL,
			data:    &snapCoreModel.UploadDocumentReq{},
			wantErr: `acquirer not found`,
		},
		{
			name:    "ERROR:Bad Request",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer: "Vendor",
			},
			wantErr: badReqResp,
		},
		{
			name:    "SUCCESS",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer:       "Vendor",
				RegistrationId: "123456",
				DocumentType:   "type",
				DocumentNumber: "doc-123456",
			},
			wantResult: &snapCoreModel.UploadDocumentResp{
				Id:      "123",
				MediaId: "mediaId-123",
			},
		},
		{
			name:    "ERROR:CreateFormFile error via test hook",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer:       "Vendor",
				RegistrationId: "123456",
				DocumentType:   "type",
				DocumentNumber: "doc-123456",
			},
			wantErr: "create form file",
			setupTestHooks: func(repo interface {
				SetTestMultipartCreateFileHook(func() error)
				SetTestMultipartCopyHook(func() error)
				SetTestMultipartCreateFieldHook(func(string) error)
				SetTestMultipartWriteToHook(func(string) error)
				SetTestMultipartCloseHook(func() error)
				SetTestHTTPNewRequestHook(func() error)
				SetTestIOReadAllHook(func() error)
			}) {
				repo.SetTestMultipartCreateFileHook(func() error {
					return assert.AnError
				})
			},
		},
		{
			name:    "ERROR:io.Copy error via test hook",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer:       "Vendor",
				RegistrationId: "123456",
				DocumentType:   "type",
				DocumentNumber: "doc-123456",
			},
			wantErr: "copy raw file",
			setupTestHooks: func(repo interface {
				SetTestMultipartCreateFileHook(func() error)
				SetTestMultipartCopyHook(func() error)
				SetTestMultipartCreateFieldHook(func(string) error)
				SetTestMultipartWriteToHook(func(string) error)
				SetTestMultipartCloseHook(func() error)
				SetTestHTTPNewRequestHook(func() error)
				SetTestIOReadAllHook(func() error)
			}) {
				repo.SetTestMultipartCopyHook(func() error {
					return assert.AnError
				})
			},
		},
		{
			name:    "ERROR:CreateFormField error via test hook",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer:       "Vendor",
				RegistrationId: "123456",
				DocumentType:   "type",
				DocumentNumber: "doc-123456",
			},
			wantErr: "create form field (text)",
			setupTestHooks: func(repo interface {
				SetTestMultipartCreateFileHook(func() error)
				SetTestMultipartCopyHook(func() error)
				SetTestMultipartCreateFieldHook(func(string) error)
				SetTestMultipartWriteToHook(func(string) error)
				SetTestMultipartCloseHook(func() error)
				SetTestHTTPNewRequestHook(func() error)
				SetTestIOReadAllHook(func() error)
			}) {
				repo.SetTestMultipartCreateFieldHook(func(name string) error {
					return assert.AnError
				})
			},
		},
		{
			name:    "ERROR:WriteTo error via test hook",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer:       "Vendor",
				RegistrationId: "123456",
				DocumentType:   "type",
				DocumentNumber: "doc-123456",
			},
			wantErr: "write form value (text)",
			setupTestHooks: func(repo interface {
				SetTestMultipartCreateFileHook(func() error)
				SetTestMultipartCopyHook(func() error)
				SetTestMultipartCreateFieldHook(func(string) error)
				SetTestMultipartWriteToHook(func(string) error)
				SetTestMultipartCloseHook(func() error)
				SetTestHTTPNewRequestHook(func() error)
				SetTestIOReadAllHook(func() error)
			}) {
				repo.SetTestMultipartWriteToHook(func(name string) error {
					return assert.AnError
				})
			},
		},
		{
			name:    "ERROR:multipart Close error via test hook",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer:       "Vendor",
				RegistrationId: "123456",
				DocumentType:   "type",
				DocumentNumber: "doc-123456",
			},
			wantErr: "close multipart",
			setupTestHooks: func(repo interface {
				SetTestMultipartCreateFileHook(func() error)
				SetTestMultipartCopyHook(func() error)
				SetTestMultipartCreateFieldHook(func(string) error)
				SetTestMultipartWriteToHook(func(string) error)
				SetTestMultipartCloseHook(func() error)
				SetTestHTTPNewRequestHook(func() error)
				SetTestIOReadAllHook(func() error)
			}) {
				repo.SetTestMultipartCloseHook(func() error {
					return assert.AnError
				})
			},
		},
		{
			name:    "ERROR:NewRequestWithContext error via test hook",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer:       "Vendor",
				RegistrationId: "123456",
				DocumentType:   "type",
				DocumentNumber: "doc-123456",
			},
			wantErr: "create request",
			setupTestHooks: func(repo interface {
				SetTestMultipartCreateFileHook(func() error)
				SetTestMultipartCopyHook(func() error)
				SetTestMultipartCreateFieldHook(func(string) error)
				SetTestMultipartWriteToHook(func(string) error)
				SetTestMultipartCloseHook(func() error)
				SetTestHTTPNewRequestHook(func() error)
				SetTestIOReadAllHook(func() error)
			}) {
				repo.SetTestHTTPNewRequestHook(func() error {
					return assert.AnError
				})
			},
		},
		{
			name:    "ERROR:io.ReadAll error via test hook",
			baseURL: server.URL,
			data: &snapCoreModel.UploadDocumentReq{
				Acquirer:       "Vendor",
				RegistrationId: "123456",
				DocumentType:   "type",
				DocumentNumber: "doc-123456",
			},
			wantErr: "read all response body",
			setupTestHooks: func(repo interface {
				SetTestMultipartCreateFileHook(func() error)
				SetTestMultipartCopyHook(func() error)
				SetTestMultipartCreateFieldHook(func(string) error)
				SetTestMultipartWriteToHook(func(string) error)
				SetTestMultipartCloseHook(func() error)
				SetTestHTTPNewRequestHook(func() error)
				SetTestIOReadAllHook(func() error)
			}) {
				repo.SetTestIOReadAllHook(func() error {
					return assert.AnError
				})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{
				SnapCoreConfig: config.SnapCoreConfig{BaseUrl: test.baseURL},
			}
			repo := New(cfg, &config.Secret{}, logger, nil)

			if test.setupTestHooks != nil {
				test.setupTestHooks(repo)
			}

			if result, err := repo.QrUploadDocument(context.Background(), test.data); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestQrFinalRegistration(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	httpReq := httpReqMock.NewIHTTPRequest(t)

	repo := New(&config.Config{}, &config.Secret{}, logger, httpReq)

	ptrRegistrationReqMockType := mock.AnythingOfType("*snapCoreModel.RegistrationReq")

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:Send HTTP request",
			setupMock: func() {
				httpReq.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), ptrRegistrationReqMockType, c.MapStrValStringMockType(),
				).Once().Return(nil, 0, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:Unmarshal JSON",
			setupMock: func() {
				httpReq.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), ptrRegistrationReqMockType, c.MapStrValStringMockType(),
				).Once().Return([]byte("page not found"), 404, nil)
			},
			wantErr: "page not found",
		},
		{
			name: "ERROR:Status code >= 400",
			setupMock: func() {
				httpReq.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), ptrRegistrationReqMockType, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"400","message":"Bad Request","error":"validation failed"}`), 400, nil)
			},
			wantErr: "snap",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				httpReq.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), ptrRegistrationReqMockType, c.MapStrValStringMockType(),
				).Return([]byte(`{"message":"OK"}`), 200, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if err := repo.QrFinalRegistration(context.Background(), &snapCoreModel.RegistrationReq{}); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestQrSyncRegistration(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	httpReq := httpReqMock.NewIHTTPRequest(t)

	repo := New(&config.Config{}, &config.Secret{}, logger, httpReq)

	ptrQrSyncRegistrationMockType := mock.AnythingOfType("*snapCoreModel.SyncRegistrationDataRequest")

	testCases := []struct {
		desc     string
		wantErr  bool
		mockFunc func()
	}{
		{
			desc:    "success sync registration",
			wantErr: false,
			mockFunc: func() {
				httpReq.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), ptrQrSyncRegistrationMockType, c.MapStrValStringMockType(),
				).Return([]byte(`{"message":"OK"}`), 200, nil).Once()
			},
		},
		{
			desc:    "error sync registration",
			wantErr: true,
			mockFunc: func() {
				httpReq.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), ptrQrSyncRegistrationMockType, c.MapStrValStringMockType(),
				).Return([]byte(`{"message":"INTERNAL SERVER ERROR"}`), 500, assert.AnError).Once()
			},
		},
		{
			desc: "ERROR:Unmarshal JSON",
			mockFunc: func() {
				httpReq.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), ptrQrSyncRegistrationMockType, c.MapStrValStringMockType(),
				).Once().Return([]byte("page not found"), 404, nil)
			},
			wantErr: true,
		},
		{
			desc:    "ERROR:Status code >= 400",
			wantErr: true,
			mockFunc: func() {
				httpReq.On(
					"POST", c.ValueCtxMockType(), c.StringMockType(), ptrQrSyncRegistrationMockType, c.MapStrValStringMockType(),
				).Once().Return([]byte(`{"code":"400","message":"Bad Request","error":"validation failed"}`), 400, nil)
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.mockFunc()

			if err := repo.QrSyncRegistration(context.Background(), &snapCoreModel.SyncRegistrationDataRequest{}); tC.wantErr {
				require.Error(t, err)

			} else {
				require.NoError(t, err)
			}
		})
	}
}
