package test

import (
	"fmt"
	"net/http"

	"github.com/h2non/gock"

	"github.com/ingka-group-digital/ocp-go-utils/stringutils"
)

// MockCall represents a mocked API call used in tests
type MockCall struct {
	Function func(config *MockConfig)
	Config   *MockConfig
}

// MockConfig configures a Mock
type MockConfig struct {
	Method     string
	StatusCode int
	UrlPath    string
	Response   string
}

// Mock is the struct that gives access to all the mocks
type Mock struct {
	baseURL string
}

// NewMock creates a new Mock
func NewMock(baseURL string) *Mock {
	return &Mock{
		baseURL: baseURL,
	}
}

// TearDown removes all the registered mocks
func (m *Mock) TearDown() {
	gock.Off()
}

// Debug is used to print the request URL and the mock returned for that particular request
func (m *Mock) Debug() {
	gock.Observe(func(req *http.Request, mock gock.Mock) {
		debug := fmt.Sprintf(
			"\n-- MOCK START\n"+
				"%s - %d \n"+
				"%s \n"+
				"-- MOCK END\n",
			req.URL, mock.Response().StatusCode, string(mock.Response().BodyBuffer),
		)

		fmt.Println(debug)
	})
}

func (m *Mock) SetJSON(response *gock.Response, config *MockConfig) {
	var f Fixtures
	if !stringutils.IsEmpty(config.Response) {
		response.JSON(
			f.ReadFixture(
				fmt.Sprintf("%s.json", config.Response),
				"mocks",
			),
		)
	}
}

func (m *Mock) MockRequest(config *MockConfig) {
	if config.StatusCode == 0 {
		config.StatusCode = http.StatusOK
	}

	request := gock.New(m.baseURL)

	switch config.Method {
	case http.MethodGet:
		request.Get(config.UrlPath)
	case http.MethodDelete:
		request.Delete(config.UrlPath)
	case http.MethodPost:
		request.Post(config.UrlPath)
	case http.MethodPut:
		request.Put(config.UrlPath)
	case http.MethodPatch:
		request.Patch(config.UrlPath)
	default:
		request.Get(config.UrlPath)
	}

	response := request.Reply(
		config.StatusCode,
	)

	m.SetJSON(response, config)
}
