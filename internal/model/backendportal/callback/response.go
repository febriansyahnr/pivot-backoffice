package callback_model

import (
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"

	"github.com/google/uuid"
)

type RegisterCallbackResponse struct {
	CallbackMasterID uuid.UUID `json:"callbackMasterId"`
	CallbackName     string    `json:"name"`
	CallbackID       uuid.UUID `json:"callbackId"`
	BaseURL          string    `json:"baseUrl"`
	URL              string    `json:"url"`
	Description      string    `json:"description"`
}

type CallbackURLSettingResp struct {
	MasterID           string       `json:"masterId" db:"master_id"`
	MasterName         string       `json:"masterName" db:"master_name"`
	CallbackID         c.NullString `json:"callbackId" db:"callback_id"`
	CallbackURL        c.NullString `json:"callbackUrl" db:"callback_url"`
	CallbackBaseURL    c.NullString `json:"callbackBaseURL" db:"callback_base_url"`
	CallbackLastUpdate c.NullTime   `json:"callbackLastUpdate" db:"updated_at"`

	CallbackTemplate callbackModel.CallbackPayloadRequest `json:"callbackTemplate" db:"-"`
}

type CallbackAPIKeyResp struct {
	APIKey  string `json:"apiKey" db:"callback_api_key"`
	Version uint   `json:"-" db:"callback_api_key_version"`
}

type TestAndSaveCallbackURLResp struct {
	Status      bool                `json:"status"`
	Information CallbackURLInfoResp `json:"information"`
	Body        interface{}         `json:"body"`
	Duration    string              `json:"duration"`
	RequestID   string              `json:"requestId"`
}

type CallbackURLInfoResp struct {
	Product        string    `json:"product"`
	Event          string    `json:"event"`
	URL            string    `json:"url"`
	Time           time.Time `json:"time"`
	CallbackID     string    `json:"callbackId"`
	CallbackToken  string    `json:"callbackToken"`
	CallbackType   string    `json:"callbackType"`
	CallbackLength string    `json:"callbackLength"`
	CallbackLogID  string    `json:"callbackLogId"`
}

type SendMerchantCallbackResponse struct {
	StatusCode     int                                 `json:"statusCode"`
	ResponseBody   []byte                              `json:"responseBody"`
	AdditionalInfo *SendMerchantCallbackAdditionalInfo `json:"additionalInfo"`
}

type SendMerchantCallbackAdditionalInfo struct {
	Headers map[string]string `json:"headers"`
	URL     string            `json:"url"`
}
