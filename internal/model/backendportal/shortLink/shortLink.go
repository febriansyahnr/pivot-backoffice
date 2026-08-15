package shortLinkModel

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

type ShortLink struct {
	UUID           string    `json:"uuid" db:"uuid"`
	Reference      string    `json:"reference" db:"reference"`
	Code           string    `json:"code" db:"code"`
	DestinationURL string    `json:"destinationUrl" db:"destination_url"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
	ExpiredAt      time.Time `json:"expiredAt" db:"expired_at"`
	ShortLinkURL   string    `json:"shortLinkUrl"`
}

type CreateShortLink struct {
	Reference          string
	DestinationURL     string
	UniqueID           string    // optional, enforce uniqueness otherwise system will define unique id
	ExpiredAt          time.Time // will parsed to UTC format
	ShortLinkURLFormat string
}

func NewShortLink(req *CreateShortLink) *ShortLink {
	now := time.Now().UTC()
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}

	code := util.GenerateShortCode(req.UniqueID)
	return &ShortLink{
		UUID:           id.String(),
		Reference:      req.Reference,
		Code:           code,
		DestinationURL: req.DestinationURL,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiredAt:      req.ExpiredAt.UTC(),
		ShortLinkURL:   fmt.Sprintf(req.ShortLinkURLFormat, code),
	}
}
