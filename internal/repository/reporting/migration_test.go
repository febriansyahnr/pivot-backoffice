package reportingRepository_test

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cdc"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/reporting"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListAccountTransactionsForMigration(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil, config.AppConfig{})

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []cdcModel.AccountTransaction
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, startDate, endDate, startDate, endDate).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, startDate, endDate, startDate, endDate).Once().Return(nil)
			},
			wantResult: []cdcModel.AccountTransaction{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.ListAccountTransactionsForMigration(t.Context(), startDate, endDate)

			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			db.AssertExpectations(t)
		})
	}
}
