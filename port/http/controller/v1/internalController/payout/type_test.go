package internalPayoutController_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

const merchantPlatformWhitelistedOldResponseFormat = "aec6636d-7a02-4d93-a4c5-006b9c235068"

func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(cwd, "backend-portal")
	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	_ = ffclient.Init(ffclient.Config{
		Retriever:    &fileretriever.Retriever{Path: targetPath},
		DataExporter: ffclient.DataExporter{},
	})
	defer ffclient.Close()

	_, _ = monitor.New("backend-portal", "0.0.0.0", "5555") // NOSONAR

	m.Run()
}

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
