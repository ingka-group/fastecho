package test

import (
	"strings"

	"github.com/ingka-group-digital/ocp-go-utils/stringutils"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type Data struct {
	Name              string
	Method            string
	Params            Params
	Handler           func(ctx echo.Context) error
	ExpectResponse    string
	ExpectErrResponse bool
	ExpectCode        int
}

func Assert(it *IntegrationTest, tests []Data) {
	for _, tt := range tests {
		it.T.Log(it.T.Name(), "/", tt.Name)

		ctx, response := Request(it, tt.Method, tt.Params)
		err := tt.Handler(*ctx)
		if err != nil {
			it.T.Log(err.Error())
		}

		if tt.ExpectErrResponse {
			require.Error(it.T, err)

			echoErr, match := err.(*echo.HTTPError)
			require.True(it.T, match)

			response.Code = echoErr.Code
		} else {
			require.NoError(it.T, err)
		}

		require.Equal(it.T, tt.ExpectCode, response.Code)

		if !stringutils.IsEmpty(tt.ExpectResponse) {
			tt.ExpectResponse = it.Fixtures.ReadResponse(tt.ExpectResponse)

			require.Equal(it.T,
				minify(tt.ExpectResponse),
				strings.TrimSpace(response.Body.String()),
			)
		}
	}
}
