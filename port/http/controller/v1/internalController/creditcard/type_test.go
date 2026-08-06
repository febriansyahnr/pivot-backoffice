package creditcard

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
)

func TestMain(m *testing.M) {
	_, _ = monitor.New("backend-portal", "0.0.0.0", "1234")

	m.Run()
}
