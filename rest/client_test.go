package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Request(t *testing.T) {
	tests := []struct {
		name       string
		givenUrl   string
		expectCode int
	}{
		{
			name:       "ok",
			givenUrl:   "https://httpbin.org/get",
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New()

			req, err := http.NewRequest("GET", tt.givenUrl, nil)
			if err != nil {
				fmt.Println(err.Error())
				t.Fail()
			}

			resp, body, err := client.Request(req)
			if err != nil {
				t.Fail()
			}

			assert.Equal(t, tt.expectCode, resp.StatusCode)
			assert.Greater(t, len(body), 0)
		})
	}
}
