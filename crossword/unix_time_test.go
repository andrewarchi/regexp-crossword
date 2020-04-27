package crossword

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUnixTimeTwoWay(t *testing.T) {
	for i, test := range []struct {
		JSON, RFC string
	}{
		{"null", "0001-01-01T00:00:00Z"},
		{"1508348908", "2017-10-18T17:48:28Z"},
		{"1136239445", "2006-01-02T22:04:05Z"},
	} {
		var u UnixTime
		if err := json.Unmarshal([]byte(test.JSON), &u); err != nil {
			t.Errorf("test %d unmarshal: %v", i, err)
			continue
		}
		if f := u.In(time.UTC).Format(time.RFC3339); f != test.RFC {
			t.Errorf("test %d unmarshal mismatch: got %s, want %s", i, f, test.RFC)
		}
		m, err := json.Marshal(&u)
		if err != nil {
			t.Errorf("test %d marshal: %v", i, err)
			continue
		}
		if string(m) != test.JSON {
			t.Errorf("test %d marshal mismatch: got %q, want %q", i, string(m), test.JSON)
		}
	}
}
