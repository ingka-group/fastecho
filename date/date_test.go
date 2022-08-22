package date

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/ocp-go-utils/testutil"
)

func TestDate_IKEAWeek(t *testing.T) {
	// load tests from the gzipped json file
	var tests []map[string]any

	testFile := "testdata/weeks.json.gz"
	payload := testutil.ReadGzippedTestFile(testFile)
	err := json.Unmarshal(payload.Bytes(), &tests)
	if err != nil {
		log.Fatalf("Cannot unmarshal json test set: %v", err)
	}

	var date string
	var week int

	for _, m := range tests {
		date, week = "", 0
		for k, v := range m {
			if k == "date" {
				date = v.(string)
			} else if k == "week" {
				week = int(v.(float64))
			} else {
				panic("unexpected key")
			}
		}

		t.Run(date, func(t *testing.T) {
			tm, err := time.Parse("2006-01-02", date)
			if err != nil {
				log.Fatalf("invalid date format for `%v`: %v", date, err)
			}

			_, gotWeek := IKEAWeek(tm.Year(), int(tm.Month()), tm.Day())
			assert.Equalf(t, gotWeek, week, "wrong week")
		})
	}
}
