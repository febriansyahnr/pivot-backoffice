package merchant_test

import (
	"context"
	"fmt"
	"mime/multipart"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	gcsModel "github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUploadDocument(t *testing.T) {
	gcs := gcsMock.NewGCSService(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIMerchantRepository(t)

	service := New(repo, logger, nil, nil, nil, nil, WithGCSService(gcs), WithServiceConfig(&config.Config{}))

	traceId := uuid.NewString()
	documentId, sha256File := uuid.NewString(), "sha256:hashed-my-file"
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	ptrDocumentMockType := mock.AnythingOfType("*merchant.Document")
	merchantID := uuid.NewString()

	tests := []struct {
		name       string
		request    *merchant.UploadDocumentReq
		setupMock  func()
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("FM: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Merchant data not found",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: c.ErrMerchantIDNotValid.Error(),
		},
		{
			name: "ERROR:Find document by type",
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchant.Merchant{}, nil).Once()

				repo.On(
					"FindDocumentByType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("FD: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Upload file from multipart",
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchant.Merchant{}, nil).Once()
				repo.On(
					"FindDocumentByType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)

				gcs.On(
					"UploadFileFromMultipart", c.ValueCtxMockType(), c.StringMockType(), c.FileHeaderMockType(), c.BoolMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("UP: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Create document",
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchant.Merchant{}, nil).Once()
				gcs.On(
					"UploadFileFromMultipart", c.ValueCtxMockType(), c.StringMockType(), c.FileHeaderMockType(), c.BoolMockType(),
				).Return(&gcsModel.UploadMultipart{}, nil)

				repo.On(
					"FindDocumentByType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
				repo.On("CreateDocument", c.ValueCtxMockType(), ptrDocumentMockType).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf("UPS: "+c.InternalErrorFmt, traceId),
		},
		{
			name: "ERROR:Update document",
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchant.Merchant{}, nil).Once()
				repo.On(
					"FindDocumentByType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchant.Document{Hash: "sha256:xxx"}, nil)
				repo.On("UpdateDocument", c.ValueCtxMockType(), ptrDocumentMockType).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrInternalServerForUser).Error(),
		},
		{
			name: "SUCCESS: Update Document",
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).Return(&merchant.Merchant{}, nil).Once()
				repo.On(
					"FindDocumentByType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&merchant.Document{Id: documentId, Hash: sha256File}, nil).Once()
				repo.On("UpdateDocument", c.ValueCtxMockType(), ptrDocumentMockType).Return(nil)
			},
			wantResult: documentId,
		},
		{
			name: "Success:Create document with empty file",
			request: &merchant.UploadDocumentReq{
				MerchantId: merchantID,
				Notes:      "Notes",
			},
			setupMock: func() {
				repo.On("FindMerchantByID", c.ValueCtxMockType(), merchantID).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil).Once()

				repo.On(
					"FindDocumentByType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
				repo.On("CreateDocument", c.ValueCtxMockType(), ptrDocumentMockType).Once().Return(nil)
			},
			wantErr:    "",
			wantResult: documentId,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			request := &merchant.UploadDocumentReq{
				Hash:  sha256File,
				File:  &multipart.FileHeader{},
				Notes: "Notes",
			}

			if test.request != nil {
				request = test.request
			}

			id, err := service.UploadDocument(ctx, request)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Empty(t, id)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			// in create document, cannot validate the id because its difficutl to mock
			if test.name == "Success:Create document with empty file" {
				assert.NotEmpty(t, id)
				return
			}
			assert.Equal(t, test.wantResult, id)
		})
	}
}
func TestGetDocuments(t *testing.T) {
	repo := repoMocks.NewIMerchantRepository(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	gcs := gcsMock.NewGCSService(t)
	service := New(repo, logger, nil, nil, nil, nil, WithGCSService(gcs))
	ctx := context.Background()

	tests := []struct {
		name       string
		request    *merchant.MerchantDocumentFilterRequest
		setupMock  func()
		wantErr    error
		wantResult *commonModel.PaginationResponse
	}{
		{
			name: "ERROR: Failed to retrieve documents",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
				Page:       1,
				PerPage:    10,
			},
			setupMock: func() {
				repo.On(
					"GetDocuments", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.MerchantDocumentFilterRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS: Retrieved documents",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
				Page:       1,
				PerPage:    10,
			},
			setupMock: func() {
				repo.On(
					"GetDocuments", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.MerchantDocumentFilterRequest"),
				).Once().Return(&commonModel.PaginationResponse{
					Data: []*merchant.DocumentFilterResponse{
						{
							DocumentID: "doc1",
							Type:       "type1",
							Identifier: "number1",
							BucketName: nil,
							URL:        util.ValueToPtr("url1"),
							Status:     "status1",
							CreatedBy:  "user1",
						},
					},
					Meta: commonModel.Meta{
						TotalItems: 1,
						Page:       1,
						PerPage:    10,
						TotalPages: 1,
					},
				}, nil)
			},
			wantResult: &commonModel.PaginationResponse{
				Data: []*merchant.DocumentFilterResponse{
					{
						DocumentID: "doc1",
						Type:       "type1",
						Identifier: "number1",
						BucketName: nil,
						URL:        util.ValueToPtr("url1"),
						Status:     "status1",
						CreatedBy:  "user1",
					},
				},
				Meta: commonModel.Meta{
					TotalItems: 1,
					Page:       1,
					PerPage:    10,
					TotalPages: 1,
				},
			},
		},
		{
			name: "SUCCESS: when have bucket name and url, then should return valid url",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
				Page:       1,
				PerPage:    10,
			},
			setupMock: func() {
				repo.On(
					"GetDocuments", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.MerchantDocumentFilterRequest"),
				).Once().Return(&commonModel.PaginationResponse{
					Data: []*merchant.DocumentFilterResponse{
						{
							DocumentID: "doc1",
							Type:       "type1",
							Identifier: "number1",
							BucketName: util.ValueToPtr("bucket1"),
							URL:        util.ValueToPtr("url1"),
							Status:     "status1",
							CreatedBy:  "user1",
						},
					},
					Meta: commonModel.Meta{
						TotalItems: 1,
						Page:       1,
						PerPage:    10,
						TotalPages: 1,
					},
				}, nil)
				gcs.On(
					"CreateSignedURL", c.ValueCtxMockType(), "url1", time.Duration(5*time.Minute),
				).Once().Return("signed-url", nil)
			},
			wantResult: &commonModel.PaginationResponse{
				Data: []*merchant.DocumentFilterResponse{
					{
						DocumentID: "doc1",
						Type:       "type1",
						Identifier: "number1",
						BucketName: util.ValueToPtr("bucket1"),
						URL:        util.ValueToPtr("signed-url"),
						Status:     "status1",
						CreatedBy:  "user1",
					},
				},
				Meta: commonModel.Meta{
					TotalItems: 1,
					Page:       1,
					PerPage:    10,
					TotalPages: 1,
				},
			},
		},
		{
			name: "SUCCESS: when have only name, then should return nil url",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
				Page:       1,
				PerPage:    10,
			},
			setupMock: func() {
				repo.On(
					"GetDocuments", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.MerchantDocumentFilterRequest"),
				).Once().Return(&commonModel.PaginationResponse{
					Data: []*merchant.DocumentFilterResponse{
						{
							DocumentID: "doc1",
							Type:       "type1",
							Identifier: "number1",
							BucketName: util.ValueToPtr("bucket1"),
							URL:        util.ValueToPtr("url1"),
							Status:     "status1",
							CreatedBy:  "user1",
						},
					},
					Meta: commonModel.Meta{
						TotalItems: 1,
						Page:       1,
						PerPage:    10,
						TotalPages: 1,
					},
				}, nil)
				gcs.On(
					"CreateSignedURL", c.ValueCtxMockType(), "url1", time.Duration(5*time.Minute),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantResult: &commonModel.PaginationResponse{
				Data: []*merchant.DocumentFilterResponse{
					{
						DocumentID: "doc1",
						Type:       "type1",
						Identifier: "number1",
						BucketName: util.ValueToPtr("bucket1"),
						URL:        nil,
						Status:     "status1",
						CreatedBy:  "user1",
					},
				},
				Meta: commonModel.Meta{
					TotalItems: 1,
					Page:       1,
					PerPage:    10,
					TotalPages: 1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetDocuments(ctx, test.request)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
