package industry

import (
	"context"
	"errors"
	"testing"
	"time"

	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	sampleIndustry := &industryModel.Industry{
		UUID:           "uuid-1",
		ParentIndustry: "Technology",
		ChildIndustry:  "Software",
		RiskLevel:      "Medium",
		MCC:            "5734",
		CommonMCC:      "5734",
		UpdatedAt:      now,
	}

	testCases := []struct {
		name      string
		industry  *industryModel.Industry
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name:     "SUCCESS: Update industry",
			industry: sampleIndustry,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:     "ERROR: Database error",
			industry: sampleIndustry,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(false, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mysqlMock)

			repo := &repository{db: mysqlMock, logger: loggerMock}
			err := repo.Update(ctx, tc.industry)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}
