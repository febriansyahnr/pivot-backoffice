package commonModel

import (
	"encoding/json"
	"time"
)

type ExportResponse struct {
	DownloadURL string    `json:"downloadURL" redis:"downloadURL"`
	ExpiresAt   time.Time `json:"expiresAt" redis:"expiresAt"`
}

func (e ExportResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ExportResponse) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, e)
}
