package routingprocessorService

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgLogger "github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/test"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func TestIntegrationBankAccountInquiry(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	type mocker struct {
		cfg               *config.Config
		snapCoreProcessor *repositoryMocks.IRoutingProcessorRepository
		flipPgProcessor   *repositoryMocks.IRoutingProcessorRepository
	}

	ctx := context.Background()

	var (
		consulActive    bool
		logger          pkgLogger.ILogger
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
				buildConsul()
				m.snapCoreProcessor.On(
					"BankAccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:    "2001800",
					ResponseMessage: "Successful",
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
					"BankAccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(nil, errors.New("error when call snap core"))
				m.flipPgProcessor.On(
					"BankAccountInquiry",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountRequest"),
				).Return(&routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:    "2001800",
					ResponseMessage: "Successful",
				}, nil)
			},
			expectedResponse: "2001800",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				cfg: &config.Config{
					Environment: "local",
				},
				snapCoreProcessor: repositoryMocks.NewIRoutingProcessorRepository(t),
				flipPgProcessor:   repositoryMocks.NewIRoutingProcessorRepository(t),
			}

			routingProcessor := map[string]repository.IRoutingProcessorRepository{
				constant.SnapCoreProcessor: m.snapCoreProcessor,
				constant.FlipPGProcessor:   m.flipPgProcessor,
			}

			tc.mockSetup(m)

			svc := New(m.cfg, pdkLogger, routingProcessor)

			transfer, err := svc.AccountInquiry(ctx, &routingProcessorModel.InquiryAccountRequest{
				BeneficiaryBankCode: "022",
				MerchantID:          "922e39ab-7565-49f6-b84f-fb56122821ae",
			})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transfer)
				assert.Equal(t, transfer.ResponseCode, tc.expectedResponse)
			}

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
