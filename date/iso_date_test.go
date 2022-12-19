package date

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestISODate_UnmarshalJSON(t *testing.T) {
	type result struct {
		Created ISODate `json:"created"`
	}

	tests := []struct {
		name          string
		given         string
		expectCreated time.Time
	}{
		{
			name:          "ok: date given",
			given:         "{\"created\": \"2021-01-10\"}",
			expectCreated: time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := result{}
			err := json.Unmarshal([]byte(tt.given), &data)
			if err != nil {
				t.Fail()
			}

			assert.Equal(t, data.Created.Time, tt.expectCreated)
		})
	}
}
