package rest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// RESTDoer is the interface that defines the way to perform HTTP requests
type RESTDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client is the struct that performs the HTTP requests
type Client struct {
	url    string
	client RESTDoer
}

// NewClient creates a new Client
func NewClient(url string) *Client {
	return &Client{
		url:    url,
		client: &http.Client{},
	}
}

func (c *Client) doRequest(r *http.Request) ([]byte, int, error) {
	resp, err := c.client.Do(r)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error while executing HTTP request: %w", err)
	}

	defer func(Body io.ReadCloser) error {
		err := Body.Close()
		if err != nil {
			return fmt.Errorf("error while closing request body: %w", err)
		}

		return nil
	}(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error while reading request body: %w", err)
	}

	return body, resp.StatusCode, nil
}

// SetClient sets the REST client, useful for mocking
func (c *Client) SetClient(r RESTDoer) *Client {
	c.client = r
	return c
}

// IsSuccessStatusCode returns whether a status code is successful
func IsSuccessStatusCode(c int) bool {
	return c >= 200 && c < 300
}
