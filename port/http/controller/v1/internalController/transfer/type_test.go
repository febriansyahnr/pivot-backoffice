package transfer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

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

	m.Run()
}
