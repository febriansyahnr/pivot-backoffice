package util

import "github.com/google/uuid"

func GenerateUUID() uuid.UUID {
	if uuidV7, err := uuid.NewV7(); err == nil {
		return uuidV7
	}
	return uuid.New()
}
