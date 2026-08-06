package adjustment_test

import (
	"context"
	"database/sql"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	adjustModel "github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/adjustment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestFindByID(t *testing.T) {
	manualAdjustment := &adjustModel.ManualAdjustmentHistory{
		UUID:      "uuid-uuid-uuid",
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	testCases := []struct {
		name        string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		input       string
		expected    *adjustModel.ManualAdjustmentHistory
		expectedErr string
		wantErr     bool
	}{
		{
			name: "SUCCESS: Find Adjustment By ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeManualAdjustmentHistoryReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					adjustPtr := args.Get(1).(*adjustModel.ManualAdjustmentHistory)
					*adjustPtr = *manualAdjustment
				})
			},
			input:       manualAdjustment.UUID,
			expected:    manualAdjustment,
			expectedErr: "",
			wantErr:     false,
		},
		{
			name: "ERROR: Adjustment Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeManualAdjustmentHistoryReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			input:    manualAdjustment.UUID,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeManualAdjustmentHistoryReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input:       manualAdjustment.UUID,
			expected:    nil,
			expectedErr: "some error",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}

			tc.mockSetup(mysqlMock)
			repo := New(mysqlMock)
			ctx := context.Background()
			res, err := repo.FindByID(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, res)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}
