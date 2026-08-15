package merchant_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mySqlExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindDocumentIdByType(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db, nil)

	documentId := uuid.NewString()

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrStringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrStringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), c.PtrStringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*string) = documentId
				}).Return(nil)
			},
			wantResult: documentId,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			id, err := repo.FindDocumentIdByType(context.Background(), uuid.NewString(), "NationalIdentityCard")
			if test.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)

			} else {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, id)
			}
		})
	}
}

func TestFindDocumentByType(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db, nil)

	fileLoc := types.JSONText(`{"bucket":"value-bucket","object":"value-object"}`)
	ptrDocumentMockType := mock.AnythingOfType("*merchant.Document")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult *merchant.Document
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrDocumentMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrDocumentMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrDocumentMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*merchant.Document) = merchant.Document{
						Location: fileLoc,
					}
				}).Return(nil)
			},
			wantResult: &merchant.Document{
				Location: fileLoc,
				ObjLocation: merchant.DocLocation{
					Bucket: "value-bucket",
					Object: "value-object",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			if result, err := repo.FindDocumentByType(context.Background(), uuid.NewString(), "doc-type"); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestCreateDocument(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db, nil)
	ptrMerchantDocMockType := mock.AnythingOfType("*merchant.Document")

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrMerchantDocMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrMerchantDocMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			if err := repo.CreateDocument(context.Background(), &merchant.Document{}); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestUpdateDocument(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*merchant.Document"),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*merchant.Document"),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			err := repo.UpdateDocument(context.Background(), &merchant.Document{})
			if test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetDocuments(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	tests := []struct {
		name      string
		request   *merchant.MerchantDocumentFilterRequest
		setupMock func(db *mySqlExt.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get documents with all filters",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID:     uuid.NewString(),
				DocumentType:   "NationalIdentityCard",
				Identifier:     "test-id",
				DocumentID:     uuid.NewString(),
				StartCreatedAt: &startDate,
				EndCreatedAt:   &endDate,
				Page:           1,
				PerPage:        10,
				SortBy:         "createdAt",
				Sort:           "desc",
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 5
				}).Return(nil).Once()

				// Mock select query
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get documents with DocumentType filter only",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID:   uuid.NewString(),
				DocumentType: "NationalIdentityCard",
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 2
				}).Return(nil).Once()

				// Mock select query
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get documents with Identifier filter",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
				Identifier: "test-identifier",
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 1
				}).Return(nil).Once()

				// Mock select query
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get documents with DocumentID filter",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
				DocumentID: uuid.NewString(),
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 1
				}).Return(nil).Once()

				// Mock select query
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get documents with StartCreatedAt only",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID:     uuid.NewString(),
				StartCreatedAt: &startDate,
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 3
				}).Return(nil).Once()

				// Mock select query
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get documents with EndCreatedAt only",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID:   uuid.NewString(),
				EndCreatedAt: &endDate,
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 3
				}).Return(nil).Once()

				// Mock select query
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything, mock.Anything,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get documents with custom pagination",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
				Page:       2,
				PerPage:    20,
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 50
				}).Return(nil).Once()

				// Mock select query
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get documents with custom sort",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
				SortBy:     "identifier",
				Sort:       "asc",
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 10
				}).Return(nil).Once()

				// Mock select query
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "ERROR: Count query fails",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query with error
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything,
				).Return(c.ErrSomeErrorForUnitTest).Once()

				// Mock select query (still needs to succeed for errgroup)
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything,
				).Return(nil).Once()
			},
			wantErr: true,
		},
		{
			name: "ERROR: Select query fails",
			request: &merchant.MerchantDocumentFilterRequest{
				MerchantID: uuid.NewString(),
			},
			setupMock: func(db *mySqlExt.IMySqlExt) {
				// Mock count query
				db.On(
					"GetContext", c.ValueCtxMockType(), mock.AnythingOfType("*int"), c.StringMockType(),
					mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 10
				}).Return(nil).Once()

				// Mock select query with error
				db.On(
					"SelectContext", c.ValueCtxMockType(), mock.AnythingOfType("*[]*merchant.DocumentFilterResponse"), c.StringMockType(),
					mock.Anything,
				).Return(c.ErrSomeErrorForUnitTest).Once()
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := mySqlExt.NewIMySqlExt(t)
			test.setupMock(db)

			repo := New(db, mockLogger)
			result, err := repo.GetDocuments(context.Background(), test.request)

			if test.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Data)
				assert.NotNil(t, result.Meta)
			}
		})
	}
}
