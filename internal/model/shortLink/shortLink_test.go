package shortLinkModel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewShortLink(t *testing.T) {
	req := &CreateShortLink{
		Reference:      "payment-ref-123",
		DestinationURL: "https://example.com/payment/123",
		UniqueID:       "unique-123",
		ExpiredAt:      time.Now().Add(24 * time.Hour),
	}

	shortLink := NewShortLink(req)

	assert.Equal(t, req.Reference, shortLink.Reference)
	assert.Equal(t, req.DestinationURL, shortLink.DestinationURL)
	assert.Equal(t, req.ExpiredAt.UTC(), shortLink.ExpiredAt)
	assert.NotEmpty(t, shortLink.Code)
}
