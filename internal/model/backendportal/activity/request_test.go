package activityModel_test

import (
	"net/http"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateActivityReq(t *testing.T) {
	req := CreateActivityReq{
		Tag:      "credential-settings",
		Activity: "User copy client ID",
	}
	r := &http.Request{
		Header: http.Header{
			"Referer": {"ABC"},
		},
	}
	data := req.Record(uuid.NewString(), uuid.NewString(), r)

	assert.NoError(t, uuid.Validate(data.ID))
	assert.Equal(t, req.Tag, data.Tag)
	assert.Equal(t, req.Activity, data.Activity)
	assert.Equal(t, "Recorded from the activities endpoint", data.ServiceName)
	assert.NoError(t, uuid.Validate(data.MerchantID))
	assert.NoError(t, uuid.Validate(*data.UserID))
}
