# core package

By using the core library you can set up the service configuration by providing your own custom env variables and specifying whether a postgres database is needed.

The service covers a lot of the basic setup like Env vars, (optional) database setup, validator, middlewares and running the service are abstracted away so that you don't have to think about it.

Note that tracing is enabled only if the `OTEL_TRACING` env var is set to true. The service context is accessible from within each handler function to add tracing:

```go
import (
	"github.com/ingka-group-digital/ocp-go-utils/api/core/context"
)

func (h *Handler) GetCountrySales(ctx echo.Context) error {
	sctx := context.GetServiceContext(ctx)
	log := sctx.ZapLogger

	...
}
```

The following example explains each step in the full service setup.

The required ENV vars are:
* GcpProjectID
* SwaggerUITitle
* ServiceName

Example usage:
```go
func Run() error {
	// provide custom env vars, optional custom props to service context and specify whether you need a DB connection set up
	props := make(map[string]interface{})
	props["myCoolProp"] = 123
	// props can also be nil if not needed

	s, err := core.NewService(config.EnvVar{
		consts.GcpProjectID: {},
		consts.SwaggerUITitle: {
			DefaultValue: "Sales Actuals API",
		},
		consts.ServiceName: {
			DefaultValue: "ffp-sales-actuals-connector-v3",
		},
	},
		props,
		consts.HasPostgresDb,
	)
	if err != nil {
		return err
	}

	// define custom validations for your handlers
	s.Validator.RegisterValidator("year_week", validator.ValidateYearWeek)
	if err != nil {
		return err
	}

	// bind the validator to echo
	s.E.Validator = s.Validator

	// write your own route config like this one:
	configureRoutes(s.E)

	// launch the service!
	return s.Run()
}

func configureRoutes(e *echo.Echo) error {
	v1 := e.Group("/v1")

	healthHandler := health.NewHealthHandler(nil)

	err := router.NewRouter(e).
		AddRoute(v1, "/health/ready", healthHandler.Ready, http.MethodGet).
		AddRoute(v1, "/health/live", healthHandler.Live, http.MethodGet).
		Init()
	if err != nil {
		return err
	}

	router.PrintRoutes(e)

	return nil
}
```

## Middlewares

Are simply added by using `s.E` which is your current instance of Echo:

```go
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	AllowOrigins: []string{"*"},
	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
}))
```

## Service context

You can inject custom properties into the service context via props. This object is of type `map[string]interface{}` so you can pass anything into your context to make it accessible in your endpoints

## Migration

This lib is using goose for migrations rather than gorm Automigrate
