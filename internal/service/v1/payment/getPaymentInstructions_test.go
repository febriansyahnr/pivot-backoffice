package paymentService

import (
	"context"
	"errors"
	"testing"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentInstructions(t *testing.T) {
	var (
		ctx             = context.Background()
		mockPaymentRepo = repositoryMocks.NewIPaymentRepository(t)
		paymentService  = PaymentService{
			paymentRepo: mockPaymentRepo,
		}
	)

	testCases := []struct {
		name          string
		paymentMethod string
		callMock      func()
		want          []paymentModel.InstructionResponse
		wantErr       error
		shouldErr     bool
	}{
		{
			name:          "when payment method does not exist, should return error",
			paymentMethod: "UNKNOWN_METHOD",
			callMock: func() {
				instructionsList := []paymentModel.InstructionResponse{
					{
						Title: "QRIS_Bank_Neo_Commerce",
					},
				}
				mockPaymentRepo.On("RetrieveInstructions", mock.Anything).
					Return(instructionsList, nil).
					Once()
			},
			wantErr:   pkgErr.New(httpResponse.HttpErrUnprocessableContent, errors.New("instructions for payment method UNKNOWN_METHOD not found")),
			shouldErr: true,
		},
		{
			name:          "when retrieval fails, should return error",
			paymentMethod: "",
			callMock: func() {
				mockPaymentRepo.On("RetrieveInstructions", mock.Anything).
					Return(nil, errors.New("repository error")).
					Once()
			},
			wantErr:   errors.New("failed to retrieve payment instructions: repository error"),
			shouldErr: true,
		},
		{
			name:          "when instructions retrieved successfully, should return all instructions",
			paymentMethod: "",
			callMock: func() {
				instructionsList := []paymentModel.InstructionResponse{
					{
						Title:       "QRIS_Bank_Neo_Commerce",
						Instruction: "<h4>How to Pay</h4><p>Open your bank or e-wallet app, scan the QR code, confirm the payment, and you're done!</p>",
						Accordion: []paymentModel.AccordionStep{
							{
								Title: "Scan via e-Wallet",
								Steps: []string{
									"Open your bank or e-wallet app that supports QRIS payment on your phone.",
									"Scan the QR code above.",
									"Ensure the total amount is correct, then click '<b>Bayar</b>'.",
									"Once successful, the payment will be automatically verified.",
								},
								Note: "Transfer fees are subject to your bank's terms and conditions.",
							},
						},
					},
				}
				mockPaymentRepo.On("RetrieveInstructions", mock.Anything).
					Return(instructionsList, nil).
					Once()
			},
			want: []paymentModel.InstructionResponse{
				{
					Title:       "QRIS_Bank_Neo_Commerce",
					Instruction: "<h4>How to Pay</h4><p>Open your bank or e-wallet app, scan the QR code, confirm the payment, and you're done!</p>",
					Accordion: []paymentModel.AccordionStep{
						{
							Title: "Scan via e-Wallet",
							Steps: []string{
								"Open your bank or e-wallet app that supports QRIS payment on your phone.",
								"Scan the QR code above.",
								"Ensure the total amount is correct, then click '<b>Bayar</b>'.",
								"Once successful, the payment will be automatically verified.",
							},
							Note: "Transfer fees are subject to your bank's terms and conditions.",
						},
					},
				},
			},
			shouldErr: false,
		},
		{
			name:          "when payment method exists, should return specific instructions",
			paymentMethod: "QRIS_BRI",
			callMock: func() {
				instructionsList := []paymentModel.InstructionResponse{
					{
						Title:       "QRIS_BRI",
						Instruction: "<h4>How to Pay</h4><p>Follow these steps to pay via QRIS BRI.</p>",
						Accordion: []paymentModel.AccordionStep{
							{
								Title: "Scan via e-Wallet",
								Steps: []string{
									"Open your bank or e-wallet app that supports QRIS payment on your phone.",
									"Scan the QR code above.",
									"Ensure the total amount is correct, then click '<b>Bayar</b>'.",
									"Once successful, the payment will be automatically verified.",
								},
								Note: "Transfer fees are subject to your bank's terms and conditions.",
							},
						},
					},
				}
				mockPaymentRepo.On("RetrieveInstructions", mock.Anything).
					Return(instructionsList, nil).
					Once()
			},
			want: []paymentModel.InstructionResponse{
				{
					Title:       "QRIS_BRI",
					Instruction: "<h4>How to Pay</h4><p>Follow these steps to pay via QRIS BRI.</p>",
					Accordion: []paymentModel.AccordionStep{
						{
							Title: "Scan via e-Wallet",
							Steps: []string{
								"Open your bank or e-wallet app that supports QRIS payment on your phone.",
								"Scan the QR code above.",
								"Ensure the total amount is correct, then click '<b>Bayar</b>'.",
								"Once successful, the payment will be automatically verified.",
							},
							Note: "Transfer fees are subject to your bank's terms and conditions.",
						},
					},
				},
			},
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			result, err := paymentService.GetPaymentInstructions(ctx, tc.paymentMethod)

			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, result)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}
