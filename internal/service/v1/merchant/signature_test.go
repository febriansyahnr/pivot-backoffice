package merchant_test

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGenOpenAPISignature(t *testing.T) {

	service := &MerchantService{}

	tests := []struct {
		name       string
		request    *merchant.GenSignatureReq
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR:Invalid private key",
			request: &merchant.GenSignatureReq{
				PrivateKey: "XXX",
			},
			wantErr: "failed to parse PEM block containing the key",
		},
		{
			name: "SUCCESS",
			request: &merchant.GenSignatureReq{
				Timestamp:  "2024-08-05T16:00:00+07:00",
				MerchantId: "3fc96de8-f65e-4b16-90a1-e2a00d1bae29",
				PrivateKey: "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC4UKt8Kj9UwwZQ\ns8ILqlPOLk1lEHwsuU5nvRiAyjx9k84JF/aRSuPRueUHg61AKoOqNODrD1LunCmS\n8+yGGPftr58ZNH0h9trTNT3HKxmSoU9YXY8Vmy69rpgkPDm3HTZD/AC/ErkCwi6J\nBIjGnAibCl8Y6ky2l/BcKNC/QD07sCNXSeBtlin9C94gmcnbiEpMC4UuzgoyWm2V\nUn8YS89hrkbi0fmcFS/0nC+6Lm3V9FX8Q5GwVmO6ZWJxFJvthXqAGP+Aj+XxfeCF\nfXYvApWG/MluFuZV3gN9op6vQ2+5vaJJDYZCGM8tSEWKnHOgbAw35jp1tIb1F6Bt\nBmSvzsCJAgMBAAECggEAQ4j/hMQAJ6E8N6beI6MaCRLLNgxny4VstrNBfrNbndHi\nLU/T/2HW/zpjsrCrczcAvoWYoliflSGwVBG/qVUNx1BR9gzXCvJmNPytscRXnvQv\nXBwP+SU+567JPYG5ziBMiXWVmm2UT+/x2C/KpCd5OcH/nWQAjuk2X7Zu4pz5stwU\nDtXx4xf3JPEEWAn2uHKn/ckvp5X4LeEh0ow5hEJ6YhM52k58bzjKRtIjPmLopouh\n5UBKw/DqL2qM85CnOeups01Yml4vOQuot1yQcWr2RqS5mIiLxzvjHE1Js1G42p37\nACvWCnhouQY9WZvyJIa65utFFdrJaweSnVGBzefjjwKBgQDtXrd3As5sr2KIb1ty\nUKSOwgLPE9oiViO3E1c43lIu5yfYw9n3f31gUkBzoNh+uXC/dapdVajQew4NGFnI\nl22AqoV+hDp7h+XORkjryBCaA27o010IvOO42ti7ZgT4U7EXwGVmTQzNYrxlZpvB\nktD6swpxWQISbeX83VQqM84p9wKBgQDGx/btyk0pGYdxASA/3F107coTQVygwq71\nMiYLU5Dr7HGhyEBT4dQ5oHIW6WCFCce7S7DgWHw/Ng77F5YuhCd9wU9MQtRgwbza\nZpfqiPAT1ceqkshGiGt/tUEFUxGSjZX6LRKZ8amk7qC5TNMY3spvPvnb+uuNAyVG\nvXH28dPJfwKBgEEX3E/yoREE94xanUU4ACh147dNxl/sJ1cpIp4huX/LPA4hh0Br\n4cHsTGhpD3WQ/O5EIjf+KZEibbQBnX14qTrDiGAteqwtlEOA2rZt4r+ZeWy3qaef\nxQMIYK7jRzGiIcpVpHjtYDlifi+Ad+4ZiN13A8IZmovbP1qch1wbYMn5AoGAKaUH\nBHZXh/7DM1eLDBX8tlyC81nEMCHZSaFB+yl8uRCGFeDAKVKshY4pmMc342dTItgO\nrFGdZhjLNquQWRpys5PmKxHtMIAmMpM/zHD36w/kjsXFk5FNBCpS/uySR+PFwe3j\nccEBS356yZdgulsiif/llMKSyq4YByP1Vkj/l70CgYEAnejQZW8Txe04JWnBxmKH\ny3yn9j4w7GO636uPKuWcMS+JzoDHFxQKiY4Iz8SDEVBFPe+k+NYPtTpqRn6nHyMH\nDkWkufwSjpb3qJRycF/2gZV+94BhO/rZcXctEVMPt/Kym2B+kPckIeUUIJpLpqL+\nLEgxw65GdFS6lLzEaMgr4LQ=\n-----END PRIVATE KEY-----",
			},
			wantResult: "O/rCIeClUOWRRKsQKr5UtvUqOEO8vCBnfoXA68fgifzPtECt5ZPeLyEXOpSpNnZyBHjPfME7saiJcmznU45W1MAqHSEv0EJ9+qBnJv2TLB1wDhGTbsjdlonuXWsaFegkRE0FF9WekjSpIXqjCAudFKfQ5usC550L1fmj2loN70SXlov64h/4x1V+2GlzfvFEL3Rkwg3w1EfI12beyX+GXZMpK21m86Pl5ZJN7Csvs6kdBpgBLB+cwp/d1WIyOkppR4+9ObaXsEmFoFij3GuyGbAWGKORhMhduIgElQp0jTIOxU9Bwf/w/GIYg/mgAhIL5L4f18AKI4SVPEJl4KOENQ==",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if pk, err := service.GenOpenAPISignature(context.Background(), test.request); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, pk)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestValidateSnapRequestSignature(t *testing.T) {

	request := &merchant.ValidateSnapSignatureRequest{
		ClientID:    "5ffd4643-d129-433f-85cb-cd5ebb3f17a6",
		Timestamp:   "2020-01-01T00:00:00+07:00",
		AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYWNrZW5kLXBvcnRhbCIsInN1YiI6IjVmZmQ0NjQzLWQxMjktNDMzZi04NWNiLWNkNWViYjNmMTdhNiIsImV4cCI6MTc0NTk4MjA4NywiY2xpZW50SWQiOiI1ZmZkNDY0My1kMTI5LTQzM2YtODVjYi1jZDVlYmIzZjE3YTYiLCJtZXJjaGFudElkIjoiNWZmZDQ2NDMtZDEyOS00MzNmLTg1Y2ItY2Q1ZWJiM2YxN2E2In0.B08bGwXSJaq-rztQ2rUPoyf5f7C2DtBtYo2uH7h66EE",
		Signature:   "QcXIRTpcp/axdPhqL4IG4utyORzkt2wDEe720Ue4akbRk4HRS9Ohr+AaIDAEso5Pd3mmpjkdoRCEvu2vvzcbeQ==",
		Url:         "/api?query=think",
		Method:      "POST",
		Body: []byte(`{
			"roast":    "duck",
			"breaking": "Bakd"
		}`),
	}

	vaultTransit := vaultMock.NewIVaultTransit(t)

	tests := []struct {
		name       string
		setup      func(merchantSvc *mocks.IMerchantRepository)
		request    *merchant.ValidateSnapSignatureRequest
		wantErr    error
		wantResult string
	}{
		{
			name:    "SUCCESS",
			request: request,
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(&merchant.MerchantAuth{
					Secret: "FYw6ZZImSm7fjHHUetgDQUmKq80eLbWQFLphvKxy",
				}, nil)
			},
		},
		{
			name:    "SUCCESS: Encrypted merchant secret",
			request: request,
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(&merchant.MerchantAuth{
					Secret:        "vault:v1:...",
					SecretVersion: 1,
				}, nil)
				vaultTransit.On(
					"Decrypt", mock.Anything, mock.Anything,
				).Once().Return(&vault.DecryptResponse{Plaintext: []byte("FYw6ZZImSm7fjHHUetgDQUmKq80eLbWQFLphvKxy")}, nil)
			},
		},
		{
			name:    "ERROR: Verify",
			request: &merchant.ValidateSnapSignatureRequest{},
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(&merchant.MerchantAuth{
					Secret: "FYw6ZZImSm7fjHHUetgDQUmKq80eLbWQFLphvKxy",
				}, nil)
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrUnprocessableContent, errors.New("empty url")),
		},
		{
			name:    "ERROR: Get Merchant Repo",
			request: request,
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(nil, errors.New("error"))
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrValidateRequestSignature),
		},
		{
			name:    "ERROR: Decrypt merchant secret",
			request: request,
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(&merchant.MerchantAuth{
					Secret:        "vault:v1:...",
					SecretVersion: 1,
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merchantRepo := mocks.NewIMerchantRepository(t)
			test.setup(merchantRepo)

			merchantSvc := New(merchantRepo, nil, nil, nil, nil, nil, WithVaultTransit(vaultTransit))

			err := merchantSvc.ValidateSnapRequestSignature(t.Context(), test.request)
			if test.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), test.wantErr.Error())
			}
			vaultTransit.AssertExpectations(t)
		})
	}
}

