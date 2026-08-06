package paymentService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInvestigationSummary(t *testing.T) {
	merchantID := uuid.NewString()
	now := time.Now()
	startDate := now.AddDate(0, 0, -30)
	endDate := now

	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	tests := []struct {
		name      string
		input     paymentModel.GetInvestigationSummaryOption
		mockSetup func(mockRepo *repoMocks.IPaymentRepository)
		expected  *paymentModel.InvestigationSummaryResponse
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get investigation summary with date range",
			input: paymentModel.GetInvestigationSummaryOption{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				expectedResponse := &paymentModel.InvestigationSummaryResponse{
					OnInvestigation: paymentModel.InvestigationSummaryItem{
						TotalAmount: "5000000",
						Currency:    "IDR",
					},
					Success: paymentModel.InvestigationSummaryItem{
						TotalAmount: "10000000",
						Currency:    "IDR",
					},
					Failed: paymentModel.InvestigationSummaryItem{
						TotalAmount: "2000000",
						Currency:    "IDR",
					},
				}
				mockRepo.On("GetInvestigationSummary", mock.Anything, mock.MatchedBy(func(opt paymentModel.GetInvestigationSummaryOption) bool {
					return opt.MerchantID == merchantID &&
						opt.StartDate.Equal(startDate) &&
						opt.EndDate.Equal(endDate)
				})).Once().Return(expectedResponse, nil)
			},
			expected: &paymentModel.InvestigationSummaryResponse{
				OnInvestigation: paymentModel.InvestigationSummaryItem{
					TotalAmount: "5000000",
					Currency:    "IDR",
				},
				Success: paymentModel.InvestigationSummaryItem{
					TotalAmount: "10000000",
					Currency:    "IDR",
				},
				Failed: paymentModel.InvestigationSummaryItem{
					TotalAmount: "2000000",
					Currency:    "IDR",
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigation summary with zero amounts",
			input: paymentModel.GetInvestigationSummaryOption{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				expectedResponse := &paymentModel.InvestigationSummaryResponse{
					OnInvestigation: paymentModel.InvestigationSummaryItem{
						TotalAmount: "0",
						Currency:    "IDR",
					},
					Success: paymentModel.InvestigationSummaryItem{
						TotalAmount: "0",
						Currency:    "IDR",
					},
					Failed: paymentModel.InvestigationSummaryItem{
						TotalAmount: "0",
						Currency:    "IDR",
					},
				}
				mockRepo.On("GetInvestigationSummary", mock.Anything, mock.MatchedBy(func(opt paymentModel.GetInvestigationSummaryOption) bool {
					return opt.MerchantID == merchantID &&
						opt.StartDate.Equal(startDate) &&
						opt.EndDate.Equal(endDate)
				})).Once().Return(expectedResponse, nil)
			},
			expected: &paymentModel.InvestigationSummaryResponse{
				OnInvestigation: paymentModel.InvestigationSummaryItem{
					TotalAmount: "0",
					Currency:    "IDR",
				},
				Success: paymentModel.InvestigationSummaryItem{
					TotalAmount: "0",
					Currency:    "IDR",
				},
				Failed: paymentModel.InvestigationSummaryItem{
					TotalAmount: "0",
					Currency:    "IDR",
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Repository returns error",
			input: paymentModel.GetInvestigationSummaryOption{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			mockSetup: func(mockRepo *repoMocks.IPaymentRepository) {
				mockRepo.On("GetInvestigationSummary", mock.Anything, mock.Anything).
					Once().Return(nil, errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repoMocks.NewIPaymentRepository(t)
			tt.mockSetup(mockRepo)

			service := New(mockRepo, mockLogger, nil, nil, nil, nil, nil)

			result, err := service.GetInvestigationSummary(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
