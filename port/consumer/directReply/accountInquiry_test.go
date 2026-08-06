package directreply

import (
	"context"
	"encoding/json"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	"testing"

	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddressingWaitReplyAccountInquiry(t *testing.T) {
	type mocker struct {
		logger              *pdkLoggerMock.ILogger
		routingProcessorSvc *serviceMocks.IRoutingProcessorService

		body []byte
	}

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(m *mocker)
	}{
		{
			desc:    "error when unmarshalling data body",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.body = []byte("invalid")

			},
		},
		{
			desc:    "error process at service",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.body, _ = json.Marshal(routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "00",
					ResponseMessage:        "Success",
					PartnerReferenceNo:     "123",
					BeneficiaryAccountName: "bejo",
					BeneficiaryAccountNo:   "123",
					BeneficiaryBankCode:    "002",
					BeneficiaryBankName:    "Permata",
					ProcessorReference:     "123",
					Status:                 "VALID",
				})

				m.routingProcessorSvc.On("AddressingReplyToAccountInquiry", mock.Anything, mock.AnythingOfType("*routingProcessorModel.InquiryAccountResponseData")).Return(assert.AnError)

			},
		},
		{
			desc:    "success process at service",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.body, _ = json.Marshal(routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:           "00",
					ResponseMessage:        "Success",
					PartnerReferenceNo:     "123",
					BeneficiaryAccountName: "bejo",
					BeneficiaryAccountNo:   "123",
					BeneficiaryBankCode:    "002",
					BeneficiaryBankName:    "Permata",
					ProcessorReference:     "123",
					Status:                 "VALID",
				})

				m.routingProcessorSvc.On("AddressingReplyToAccountInquiry", mock.Anything, mock.AnythingOfType("*routingProcessorModel.InquiryAccountResponseData")).Return(nil)

			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				logger:              pdkLoggerMock.NewILogger(t),
				routingProcessorSvc: serviceMocks.NewIRoutingProcessorService(t),
			}

			tc.mockSetup(m)

			ctrl := New(m.logger, m.routingProcessorSvc)
			err := ctrl.AddressingWaitReplyAccountInquiry(context.Background(), m.body, "")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}
