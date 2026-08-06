package tnc_test

import (
	"context"
	"database/sql"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/tnc"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTNCVersionByID(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrTNCMockType := mock.AnythingOfType("*tnc.TNC")

	repo := New(logger, db)

	tests := []struct {
		name    string
		id      string
		setup   func()
		wantErr string
		wantNil bool
	}{
		{
			name: "SUCCESS: returns version by id",
			id:   "tnc-uuid",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType(), c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						v := args[1].(*tncModel.TNC)
						v.UUID = "tnc-uuid"
						v.Version = "1.0.0"
					}).Return(nil)
			},
		},
		{
			name: "SUCCESS: no rows returns nil",
			id:   "missing-uuid",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType(), c.StringMockType()).
					Once().Return(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name: "ERROR: database failure",
			id:   "tnc-uuid",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType(), c.StringMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
			wantNil: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			got, err := repo.GetTNCVersionByID(context.Background(), test.id)
			if test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
			if test.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, test.id, got.UUID)
			}
		})
	}
}

func TestGetTNCVersionByVersion(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrTNCMockType := mock.AnythingOfType("*tnc.TNC")

	repo := New(logger, db)

	tests := []struct {
		name    string
		version string
		setup   func()
		wantErr string
		wantNil bool
	}{
		{
			name:    "SUCCESS: returns version by version string",
			version: "1.0.0",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType(), c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						v := args[1].(*tncModel.TNC)
						v.UUID = "tnc-uuid"
						v.Version = "1.0.0"
					}).Return(nil)
			},
		},
		{
			name:    "SUCCESS: no rows returns nil",
			version: "9.9.9",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType(), c.StringMockType()).
					Once().Return(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name:    "ERROR: database failure",
			version: "1.0.0",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType(), c.StringMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
			wantNil: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			got, err := repo.GetTNCVersionByVersion(context.Background(), test.version)
			if test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
			if test.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, test.version, got.Version)
			}
		})
	}
}

func TestListTNCVersions(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrIntMockType := mock.AnythingOfType("*int")
	ptrSliceTNCMockType := mock.AnythingOfType("*[]*tnc.TNC")
	int64MockType := mock.AnythingOfType("int64")

	repo := New(logger, db)

	tests := []struct {
		name    string
		setup   func()
		wantErr string
	}{
		{
			name: "SUCCESS: returns versions with total",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						total := args[1].(*int)
						*total = 2
					}).Return(nil)
				db.On("SelectContext", c.ValueCtxMockType(), ptrSliceTNCMockType, c.StringMockType(), int64MockType, int64MockType).
					Once().
					Run(func(args mock.Arguments) {
						v := args[1].(*[]*tncModel.TNC)
						*v = []*tncModel.TNC{
							{UUID: "v1", Version: "1.0.0"},
							{UUID: "v2", Version: "1.1.0"},
						}
					}).Return(nil)
			},
		},
		{
			name: "ERROR: count query fails",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR: data query fails",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						total := args[1].(*int)
						*total = 2
					}).Return(nil)
				db.On("SelectContext", c.ValueCtxMockType(), ptrSliceTNCMockType, c.StringMockType(), int64MockType, int64MockType).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			versions, total, err := repo.ListTNCVersions(context.Background(), &tncModel.TNCVersionQuery{Page: 1, PageSize: 10})
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, 2, total)
				require.Len(t, versions, 2)
				assert.Equal(t, "v1", versions[0].UUID)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				assert.Nil(t, versions)
				assert.Equal(t, 0, total)
			}
		})
	}
}

func TestGetActiveTNCVersion(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrTNCMockType := mock.AnythingOfType("*tnc.TNC")

	repo := New(logger, db)

	tests := []struct {
		name    string
		setup   func()
		wantErr string
		wantNil bool
	}{
		{
			name: "SUCCESS: returns active version",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						v := args[1].(*tncModel.TNC)
						v.UUID = "tnc-uuid"
						v.Version = "1.0.0"
					}).Return(nil)
			},
		},
		{
			name: "SUCCESS: no rows returns nil",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType()).
					Once().Return(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name: "ERROR: database failure",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrTNCMockType, c.StringMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
			wantNil: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			got, err := repo.GetActiveTNCVersion(context.Background())
			if test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
			if test.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, "1.0.0", got.Version)
			}
		})
	}
}

