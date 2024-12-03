# FastEcho

<img src="fastecho.png" alt="drawing" style="width:100%;"/>

By using the core library you can set up the service configuration by providing your own custom env variables and specifying whether a postgres database is needed.

The service covers a lot of the basic setup like Env vars, (optional) database setup, validator, middlewares and running the service are abstracted away so that you don't have to think about it.

Note that tracing is enabled only if the `OTEL_TRACING` env var is set to true. The service context is accessible from within each handler function to add tracing:

```go
func (h *Handler) GetData(ctx echo.Context) error {
    sctx := context.GetServiceContext[any](ctx)
    log := sctx.ZapLogger

    ...
}
```

The following example explains each step in the full service setup.

The required ENV vars are:
* SwaggerUITitle
* ServiceName

However, you can pass extra variables if required by your service.

Example usage:
```go
func Run() {
    // provide custom env vars, optional custom props to service context and specify whether you need a DB connection set up
    props := make(map[string]interface{})
    props["myCoolProp"] = 123
    // props can also be nil if not needed

    s, err := core.NewServer(config.EnvVar{
        "SWAGGER_UI_TITLE": {
            DefaultValue: "My Service",
        },
        "SERVICE_NAME": {
            DefaultValue: "my-service",
        }
        "EXTRA_VAR_1": {
            DefaultValue: "value",
        },
        "EXTRA_VAR_2": {
            DefaultValue: "value",
        },
    },
        props,
        true,   // withPostgres
    )
    if err != nil {
        log.Fatalf("Failed to initialize server! \n %s", err)
    }

    // define custom validations for your handlers
    registerValidations(s.Validator)

    // write your own route config like this one:
    configureRoutes(s.Echo)
    if err != nil {
        log.Fatalf("Failed to configure routes! \n %s", err)
    }

    // launch the service!
    log.Println("Starting service...")
    if err := s.Run(); err != nil {
        log.Fatalf("Service stopped! \n %s", err)
    }
}

func configureRoutes(e *echo.Echo) error {
    v1 := e.Group("/v1")

    healthHandler := health.NewHealthHandler(nil)

    err := router.NewRouter(e).
        AddRoute(v1, "/health/ready", healthHandler.Ready, http.MethodGet).
        AddRoute(v1, "/health/live", healthHandler.Live, http.MethodGet).
        AddMetrics(e).
        AddSwagger(e).
        Init()
    if err != nil {
        return err
    }

    router.PrintRoutes(e)

    return nil
}

func registerValidations(validator *router.Validator) {
    validator.Vdt.RegisterStructValidation(daterange.ValidateISODateRangeBasic(), daterange.ISODateRangeBasic{})
}
```

## Service context

You can inject custom properties into the service context via props. This object is of type `any` so you can pass anything into your context to make it accessible in your endpoints

## Migration

This lib is using `goose` for migrations rather than gorm Automigrate. The migrations are expected to be under `db/migrations` in the root of your microservice.


## Additional middlewares

Middlewares can be injected freely in the available router callback function.

## Environment variables

Environment variables are read from the environment or from a `.env` file in the rood of the directory.

## Plugin

Plugins are a set of handlers and their binded components(validators, middlewares, etc) which can be reused across multiple services using fastecho.

### Usage

```go
fastechoConfig.Use(<pluginConfig>)
```

## TODO

- Which libraries are we using?
- Example on how to setup a service with DB
- Example on how to setup a service without DB
- Example on ENV vars the service uses
- Example on passing extra ENV vars
- Example configuration options
- Example features
- Configure repo for release
- Configure release pipelines
- JFrog automation scanning workflow
- Make sure at least one internal service uses the open-source version and it works
- Test cases & Code coverage badge
- Gitstream to automate dependabot and releases
- Branch protection
