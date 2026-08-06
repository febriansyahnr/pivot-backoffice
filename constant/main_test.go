package constant_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

var cfg *config.Config

func TestMain(m *testing.M) {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("/%d-test-config.yaml", time.Now().Unix()))
	//nolint:gofumpt
	_ = os.WriteFile(tmpFile, []byte(`
ENVIRONMENT: "staging"
PAPER_COMMUNICATION:
    EMAIL_SENDER: "sender name <email@example.com>"
USER_OTP_CONFIG:
    MAX_SEND_RESET_PWD: 5
    MAX_SEND_RESET_PIN: 10
    MAX_SEND_CHANGE_PWD: 5
    MAX_SEND_USER_LOGIN: 10
    EXPIRATION_SECONDS_FORGOT_PASSWORD: 300
    EXPIRATION_SECONDS_RESET_PIN: 300
    EXPIRATION_SECONDS_CHANGE_PASSWORD: 300
    EXPIRATION_SECONDS_USER_LOGIN: 300
    EXPIRATION_SECONDS_FIRST_TIME_LOGIN: 300
    RESEND_DELAY_SECONDS_FORGOT_PASSWORD: 241
    RESEND_DELAY_SECONDS_RESET_PIN: 242
    RESEND_DELAY_SECONDS_CHANGE_PASSWORD: 243
    RESEND_DELAY_SECONDS_USER_LOGIN: 244
    RESEND_DELAY_SECONDS_FIRST_TIME_LOGIN: 245
    RESEND_DELAY_SECONDS_DEFAULT: 30
    MAX_FAILED_VERIFY_RESET_PWD: 3
    MAX_FAILED_VERIFY_RESET_PIN: 6
    MAX_FAILED_VERIFY_CHANGE_PWD: 3
    MAX_FAILED_VERIFY_USER_LOGIN: 3
    USER_LOGIN_WAIT_AFTER_SEND: 3
    USER_LOGIN_WAIT_TIME_MINUTE: 60
    FIRST_TIME_LOGIN_MAX_SEND: 21
    FIRST_TIME_LOGIN_MAX_FAILED_VERIFY: 22
    FIRST_TIME_LOGIN_WAIT_AFTER_SEND: 23
    FIRST_TIME_LOGIN_WAIT_TIME_MINUTE: 24
`), 0777)
	defer os.Remove(tmpFile)

	var err error

	cfg, _, err = config.LoadConfig(tmpFile, tmpFile)
	if err != nil {
		panic("Test: Load config failed: " + err.Error())
	}

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
