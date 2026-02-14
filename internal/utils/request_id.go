package utils

import (
	guuid "github.com/google/uuid"
)

func GenerateRequestID() string {
	id := guuid.New()
	return id.String()
}
