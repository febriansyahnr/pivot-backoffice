package credential_test

import (
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/credential"

	"github.com/stretchr/testify/assert"
)

func TestCredentialDashboard(t *testing.T) {
	data := CredentialDashboard{
		ClientID: "unique-client-id",
		ClientSecrets: []ClientSecretSummary{
			{
				ID:         "unique-secret-id",
				KeyName:    "Client Secret 1",
				LastUpdate: time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	assert.Equal(t, &CredentialDashboardResp{
		ClientID: "unique-client-id",
		ClientSecrets: []ClientSecretSummary{
			{
				ID:         "unique-secret-id",
				KeyName:    "Client Secret 1",
				LastUpdate: time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
			},
		},
	}, data.ToResponse())
}
