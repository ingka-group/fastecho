package rest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// RESTDoer is the interface that defines the way to perform HTTP requests
type RESTDoer interface {
	DoRequest(p *Params) (*HTTPResponse, []byte, error)
	Request(req *http.Request) (*HTTPResponse, []byte, error)
}

// Client is the struct that implements the RESTDoer
type Client struct {
	client *http.Client
}

// New creates a new REST Client
func New() *Client {
	return &Client{
		client: &http.Client{},
	}
}

// Params describes the request parameters
type Params struct {
	RequestType        string
	RequestHeaders     map[string]string
	RequestURL         string
	RequestQueryParams map[string]string
	RequestBody        interface{}
	RequestID          *string
}

// DoRequest builds and performs a request given the rest.Params
func (c *Client) DoRequest(p *Params) (*HTTPResponse, []byte, error) {
	r, err := NewRequest(p)
	if err != nil {
		return nil, nil, err
	}

	r.SetQueryParams(p)
	r.SetRequestHeaders(p)
	err = r.SetRequestBodyJSON(p)
	if err != nil {
		return nil, nil, err
	}

	resp, body, err := c.Request(
		(*http.Request)(r),
	)
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}

// Request performs an HTTP request
func (c *Client) Request(req *http.Request) (*HTTPResponse, []byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error while executing HTTP request: %w", err)
	}

	defer func(body io.Closer) {
		if err = body.Close(); err != nil {
			err = fmt.Errorf("error while closing request body: %w", err)
		}
	}(resp.Body)

	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error while reading request body: %w", err)
	}

	return (*HTTPResponse)(resp), body, nil
}
