package crossword

import (
	"strconv"
	"time"
)

// UnixTime is a time formatted as a Unix timestamp.
type UnixTime struct {
	time.Time
}

// MarshalJSON implements the json.Marshaler interface.
// The time is a number representing a Unix timestamp.
func (t UnixTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(t.Unix(), 10)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a number representing a Unix timestamp.
func (t *UnixTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	sec, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*t = UnixTime{time.Unix(sec, 0)}
	return nil
}
