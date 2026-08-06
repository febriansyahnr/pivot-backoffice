package otp_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/test"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

var (
	cfg           *config.Config
	loggerMock    logger.ILogger
	pdkLoggerMock pdkLogger.ILogger
)

func TestMain(m *testing.M) {
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("/%d-otp-config.yaml", time.Now().Unix()))

	_ = os.WriteFile(tmpFile, []byte(`
PAPER_COMMUNICATION:
    EMAIL_SENDER: "sender name <email@example.com>"
    EMAIL_LOGO_URL: ""
USER_OTP_CONFIG:
    MAX_SEND_RESET_PWD: 3
    MAX_SEND_RESET_PIN: 3
    MAX_SEND_USER_LOGIN: 3
    FIRST_TIME_LOGIN_MAX_SEND: 3
    USER_LOGIN_WAIT_AFTER_SEND: 3
    USER_LOGIN_WAIT_TIME_MINUTE: 6
    MAX_FAILED_VERIFY_USER_LOGIN: 3
`), 0777)
	defer os.Remove(tmpFile)

	err := error(nil)
	cfg, _, _ = config.LoadConfig(tmpFile, tmpFile)

	// Setup logger
	if loggerMock, pdkLoggerMock, err = test.SetupLogger(); err != nil {
		panic(err)
	}
	defer func() {
		_ = pdkLoggerMock.Sync()
		_ = loggerMock.Sync()
	}()

	m.Run()
}
