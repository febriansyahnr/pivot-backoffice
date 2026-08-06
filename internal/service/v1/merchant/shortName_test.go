package merchant_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	merchantService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	redisExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMerchantShortNameValidation(t *testing.T) {
	redisClient := redisExtMock.NewIRedisExt(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIMerchantRepository(t)

	service := merchantService.New(repo, logger, nil, nil, nil, nil, merchantService.WithRedisClient(redisClient), merchantService.WithServiceConfig(&config.Config{}))

	// Get the concrete service to access the methods
	concreteService := service.(*merchantService.MerchantService)

	traceId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	shortName := "TEST"
	parentMerchantID := "merchant-123"

	tests := []struct {
		name       string
		shortName  string
		merchantID string
		setupMock  func()
		wantErr    string
	}{
		{
			name:       "ERROR:Empty short name",
			shortName:  "",
			merchantID: parentMerchantID,
			setupMock:  func() {},
			wantErr:    c.ErrMerchantShortNameInvalid.Error(),
		},
		{
			name:       "ERROR:CheckOrSetReservedShortNames fails",
			shortName:  shortName,
			merchantID: parentMerchantID,
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(0, c.ErrSomeErrorForUnitTest)).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:       "ERROR:HGetScan fails",
			shortName:  shortName,
			merchantID: parentMerchantID,
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				redisClient.On("HGetScan", c.ValueCtxMockType(), c.MerchantReservedShortNameKey, "TEST", mock.AnythingOfType("*string")).Return(c.ErrSomeErrorForUnitTest).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:       "SUCCESS:Short name not reserved",
			shortName:  shortName,
			merchantID: parentMerchantID,
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				redisClient.On("HGetScan", c.ValueCtxMockType(), c.MerchantReservedShortNameKey, "TEST", mock.AnythingOfType("*string")).Return(redisExt.ErrNil).Once()
			},
			wantErr: "",
		},
		{
			name:       "SUCCESS:Merchant allowed to use reserved short name",
			shortName:  shortName,
			merchantID: parentMerchantID,
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				redisClient.On("HGetScan", c.ValueCtxMockType(), c.MerchantReservedShortNameKey, "TEST", mock.AnythingOfType("*string")).Run(func(args mock.Arguments) {
					val := args.Get(3).(*string)
					*val = "merchant-123,merchant-456"
				}).Return(nil).Once()
			},
			wantErr: "",
		},
		{
			name:       "ERROR:Merchant not allowed to use reserved short name",
			shortName:  shortName,
			merchantID: "other-merchant",
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
				redisClient.On("HGetScan", c.ValueCtxMockType(), c.MerchantReservedShortNameKey, "TEST", mock.AnythingOfType("*string")).Run(func(args mock.Arguments) {
					val := args.Get(3).(*string)
					*val = "merchant-123,merchant-456"
				}).Return(nil).Once()
			},
			wantErr: c.ErrMerchantReservedShortName.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := concreteService.MerchantShortNameValidation(ctx, test.shortName, test.merchantID)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCheckOrSetReservedShortNames(t *testing.T) {
	redisClient := redisExtMock.NewIRedisExt(t)
	gcs := gcsMock.NewGCSService(t)
	excel := xlsxMock.NewExceler(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIMerchantRepository(t)

	cfg := &config.Config{
		GCSConfig: config.GCSConfig{
			MerchantReservedSortName: "reserved-names",
			ServiceBucketName:        "test-bucket",
		},
	}

	service := merchantService.New(repo, logger, nil, nil, nil, nil, merchantService.WithRedisClient(redisClient), merchantService.WithGCSService(gcs), merchantService.WithExcelLibrary(excel), merchantService.WithServiceConfig(cfg))
	concreteService := service.(*merchantService.MerchantService)

	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:Redis Exists fails",
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(0, c.ErrSomeErrorForUnitTest)).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS:Key already exists",
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(1, nil)).Once()
			},
			wantErr: "",
		},
		{
			name: "ERROR:ReadReservedShortName fails",
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(0, nil)).Once()
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return(nil, c.ErrSomeErrorForUnitTest).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:SetReservedSubMerchantShortName fails",
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(0, nil)).Once()
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return([]byte("file-content"), nil).Once()

				mockFile := xlsxMock.NewFiler(t)
				excel.On("OpenReader", mock.AnythingOfType("*bytes.Buffer")).Return(mockFile, nil).Once()
				mockFile.On("Close").Return(nil).Once()
				mockFile.On("GetRows", "data", xlsx.Options{RawCellValue: true}).Return([][]string{
					{"header1", "header2", "header3"},
					{"TEST1", "BANK", "merchant-1,merchant-2"},
				}, nil).Once()

				redisClient.On("HSet", c.ValueCtxMockType(), c.MerchantReservedShortNameKey, "TEST1", "merchant-1,merchant-2").Return(redis.NewIntResult(0, c.ErrSomeErrorForUnitTest)).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS:Full flow",
			setupMock: func() {
				redisClient.On("Exists", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(0, nil)).Once()
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return([]byte("file-content"), nil).Once()

				mockFile := xlsxMock.NewFiler(t)
				excel.On("OpenReader", mock.AnythingOfType("*bytes.Buffer")).Return(mockFile, nil).Once()
				mockFile.On("Close").Return(nil).Once()
				mockFile.On("GetRows", "data", xlsx.Options{RawCellValue: true}).Return([][]string{
					{"header1", "header2", "header3"},
					{"TEST1", "BANK", "merchant-1,merchant-2"},
				}, nil).Once()

				redisClient.On("HSet", c.ValueCtxMockType(), c.MerchantReservedShortNameKey, "TEST1", "merchant-1,merchant-2").Return(redis.NewIntResult(1, nil)).Once()
			},
			wantErr: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := concreteService.CheckOrSetReservedShortNames(ctx)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSetReservedSubMerchantShortName(t *testing.T) {
	redisClient := redisExtMock.NewIRedisExt(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIMerchantRepository(t)

	service := merchantService.New(repo, logger, nil, nil, nil, nil, merchantService.WithRedisClient(redisClient))
	concreteService := service.(*merchantService.MerchantService)

	ctx := context.Background()

	tests := []struct {
		name      string
		request   []merchant.ReservedMerchantShortNameItem
		setupMock func()
		wantErr   string
	}{
		{
			name:    "SUCCESS:Empty request",
			request: []merchant.ReservedMerchantShortNameItem{},
			setupMock: func() {
				redisClient.On("HSet", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(0, nil)).Once()
			},
			wantErr: "",
		},
		{
			name:    "ERROR:HSet fails for empty request",
			request: []merchant.ReservedMerchantShortNameItem{},
			setupMock: func() {
				redisClient.On("HSet", c.ValueCtxMockType(), c.MerchantReservedShortNameKey).Return(redis.NewIntResult(0, c.ErrSomeErrorForUnitTest)).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS:With data",
			request: []merchant.ReservedMerchantShortNameItem{
				{
					ShortName:        "TEST1",
					AllowedMerchants: []string{"merchant-1", "merchant-2"},
				},
				{
					ShortName:        "TEST2",
					AllowedMerchants: []string{"merchant-3"},
				},
			},
			setupMock: func() {
				redisClient.On("HSet", c.ValueCtxMockType(), c.MerchantReservedShortNameKey, "TEST1", "merchant-1,merchant-2", "TEST2", "merchant-3").Return(redis.NewIntResult(2, nil)).Once()
			},
			wantErr: "",
		},
		{
			name: "ERROR:HSet fails with data",
			request: []merchant.ReservedMerchantShortNameItem{
				{
					ShortName:        "TEST1",
					AllowedMerchants: []string{"merchant-1", "merchant-2"},
				},
			},
			setupMock: func() {
				redisClient.On("HSet", c.ValueCtxMockType(), c.MerchantReservedShortNameKey, "TEST1", "merchant-1,merchant-2").Return(redis.NewIntResult(0, c.ErrSomeErrorForUnitTest)).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := concreteService.SetReservedSubMerchantShortName(ctx, test.request)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestReadReservedShortName(t *testing.T) {
	gcs := gcsMock.NewGCSService(t)
	excel := xlsxMock.NewExceler(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIMerchantRepository(t)

	cfg := &config.Config{
		GCSConfig: config.GCSConfig{
			MerchantReservedSortName: "reserved-names",
			ServiceBucketName:        "test-bucket",
		},
	}

	service := merchantService.New(repo, logger, nil, nil, nil, nil, merchantService.WithGCSService(gcs), merchantService.WithExcelLibrary(excel), merchantService.WithServiceConfig(cfg))
	concreteService := service.(*merchantService.MerchantService)

	ctx := context.Background()

	tests := []struct {
		name       string
		setupMock  func()
		wantResult []merchant.ReservedMerchantShortNameItem
		wantErr    string
	}{
		{
			name: "ERROR:GCS ReadAll fails",
			setupMock: func() {
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return(nil, c.ErrSomeErrorForUnitTest).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:Excel OpenReader fails",
			setupMock: func() {
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return([]byte("invalid-excel"), nil).Once()
				excel.On("OpenReader", mock.AnythingOfType("*bytes.Buffer")).Return(nil, c.ErrSomeErrorForUnitTest).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:GetRows fails",
			setupMock: func() {
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return([]byte("file-content"), nil).Once()

				mockFile := xlsxMock.NewFiler(t)
				excel.On("OpenReader", mock.AnythingOfType("*bytes.Buffer")).Return(mockFile, nil).Once()
				mockFile.On("Close").Return(nil).Once()
				mockFile.On("GetRows", "data", xlsx.Options{RawCellValue: true}).Return(nil, c.ErrSomeErrorForUnitTest).Once()
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS:Empty file (less than 2 rows)",
			setupMock: func() {
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return([]byte("file-content"), nil).Once()

				mockFile := xlsxMock.NewFiler(t)
				excel.On("OpenReader", mock.AnythingOfType("*bytes.Buffer")).Return(mockFile, nil).Once()
				mockFile.On("Close").Return(nil).Once()
				mockFile.On("GetRows", "data", xlsx.Options{RawCellValue: true}).Return([][]string{
					{"header1", "header2", "header3"},
				}, nil).Once()
			},
			wantResult: []merchant.ReservedMerchantShortNameItem{},
			wantErr:    "",
		},
		{
			name: "SUCCESS:With data (2 columns)",
			setupMock: func() {
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return([]byte("file-content"), nil).Once()

				mockFile := xlsxMock.NewFiler(t)
				excel.On("OpenReader", mock.AnythingOfType("*bytes.Buffer")).Return(mockFile, nil).Once()
				mockFile.On("Close").Return(nil).Once()
				mockFile.On("GetRows", "data", xlsx.Options{RawCellValue: true}).Return([][]string{
					{"header1", "header2"},
					{"TEST1", "bank"},
				}, nil).Once()
			},
			wantResult: []merchant.ReservedMerchantShortNameItem{
				{
					ShortName:        "TEST1",
					AllowedMerchants: []string{},
				},
			},
			wantErr: "",
		},
		{
			name: "SUCCESS:With data (3 columns)",
			setupMock: func() {
				gcs.On("ReadAll", c.ValueCtxMockType(), "test-bucket", "reserved-names/active_reserved_shortname.xlsx").Return([]byte("file-content"), nil).Once()

				mockFile := xlsxMock.NewFiler(t)
				excel.On("OpenReader", mock.AnythingOfType("*bytes.Buffer")).Return(mockFile, nil).Once()
				mockFile.On("Close").Return(nil).Once()
				mockFile.On("GetRows", "data", xlsx.Options{RawCellValue: true}).Return([][]string{
					{"header1", "header2", "header3"},
					{"TEST1", "bank", "merchant-1, merchant-2"},
					{"TEST2", "bank", "merchant-3"},
				}, nil).Once()
			},
			wantResult: []merchant.ReservedMerchantShortNameItem{
				{
					ShortName:        "TEST1",
					AllowedMerchants: []string{"merchant-1", "merchant-2"},
				},
				{
					ShortName:        "TEST2",
					AllowedMerchants: []string{"merchant-3"},
				},
			},
			wantErr: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := concreteService.ReadReservedShortName(ctx)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
