package fraudruleservice

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IFraudRulesRepository)
		uuid    string
		wantErr bool
		errType string
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				// Mock GetByID to return a valid fraud rule
				repo.On("GetByID", mock.Anything, "01967b9c-6217-77e3-a754-531365a6f5f2").Return(&fraudrulesmodel.FraudRules{
					UUID:     "01967b9c-6217-77e3-a754-531365a6f5f2",
					RuleName: "Test Rule",
				}, nil)

				// Mock Delete to return success
				repo.On("Delete", mock.Anything, "01967b9c-6217-77e3-a754-531365a6f5f2").Return(nil)
			},
			uuid:    "01967b9c-6217-77e3-a754-531365a6f5f2",
			wantErr: false,
		},
		{
			name: "ERROR: Failed to get fraud rule",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				// Mock GetByID to return an error
				repo.On("GetByID", mock.Anything, "01967b9c-6217-77e3-a754-531365a6f5f2").Return(nil, errors.New("database error"))
			},
			uuid:    "01967b9c-6217-77e3-a754-531365a6f5f2",
			wantErr: true,
			errType: response.HttpErrInternal,
		},
		{
			name: "ERROR: Fraud rule not found",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				// Mock GetByID to return nil rule (not found)
				repo.On("GetByID", mock.Anything, "01967b9c-6217-77e3-a754-531365a6f5f2").Return(nil, nil)
			},
			uuid:    "01967b9c-6217-77e3-a754-531365a6f5f2",
			wantErr: true,
			errType: response.HttpErrRequest,
		},
		{
			name: "ERROR: Failed to delete fraud rule",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				// Mock GetByID to return a valid fraud rule
				repo.On("GetByID", mock.Anything, "01967b9c-6217-77e3-a754-531365a6f5f2").Return(&fraudrulesmodel.FraudRules{
					UUID:     "01967b9c-6217-77e3-a754-531365a6f5f2",
					RuleName: "Test Rule",
				}, nil)

				// Mock Delete to return an error
				repo.On("Delete", mock.Anything, "01967b9c-6217-77e3-a754-531365a6f5f2").Return(errors.New("database error"))
			},
			uuid:    "01967b9c-6217-77e3-a754-531365a6f5f2",
			wantErr: true,
			errType: response.HttpErrInternal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			repo := mocksRepo.NewIFraudRulesRepository(t)
			logger, _ := logger.NewZapLogger(logger.Config{})

			// Setup expectations
			if tc.setup != nil {
				tc.setup(repo)
			}

			// Create service
			svc := New(logger, repo)

			// Call method
			err := svc.Delete(context.Background(), tc.uuid)

			// Verify results
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			repo.AssertExpectations(t)
		})
	}
}
