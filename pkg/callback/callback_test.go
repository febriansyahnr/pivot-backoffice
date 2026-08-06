package callback_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	. "github.com/paper-indonesia/pivot-backoffice/pkg/callback"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCallback(t *testing.T) {
	const exampleApiKey = "dcf81a2f-7efd-41ac-894e-f81857e6e04d"

	var requestCount atomic.Int64

	mux := http.NewServeMux()
	mux.HandleFunc("POST /test", func(w http.ResponseWriter, r *http.Request) {
		if requestCount.Load() == 2 {
			time.Sleep(time.Second)
		}

		statusCode, response := http.StatusOK, `{"message":"OK"}`
		if r.Header.Get(constant.HeaderXAPIKey) != exampleApiKey {
			statusCode, response = http.StatusUnauthorized, `{"message":"unauthorized"}`
		}
		w.WriteHeader(statusCode)
		w.Write([]byte(response))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	log := mocks.NewILogger(t)
	httpClient := New(log, httputil.NewHTTPClient(httputil.HTTPClientConfig{
		RequestTimeout: 500 * time.Millisecond,
	}))

	tests := []struct {
		name        string
		url         string
		headers     map[string]string
		setupMock   func()
		assertError func(t *testing.T, err error)
		wantResult  string
	}{
		{
			name: "ERROR:Invalid request", // NOSONAR
			url:  "invalid",
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed to send callback to merchant", mock.Anything, mock.Anything).Once().Return()
			},
			assertError: func(t *testing.T, err error) {
				if !assert.Error(t, err) {
					return
				}
				assert.Equal(t, "ERROR_INTERNAL | an error occurred on the server. please try again later", err.Error())
			},
		},
		{
			name: "ERROR:Timeout", // NOSONAR
			url:  svr.URL + "/test",
			setupMock: func() {
				log.On("Error", mock.Anything, "Failed to send callback to merchant", mock.Anything, mock.Anything).Once().Return()
			},
			assertError: func(t *testing.T, err error) {
				if !assert.Error(t, err) {
					return
				}
				assert.Equal(t, "ERROR_REQUEST | request timeout", err.Error())
			},
		},
		{
			name: "ERROR:Invalid api key", // NOSONAR
			url:  svr.URL + "/test",
			setupMock: func() {
				log.On("Info", mock.Anything, "Failed to send callback to merchant: Received non-2xx response", mock.Anything, mock.Anything).Once().Return()
			},
			assertError: func(t *testing.T, err error) {
				if !assert.Error(t, err) {
					return
				}
				errHttp, ok := err.(*ErrHttpClient)
				if !assert.Truef(t, ok, "non-ErrHttpClient error type") {
					return
				}
				assert.Equal(t, http.StatusUnauthorized, errHttp.StatusCode())
				assert.JSONEq(t, `{"message":"unauthorized"}`, string(errHttp.ResponseBody()))
				assert.Equal(t, constant.ErrInvokeClientWebhook.Error(), errHttp.Error())
			},
			wantResult: `{"message":"unauthorized"}`,
		},
		{
			name:    "SUCCESS", // NOSONAR
			url:     svr.URL + "/test",
			headers: map[string]string{constant.HeaderXAPIKey: exampleApiKey},
			setupMock: func() {
				log.On("Info", mock.Anything, "Response from callback", mock.Anything).Once().Return()
			},
			assertError: func(t *testing.T, err error) { /* Empty Function */ },
			wantResult:  `{"message":"OK"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			requestCount.Add(1)

			result, err := httpClient.Callback(
				t.Context(), callbackModel.CallbackRequest{URL: test.url}, test.headers,
			)
			test.assertError(t, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