func TestGenerateSnapRequestSignature(t *testing.T) {

	vaultTransit := vaultMock.NewIVaultTransit(t)

	request := &merchant.GenerateSnapSignatureRequest{
		ClientID:    "5ffd4643-d129-433f-85cb-cd5ebb3f17a6",
		Timestamp:   "2020-01-01T00:00:00+07:00",
		AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJiYWNrZW5kLXBvcnRhbCIsInN1YiI6IjVmZmQ0NjQzLWQxMjktNDMzZi04NWNiLWNkNWViYjNmMTdhNiIsImV4cCI6MTc0NTk4MjA4NywiY2xpZW50SWQiOiI1ZmZkNDY0My1kMTI5LTQzM2YtODVjYi1jZDVlYmIzZjE3YTYiLCJtZXJjaGFudElkIjoiNWZmZDQ2NDMtZDEyOS00MzNmLTg1Y2ItY2Q1ZWJiM2YxN2E2In0.B08bGwXSJaq-rztQ2rUPoyf5f7C2DtBtYo2uH7h66EE",
		Url:         "/api?query=think",
		Method:      "POST",
		Body: []byte(`{
			"roast": "duck", "breaking": "Bakd"
		}`),
	}

	tests := []struct {
		name       string
		setup      func(merchantSvc *mocks.IMerchantRepository)
		request    *merchant.GenerateSnapSignatureRequest
		wantErr    error
		wantResult string
	}{
		{
			name:    "SUCCESS",
			request: request,
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(&merchant.MerchantAuth{
					Secret: "FYw6ZZImSm7fjHHUetgDQUmKq80eLbWQFLphvKxy",
				}, nil)
			},
		},
		{
			name:    "SUCCESS: Encrypted merchant secret",
			request: request,
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(&merchant.MerchantAuth{
					Secret:        "vault:v1:...",
					SecretVersion: 1,
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(&vault.DecryptResponse{
					Plaintext: []byte("FYw6ZZImSm7fjHHUetgDQUmKq80eLbWQFLphvKxy"),
				}, nil)
			},
		},
		{
			name:    "ERROR: Get Merchant Repo",
			request: request,
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(nil, errors.New("error"))
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrGenerateRequestSignature),
		},
		{
			name:    "ERROR: Decrypt merchant secret",
			request: request,
			setup: func(merchantSvc *mocks.IMerchantRepository) {
				merchantSvc.On(
					"GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything,
				).Return(&merchant.MerchantAuth{
					Secret:        "vault:v1:...",
					SecretVersion: 1,
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErrs.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merchantRepo := mocks.NewIMerchantRepository(t)
			test.setup(merchantRepo)

			merchantSvc := New(merchantRepo, nil, nil, nil, nil, nil, WithVaultTransit(vaultTransit))

			_, err := merchantSvc.GenerateSnapRequestSignature(t.Context(), test.request)
			if test.wantErr == nil {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), test.wantErr.Error())
			}
			vaultTransit.AssertExpectations(t)
		})
	}
}
