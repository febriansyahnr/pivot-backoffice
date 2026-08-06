package creditcardCoreProcessorRepository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/creditcardCoreProcessor"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMIDByAcquirerMID(t *testing.T) {
	successResponse := `{"data": {"id": "123", "acquirer_mid": "test_mid"}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Get MID by Acquirer MID",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.GetMIDByAcquirerMID(context.Background(), "test_acquirer_mid")
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestCreateMID(t *testing.T) {
	successResponse := `{"data": {"id": "123", "name": "test_mid"}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Create MID",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.CreateMID(context.Background(), &creditcardCoreProcessorModel.CreateMIDRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestCreateMIDMap(t *testing.T) {
	successResponse := `{"data": {"id": "123", "mid_id": "test_mid_id"}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Create MID Map",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.CreateMIDMap(context.Background(), &creditcardCoreProcessorModel.CreateMIDMapRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestGetMIDList(t *testing.T) {
	successResponse := `{"data": {"mids": []}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Get MID List",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.GetMIDList(context.Background(), &creditcardCoreProcessorModel.GetMIDListRequest{
				IsDefault: util.BoolPtr(true),
				IsActive:  util.BoolPtr(false),
			})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestGetMIDMapList(t *testing.T) {
	successResponse := `{"data": {"mid_maps": []}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Get MID Map List",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.GetMIDMapList(context.Background(), &creditcardCoreProcessorModel.GetMIDMapListRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestUpdateMIDMapPriority(t *testing.T) {
	successResponse := `{"data": {"id": "123", "status": "updated"}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Update MID Map Priority",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.UpdateMIDMapPriority(context.Background(), creditcardCoreProcessorModel.UpdateMIDMapPriorityRequest{MidMapID: uuid.New()})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestFindMIDMapByMerchant(t *testing.T) {
	successResponse := `{"data": {"mid_maps": []}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Get MID Map List",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.FindMIDMapByMerchant(context.Background(), &creditcardCoreProcessorModel.FindMIDMapByMerchantRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestUpdateMID(t *testing.T) {
	successResponse := `{"data": {"id": "123", "name": "updated_mid"}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Update MID",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"PUT",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.UpdateMID(context.Background(), &creditcardCoreProcessorModel.UpdateMIDRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestGetMID(t *testing.T) {
	successResponse := `{"data": {"id": "123", "name": "updated_mid"}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Get MID",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 403 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"403","error":"error message"}`), 403, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"GET",
					mock.Anything,
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			_, err := repo.GetMID(context.Background(), "")
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}

func TestValidateMidInstallmentBins(t *testing.T) {
	successResponse := `{"data": {"id": "123", "name": "updated_mid"}, "message": "success"}`

	testCases := []struct {
		name      string
		wantError bool
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
	}{
		{
			name:      "SUCCESS: Validate MID Installment Bins",
			wantError: false,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 200, nil)
			},
		},
		{
			name:      "ERROR: HTTP status not ok",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(successResponse), 400, nil)
			},
		},
		{
			name:      "ERROR: error unmarshal response",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{error unmarshal}`), 200, nil)
			},
		},
		{
			name:      "ERROR: error other",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(``), 500, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: got 400 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"400","error":"error message"}`), 400, nil)
			},
		},
		{
			name:      "ERROR: got 500 with field error json from credit core",
			wantError: true,
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On(
					"POST",
					mock.Anything,
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code":"500","error":"error message"}`), 500, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockHttp := httpMocks.NewIHTTPRequest(t)
			tc.setupMock(mockHttp)

			conf := &config.Config{CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{BaseUrl: ""}}
			secret := &config.Secret{}

			repo := New(conf, secret, mockLogger, mockHttp)
			err := repo.ValidateMidInstallmentBins(context.Background(), &creditcardCoreProcessorModel.ValidateMIDInstallmentBinsRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockHttp.AssertExpectations(t)
		})
	}
}
