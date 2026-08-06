package openApi

import (
	"context"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pdk/go/snap"
)

var snapApiServiceCode = map[string]string{
	"/access-token/b2b":            snap.SNAP_SERVICE_B2B,
	"/transfer-va/create-va":       snap.SNAP_SERVICE_VIRTUAL_ACCOUNT,
	"/transfer-va/get-va":          constant.GetVirtualAccountSnapApiCode,
	"/transfer-va/update-va":       constant.UpdateVirtualAccountSnapApiCode,
	"/qr/qr-mpm-generate":          constant.GenerateQrisMPMSnapApiCode,
	"/qr/qr-mpm-query":             constant.QueryPaymentDynamicQrisMPMSnapApiCode,
	"/qr/transaction-history-list": constant.QueryPaymentStaticQrisMPMSnapApiCode,
}

func IdentitySnapServiceCode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		ctx, segment := otelTracer.Start(req.Context(), "http/middleware/openApi/IdentitySnapServiceCode")
		defer segment.End()

		serviceCode := snap.SNAP_SERVICE_B2B

		for key, value := range snapApiServiceCode {
			if strings.Contains(req.Header.Get(constant.HeaderXSnapPath), key) {
				serviceCode = value

				break
			}
		}
		req = req.WithContext(context.WithValue(ctx, constant.CtxSnapApiName, serviceCode))

		next.ServeHTTP(wr, req)
	})
}
