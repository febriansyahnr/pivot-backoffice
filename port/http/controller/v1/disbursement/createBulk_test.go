package disbursementController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	funcMock "github.com/paper-indonesia/pivot-backoffice/mocks/func"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	res "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/disbursement"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateBulk(t *testing.T) {
	disbursementSvc := serviceMocks.NewIDisbursementService(t)

	service := Services{
		DisbursementSvc: disbursementSvc,
	}

	handler := New(&config.Config{}, validator.New(), nil, service, nil, nil)

	router := chi.NewRouter()
	router.Post("/bulk/upload", handler.CreateBulk)
	strToPtrStr := func(s string) *string { return &s }

	tests := []struct {
		name           string
		userClams      *user.UserTokenClaims
		filename       string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, res.ErrTypeAPI, "user not found"),
		},
		{
			name:           "ERROR:Parse file",
			userClams:      &user.UserTokenClaims{},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, res.ErrTypeAPI, "http: no such file"),
		},
		{
			name:      "ERROR:Some error",
			userClams: &user.UserTokenClaims{},
			filename:  "test.txt",
			setupMock: func() {
				disbursementSvc.On(
					"BulkCreate", mock.Anything, mock.AnythingOfType("*disbursementModel.BulkCreateRequest"), // NOSONAR
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, res.ErrTypeUnknown, "some error"),
		},
		{
			name:      "SUCCESS",
			userClams: &user.UserTokenClaims{},
			filename:  "test.txt",
			setupMock: func() {
				disbursementSvc.On(
					"BulkCreate", mock.Anything, mock.AnythingOfType("*disbursementModel.BulkCreateRequest"), // NOSONAR
				).Return(&disbursementModel.BulkCreateResponse{
					UUID:       "uuid",
					MerchantID: "123456",
					File:       "http://falid.file",
					FileFailed: strToPtrStr("http://invalid.file"),
					Status:     "UPLOADING",
					CreatedBy:  strToPtrStr("JOHN"),
					TotalData:  1, TotalAmount: 15_000,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"uuid":"uuid","merchantId":"123456","file":"http://falid.file","fileFailed":"http://invalid.file","status":"UPLOADING","createdBy":"JOHN","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z","totalData":1,"totalAmount":15000}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			body, contentType, err := funcMock.CreateMultipartFormFile("file", test.filename, "")
			require.NoError(t, err)
			require.NotEmpty(t, contentType)
			defer body.Reset()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/bulk/upload", body)

			if test.userClams != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClams))
			}
			req.Header.Add(c.HeaderContentType, contentType)

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
