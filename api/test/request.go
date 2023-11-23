package test

import (
	"io"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"
)

type Params struct {
	Path  map[string]string
	Query map[string]string
	Body  *string
}

// Request creates a new request and a new test service context to which it passes the required parameters
func Request(it *IntegrationTest, method string, params Params) (*echo.Context, *httptest.ResponseRecorder) {
	var reader io.Reader
	if params.Body != nil {
		reader = strings.NewReader(*params.Body)
	}

	// 2nd parameter is supposed to be the URI but since we inject everything via context, we can ignore this
	req := httptest.NewRequest(method, "/", reader)
	req.Header.Set(
		echo.HeaderContentType,
		echo.MIMEApplicationJSON,
	)

	response := httptest.NewRecorder()
	ctx := it.Echo.NewContext(req, response)

	if params.Path != nil {
		var paramNames []string
		var paramValues []string

		for name, value := range params.Path {
			paramNames = append(paramNames, name)
			paramValues = append(paramValues, value)
		}

		ctx.SetParamNames(paramNames...)
		ctx.SetParamValues(paramValues...)
	}

	if params.Query != nil {
		q := ctx.QueryParams()
		for name, value := range params.Query {
			q.Add(name, value)
		}
	}

	return &ctx, response
}
