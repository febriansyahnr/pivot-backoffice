package paymentRepository_test

import (
	"database/sql"
	"testing"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/payment"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindPendingSubsequentCardFundedPayout(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	var (
		merchantID  = "72f85bf4-9075-4c0b-8047-d09a2a242c9e"
		referenceID = "467d3d65-512a-46cd-9c0c-7db14b931dbf"
	)
	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []model.CardFundedPayment
	}{
		{
			name: "ERROR: Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, merchantID, referenceID,
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS: Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, merchantID, referenceID,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS: Data found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, merchantID, referenceID,
				).Once().Return(nil)
			},
			wantResult: []model.CardFundedPayment{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := repo.FindPendingSubsequentCardFundedPayout(t.Context(), merchantID, referenceID)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.wantResult, result)

			db.AssertExpectations(t)
		})
	}
}

func TestGetCardFundedPayoutFundingSummary(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	var (
		merchantID     = "a4f4b09b-9097-4b51-8c9c-a6691b1f10d5"
		referenceID    = "9cf6d420-55c7-4b32-8ad2-f5635e0e4e9c"
		maxCreatedDays = 14
	)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *model.CardFundedPayoutFundingSummary
	}{
		{
			name: "ERROR: Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, maxCreatedDays, merchantID, referenceID, maxCreatedDays,
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, maxCreatedDays, merchantID, referenceID, maxCreatedDays,
				).Once().Return(nil)
			},
			wantResult: &model.CardFundedPayoutFundingSummary{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := repo.GetCardFundedPayoutFundingSummary(t.Context(), merchantID, referenceID, maxCreatedDays)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.wantResult, result)

			db.AssertExpectations(t)
		})
	}
}

func TestHardDeleteCardFundedPayoutPayments(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	var (
		merchantID  = "c4f7d895-e7a7-4ffb-bf90-8064a1da2b27"
		referenceID = "8389a85a-f877-4e1d-b5cd-4d2ffca8cd2a"
	)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR: Some error", // NOSONAR
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, merchantID, referenceID).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, merchantID, referenceID).Once().Return(true, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			assert.Equal(t, tt.wantError, repo.HardDeleteCardFundedPayoutPayments(t.Context(), merchantID, referenceID))

			db.AssertExpectations(t)
		})
	}
}
