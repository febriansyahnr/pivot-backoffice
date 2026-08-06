package routingprocessorService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransferById(t *testing.T) {
	type mocker struct {
		routingProcessorSvc *repositoryMocks.IRoutingProcessorRepository
		proccNameReq        string
	}

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(m *mocker)
	}{
		{
			desc:    "error when do getTransferByID",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.routingProcessorSvc.On("GetTransferById", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
				m.proccNameReq = constant.SnapCoreProcessor
			},
		},
		{
			desc:    "error processor not found",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.proccNameReq = "DANA"
			},
		},
		{
			desc:    "success getTransferByID",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.routingProcessorSvc.On("GetTransferById", mock.Anything, mock.Anything, mock.Anything).Return(&routingProcessorModel.BankTransferResponseData{}, nil)
				m.proccNameReq = constant.SnapCoreProcessor
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				routingProcessorSvc: repositoryMocks.NewIRoutingProcessorRepository(t),
			}

			tc.mockSetup(m)

			svc := &routingProcessorService{
				routingProcessor: map[string]repository.IRoutingProcessorRepository{
					constant.SnapCoreProcessor: m.routingProcessorSvc,
					constant.FlipPGProcessor:   m.routingProcessorSvc,
				},
			}

			get, err := svc.GetTransferByID(context.Background(), &orchestrator_model.AccountTransactionWithUseCase{
				UUID:               uuid.New(),
				ProcessorReference: m.proccNameReq,
			}, false)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, get)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, get)
			}
		})
	}
}
