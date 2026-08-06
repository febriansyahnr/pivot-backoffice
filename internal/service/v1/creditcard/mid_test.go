package creditcard

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMIDList(t *testing.T) {
	conf := &config.Config{}
	ctx := context.Background()

	mockResponse := &creditcardCoreProcessorModel.MIDListResponseData{
		Results: []creditcardCoreProcessorModel.MIDResponseData{
			{
				Uuid:               uuid.New(),
				Mid:                "TEST_MID_001",
				Name:               "Test MID 1",
				Type:               "REGULAR",
				Processor:          "HARSYA",
				PrincipalAvailable: []string{"VISA", "MASTERCARD"},
				IsActive:           true,
				IsDefault:          false,
				BaseURL:            "https://test.example.com",
			},
		},
		Pagination: creditcardCoreProcessorModel.PaginationResponse{
			PageLimit:   10,
			PageNumber:  1,
			TotalRecord: 1,
			TotalPage:   1,
		},
	}

	expectedResponse := &commonModel.PaginationResponse{
		Data: []interface{}{
			map[string]interface{}{
				"uuid":               mockResponse.Results[0].Uuid,
				"mid":                "TEST_MID_001",
				"name":               "Test MID 1",
				"type":               "REGULAR",
				"processor":          "HARSYA",
				"principalAvailable": []string{"VISA", "MASTERCARD"},
				"isActive":           true,
				"isDefault":          false,
				"baseURL":            "https://test.example.com",
			},
		},
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    10,
			TotalPages: 1,
			TotalItems: 1,
		},
	}

	testCases := []struct {
		name      string
		request   *creditcardModel.GetMIDListRequest
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository)
	}{
		{
			name: "SUCCESS",
			request: &creditcardModel.GetMIDListRequest{
				Page:    constant.DefaultPage,
				PerPage: constant.DefaultPaginationPageSize,
			},
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("GetMIDList", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.GetMIDListRequest")).Return(mockResponse, nil)
			},
		},
		{
			name: "ERROR: Repository Error",
			request: &creditcardModel.GetMIDListRequest{
				Page:    constant.DefaultPage,
				PerPage: constant.DefaultPaginationPageSize,
			},
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("GetMIDList", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.GetMIDListRequest")).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

			tc.mockSetup(mockRepo)

			svc := New(conf, mockLogger, nil, nil, nil, mockRepo)
			response, err := svc.GetMIDList(ctx, tc.request)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, expectedResponse.Meta.Page, response.Meta.Page)
				assert.Equal(t, expectedResponse.Meta.PerPage, response.Meta.PerPage)
				assert.Equal(t, expectedResponse.Meta.TotalPages, response.Meta.TotalPages)
				assert.Equal(t, expectedResponse.Meta.TotalItems, response.Meta.TotalItems)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetMIDMapList(t *testing.T) {
	conf := &config.Config{}
	ctx := context.Background()

	mockResponse := &creditcardCoreProcessorModel.MIDMapListResponseData{
		Results: []creditcardCoreProcessorModel.MIDMapResponseData{
			{
				Uuid:         uuid.New(),
				MerchantId:   "merchant-123",
				MerchantName: "Test Merchant",
				IsActive:     true,
				Priority:     1,
			},
		},
		Pagination: creditcardCoreProcessorModel.PaginationResponse{
			PageLimit:   10,
			PageNumber:  1,
			TotalRecord: 1,
			TotalPage:   1,
		},
	}

	testCases := []struct {
		name       string
		limit      int
		page       int
		merchantId string
		wantErr    bool
		mockSetup  func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository)
	}{
		{
			name:       "SUCCESS",
			limit:      10,
			page:       1,
			merchantId: "",
			wantErr:    false,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("GetMIDMapList", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.GetMIDMapListRequest")).Return(mockResponse, nil)
			},
		},
		{
			name:       "SUCCESS with merchantId filter",
			limit:      10,
			page:       1,
			merchantId: "merchant-123",
			wantErr:    false,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("GetMIDMapList", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.GetMIDMapListRequest")).Return(mockResponse, nil)
			},
		},
		{
			name:       "ERROR: Repository Error",
			limit:      10,
			page:       1,
			merchantId: "",
			wantErr:    true,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("GetMIDMapList", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.GetMIDMapListRequest")).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

			tc.mockSetup(mockRepo)

			svc := New(conf, mockLogger, nil, nil, nil, mockRepo)
			response, err := svc.GetMIDMapList(ctx, tc.limit, tc.page, tc.merchantId)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, int64(1), response.Meta.Page)
				assert.Equal(t, int64(10), response.Meta.PerPage)
				assert.Equal(t, int64(1), response.Meta.TotalPages)
				assert.Equal(t, int64(1), response.Meta.TotalItems)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateMID(t *testing.T) {
	conf := &config.Config{}
	ctx := context.Background()

	validRequest := &creditcardModel.CreateMIDRequest{
		Mid:                "TEST_MID_001",
		Name:               "Test MID 1",
		Type:               "REGULAR",
		Processor:          "HARSYA",
		PrincipalAvailable: []string{"VISA", "MASTERCARD"},
		IsActive:           true,
		IsDefault:          false,
		BaseURL:            "https://test.example.com",
		Password:           "test123",
	}

	mockResponse := &creditcardCoreProcessorModel.CreateMIDResponseData{
		Uuid:    uuid.New(),
		Created: true,
	}

	testCases := []struct {
		name      string
		request   *creditcardModel.CreateMIDRequest
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository)
	}{
		{
			name:    "SUCCESS",
			request: validRequest,
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("CreateMID", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.CreateMIDRequest")).Return(mockResponse, nil)
			},
		},
		{
			name:    "ERROR: Repository Error",
			request: validRequest,
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("CreateMID", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.CreateMIDRequest")).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

			tc.mockSetup(mockRepo)

			svc := New(conf, mockLogger, nil, nil, nil, mockRepo)
			err := svc.CreateMID(ctx, tc.request)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateMID(t *testing.T) {
	conf := &config.Config{}
	ctx := context.Background()

	validRequest := &creditcardModel.UpdateMIDRequest{
		Mid:                "TEST_MID_001",
		Name:               "Updated Test MID",
		Type:               "REGULAR",
		Processor:          "HARSYA",
		PrincipalAvailable: []string{"VISA", "MASTERCARD"},
		IsActive:           true,
		IsDefault:          false,
		BaseURL:            "https://test.example.com",
		Password:           "test123",
		UUID:               uuid.NewString(),
	}

	mockResponse := &creditcardCoreProcessorModel.UpdateMIDResponseData{
		Uuid:    uuid.New(),
		Updated: true,
	}

	testCases := []struct {
		name      string
		request   *creditcardModel.UpdateMIDRequest
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository)
	}{
		{
			name:    "SUCCESS",
			request: validRequest,
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("UpdateMID", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.UpdateMIDRequest")).Return(mockResponse, nil)
			},
		},
		{
			name:    "ERROR: Repository Error",
			request: validRequest,
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("UpdateMID", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.UpdateMIDRequest")).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

			tc.mockSetup(mockRepo)

			svc := New(conf, mockLogger, nil, nil, nil, mockRepo)
			err := svc.UpdateMID(ctx, tc.request)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetMIDDetail(t *testing.T) {
	conf := &config.Config{}
	ctx := context.Background()

	mockResponse := &creditcardCoreProcessorModel.MIDResponseData{
		Uuid: uuid.New(),
		Mid:  "TEST_MID_001",
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository)
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("GetMID", mock.Anything, mock.Anything).Return(mockResponse, nil)
			},
		},
		{
			name:    "ERROR: Repository Error",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("GetMID", mock.Anything, mock.Anything).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

			tc.mockSetup(mockRepo)

			svc := New(conf, mockLogger, nil, nil, nil, mockRepo)
			_, err := svc.GetMIDDetail(ctx, uuid.NewString())

			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestValidateMIDInstallmentBins(t *testing.T) {
	conf := &config.Config{}
	ctx := context.Background()
	request := &creditcardModel.ValidateMIDInstallmentBinsRequest{
		MidID: "TEST_MID_001",
		Bins:  []string{"1234567890123456"},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository)
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("ValidateMidInstallmentBins", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "ERROR: Repository Error",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.ICreditcardCoreProcessorRepository) {
				mockRepo.On("ValidateMidInstallmentBins", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

			tc.mockSetup(mockRepo)

			svc := New(conf, mockLogger, nil, nil, nil, mockRepo)
			err := svc.ValidateMIDInstallmentBins(ctx, request)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
