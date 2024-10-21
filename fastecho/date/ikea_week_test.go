package date

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ingka-group-digital/ocp-go-utils/gzip"
)

func TestDate_IKEAWeek(t *testing.T) {
	// load tests from the gzipped json file
	var tests []map[string]any

	testFile := "testdata/weeks.json.gz"
	payload := gzip.ReadFile(testFile)
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

		tm, err := time.Parse("2006-01-02", date)
		if err != nil {
			log.Fatalf("invalid date format for `%v`: %v", date, err)
		}

		_, gotWeek := IKEAWeek(tm.Year(), int(tm.Month()), tm.Day())
		assert.Equalf(t, gotWeek, week, "wrong week - %s", date)
	}
}

func TestDate_IKEAWeekFirstDay(t *testing.T) {
	// load tests from the gzipped json file
	var tests []map[string]string

	testFile := "testdata/firstday.json"
	fileData, err := os.ReadFile(testFile)
	if err != nil {
		log.Fatalf("Cannot read file: %v", err)
	}
	err = json.Unmarshal(fileData, &tests)
	if err != nil {
		log.Fatalf("Cannot unmarshal json test set: %v", err)
	}

	var date string
	var week int
	var year int

	for _, m := range tests {
		date, year, week = "", 0, 0
		for k, v := range m {
			if k == "date" {
				date = v
			} else if k == "yearweek" {
				yw := strings.Split(v, "-")
				year, _ = strconv.Atoi(yw[0])
				week, _ = strconv.Atoi(yw[1])
			} else {
				panic("unexpected key")
			}
		}

		t.Run(date, func(t *testing.T) {
			expected, err := time.Parse("2006-01-02", date)
			if err != nil {
				log.Fatalf("invalid date format for `%v`: %v", date, err)
			}

			firstDay := IKEAWeekFirstDay(year, week)
			assert.Equalf(t, firstDay, expected, "wrong first day")
		})
	}
}
