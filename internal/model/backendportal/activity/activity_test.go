package activityModel_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/activity"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var userID = uuid.NewString()
var formatTimeLayout = "2006-01-02 15:04:05 MST"

func TestActivitySerialization(t *testing.T) {
	// Create an instance of the Activity struct
	originalActivity := activityModel.Activity{
		ID:          uuid.NewString(),
		MerchantID:  uuid.NewString(),
		UserID:      &userID,
		Tag:         constant.TagAccount,
		Activity:    constant.ActivityUserLogin,
		ServiceName: "DoLogin",
		Parameter: &map[string]any{
			"email":    "jay@paper.id",
			"password": "*****",
		},
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(originalActivity)
	if err != nil {
		t.Fatalf("Error serializing to JSON: %v", err)
	}

	// Deserialize from JSON
	var decodedActivity activityModel.Activity
	err = json.Unmarshal(jsonData, &decodedActivity)
	if err != nil {
		t.Fatalf("Error deserializing from JSON: %v", err)
	}

	assertActivityModel(t, originalActivity, decodedActivity)
}

func TestActivityBSONSerialization(t *testing.T) {
	// Create an instance of the Activity struct
	originalActivity := activityModel.Activity{
		ID:          uuid.NewString(),
		MerchantID:  uuid.NewString(),
		UserID:      &userID,
		Tag:         constant.TagAccount,
		Activity:    constant.ActivityUserLogin,
		ServiceName: "DoLogin",
		Parameter: &map[string]any{
			"email":    "jay@paper.id",
			"password": "*****",
		},
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	// Serialize to BSON
	bsonData, err := bson.Marshal(originalActivity)
	if err != nil {
		t.Fatalf("Error serializing to BSON: %v", err)
	}

	// Deserialize from BSON
	var decodedActivity activityModel.Activity
	err = bson.Unmarshal(bsonData, &decodedActivity)
	if err != nil {
		t.Fatalf("Error deserializing from BSON: %v", err)
	}

	assertActivityModel(t, originalActivity, decodedActivity)
}

func TestToDTO(t *testing.T) {
	activityID := uuid.NewString()
	merchantID := uuid.NewString()

	testCases := []struct {
		Name     string
		Input    *activityModel.Activity
		Expected *activityModel.ActivityDTO
	}{
		{
			Name: "it should return activity dto",
			Input: &activityModel.Activity{
				ID:          activityID,
				MerchantID:  merchantID,
				UserID:      &userID,
				Tag:         constant.TagAccount,
				Activity:    constant.ActivityUserLogin,
				ServiceName: "DoLogin",
				Parameter: &map[string]any{
					"email":    "jay@paper.id",
					"password": "*****",
				},
				CreatedAt: util.TimeNow,
				UpdatedAt: util.TimeNow,
			},
			Expected: &activityModel.ActivityDTO{
				ID:          activityID,
				MerchantID:  merchantID,
				UserID:      &userID,
				Tag:         constant.TagAccount,
				Activity:    constant.ActivityUserLogin,
				ServiceName: "DoLogin",
				Parameter:   `{"email":"jay@paper.id","password":"*****"}`,
				CreatedAt:   util.TimeNow,
				UpdatedAt:   util.TimeNow,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			activityDTO := tc.Input.ToDTO()
			require.Equal(t, tc.Expected, activityDTO)
		})
	}
}

func TestFromDTO(t *testing.T) {
	activityID := uuid.NewString()
	merchantID := uuid.NewString()

	testCases := []struct {
		Name     string
		Input    *activityModel.ActivityDTO
		Expected *activityModel.Activity
	}{
		{
			Name: "it should return activity dto",
			Expected: &activityModel.Activity{
				ID:          activityID,
				MerchantID:  merchantID,
				UserID:      &userID,
				Tag:         constant.TagAccount,
				Activity:    constant.ActivityUserLogin,
				ServiceName: "DoLogin",
				Parameter: &map[string]any{
					"email":    "jay@paper.id",
					"password": "*****",
				},
				CreatedAt: util.TimeNow,
				UpdatedAt: util.TimeNow,
			},
			Input: &activityModel.ActivityDTO{
				ID:          activityID,
				MerchantID:  merchantID,
				UserID:      &userID,
				Tag:         constant.TagAccount,
				Activity:    constant.ActivityUserLogin,
				ServiceName: "DoLogin",
				Parameter:   `{"email":"jay@paper.id","password":"*****"}`,
				CreatedAt:   util.TimeNow,
				UpdatedAt:   util.TimeNow,
			},
		},
		{
			Name: "it should return activity dto with empty parameter",
			Expected: &activityModel.Activity{
				ID:          activityID,
				MerchantID:  merchantID,
				UserID:      &userID,
				Tag:         constant.TagAccount,
				Activity:    constant.ActivityUserLogin,
				ServiceName: "DoLogin",
				Parameter:   nil,
				CreatedAt:   util.TimeNow,
				UpdatedAt:   util.TimeNow,
			},
			Input: &activityModel.ActivityDTO{
				ID:          activityID,
				MerchantID:  merchantID,
				UserID:      &userID,
				Tag:         constant.TagAccount,
				Activity:    constant.ActivityUserLogin,
				ServiceName: "DoLogin",
				Parameter:   `{"}`,
				CreatedAt:   util.TimeNow,
				UpdatedAt:   util.TimeNow,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			activity := &activityModel.Activity{}
			activity.FromDTO(tc.Input)
			require.Equal(t, tc.Expected, activity)
		})
	}
}

func assertActivityModel(t *testing.T, originalActivity, decodedActivity activityModel.Activity) {
	assert.Equal(t, originalActivity.ID, decodedActivity.ID)
	assert.Equal(t, originalActivity.MerchantID, decodedActivity.MerchantID)
	assert.Equal(t, originalActivity.UserID, decodedActivity.UserID)
	assert.Equal(t, originalActivity.Tag, decodedActivity.Tag)
	assert.Equal(t, originalActivity.Activity, decodedActivity.Activity)
	assert.Equal(t, originalActivity.ServiceName, decodedActivity.ServiceName)
	assert.Equal(t, originalActivity.Parameter, decodedActivity.Parameter)
	assert.Equal(t, originalActivity.CreatedAt.Format(formatTimeLayout), decodedActivity.CreatedAt.Format(formatTimeLayout))
	assert.Equal(t, originalActivity.UpdatedAt.Format(formatTimeLayout), decodedActivity.UpdatedAt.Format(formatTimeLayout))
}
