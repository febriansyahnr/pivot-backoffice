package liveFeature

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/liveFeature"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/liveFeature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLiveFeatureController_GetList(t *testing.T) {
	data := make([]liveFeature.LiveFeature, 0)
	data = append(data, liveFeature.LiveFeature{
		UUID:      uuid.New().String(),
		Name:      "test",
		Category:  "payout",
		Channel:   "card",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	testCases := []struct {
		name      string
		mockSetup func(
			mockService *mocks.ILiveFeaturesService,
		)
		expectedStatus int
	}{
		{
			name: "SUCCESS: Get List",
			mockSetup: func(mockService *mocks.ILiveFeaturesService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Return(data, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ERROR: Got error 500 on get list caused by invalid service",
			mockSetup: func(mockService *mocks.ILiveFeaturesService) {
				mockService.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
				).Return(nil, errors.New("some-error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := mocks.NewILiveFeaturesService(t)

			tc.mockSetup(mockService)

			controller := New(mockService)

			baseUrl := "/api/v1/services"

			chiRouterCtx := chi.NewRouteContext()
			req := httptest.NewRequest(http.MethodGet, baseUrl, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.GetList)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			mockService.AssertExpectations(t)
		})
	}
}
