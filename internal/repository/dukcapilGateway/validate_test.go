package dukcapilgatewayrepository

import (
	"context"
	"errors"
	"net/http"
	"testing"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/config"
	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/dukcapil"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(httpReq *mocks.IHTTPRequest)
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(mockHttp *mocks.IHTTPRequest) {
				mockHttp.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(
					[]byte(`{
						"content": [
							{
							"ALAMAT": "your_address",
							"JENIS_KLMIN": "your_gender",
							"JENIS_PKRJN": "your_occupation",
							"KAB_NAME": "your_district",
							"KEC_NAME": "your_sub_district",
							"KEL_NAME": "your_sub_district2",
							"PROP_NAME": "your_province",
							"NAMA_LGKP": "your_full_name",
							"NO_KAB": "your_district_no",
							"NO_KEC": "your_sub_district_no",
							"NO_KEL": "your_sub_district2_no",
							"NO_PROP": "your_province_no",
							"NO_RT": "your_rt",
							"NO_RW": "your_rw",
							"STATUS_KAWIN": "your_marital_status",
							"TGL_LHR": "your_birth_date",
							"TMPT_LHR": "your_birth_place"
							}
						],
						"lastPage": true,
						"numberOfElements": 1,
						"totalElements": 10,
						"firstPage": true,
						"number": 0,
						"size": 10,
						"quotaLimiter": 100
					}`),
					http.StatusOK,
					nil,
				)
			},
		},
		{
			name: "ERROR: HTTP Error",
			setup: func(mockHttp *mocks.IHTTPRequest) {
				mockHttp.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(
					[]byte(``),
					http.StatusBadGateway,
					errors.New("error http call"),
				)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid JSON Response",
			setup: func(mockHttp *mocks.IHTTPRequest) {
				mockHttp.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(
					[]byte(`{invalid json response}`),
					http.StatusOK,
					nil,
				)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Empty Content",
			setup: func(mockHttp *mocks.IHTTPRequest) {
				mockHttp.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(
					[]byte(`{
						"content": [
							
						],
						"lastPage": true,
						"numberOfElements": 0,
						"totalElements": 10,
						"firstPage": true,
						"number": 0,
						"size": 10,
						"quotaLimiter": 100
					}`),
					http.StatusOK,
					nil,
				)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Empty Content Alternative",
			setup: func(mockHttp *mocks.IHTTPRequest) {
				mockHttp.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(
					[]byte(`{
						"content": [
							
						],
						"lastPage": true,
						"numberOfElements": 0,
						"totalElements": 10,
						"firstPage": true,
						"number": 0,
						"size": 10,
						"quotaLimiter": 100
					}`),
					http.StatusOK,
					nil,
				)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid Setup Configuration ",
			setup: func(mockHttp *mocks.IHTTPRequest) {
				mockHttp.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(
					[]byte(`{
						"content": [
							{
							"RESPONSE_CODE": "04",
							"RESPONSE_DESC": "IP Tidak Sesuai",
							"RESPONSE": "IP Tidak Sesuai"
							}
						],
						"lastPage": true,
						"numberOfElements": 1,
						"totalElements": 10,
						"firstPage": true,
						"number": 0,
						"size": 10,
						"quotaLimiter": 100
					}`),
					http.StatusOK,
					nil,
				)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid NIK ",
			setup: func(mockHttp *mocks.IHTTPRequest) {
				mockHttp.On(
					"POST", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(
					[]byte(`{
						"content": [
							{
							"RESPONSE_CODE": "11",
							"RESPONSE_DESC": "meninggal dunia",
							"RESPONSE": "meninggal dunia"
							}
						],
						"lastPage": true,
						"numberOfElements": 1,
						"totalElements": 10,
						"firstPage": true,
						"number": 0,
						"size": 10,
						"quotaLimiter": 100
					}`),
					http.StatusOK,
					nil,
				)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		httpExt := mocks.NewIHTTPRequest(t)
		logger, _ := pdkLogger.NewZapLogger(pdkLogger.Config{})

		tc.setup(httpExt)

		repo := New(&config.Config{}, &config.Secret{}, logger, httpExt)
		_, err := repo.VerifyIdentity(context.Background(), &dukcapilmodel.VerifyRequest{})
		if tc.wantErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
