package activityModel

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ActivityFilterRequest struct {
	MerchantID     *string    `json:"merchantId"`
	StartCreatedAt *time.Time `json:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt"`
}

type CreateActivityReq struct {
	Tag      string         `json:"tag" validate:"required"`
	Activity string         `json:"activity" validate:"required"`
	Params   map[string]any `json:"params" validate:"omitnil,dive,required"`
}

func (c *CreateActivityReq) Record(merchantID, userID string, r *http.Request) *Activity {
	if c.Params == nil {
		c.Params = map[string]any{}
	}
	if r.Referer() != "" {
		c.Params["referer"] = r.Referer()
	}
	c.Params["user_agent"] = r.UserAgent()
	c.Params["remote_addr"] = r.RemoteAddr

	return &Activity{
		ID:          uuid.NewString(),
		MerchantID:  merchantID,
		UserID:      &userID,
		Tag:         c.Tag,
		Activity:    c.Activity,
		ServiceName: "Recorded from the activities endpoint",
		Parameter:   &c.Params,
		CreatedAt:   time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
}
