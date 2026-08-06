package passwordHistories

import "time"

type PasswordHistories struct {
	UUID           string    `json:"uuid" db:"uuid"`
	UserID         string    `json:"userId" db:"user_id"`
	PasswordHashes string    `json:"passwordHash" db:"password_hash"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
}
