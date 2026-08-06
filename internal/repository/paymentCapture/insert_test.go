package paymentCapture

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentCapture"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInsert(t *testing.T) {
	processorRefID := "processor-ref-123"
	capture := &paymentCaptureModel.PaymentCapture{
		ID:                     uuid.NewString(),
		PaymentID:              uuid.NewString(),
		ProcessorReferenceID:   &processorRefID,
		Status:                 "CAPTURED",
		ReleaseRemainingAmount: true,
		Currency:               "IDR",
		Amount:                 10000,
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
			name:    "SUCCESS: Insert payment capture",
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
			name:    "ERROR: Database error during insert",
			capture: capture,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*paymentCaptureModel.PaymentCapture"),
				).Return(false, errors.New("duplicate key error"))
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
			err := repo.Insert(context.Background(), tc.capture)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
