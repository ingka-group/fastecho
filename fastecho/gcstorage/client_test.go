package gcstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeconstructGsURI(t *testing.T) {
	tests := []struct {
		name      string
		given     string
		expect    []string
		expectErr bool
	}{
		{
			name:   "ok",
			given:  "gs://ocp-mlflow-dev/xfer/weekly-forecasts/output/scheduled__2022-01-25T22:00:00+00:00/fcp-baseline-latest-NL.json",
			expect: []string{"ocp-mlflow-dev", "xfer/weekly-forecasts/output/scheduled__2022-01-25T22:00:00+00:00/fcp-baseline-latest-NL.json"},
		},
		{
			name:      "error: invalid gs uri",
			given:     "ocp-mlflow-dev/xfer/weekly-forecasts/output/scheduled__2022-01-25T22:00:00+00:00/fcp-baseline-latest-NL.json",
			expectErr: true,
		},
		{
			name:      "error: non gs uri",
			given:     "https://ocp-mlflow-dev/xfer/weekly-forecasts/output/scheduled__2022-01-25T22:00:00+00:00/fcp-baseline-latest-NL.json",
			expectErr: true,
		},
		{
			name:      "error: empty gs uri",
			given:     "gs:// / ",
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucketName, filePath, err := DeconstructGsURI(tt.given)
			if err != nil {
				assert.True(t, tt.expectErr)
			} else {
				assert.False(t, tt.expectErr)
				assert.Equal(t, bucketName, tt.expect[0])
				assert.Equal(t, filePath, tt.expect[1])
			}
		})
	}
}

func TestDeconstructPath(t *testing.T) {
	tests := []struct {
		name      string
		given     string
		expect    []string
		expectErr bool
	}{
		{
			name:   "ok",
			given:  "xfer/2022-01-01T00:00:00/forecast-NL.json",
			expect: []string{"xfer/2022-01-01T00:00:00", "forecast-NL.json"},
		},
		{
			name:      "error: invalid file path",
			given:     "xfer-2022-01-01T00:00:00-forecast-NL.json",
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, fileName, err := DeconstructPath(tt.given)
			if err != nil {
				assert.True(t, tt.expectErr)
			} else {
				assert.False(t, tt.expectErr)
				assert.Equal(t, location, tt.expect[0])
				assert.Equal(t, fileName, tt.expect[1])
			}
		})
	}
}

func TestConstructGsURI(t *testing.T) {
	type args struct {
		bucketName string
		filePath   string
	}

	tests := []struct {
		name   string
		given  args
		expect string
	}{
		{
			name: "ok",
			given: args{
				bucketName: "ocp-bff-v1",
				filePath:   "xfer/2022-01-01T00:00:00/forecast-NL.json",
			},
			expect: "gs://ocp-bff-v1/xfer/2022-01-01T00:00:00/forecast-NL.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gsUri := ConstructGsURI(tt.given.bucketName, tt.given.filePath)
			assert.Equal(t, gsUri, tt.expect)
		})
	}
}
