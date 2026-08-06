package creditcardCoreProcessorRepository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/creditcardCoreProcessor"
	mockHttpRequest "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestBlockCard(t *testing.T) {
	t.Run("success block card", func(t *testing.T) {
		mockHttp := mockHttpRequest.NewIHTTPRequest(t)
		mockLog, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		cfg := &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: "https://test.com",
			},
		}

		repo := New(cfg, nil, mockLog, mockHttp)

		request := &creditcardCoreProcessorModel.BlockCardRequest{
			CardUUID:    "test-card-uuid",
			IsBlocked:   true,
			BlockReason: "Security concern",
		}

		mockHttp.On("PUT", mock.Anything, "https://test.com/crm/v1/card/block", request, mock.Anything).
			Return([]byte(`{"code":"00","message":"success"}`), 200, nil)

		err := repo.BlockCard(context.Background(), request)

		assert.NoError(t, err)
	})

	t.Run("http request error", func(t *testing.T) {
		mockHttp := mockHttpRequest.NewIHTTPRequest(t)
		mockLog, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		cfg := &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: "https://test.com",
			},
		}

		repo := New(cfg, nil, mockLog, mockHttp)

		request := &creditcardCoreProcessorModel.BlockCardRequest{
			CardUUID:    "test-card-uuid",
			IsBlocked:   true,
			BlockReason: "Security concern",
		}

		mockHttp.On("PUT", mock.Anything, mock.Anything, request, mock.Anything).
			Return([]byte(nil), 0, errors.New("network error"))

		err := repo.BlockCard(context.Background(), request)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("400 bad request error", func(t *testing.T) {
		mockHttp := mockHttpRequest.NewIHTTPRequest(t)
		mockLog, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		cfg := &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: "https://test.com",
			},
		}

		repo := New(cfg, nil, mockLog, mockHttp)

		request := &creditcardCoreProcessorModel.BlockCardRequest{
			CardUUID:    "test-card-uuid",
			IsBlocked:   true,
			BlockReason: "Security concern",
		}

		mockHttp.On("PUT", mock.Anything, mock.Anything, request, mock.Anything).
			Return([]byte(`{"code":"40","error":"Invalid card UUID"}`), 400, nil)

		err := repo.BlockCard(context.Background(), request)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid card UUID")
	})

	t.Run("500 internal server error", func(t *testing.T) {
		mockHttp := mockHttpRequest.NewIHTTPRequest(t)
		mockLog, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		cfg := &config.Config{
			CreditcardCoreProcessorConfig: config.CreditcardCoreProcessorConfig{
				BaseUrl: "https://test.com",
			},
		}

		repo := New(cfg, nil, mockLog, mockHttp)

		request := &creditcardCoreProcessorModel.BlockCardRequest{
			CardUUID:    "test-card-uuid",
			IsBlocked:   true,
			BlockReason: "Security concern",
		}

		mockHttp.On("PUT", mock.Anything, mock.Anything, request, mock.Anything).
			Return([]byte(`{"code":"50","error":"Internal server error"}`), 500, nil)

		err := repo.BlockCard(context.Background(), request)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Internal server error")
	})
}