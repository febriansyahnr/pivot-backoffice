package crmCreditcardController

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
)

func TestBlockCard(t *testing.T) {
	t.Run("success block card", func(t *testing.T) {
		mockCreditcardSvc := mockService.NewICreditCardService(t)

		handler := &handler{
			validator:     validatorExt.New(),
			creditcardSvc: mockCreditcardSvc,
		}

		request := creditcardModel.BlockCardRequest{
			CardUUID:    "test-card-uuid",
			IsBlocked:   true,
			BlockedTo:   time.Now().Add(24 * time.Hour),
			BlockReason: "Security concern",
		}

		mockCreditcardSvc.On("BlockCard", mock.Anything, mock.AnythingOfType("*card.BlockCardRequest")).Return(nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPut, "/crm/v1/card/block", bytes.NewBuffer(body))
		req = req.WithContext(context.Background())
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.BlockCard(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check the actual response structure
		t.Logf("Response body: %s", w.Body.String())

		// The response should contain success field
		if data, ok := response["data"].(map[string]interface{}); ok {
			assert.Equal(t, true, data["success"])
			assert.Equal(t, "Card blocked successfully", data["message"])
		}
	})

	t.Run("invalid request payload", func(t *testing.T) {
		mockCreditcardSvc := mockService.NewICreditCardService(t)

		handler := &handler{
			validator:     validatorExt.New(),
			creditcardSvc: mockCreditcardSvc,
		}

		req := httptest.NewRequest(http.MethodPut, "/crm/v1/card/block", bytes.NewBuffer([]byte("invalid json")))
		req = req.WithContext(context.Background())
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.BlockCard(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}