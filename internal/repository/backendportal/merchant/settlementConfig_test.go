package merchant_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSettlementConfig(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	var (
		merchantID = "merchant-id-12345" // NOSONAR
		reference  = "REFERENCE"         // NOSONAR
		method     = "METHOD"            // NOSONAR
		channel    = "CHANNEL"           // NOSONAR
	)
	args := []any{
		mock.Anything, mock.Anything, mock.Anything, merchantID, reference, &method, &channel, merchantID, reference, &method,
	}
	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *merchant.SettlementConfig
		request    *merchant.GetSettlementConfigRequest
	}{
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data not found for settlement Instant", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, reference, &method, &channel, constant.SettlementTypeInstant, merchantID, reference, &method, constant.SettlementTypeInstant,
				).Once().Return(sql.ErrNoRows)
			},
			request: &merchant.GetSettlementConfigRequest{
				MerchantId:       merchantID,
				Reference:        reference,
				Method:           &method,
				Channel:          &channel,
				SettlementMethod: constant.SettlementTypeInstant,
			},
		},
		{
			name: "SUCCESS:Data not found for settlement non Instant", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, reference, &method, &channel, constant.SettlementTypeStandard, merchantID, reference, &method, constant.SettlementTypeStandard,
				).Once().Return(sql.ErrNoRows)
			},
			request: &merchant.GetSettlementConfigRequest{
				MerchantId:       merchantID,
				Reference:        reference,
				Method:           &method,
				Channel:          &channel,
				SettlementMethod: constant.SettlementTypeStandard,
			},
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Return(assert.AnError)
			},
			wantError: fmt.Errorf("get settlement config: %v", assert.AnError),
		},
		{
			name: "ERROR:JSON unmarshal", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*types.JSONText) = []byte(`B`)
				}).Return(nil)
			},
			wantError: errors.New("unmarshal json: invalid character 'B' looking for beginning of value"),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*types.JSONText) = []byte(`{"type": "T+1"}`)
				}).Return(nil)
			},
			wantResult: &merchant.SettlementConfig{Type: "T+1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			request := merchant.GetSettlementConfigRequest{
				MerchantId: merchantID,
				Reference:  reference,
				Method:     &method,
				Channel:    &channel,
			}
			if test.request != nil {
				request = *test.request
			}
			result, err := repo.GetSettlementConfig(t.Context(), request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
