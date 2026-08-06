package v1InternalRefundController_test

import (
	"context"
	"fmt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"strconv"
)

func wrapErrOpenApiNonSnap(code int, msg string, respType ...string) string {
	var errorType string
	if len(respType) > 0 {
		errorType = respType[0]
	} else {
		errorType = "ERROR_REQUEST"
	}

	errorResponse := response.HandleDetailedError(context.Background(), strconv.Itoa(code), msg, errorType)

	return fmt.Sprintf(
		`{"code":"%s","message":"%s","error":{"type":"%s","details":[{"field":"%s","message":"%s"}],"traceId":"%s"}}`,
		errorResponse.Code, errorResponse.Message, errorResponse.Error.Type, errorResponse.Error.Details[0].Field, errorResponse.Error.Details[0].Message, errorResponse.Error.TraceId,
	)
}
