package adjustment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/adjustment"
	sqlPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateManualTopup(t *testing.T) {
	sqlMock := sqlPkgMock.NewIMySqlExt(t)

	repo := New(sqlMock)

	tests := []struct {
		mockSetup func(s *sqlPkgMock.IMySqlExt)
		wantErr   string
	}{
		{
			mockSetup: func(s *sqlPkgMock.IMySqlExt) {
				s.On(
					"NamedExecContext", constant.ValueCtxMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("*adjustment.ManualAdjustmentHistory"),
				).Once().Return(false, errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			mockSetup: func(s *sqlPkgMock.IMySqlExt) {
				s.On(
					"NamedExecContext", constant.ValueCtxMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("*adjustment.ManualAdjustmentHistory"),
				).Once().Return(true, nil)
			},
		},
	}
	for _, test := range tests {

		test.mockSetup(sqlMock)

		if err := repo.CreateAdjustment(context.Background(), &adjustment.ManualAdjustmentHistory{}); test.wantErr == "" {
			require.NoError(t, err)

		} else {
			require.Error(t, err)
			require.ErrorContains(t, err, test.wantErr)
		}
	}
}
