package feeService

import (
	"log"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
)

func TestMain(m *testing.M) {
	contentConfig := `
CREDIT_CARD_REFERENCES:
  DEFAULT_FEE:
    OTHER_CHANNEL:
      AMOUNT: 2000
      PERCENTAGE: 2.5
    CUSTOM_CHANNEL:
      LOCAL_VISA:
        AMOUNT: 1500
        PERCENTAGE: 2.75
`
	f, err := os.CreateTemp(os.TempDir(), "*.yaml")
	if err != nil {
		log.Fatalln("Failed create file config:", err.Error())
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, err = f.WriteString(contentConfig)
	if err != nil {
		log.Fatalln("Failed put file config:", err.Error())
	}

	_, _, err = config.LoadConfig(f.Name(), f.Name())
	if err != nil {
		log.Fatalln("Failed load file config:", err.Error())
	}

	m.Run()
}
