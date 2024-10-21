package rest

import (
	"fmt"

	"github.com/stretchr/testify/assert"

	"net/http"
	"testing"
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

			req, err := http.NewRequest(http.MethodGet, tt.givenUrl, nil)
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

func TestClient_DoRequest(t *testing.T) {
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

			p := Params{
				RequestType: http.MethodGet,
				RequestURL:  tt.givenUrl,
			}

			resp, body, err := client.DoRequest(&p)
			if err != nil {
				t.Fail()
			}

			assert.Equal(t, tt.expectCode, resp.StatusCode)
			assert.Greater(t, len(body), 0)
		})
	}
}

func TestClientConfig_DoRequest(t *testing.T) {

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

			transport := MockTransport{}
			client := NewWithClientConfig(&ClientConfig{
				Transport: &transport,
			})
			p := Params{
				RequestType: http.MethodGet,
				RequestURL:  tt.givenUrl,
			}

			_, _, err := client.DoRequest(&p)
			if err != nil {
				t.Fail()
			}
			assert.True(t, transport.called)
		})
	}
}
