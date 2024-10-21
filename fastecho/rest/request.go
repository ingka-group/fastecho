package rest

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/labstack/echo/v4"

	"io"
	"net/http"
)

type HTTPRequest http.Request

func NewRequest(p *Params) (*HTTPRequest, error) {
	r, err := http.NewRequest(p.RequestType, p.RequestURL, nil)
	if err != nil {
		return nil, err
	}

	r.Close = true

	return (*HTTPRequest)(r), nil
}

func NewRequestWithContext(ctx context.Context, p *Params) (*HTTPRequest, error) {
	r, err := http.NewRequestWithContext(ctx, p.RequestType, p.RequestURL, nil)
	if err != nil {
		return nil, err
	}

	r.Close = true

	return (*HTTPRequest)(r), nil
}

// SetQueryParams sets the query parameters given in the rest.Params to the request
func (r *HTTPRequest) SetQueryParams(p *Params) {
	if p.RequestQueryParams != nil {
		query := r.URL.Query()
		for _, param := range p.RequestQueryParams {
			query.Add(param.Key, param.Value)
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

		r.Body = io.NopCloser(bytes.NewBuffer(data))
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
