package userLoggedInDeviceModel

import "time"

type UserLoggedInDevice struct {
	UUID             string    `json:"uuid" db:"uuid"`
	UserID           string    `json:"userId" db:"user_id"`
	DeviceIdentifier string    `json:"deviceIdentifier" db:"device_identifier"`
	Status           string    `json:"status" db:"status"`
	AdditionalInfo   *string   `json:"additionalInfo" db:"additional_info"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time `json:"updatedAt" db:"updated_at"`
}

type UserLoggedInDeviceMetadata struct {
	IsRemember    bool      `json:"isRemember"`
	RememberUntil time.Time `json:"rememberUntil"`
}
