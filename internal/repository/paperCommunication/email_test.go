package paperCommunication_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/paperCommunication"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/paperCommunication"
	httpReqPkgMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	logPkgMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSendEmailV1(pt *testing.T) {
	httpReq := httpReqPkgMock.NewIHTTPRequest(pt)
	logger, _ := logPkgMock.NewZapLogger(logPkgMock.Config{})

	client := New(&config.PaperCommunication{BaseURL: "http://example.id"}, logger, httpReq)

	tests := []struct {
		name      string
		mockSetup func(r *httpReqPkgMock.IHTTPRequest)
		wantErr   string
	}{
		{
			name: "ERROR:Execute http request",
			mockSetup: func(r *httpReqPkgMock.IHTTPRequest) {
				r.On(
					"POST", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("*paperCommunication.Email"), mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Once().Return(nil, 0, errors.New("failed to prepare request"))
			},
			wantErr: "failed to prepare request",
		},
		{
			name: "ERROR:Internal server error",
			mockSetup: func(r *httpReqPkgMock.IHTTPRequest) {
				r.On(
					"POST", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("*paperCommunication.Email"), mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Once().Return([]byte(`{"code": "99", "errors": "Internal server error"}`), http.StatusInternalServerError, nil)
			},
			wantErr: "failed to send email, see log for more details",
		}, {
			name: "SUCCESS",
			mockSetup: func(r *httpReqPkgMock.IHTTPRequest) {
				r.On(
					"POST", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("*paperCommunication.Email"), mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"code": "20"`), http.StatusOK, nil)
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup(httpReq)

			err := client.SendEmailV1(context.Background(), "", &paperCommunication.Email{})
			if test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
