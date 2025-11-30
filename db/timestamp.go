package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Timestamp struct {
	time.Time
}

var (
	formats []string = []string{
		time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
		"2006-01-02 15:04:05", // Format SQLite par défaut
		"2006-01-02",          // Date seule
	}
)

// (DB → Go)
func (t *Timestamp) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into Timestamp", value)
	}

	var parsed time.Time
	var err error

	for _, layout := range formats {
		parsed, err = time.Parse(layout, str)
		if err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("cannot parse %q as timestamp: %w", str, err)
}

func (t Timestamp) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time.Format(time.RFC3339Nano), nil
}