func TestGetLatestSigningByMerchant(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrHistoryMockType := mock.AnythingOfType("*tnc.MerchantTNCSigningHistory")

	repo := New(logger, db)

	tests := []struct {
		name       string
		merchantID string
		setup      func()
		wantErr    string
		wantNil    bool
	}{
		{
			name:       "SUCCESS: returns latest signing",
			merchantID: "merchant-1",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrHistoryMockType, c.StringMockType(), c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						h := args[1].(*tncModel.MerchantTNCSigningHistory)
						h.UUID = "hist-1"
						h.MerchantID = "merchant-1"
						h.Version = "1.0.0"
					}).Return(nil)
			},
		},
		{
			name:       "SUCCESS: no rows returns nil",
			merchantID: "merchant-1",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrHistoryMockType, c.StringMockType(), c.StringMockType()).
					Once().Return(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name:       "ERROR: database failure",
			merchantID: "merchant-1",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrHistoryMockType, c.StringMockType(), c.StringMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
			wantNil: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			got, err := repo.GetLatestSigningByMerchant(context.Background(), test.merchantID)
			if test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
			if test.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, test.merchantID, got.MerchantID)
			}
		})
	}
}

func TestGetSigningByMerchantAndVersion(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrHistoryMockType := mock.AnythingOfType("*tnc.MerchantTNCSigningHistory")

	repo := New(logger, db)

	tests := []struct {
		name       string
		merchantID string
		version    string
		setup      func()
		wantErr    string
		wantNil    bool
	}{
		{
			name:       "SUCCESS: returns signing",
			merchantID: "merchant-1",
			version:    "1.0.0",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrHistoryMockType, c.StringMockType(), c.StringMockType(), c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						h := args[1].(*tncModel.MerchantTNCSigningHistory)
						h.UUID = "hist-1"
						h.MerchantID = "merchant-1"
						h.Version = "1.0.0"
					}).Return(nil)
			},
		},
		{
			name:       "SUCCESS: no rows returns nil",
			merchantID: "merchant-1",
			version:    "1.0.0",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrHistoryMockType, c.StringMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name:       "ERROR: database failure",
			merchantID: "merchant-1",
			version:    "1.0.0",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrHistoryMockType, c.StringMockType(), c.StringMockType(), c.StringMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
			wantNil: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			got, err := repo.GetSigningByMerchantAndVersion(context.Background(), test.merchantID, test.version)
			if test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
			if test.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, test.merchantID, got.MerchantID)
				assert.Equal(t, test.version, got.Version)
			}
		})
	}
}

func TestListSigningHistories(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	ptrIntMockType := mock.AnythingOfType("*int")
	ptrSliceHistoryMockType := mock.AnythingOfType("*[]*tnc.MerchantTNCSigningHistory")
	int64MockType := mock.AnythingOfType("int64")

	repo := New(logger, db)

	tests := []struct {
		name    string
		setup   func()
		wantErr string
	}{
		{
			name: "SUCCESS: returns histories with total",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						total := args[1].(*int)
						*total = 2
					}).Return(nil)
				db.On("SelectContext", c.ValueCtxMockType(), ptrSliceHistoryMockType, c.StringMockType(), int64MockType, int64MockType).
					Once().
					Run(func(args mock.Arguments) {
						h := args[1].(*[]*tncModel.MerchantTNCSigningHistory)
						*h = []*tncModel.MerchantTNCSigningHistory{
							{UUID: "h1", MerchantID: "m1"},
							{UUID: "h2", MerchantID: "m1"},
						}
					}).Return(nil)
			},
		},
		{
			name: "ERROR: count query fails",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType()).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR: data query fails",
			setup: func() {
				db.On("GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType()).
					Once().
					Run(func(args mock.Arguments) {
						total := args[1].(*int)
						*total = 2
					}).Return(nil)
				db.On("SelectContext", c.ValueCtxMockType(), ptrSliceHistoryMockType, c.StringMockType(), int64MockType, int64MockType).
					Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			histories, total, err := repo.ListSigningHistories(context.Background(), &tncModel.SigningHistoryQuery{Page: 1, PageSize: 10})
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, 2, total)
				require.Len(t, histories, 2)
				assert.Equal(t, "h1", histories[0].UUID)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				assert.Nil(t, histories)
				assert.Equal(t, 0, total)
			}
		})
	}
}
