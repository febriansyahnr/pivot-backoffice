package purpose_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/purpose"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/list", purpose.New(&config.Config{}).List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/list", nil)

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Result().StatusCode)
	assert.JSONEq(t, `{"code":"00","message":"OK","data":"Unimplemented"}`, rec.Body.String())
}
