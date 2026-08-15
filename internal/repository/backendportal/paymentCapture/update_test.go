package paymentCapture

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentCapture"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	processorRefID := "processor-ref-updated"
	capture := &paymentCaptureModel.PaymentCapture{
		ID:                     uuid.NewString(),
		PaymentID:              uuid.NewString(),
		ProcessorReferenceID:   &processorRefID,
		Status:                 "CAPTURED",
		ReleaseRemainingAmount: false,
		Currency:               "IDR",
		Amount:                 15000,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	testCase := []struct {
		name      string
		capture   *paymentCaptureModel.PaymentCapture
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name:    "SUCCESS: Update payment capture",
			capture: capture,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*paymentCaptureModel.PaymentCapture"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Database error during update",
			capture: capture,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*paymentCaptureModel.PaymentCapture"),
				).Return(false, errors.New("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.Update(context.Background(), tc.capture)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
