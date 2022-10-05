package rest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HTTPRequest http.Request

func NewRequest(p *Params) (*HTTPRequest, error) {
	r, err := http.NewRequest(p.RequestType, p.RequestURL, nil)
	if err != nil {
		return nil, err
	}

	return (*HTTPRequest)(r), nil
}

// SetQueryParams sets the query parameters given in the rest.Params to the request
func (r *HTTPRequest) SetQueryParams(p *Params) {
	if p.RequestQueryParams != nil {
		query := r.URL.Query()
		for name, value := range p.RequestQueryParams {
			query.Add(name, value)
		}

		r.URL.RawQuery = query.Encode()
	}
}

// SetRequestBodyJSON sets the request body as a JSON, to the http.Request given the rest.Params
func (r *HTTPRequest) SetRequestBodyJSON(p *Params) error {
	if p.RequestBody != nil {
		data, err := json.Marshal(p.RequestBody)
		if err != nil {
			return err
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}

	return nil
}

// SetRequestHeaders sets all the headers given in the rest.Params to the request
func (r *HTTPRequest) SetRequestHeaders(p *Params) {
	if p.RequestHeaders != nil {
		for name, value := range p.RequestHeaders {
			r.Header.Set(name, value)
		}
	}

	if p.RequestID != nil {
		r.Header.Set(echo.HeaderXRequestID, *p.RequestID)
	}
}
