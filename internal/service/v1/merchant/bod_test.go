package merchant_test

import (
	"context"
	"fmt"
	"mime/multipart"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	redisPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	gcsModel "github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpsertMerchantBOD(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	gcs := gcsMock.NewGCSService(t)
	redis := redisPkgMock.NewIRedisExt(t)
	repo := repoMocks.NewIMerchantRepository(t)
	cfg := &config.Config{GCSConfig: config.GCSConfig{MerchantDocumentFolderName: "merchants"}}

	service := New(
		repo, logger, nil, nil, nil, nil, WithServiceConfig(cfg), WithGCSService(gcs), WithRedisClient(redis),
	)

	bodId := uuid.NewString()
	traceId := uuid.NewString()
	dummyHash := "sha256:xxx"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)
	ptrBoardOfDirectorMockType := mock.AnythingOfType("*merchant.BoardOfDirector")
	ptrUpsertBoardOfDirectorReqMockType := mock.AnythingOfType("*merchant.UpsertBoardOfDirectorReq")

	tests := []struct {
		name       string
		input      *merchant.UpsertBoardOfDirectorReq
		setupMock  func()
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("FM: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Once().Return(nil, nil)
			},
			wantErr: c.ErrMerchantIDNotValid.Error(),
		},
		{
			name: "ERROR:Validate merchant bod data",
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchant.Merchant{}, nil)

				repo.On(
					"ValidateMerchantBODData", c.ValueCtxMockType(), ptrUpsertBoardOfDirectorReqMockType,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("VLD: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Invalid data/Action POST",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method: c.ActionPost,
			},
			setupMock: func() {
				repo.On(
					"ValidateMerchantBODData", c.ValueCtxMockType(), ptrUpsertBoardOfDirectorReqMockType,
				).Once().Return(&merchant.BODValidation{Valid: false}, nil)
			},
			wantErr: c.ErrDuplicateData.Error(),
		},
		{
			name: "ERROR:Invalid data/Action PUT",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method: c.ActionPut,
			},
			setupMock: func() {
				repo.On(
					"ValidateMerchantBODData", c.ValueCtxMockType(), ptrUpsertBoardOfDirectorReqMockType,
				).Once().Return(&merchant.BODValidation{Valid: false}, nil)
			},
			wantErr: c.ErrDataNotFound.Error(),
		},
		{
			name: "ERROR:Upload file from multipart",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method:       c.ActionPut,
				IdentityFile: &multipart.FileHeader{},
				Hash:         dummyHash,
			},
			setupMock: func() {
				repo.On(
					"ValidateMerchantBODData", c.ValueCtxMockType(), ptrUpsertBoardOfDirectorReqMockType,
				).Once().Return(&merchant.BODValidation{Valid: true, IsCreate: false}, nil)

				gcs.On(
					"UploadFileFromMultipart", c.ValueCtxMockType(), c.StringMockType(), c.PtrFileHeaderMockType(), c.BoolMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("UP: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Upsert merchant bod",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method:       c.ActionPost,
				IdentityFile: &multipart.FileHeader{},
				Hash:         dummyHash,
			},
			setupMock: func() {
				gcs.On(
					"UploadFileFromMultipart", c.ValueCtxMockType(), c.StringMockType(), c.PtrFileHeaderMockType(), c.BoolMockType(),
				).Return(&gcsModel.UploadMultipart{}, nil)

				repo.On(
					"ValidateMerchantBODData", c.ValueCtxMockType(), ptrUpsertBoardOfDirectorReqMockType,
				).Once().Return(&merchant.BODValidation{Valid: true, IsCreate: true}, nil)
				repo.On(
					"UpsertMerchantBOD", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("BOD: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "SUCCESS:Update data",
			input: &merchant.UpsertBoardOfDirectorReq{
				Id:           bodId,
				Method:       c.ActionPut,
				IdentityFile: &multipart.FileHeader{},
				Hash:         dummyHash,
			},
			setupMock: func() {
				repo.On(
					"ValidateMerchantBODData", c.ValueCtxMockType(), ptrUpsertBoardOfDirectorReqMockType,
				).Once().Return(&merchant.BODValidation{Valid: true, IsCreate: false, Hash: dummyHash}, nil)
				repo.On(
					"UpsertMerchantBOD", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Once().Return(nil)
				redis.On(
					"Del", c.ValueCtxMockType(), c.StringMockType(),
				).Return(nil)
			},
			wantResult: bodId,
		},
		{
			name: "SUCCESS:Create data",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method:       c.ActionPost,
				IdentityFile: &multipart.FileHeader{},
				Hash:         dummyHash,
			},
			setupMock: func() {
				repo.On(
					"ValidateMerchantBODData", c.ValueCtxMockType(), ptrUpsertBoardOfDirectorReqMockType,
				).Once().Return(&merchant.BODValidation{Valid: true, IsCreate: true}, nil)
				repo.On(
					"UpsertMerchantBOD", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Run(func(args mock.Arguments) {
					args.Get(2).(*merchant.BoardOfDirector).Id = bodId
				}).Return(nil)
			},
			wantResult: bodId,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if test.input == nil {
				test.input = &merchant.UpsertBoardOfDirectorReq{}
			}
			if resp, err := service.UpsertMerchantBOD(ctx, test.input); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.Equal(t, test.wantResult, resp)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetListMerchantBODs(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	gcs := gcsMock.NewGCSService(t)
	db, clientMock := redismock.NewClientMock()

	repo := repoMocks.NewIMerchantRepository(t)
	cfg := &config.Config{GCSConfig: config.GCSConfig{MerchantDocumentFolderName: "merchants"}}

	service := New(
		repo, logger, nil, nil, nil, nil, WithServiceConfig(cfg), WithGCSService(gcs), WithRedisClient(redisExt.WrapRedisClient(db, nil)),
	)

	traceId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	signedURL := "http://storage.com/yourfile.jpg"
	patternRedisKey := "backend-portal:signed-urls:merchants"

	boardOfDirectors := []merchant.BoardOfDirectorResp{
		{
			Id:           uuid.NewString(),
			Name:         "John Wick",
			Position:     "Director",
			File:         []byte(`{"object":"/storage/yourfile.jpg"}`),
			IdentityFile: signedURL,
			Shares:       10,
		},
		{
			Id:           uuid.NewString(),
			Name:         "Henru",
			Position:     "Commissioner",
			File:         []byte(`{"object":"/storage/yourfile.jpg"}`),
			IdentityFile: signedURL,
			Shares:       20,
		},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult []merchant.BoardOfDirectorResp
	}{
		{
			name: "ERROR:Get list merchant bods",
			setupMock: func() {
				repo.On(
					"GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("GLS: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				repo.On(
					"GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantResult: nil,
		},
		{
			name: "ERROR:Unmarshal JSON",
			setupMock: func() {
				repo.On(
					"GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return([]merchant.BoardOfDirectorResp{{File: []byte(`B`)}}, nil)

				clientMock.Regexp().ExpectGet(patternRedisKey).RedisNil()
			},
			wantErr: fmt.Sprintf("UNM: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Create signed URL",
			setupMock: func() {
				clientMock.ClearExpect()
				clientMock.Regexp().ExpectGet(patternRedisKey).SetVal("")

				repo.On("GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType()).Return(boardOfDirectors, nil).Once()

				gcs.On(
					"CreateSignedURL", c.ValueCtxMockType(), c.StringMockType(), c.DurationMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("GEN: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				clientMock.ClearExpect()
				clientMock.Regexp().ExpectGet(patternRedisKey).SetVal("")

				repo.On("GetListMerchantBODs", c.ValueCtxMockType(), c.StringMockType()).Return(boardOfDirectors, nil).Once()

				gcs.On(
					"CreateSignedURL", c.ValueCtxMockType(), c.StringMockType(), c.DurationMockType(),
				).Return(signedURL+" s", nil)

			},
			wantResult: boardOfDirectors,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			if resp, err := service.GetListMerchantBODs(ctx, uuid.NewString()); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.Equal(t, test.wantResult, resp)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
