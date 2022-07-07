package rest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// HTTPResponse is an alias for http.Response
type HTTPResponse http.Response

// RESTDoer is the interface that defines the way to perform HTTP requests
type RESTDoer interface {
	Request(req *http.Request) (*HTTPResponse, []byte, error)
}

// Client is the struct that implements the RESTDoer
type Client struct {
	client *http.Client
}

// NewClient creates a new Client
func NewClient() *Client {
	return &Client{
		client: &http.Client{},
	}
}

// Request performs an HTTP request
func (c *Client) Request(req *http.Request) (*HTTPResponse, []byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error while executing HTTP request: %w", err)
	}

	defer func(Body io.ReadCloser) error {
		err = Body.Close()
		if err != nil {
			return fmt.Errorf("error while closing request body: %w", err)
		}

		return nil
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

// HasSuccessStatusCode returns whether a status code is successful
func (r *HTTPResponse) HasSuccessStatusCode() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
