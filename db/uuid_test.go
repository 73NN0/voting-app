package db

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUIDToString(t *testing.T) {
	uuid1 := uuid.New()
	uuid1Str := uuid1.String()
	tests := []struct {
		name string
		want string
		got  uuid.UUID
	}{
		{name: "uuid to string", want: uuid1Str, got: uuid1},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			temp := uuidToString(tt.got)

			if temp != tt.want {
				t.Fatalf("got %s, want %s", tt.got, tt.want)
			}
		})
	}
}

func TestStringToUUID(t *testing.T) {
	uuid1 := uuid.New()
	uuid1Str := uuid1.String()
	tests := []struct {
		name string
		want uuid.UUID
		got  string
	}{
		{name: "uuid to string", want: uuid1, got: uuid1Str},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			temp := stringToUuid(tt.got)

			if temp != tt.want {
				t.Fatalf("got %s, want %s", tt.got, tt.want)
			}
		})
	}
}
