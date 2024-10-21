package rest

import "net/http"

// HTTPResponse is an alias for http.Response
type HTTPResponse http.Response

// HasSuccessCode returns whether a status code is successful
func (r *HTTPResponse) HasSuccessCode() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
