package fdscommon

type UpdateResponse struct {
	Success bool       `json:"success"`
	Code    *string    `json:"code,omitempty"`
	Source  *string    `json:"source,omitempty"`
	Message any        `json:"message,omitempty"`
	Data    UpdateData `json:"data"`
}

type UpdateData struct {
	ID    string `json:"id"`
	Link  string `json:"link"`
	Timer int    `json:"timer"`
}
