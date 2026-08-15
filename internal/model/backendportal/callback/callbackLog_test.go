package callback_model_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
)

const ExpectedNoErr = "Expected no error, got %v"

func TestMarshalJSON(t *testing.T) {
	event := constant.CallbackEventPayoutDone
	response := "{}"
	c := callbackModel.CallbackLogWithMaster{
		Event:      &event,
		UUID:       uuid.New(),
		CallbackID: uuid.New(),
		Type:       constant.CallbackNameDisbursement,
		BaseURL:    nil,
		URL:        "http://localhost",
		Request:    "{}",
		Response:   &response,
		Status:     constant.CallbackStatusDelivered,
		Retry:      0,
		CreatedAt:  util.TimeNow,
		UpdatedAt:  util.TimeNow,
	}

	expected := `{"event":"PAYOUT.DONE","eventTitle":"Payout Done"}`

	// Marshal the CallbackLogWithMaster instance to JSON
	jsonData, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf(ExpectedNoErr, err)
	}

	// Check if the output matches the expected JSON string
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf(ExpectedNoErr, err)
	}

	var expectedResult map[string]interface{}
	err = json.Unmarshal([]byte(expected), &expectedResult)
	if err != nil {
		t.Fatalf(ExpectedNoErr, err)
	}

	for key, expectedValue := range expectedResult {
		if result[key] != expectedValue {
			t.Errorf("For key %v, expected %v, got %v", key, expectedValue, result[key])
		}
	}
}

func TestToCallbackLLog(t *testing.T) {
	callbackUUID := uuid.New()
	clm := callbackModel.CallbackLogWithMaster{
		UUID: callbackUUID,
	}

	expectedCallbackLog := &callbackModel.CallbackLog{
		UUID: callbackUUID,
	}
	assert.Equal(t, expectedCallbackLog, clm.ToCallbackLog())
}
