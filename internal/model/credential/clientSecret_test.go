package credential_test

import (
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/credential"

	"github.com/stretchr/testify/assert"
)

func TestClientSecret(t *testing.T) {
	data := ClientSecret{
		Secret:    "client-secret",
		UpdatedAt: time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
	}
	resp := data.ToResponse()
	assert.Equal(t, "client-secret", resp.Secret)
	assert.Equal(t, time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC), resp.LastUpdate)
	assert.Greater(t, resp.Time, time.Now().Add(-time.Minute).UTC().Unix())
}
