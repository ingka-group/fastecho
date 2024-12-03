<img src="fastecho.png" alt="drawing" style="width:100%;"/>

Fastecho is a Go library that provides an easily configurable, ready-to-use echo server. It is a wrapper on top of the echo framework and it adds extra functionalities that are often required when setting up web servers.

# How to run
For specifics, check the detailed features below.
```go
    // load env vars
    err := core.Envs.SetEnv()
	if err != nil {
		log.Fatalf("failed to set environment variables: %s", err)
	}

    // set up a DB and pass it to handler as you like
	db, err := fastecho.NewDB()
	if err != nil {
		log.Fatalf("failed to connect to the database: %s", err)
	}

	config := fastecho.Config{
		ExtraEnvs:           core.Envs,
		ValidationRegistrar: validator.RegisterValidations,
		Routes: func(e *echo.Echo, r *router.Router) error {
			return configureRoutes(e, r, db)
		},
		ContextProps: map[string]interface{}{
			"my_property": "",
		},
		Opts: fastecho.Opts{
			Tracing: fastecho.TracingOpts{
				Skip:        !core.Envs[consts.OtelTracing].BooleanValue,
				ServiceName: core.Envs[consts.OtelServiceName].Value,
			},
			HealthChecks: fastecho.HealthChecksOpts{
				Skip: false,
				DB:   db,
			},
		},
	}

	// Starting service...
	if err := fastecho.Run(&config); err != nil {
		log.Fatalf("Service stopped! \n %s", err)
	}
```

# Features:

### Logger
We integrated `go.uber.org/zap`
### Dynamic request context
You can inject custom properties into the service context via props. This object is of type `any` so you can pass anything into your context to make it accessible in your endpoints
```go
func (h *Handler) GetData(ctx echo.Context) error {
    sctx := context.GetServiceContext[any](ctx)
    log := sctx.ZapLogger

    ...
}
```
### Endpoint router
The router is providing a couple of preset endpoints for swagger, monitoring and health checks but custom endpoints can also be injected. The router wrapper in the example above can be used to register additional endpoints.

Example:
```go
func configureRoutes(e *echo.Echo, r *router.Router, db *gorm.DB) error {
	myHandler := NewHandler(db)

	v1 := e.Group("/v1")
	myGroup := v1.Group("/example")

	router.AddRoute(r, myGroup, "/data", myHandler, http.MethodGet)
	return nil
}
```
### Request validation
Custom validation can be registered using the provided validator. You need to define a function in which you register custom validations and then add it to the config.
```go
func RegisterValidations(validator *router.Validator) error {
	validator.Vdt.RegisterStructValidation(daterange.ValidateBasicDateRange(), daterange.BasicDateRange{})

	return nil
}
```
### Middleware injection
Custom middleware can be injected easily just like routes.
```go
func configureRoutes(e *echo.Echo, r *router.Router, db *gorm.DB) error {
	v1 := e.Group("/v1")
	myGroup := v1.Group("/data")

	myGroup.Use(middleware.MyCustomMiddleware())

	return nil
}
```
### Swagger
Swagger is baked into the router wrapper. The title and path can be used via these environment variables:

`swaggerUITitle`
`swaggerJSONPath`

The swagger documentation is configured on the root path suffixed with `/swagger/`.
### Health probe endpoints
The health endpoints are configured on the root path suffixed with `/health/live` and `/health/ready`.
### Environment variables
Environment variables are read by default from the environment or from a `.env` file in the root of the directory.

The required ENV vars are:
* SwaggerUITitle
* ServiceName

Here's an example on how to define env vars and how to load them before starting fastecho:
```go
// define variables
var (
	Envs = env.Map{
		consts.SwaggerUITitle: {
			DefaultValue: "FFP Sales forecast connector API",
		},
		consts.OtelServiceName: {
			DefaultValue: "ffp-sales-forecast-connector-v1",
		},
		consts.OtelTracing: {
			DefaultValue: "false",
			IsBoolean:    true,
		},
		// The variables related to the DB are already defined in fastecho
	}
)

// load them
    err := core.Envs.SetEnv()
	if err != nil {
		log.Fatalf("failed to set environment variables: %s", err)
	}
```
### OTEL tracing (optional)
Tracing is enabled only if the `OTEL_TRACING` env var is set to true.
### Database (optional)
Fastecho has an optional postgres DB connection baked into it using `gorm`. We are using `goose` for migrations rather than gorm Automigrate. The migrations are expected to be under `db/migrations` in the root of your folder.

# Miscellaneous
Fastecho is fully compatible with [Echoprobe](https://github.com/ingka-group/echoprobe)


## Contributing
Please read [CONTRIBUTING](./CONTRIBUTING.md) for more details about making a contribution to this open source project and ensure that you follow our [CODE_OF_CONDUCT](./CODE_OF_CONDUCT.md).


## Contact
If you have any other issues or questions regarding this project, feel free to contact one of the [code owners/maintainers](.github/CODEOWNERS) for a more in-depth discussion.


## Licence
This open source project is licensed under the "Apache-2.0", read the [LICENCE](./LICENCE.md) terms for more details.
