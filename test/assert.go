package test

import (
	"github.com/ingka-group/echoprobe"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/ingka-group/fastecho/context"
	"github.com/ingka-group/fastecho/router"
)

// AssertAllWithCustomContext is a helper function to run multiple tests in a single test function.
// Before asserting a test, the function prepares the custom context.ServiceContext and calls the handler function.
func AssertAllWithCustomContext(
	it *echoprobe.IntegrationTest,
	tt []echoprobe.Data,
	registerValidator func(validator *router.Validator) error,
	middlewares []func() echo.MiddlewareFunc,
	props map[string]interface{},
) {
	zapLogger := zap.NewNop()

	// register validator and custom validations
	v, err := router.NewValidator()
	if err != nil {
		it.T.Fatalf("validator error: %v", err)
	}
	registerValidator(v)
	it.Echo.Validator = v

	for _, t := range tt {
		ctx, response := echoprobe.Request(it, t.Method, t.Params)

		// middleware to set the service context
		sctxMiddlewareFn := context.ServiceContextMiddleware[any](zapLogger, nil, props)

		// register custom middlewares
		for _, mw := range middlewares {
			next := sctxMiddlewareFn
			sctxMiddlewareFn = func(handler echo.HandlerFunc) echo.HandlerFunc {
				return mw()(next(handler))
			}
		}

		// Bind the middlewares to the handler function
		// NOTE: This is not a good approach as if we write more tests
		// for other handlers, the "SelectForecastTableMiddleware" can be
		// obsolete.
		h := sctxMiddlewareFn(t.Handler)

		// Execute the handler function
		// Since the handler will use the ServiceContextMiddleware, which converts the echo.Context
		// to context.ServiceContext we can pass the echo.Context to the handler.
		err = h(ctx)

		if err != nil {
			it.T.Log(err.Error())
		}

		echoprobe.Assert(it, &t, &echoprobe.HandlerResult{
			Err:      err,
			Response: response,
		})
	}
}
