package merchantConsumerController

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessBulkCreateSubMerchant(t *testing.T) {
	input := merchantModel.ProcessBulkCreateSubMerchantRequest{
		ID:         "id",
		MerchantId: "merchantId",
		KYCType:    "KYC",
		Batch:      1,
		FileName:   "fileName.csv",
	}

	inputB, err := json.Marshal(input)
	assert.NoError(t, err)

	testCases := []struct {
		name       string
		input      []byte
		mocksSetup func(merchantSvc *serviceMocks.IMerchantService)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: Process Bulk Create Submerchant Request",
			input: inputB,
			mocksSetup: func(merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.On(
					"ProcessBulkCreateSubMerchant",
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: when unmarshal data request",
			input: []byte("invalid input"),
			mocksSetup: func(merchantSvc *serviceMocks.IMerchantService) {
				// do nothing
			},
			wantErr: true,
		},
		{
			name:  "ERROR: Process Bulk Create Submerchant Request",
			input: inputB,
			mocksSetup: func(merchantSvc *serviceMocks.IMerchantService) {
				merchantSvc.On(
					"ProcessBulkCreateSubMerchant",
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantSvc := serviceMocks.NewIMerchantService(t)
			logger := logger.NewSlogger(logger.Config{})
			conf := &config.Config{}

			tc.mocksSetup(merchantSvc)

			consumer := New(conf, logger, merchantSvc)
			ctx := context.Background()
			err := consumer.ProcessBulkCreateSubMerchant(ctx, tc.input, "")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}

}
