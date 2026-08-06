package vendor

import (
	"testing"

	"github.com/go-playground/validator/v10"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	svc := mockSvc.NewIVendorService(t)
	validate := validator.New()

	ctrl := New(svc, validate)

	assert.NotNil(t, ctrl)
}
