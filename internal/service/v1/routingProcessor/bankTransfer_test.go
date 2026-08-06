package routingprocessorService

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/test"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func TestIntegrationBankTransfer(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	type mocker struct {
		cfg               *config.Config
		snapCoreProcessor *repositoryMocks.IRoutingProcessorRepository
		flipPgProcessor   *repositoryMocks.IRoutingProcessorRepository
		merchantID        string
	}

	ctx := context.Background()

	var (
		consulActive    bool
		logger          logger.ILogger
		pdkLogger       pdkLogger.ILogger
		err             error
		consulContainer testcontainers.Container
		consulURL       string
	)

	buildConsul := func() {
		logger, pdkLogger, err = test.SetupLogger()
		assert.NoError(t, err)
		consulContainer, consulURL, err = test.SetupConsul(ctx)
		assert.NoError(t, err)
		test.SetupFeatureFlag(consulURL)
		test.SetupGoff(ctx, consulURL, pdkLogger)
		consulActive = true
	}

	payloadTransfer := &routingProcessorModel.BankTransferRequest{
		Beneficiary: routingProcessorModel.SubjectRequest{
			BankCode: "022",
		},
	}

	testCases := []struct {
		desc             string
		wantErr          bool
		mockSetup        func(m *mocker)
		expectedResponse string
	}{
		{
			desc:    "success when call first priority no reroute to next priority",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.merchantID = ""
				buildConsul()
				m.snapCoreProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode:    "2001800",
					ResponseMessage: "Successful",
					Status:          "SUCCESS",
				}, nil)
			},
			expectedResponse: "2001800",
		},
		{
			desc:    "error when call first priority but success when call second priority",
			wantErr: false,
			mockSetup: func(m *mocker) {
				buildConsul()
				m.snapCoreProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(nil, errors.New("error when call snap core"))
				m.flipPgProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode: "2001800",
				}, nil)
			},
			expectedResponse: "2001800",
		},
		{
			desc:    "error when call all priority",
			wantErr: true,
			mockSetup: func(m *mocker) {
				buildConsul()
				m.snapCoreProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode: "4001802",
				}, errors.New("error when call snap-core"))
				m.flipPgProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode: "4001802",
				}, errors.New("error when call flip pg"))
			},
			expectedResponse: "4001802",
		},
		{
			desc:    "error when call first priority with pending status",
			wantErr: true,
			mockSetup: func(m *mocker) {
				buildConsul()
				m.snapCoreProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode: "2021802",
					Status:       "PENDING",
				}, errors.New("error when call snap-core"))

			},
			expectedResponse: "2021802",
		},
		{
			desc:    "error InternalError when call second priority",
			wantErr: true,
			mockSetup: func(m *mocker) {
				buildConsul()
				m.snapCoreProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode: "4001802",
				}, errors.New("error when call snap-core"))
				m.flipPgProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode: "5001800",
					Status:       "UNKNOWN",
				}, errors.New("error when call snap-core"))
			},
			expectedResponse: "5001800",
		},
		{
			desc:    "consul not active success when call default priority",
			wantErr: false,
			mockSetup: func(m *mocker) {
				m.snapCoreProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode:    "2001800",
					ResponseMessage: "Successful",
				}, nil)
			},
			expectedResponse: "2001800",
		},
		{
			desc:    "consul not active error when call default priority",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.snapCoreProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(nil, errors.New("error when call snap core"))
			},
			expectedResponse: "",
		},
		{
			desc:    "success with allowedDestination bank code",
			wantErr: false,
			mockSetup: func(m *mocker) {
				buildConsul()
				m.merchantID = "00000000-0000-0000-0000-000000000001"
				m.flipPgProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode:    "2001800",
					ResponseMessage: "Successful",
					Status:          "SUCCESS",
				}, nil)
			},
			expectedResponse: "2001800",
		},
		{
			desc:    "success with beneficiary not in list allowedDestination bank code",
			wantErr: false,
			mockSetup: func(m *mocker) {
				payloadTransfer.Beneficiary.BankCode = "002"
				buildConsul()
				m.merchantID = "00000000-0000-0000-0000-000000000001"
				m.snapCoreProcessor.On(
					"TriggerTransfer",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferRequest"),
				).Return(&routingProcessorModel.BankTransferResponseData{
					ResponseCode:    "2001800",
					ResponseMessage: "Successful",
					Status:          "SUCCESS",
				}, nil)
			},
			expectedResponse: "2001800",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				cfg: &config.Config{
					Environment: constant.EnvironmentProduction,
				},

				snapCoreProcessor: repositoryMocks.NewIRoutingProcessorRepository(t),
				flipPgProcessor:   repositoryMocks.NewIRoutingProcessorRepository(t),
				merchantID:        "922e39ab-7565-49f6-b84f-fb56122821ae",
			}

			routingProcessor := map[string]repository.IRoutingProcessorRepository{
				constant.SnapCoreProcessor: m.snapCoreProcessor,
				constant.FlipPGProcessor:   m.flipPgProcessor,
			}

			tc.mockSetup(m)

			svc := New(m.cfg, pdkLogger, routingProcessor)

			payloadTransfer.HeaderRequest = snapCoreModel.BankTransferHeaderRequest{
				MerchantId: m.merchantID,
			}

			transfer, err := svc.BankTransfer(ctx, payloadTransfer)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transfer)
			}
			assert.Equal(t, transfer.ResponseCode, tc.expectedResponse)

			if consulActive {
				defer pdkLogger.Sync()
				defer logger.Sync()
				defer ffclient.Close()
				defer consulContainer.Terminate(ctx)
				consulActive = false
			}
		})
	}
}

func TestIsChannelCodeAllowed(t *testing.T) {
	testCases := []struct {
		input    string
		list     []string
		expected bool
	}{
		{
			input:    "001",
			list:     []string{"002"},
			expected: false,
		},
		{
			input:    "002",
			list:     []string{"002"},
			expected: true,
		},
		{
			input:    "001",
			list:     []string{"002", "003", "004"},
			expected: false,
		},
		{
			input:    "002",
			list:     []string{"002", "003", "004"},
			expected: true,
		},
		{
			input:    "003",
			list:     []string{"002", "003"},
			expected: true,
		},
		{
			input:    "004",
			list:     []string{"002", "003"},
			expected: false,
		},
		{
			input:    "001",
			list:     []string{"!001"},
			expected: false,
		},
		{
			input:    "002",
			list:     []string{"!001"},
			expected: true,
		},
		{
			input:    "001",
			list:     []string{"!001", "!002", "!003"},
			expected: false,
		},
		{
			input:    "002",
			list:     []string{"!001", "!002", "!003"},
			expected: false,
		},
		{
			input:    "003",
			list:     []string{"!001", "!002", "!003"},
			expected: false,
		},
		{
			input:    "004",
			list:     []string{"!001", "!002", "!003"},
			expected: true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Test %s with list %v expected %v", tc.input, tc.list, tc.expected), func(t *testing.T) {
			repo := &routingProcessorService{}

			assert.Equal(t, repo.isChannelCodeAllowed(tc.list, tc.input), tc.expected)
		})
	}
}
