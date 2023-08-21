package rest

import "net/http"

type MockTransport struct {
	called bool
}

func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.called = true
	return http.DefaultTransport.RoundTrip(req)
}
