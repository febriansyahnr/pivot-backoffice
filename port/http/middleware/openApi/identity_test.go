package openApi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestSnapApiServiceCode(t *testing.T) {

	tests := []struct {
		name    string
		service string
		want    string
	}{
		{
			name:    "Access Token B2B",
			service: "/api/snap/v1.0/access-token/b2b",
			want:    "73",
		},
		{
			name:    "Create Virtual Account",
			service: "/api/snap/v1.0/transfer-va/create-va",
			want:    "27",
		},
		{
			name:    "Get Virtual Account",
			service: "/api/snap/v1.0/transfer-va/get-va",
			want:    "30",
		},
		{
			name:    "Update Virtual Account",
			service: "/api/snap/v1.0/transfer-va/update-va",
			want:    "28",
		},
		{
			name:    "Generate QR MPM",
			service: "/api/snap/v1.0/qr/qr-mpm-generate",
			want:    "47",
		},
		{
			name:    "Query Payment - Dynamic QR MPM",
			service: "/api/snap/v1.0/qr/qr-mpm-query",
			want:    "51",
		},
		{
			name:    "Query Payment - Static QR MPM",
			service: "/api/snap/v1.0/qr/transaction-history-list",
			want:    "12",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Use(IdentitySnapServiceCode)
			router.Get("/code", func(w http.ResponseWriter, r *http.Request) {
				code, _ := r.Context().Value(constant.CtxSnapApiName).(string)
				w.Write([]byte(code))
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/code", nil)
			req.Header.Set(constant.HeaderXSnapPath, test.service)

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.want, rec.Body.String())
		})
	}
}
