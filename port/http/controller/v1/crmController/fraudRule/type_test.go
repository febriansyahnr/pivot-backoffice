package crmfraudrulecontroller

import (
	"testing"

	"github.com/go-playground/validator/v10"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	mockSvc := mocks.NewIFraudRuleService(t)
	mockValidator := validator.New()

	controller := New(mockSvc, mockValidator)

	assert.NotNil(t, controller)
}
