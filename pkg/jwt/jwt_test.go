package jwt

import (
	"log"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
)

func TestMain(m *testing.M) {

	f, err := os.CreateTemp(os.TempDir(), "test-jwt-config-*.yml")
	if err != nil {
		log.Fatalln("Test: Create temporery config:", err.Error())
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, _ = f.Write([]byte(`
USER_OTP_CONFIG:
    EXPIRATION_SECONDS_FORGOT_PASSWORD: 300
`))

	if _, _, err = config.LoadConfig(f.Name(), f.Name()); err != nil {
		panic("Test: Load config failed: " + err.Error())
	}

	m.Run()
}
