package merchant_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidateMerchantBODData(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)
	ptrBODValidationMockType := mock.AnythingOfType("*merchant.BODValidation")

	tests := []struct {
		name       string
		input      *merchant.UpsertBoardOfDirectorReq
		setupMock  func()
		wantErr    string
		wantResult *merchant.BODValidation
	}{
		{
			name: "ERROR:Some error action POST",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method: c.ActionPost,
			},
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrBODValidationMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS:Action POST/Valid",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method: c.ActionPost,
			},
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrBODValidationMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					args.Get(1).(*merchant.BODValidation).Valid = true
				}).Return(nil)
			},
			wantResult: &merchant.BODValidation{Valid: true, IsCreate: true},
		},
		{
			name: "SUCCESS:Action POST/Not Valid",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method:   c.ActionPost,
				Position: c.MerchantBODPositionShareholder,
				Name:     "name",
			},
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrBODValidationMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil)
			},
			wantResult: &merchant.BODValidation{IsCreate: true},
		},
		{
			name: "SUCCESS:Action PUT/Valid",
			input: &merchant.UpsertBoardOfDirectorReq{
				Method: c.ActionPut,
			},
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrBODValidationMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					args.Get(1).(*merchant.BODValidation).Valid = true
				}).Return(nil)
			},
			wantResult: &merchant.BODValidation{Valid: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			if valid, err := repo.ValidateMerchantBODData(context.Background(), test.input); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, valid)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestUpsertMerchantBOD(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	var shares float64 = 90

	ptrBoardOfDirectorMockType := mock.AnythingOfType("*merchant.BoardOfDirector")

	tests := []struct {
		name      string
		action    string
		data      *merchant.BoardOfDirector
		setupMock func(db *mySqlExtMock.IMySqlExt)
		wantErr   string
	}{
		{
			name:   "ERROR:Some error when create data",
			action: c.ActionPost,
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:   "SUCCESS:Create new data",
			action: c.ActionPost,
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Once().Return(true, nil)
			},
		},
		{
			name:   "SUCCESS:Update data",
			action: c.ActionPut,
			data: &merchant.BoardOfDirector{
				Id:   uuid.NewString(),
				Name: "Updated Name",
			},
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Return(true, nil)
			},
		},
		{
			name:   "SUCCESS:Update data with changes",
			action: c.ActionPut,
			data: &merchant.BoardOfDirector{
				Id:             uuid.NewString(),
				Position:       "new-position",
				Name:           "new-name",
				Shares:         &shares,
				IdentityNumber: "new-identity-number",
				IdentityFile:   []byte("new-identity-file"),
				Hash:           "new-hash",
				PositionLong:   "new-position-long",
				UpdatedAt:      time.Now(),
			},
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Return(true, nil)
			},
		},
		{
			name:   "SUCCESS:Update with empty IdentityFile (still updates)",
			action: c.ActionPut,
			data: &merchant.BoardOfDirector{
				Id:           uuid.NewString(),
				IdentityFile: []byte{}, // Empty but not nil, will have JSON representation
			},
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				// With len(b) > 0, even empty IdentityFile creates an update
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Return(true, nil)
			},
		},
		{
			name:   "SUCCESS:Update with nil IdentityFile (still creates update due to MarshalJSON)",
			action: c.ActionPut,
			data: &merchant.BoardOfDirector{
				Id: uuid.NewString(),
				// Even with nil IdentityFile, MarshalJSON() returns bytes with len > 0
			},
			setupMock: func(db *mySqlExtMock.IMySqlExt) {
				// Even nil IdentityFile triggers update with len(b) > 0
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), ptrBoardOfDirectorMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := mySqlExtMock.NewIMySqlExt(t)
			repo := New(db, mockLogger)

			test.setupMock(db)
			if test.data == nil {
				test.data = &merchant.BoardOfDirector{}
			}
			if err := repo.UpsertMerchantBOD(context.Background(), test.action, test.data); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}

	// Test case for early return when setQuery is empty (100% coverage)
	t.Run("SUCCESS:Update with no changes - early return via injection", func(t *testing.T) {
		db := mySqlExtMock.NewIMySqlExt(t)

		// Inject custom marshal func that returns empty bytes to trigger early return
		repo := New(db, mockLogger, WithJSONMarshalFunc(func(v interface{}) ([]byte, error) {
			return []byte{}, nil // Return empty bytes to make len(b) == 0
		}))

		data := &merchant.BoardOfDirector{
			Id: uuid.NewString(),
			// All other fields are zero values
		}

		// No mock setup needed - function should return early
		err := repo.UpsertMerchantBOD(context.Background(), c.ActionPut, data)
		require.NoError(t, err)
	})
}

func TestGetListMerchantBODs(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	result := []merchant.BoardOfDirectorResp{
		{
			Id:       uuid.NewString(),
			Name:     "John Wick",
			Position: "Director",
		},
	}
	ptrSliceBoardOfDirectorRespMockType := mock.AnythingOfType("*[]merchant.BoardOfDirectorResp")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult []merchant.BoardOfDirectorResp
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrSliceBoardOfDirectorRespMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrSliceBoardOfDirectorRespMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), ptrSliceBoardOfDirectorRespMockType, c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchant.BoardOfDirectorResp)) = result
				}).Return(nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			if resp, err := repo.GetListMerchantBODs(context.Background(), uuid.NewString()); test.wantErr == "" {
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
