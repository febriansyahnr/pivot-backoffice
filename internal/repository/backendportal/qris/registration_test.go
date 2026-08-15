package qris_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/qris"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/qris"
	mySqlExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInitRegistration(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*qris.Registration"),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*qris.Registration"),
				).Return(true, nil)
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if err := repo.InitRegistration(context.Background(), &qris.Registration{}); test.wantErr == nil {
				assert.NoError(t, err)

			} else {
				require.Error(t, err)
				require.Equal(t, test.wantErr, err)
			}
		})
	}
}

func TestUpdateUploadedDocument(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)

	tests := []struct {
		name      string
		data      *qris.UpdateDocument
		setupMock func()
		wantErr   error
	}{
		{
			name:    "ERROR:Document not found",
			data:    &qris.UpdateDocument{},
			wantErr: errors.New("document type not found"),
		},
		{
			name: "ERROR:NationalIdentityCard",
			data: &qris.UpdateDocument{
				Type: "NationalIdentityCard",
			},
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:CertificateIncorporation",
			data: &qris.UpdateDocument{
				Type: "CertificateIncorporation",
			},
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			data: &qris.UpdateDocument{
				Type:   "CertificateIncorporation",
				Number: "123",
			},
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			if err := repo.UpdateUploadedDocument(context.Background(), uuid.NewString(), test.data); test.wantErr == nil {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				require.Equal(t, test.wantErr, err)
			}
		})
	}
}

func TestUpdateRegistrationStatus(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(true, nil)
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			if err := repo.UpdateRegistrationStatus(context.Background(), c.StatusSuccess, uuid.NewString()); test.wantErr == nil {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				require.Equal(t, test.wantErr, err)
			}
		})
	}
}

func TestFindQrRegistrationForValidationById(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)

	data := &qris.Registration{
		Id: uuid.NewString(),
	}
	ptrRegistrationMockType := mock.AnythingOfType("*qris.Registration")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult *qris.Registration
	}{
		{
			name: "ERROR:" + c.ErrSomeErrorForUnitTest.Error(),
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrRegistrationMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:data not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrRegistrationMockType, c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrRegistrationMockType, c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*qris.Registration) = *data
				}).Return(nil)
			},
			wantResult: data,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.FindQrRegistrationForValidationById(context.Background(), uuid.NewString())
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestUpdateCallbackQrRegistration(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)

	repo := New(db)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   string
	}{
		{
			name: "ERROR:" + c.ErrSomeErrorForUnitTest.Error(),
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.AnythingOfType("types.JSONText"), c.TimeMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.AnythingOfType("types.JSONText"), c.TimeMockType(), c.StringMockType(),
				).Once().Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.UpdateCallbackQrRegistration(context.Background(), uuid.NewString(), &qris.RegistrationCallback{})
			if test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestUpdateQrRegistration(t *testing.T) {
	db := mySqlExt.NewIMySqlExt(t)
	repo := New(db)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.AnythingOfType("time.Time"), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.AnythingOfType("time.Time"), c.StringMockType(),
				).Return(true, nil)
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.UpdateQrRegistration(context.Background(), "reg-id", "merchant-id", "terminal-id")
			if test.wantErr == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, test.wantErr, err)
			}
		})
	}
}
