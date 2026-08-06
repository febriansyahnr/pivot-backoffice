package callbackRepository

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	callback_model "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateCallbackLog(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	repo := New(db, logger)
	tests := []struct {
		name       string
		setupMocks func()
		wantErr    string
	}{
		{
			name: "ERROR:Some internal error",
			setupMocks: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*callback_model.CallbackLog"),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), mock.AnythingOfType("*callback_model.CallbackLog"),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if err := repo.UpdateCallbackLog(context.Background(), &callback_model.CallbackLog{}); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestUpdateCallbackURLById(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)
	tests := []struct {
		name       string
		setupMocks func()
		wantErr    string
	}{
		{
			name: "ERROR:Some internal error",
			setupMocks: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if err := repo.UpdateCallbackURLById(context.Background(), uuid.NewString(), "https://"); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
func TestUpdateCallbackBaseURLById(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)
	tests := []struct {
		name       string
		setupMocks func()
		wantErr    string
	}{
		{
			name: "ERROR:Some internal error",
			setupMocks: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if err := repo.UpdateCallbackBaseURLById(context.Background(), uuid.NewString(), "https://example.com"); test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
