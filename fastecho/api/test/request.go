package test

import (
	"io"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ingka-group-digital/ocp-go-utils/stringutils"
)

// Params define the parameters of a request.
type Params struct {
	Path  map[string]string
	Query map[string][]string
	Body  string
}

// Request creates a new request and a new test service context to which it passes the required parameters.
func Request(it *IntegrationTest, method string, params Params) (echo.Context, *httptest.ResponseRecorder) {
	var reader io.Reader

	// If the body is not empty, read the body fixture and create a reader from it.
	// NOTE: The body expects the filename of the fixture, not the content.
	if !stringutils.IsEmpty(params.Body) {
		params.Body = it.Fixtures.ReadRequestBody(params.Body)
		reader = strings.NewReader(params.Body)
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

	// params.Query is a map with value as a slice of strings
	// This is required in case we want to pass multiple values for
	// the same query parameter. For example /v1/sales?status=active&status=inactive
	if params.Query != nil {
		q := ctx.QueryParams()
		for name, value := range params.Query {
			for i := range value {
				q.Add(name, value[i])
			}
		}
	}

	return ctx, response
}
