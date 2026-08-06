package feeRepository

import (
	"testing"

	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/test"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Setup
	mockDB := mysqlMocks.NewIMySqlExt(t)
	_, pdkLog, _ := test.SetupLogger()

	// Test
	repo := New(mockDB, pdkLog)

	// Assertions
	assert.NotNil(t, repo, "Repository should not be nil")

	// Check if it's the correct type
	_, ok := repo.(*feeRepository)
	assert.True(t, ok, "Repository should be of type *feeRepository")
}
