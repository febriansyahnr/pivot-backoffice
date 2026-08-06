package merchant_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	funcMock "github.com/paper-indonesia/pivot-backoffice/mocks/func"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/merchant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpsertMerchantBOD(t *testing.T) {
	merchantSvc := serviceMocks.NewIMerchantService(t)

	handler := New(merchantSvc, validatorExt.New(), nil)

	router := chi.NewRouter()
	router.Post("/merchants/{id}/bods", handler.CreateMerchantBOD)
	router.Put("/merchants/{id}/bods/{bod_id}", handler.UpdateMerchantBOD)

	requestFormData := func(r *http.Request) {
		r.PostForm = url.Values{
			"position":       {"Director"},
			"name":           {"John Wick"},
			"identityNumber": {"123456789"},
			"positionLong":   {""},
			"createdBy":      {"Hendru"},
		}
	}
	fileName := "ktp.jpt"
	bodId := uuid.NewString()
	merchantId := uuid.NewString()

	tests := []struct {
		name           string
		bodId          string
		merchantId     string
		filename       string
		setupMock      func()
		setupRequest   func(r *http.Request)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:No such file",
			bodId:          "X",
			filename:       "invalid",
			setupRequest:   requestFormData,
			merchantId:     "X",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "multipart: NextPart: EOF"),
		},
		{
			name:           "ERROR:Invalid merchant id",
			bodId:          bodId,
			merchantId:     "X",
			setupRequest:   requestFormData,
			filename:       fileName,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantId":"Key: 'UpsertBoardOfDirectorReq.MerchantId' Error:Field validation for 'MerchantId' failed on the 'uuid' tag"}}`,
		},
		{
			name:       "ERROR: Failed Validation",
			bodId:      bodId,
			merchantId: merchantId,
			filename:   fileName,
			setupRequest: func(r *http.Request) {
				r.PostForm = url.Values{
					"position":       {"Shareholder"},
					"name":           {"John Wick"},
					"identityNumber": {"123456789"},
					"positionLong":   {""},
					"createdBy":      {"Hendru"},
					"shares":         {"-100"},
				}
			},
			setupMock: func() {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid shares value"),
		},
		{
			name:         "ERROR:Some error when run function upsert",
			bodId:        bodId,
			merchantId:   merchantId,
			filename:     fileName,
			setupRequest: requestFormData,
			setupMock: func() {
				merchantSvc.On(
					"UpsertMerchantBOD", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.UpsertBoardOfDirectorReq"),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:         "SUCCESS",
			bodId:        bodId,
			merchantId:   merchantId,
			filename:     fileName,
			setupRequest: requestFormData,
			setupMock: func() {
				merchantSvc.On(
					"UpsertMerchantBOD", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.UpsertBoardOfDirectorReq"),
				).Return(bodId, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"id":"%s"}}`, bodId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for i := 0; i < 2; i++ {
				rec := httptest.NewRecorder()

				body, contentType, err := funcMock.CreateMultipartFormFile("identityFile", test.filename, "")
				require.NoError(t, err)
				require.NotEmpty(t, contentType)
				defer body.Reset()

				var req *http.Request
				if test.filename == "invalid" {
					body = bytes.NewBufferString("invalid multipart data")
				}
				if i == 0 {
					req = httptest.NewRequest(http.MethodPost, "/merchants/"+test.merchantId+"/bods", body)

				} else {
					req = httptest.NewRequest(http.MethodPut, "/merchants/"+test.merchantId+"/bods/"+test.bodId, body)
				}
				req.Header.Add(c.HeaderContentType, contentType)

				test.setupRequest(req)

				if test.setupMock != nil {
					test.setupMock()
				}
				router.ServeHTTP(rec, req)
				if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
					t.Log("Output:", rec.Body.String())
				}
				assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			}
		})
	}
}

func TestGetListMerchantBOD(t *testing.T) {
	merchantSvc := serviceMocks.NewIMerchantService(t)

	handler := New(merchantSvc, validatorExt.New(), nil)

	router := chi.NewRouter()
	router.Get("/merchants/{id}/bods", handler.GetListMerchantBOD)

	bodId := uuid.NewString()
	merchantId := uuid.NewString()

	tests := []struct {
		name           string
		merchantId     string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid merchant id",
			merchantId:     "X",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid merchant id format"),
		},
		{
			name:       "ERROR:Some error",
			merchantId: merchantId,
			setupMock: func() {
				merchantSvc.On("GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:       "SUCCESS",
			merchantId: merchantId,
			setupMock: func() {
				merchantSvc.On("GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType()).Return([]merchant.BoardOfDirectorResp{
					{
						Id:             bodId,
						Name:           "John Wick",
						PositionLong:   "-",
						IdentityNumber: "123456789",
						IdentityFile:   "http://mystorage.id/ktp.jpg",
						Position:       "Striker",
						Shares:         100,
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":[{"id":"%s","name":"John Wick","position":"Striker","positionLong":"-","identityNumber":"123456789","identityFile":"http://mystorage.id/ktp.jpg","shares":100,"createdBy":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`, bodId),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/merchants/"+test.merchantId+"/bods", nil)

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
