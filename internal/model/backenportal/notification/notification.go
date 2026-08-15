package notification

import "time"

type PushNotification struct {
	RoutingKey string
	Payload    PushNotificationPayload
	Priority   uint8 // 0 to 9
	RetryCount int
}

type PushNotificationPayload struct {
	ID        string      `json:"id"`         // Unique ID (UUID)
	Subject   string      `json:"subject"`    // Notification Subject (Title)
	Type      string      `json:"type"`       // Notification Type
	Message   string      `json:"message"`    // Notification Message
	Data      interface{} `json:"data"`       // Data Sent Is Optional (can be null)
	CreatedAt time.Time   `json:"created_at"` // Created At
	Status    string      `json:"status"`
}
