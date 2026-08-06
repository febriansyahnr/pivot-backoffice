package basicsql_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	mysqlPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBeginTransaction(t *testing.T) {
	mySQLMock := mysqlPkgMock.NewIMySqlExt(t)

	ctx, span := otel.Tracer("PropertiesRepository").Start(context.Background(), "internal/repository/v1/basicsql/BeginTransaction")
	defer span.End()

	repo := NewBasicSQLProperties(mySQLMock)

	tests := []struct {
		name      string
		mockSetup func(db *mysqlPkgMock.IMySqlExt)
		wantErr   string
	}{
		{
			name: "ERROR:Invalid session",
			mockSetup: func(db *mysqlPkgMock.IMySqlExt) {
				db.On(
					"BeginTxx", mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Return(nil, errors.New("Invalid db session")).Once()
			},
			wantErr: "Invalid db session",
		},
		{
			name: "SUCCESS",
			mockSetup: func(db *mysqlPkgMock.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), mySqlExt.CtxSqlTx, tx)
				db.On(
					"BeginTxx", constant.ValueCtxMockType(),
				).Return(ctx, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(mySQLMock)

			if _, err := repo.BeginTransaction(ctx); test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)

			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestCommitTransaction(t *testing.T) {
	mySQLMock := mysqlPkgMock.NewIMySqlExt(t)

	repo := NewBasicSQLProperties(mySQLMock)

	tests := []struct {
		name      string
		mockSetup func(db *mysqlPkgMock.IMySqlExt)
		wantErr   string
	}{
		{
			name: "ERROR:Invalid session",
			mockSetup: func(db *mysqlPkgMock.IMySqlExt) {
				db.On(
					"Commit", constant.ValueCtxMockType(),
				).Once().Return(errors.New("Invalid db session"))
			},
			wantErr: "Invalid db session",
		},
		{
			name: "SUCCESS",
			mockSetup: func(db *mysqlPkgMock.IMySqlExt) {
				db.On(
					"Commit", constant.ValueCtxMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(mySQLMock)

			if err := repo.CommitTransaction(context.Background()); test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestRollbackTransaction(t *testing.T) {
	mySQLMock := mysqlPkgMock.NewIMySqlExt(t)

	repo := NewBasicSQLProperties(mySQLMock)

	tests := []struct {
		name      string
		mockSetup func(db *mysqlPkgMock.IMySqlExt)
		wantErr   string
	}{
		{
			name: "ERROR:Invalid session",
			mockSetup: func(db *mysqlPkgMock.IMySqlExt) {
				db.On(
					"Rollback", constant.ValueCtxMockType(),
				).Once().Return(errors.New("Invalid db session"))
			},
			wantErr: "Invalid db session",
		},
		{
			name: "SUCCESS",
			mockSetup: func(db *mysqlPkgMock.IMySqlExt) {
				db.On(
					"Rollback", constant.ValueCtxMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockSetup(mySQLMock)

			if err := repo.RollbackTransaction(context.Background()); test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
