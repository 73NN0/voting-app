package db

import "github.com/google/uuid"

func stringToUuid(s string) uuid.UUID {
	return uuid.MustParse(s)
}

func uuidToString(u uuid.UUID) string {
	return u.String()
}
