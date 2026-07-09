# Fastecho

Fastecho is a Go library that provides an easily configurable, ready-to-use echo server. It is a wrapper on top of the echo framework and it adds extra functionalities that are often required when setting up web servers.

## Getting started

For specifics, check the detailed features below.
```go
// load env vars
err := envs.SetEnv()
if err != nil {
    log.Fatalf("failed to set environment variables: %s", err)
}

// set up a DB and pass it to handler as you like
// optionally you can provide a gorm config to customize your gorm instance
db, err := fastecho.NewDB(&gorm.Config{
    Logger: newLogger,
})
if err != nil {
    log.Fatalf("failed to connect to the database: %s", err)
}

config := fastecho.Config{
    ExtraEnvs:           envs,
    ValidationRegistrar: validator.RegisterValidations,
    Routes: func(e *echo.Echo, r *router.Router) error {
        return configureRoutes(e, r, db)
    },
    Opts: fastecho.Opts{
        Tracing: fastecho.TracingOpts{
            Skip: !envs["OTEL_ENABLED"].BooleanValue,
        },
        Metrics: fastecho.MetricsOpts{
            Skip: !envs["OTEL_ENABLED"].BooleanValue,
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

## Features

- [Logger](#logger)
- [Request context](#request-context)
- [Routing](#routing)
- [Request validation](#request-validation)
- [Middleware](#middleware)
- [Swagger](#swagger)
- [Health probe endpoints](#health-probe-endpoints)
- [Environment variables](#environment-variables)
- [Observability (OpenTelemetry)](#observability-opentelemetry)
- [Database](#database-optional)
- [Background workers](#background-workers)
- [Plugins](#plugins)
- [Access Echo instance](#access-echo-instance)

### Logger

We integrated `go.uber.org/zap`

### Request context

Fastecho injects a request-scoped logger, tracer, and request ID into the request's `context.Context`. Access them via the `fctx` package:

```go
import "github.com/ingka-group/fastecho/fctx"

func (h *Handler) GetData(ec echo.Context) error {
    ctx := fctx.From(ec)
    log := fctx.Logger(ctx)
    reqID := fctx.RequestID(ctx)

    log.Info("handling request", zap.String("request_id", reqID))
    // ...
}
```

The logger automatically includes `trace_id`, `span_id`, and `request_id` fields when available.

`trace_id` and `request_id` answer different questions: a trace covers a whole
user action, while a request id names exactly one HTTP call within it — a page
load fanning out to five API calls shares one `trace_id`, but each call gets
its own `request_id`, echoed back in the `X-Request-Id` response header so a
failing call can be reported and grepped for. See
[Correlation IDs](docs/observability.md#correlation-ids) for a worked example.

### Routing

The router provides preset endpoints for swagger, monitoring and health checks. Custom endpoints can be injected via the `Routes` function in the config.

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

Custom validation can be registered using the provided validator. Define a function in which you register custom validations and add it to the config.
```go
func RegisterValidations(validator *router.Validator) error {
	validator.Vdt.RegisterStructValidation(daterange.ValidateBasicDateRange(), daterange.BasicDateRange{})

	return nil
}
```

### Middleware

Custom middleware can be injected via the route configuration function.
```go
func configureRoutes(e *echo.Echo, r *router.Router, db *gorm.DB) error {
	v1 := e.Group("/v1")
	myGroup := v1.Group("/data")

	myGroup.Use(middleware.MyCustomMiddleware())

	return nil
}
```

### Swagger

Swagger is baked into the router wrapper. The title and JSON path can be configured via environment variables:

`SWAGGER_UI_TITLE`
`SWAGGER_JSON_PATH`

The swagger documentation is configured on the root path suffixed with `/swagger/`.

### Health probe endpoints

The health endpoints are configured on the root path suffixed with `/health/live` and `/health/ready`.

### Environment variables

Environment variables are read by default from the environment or from a `.env` file in the root of the directory.

Fastecho defines the following env vars internally:

| Variable            | Default                 | Notes                         |
|---------------------|-------------------------|-------------------------------|
| `HOSTNAME`          | `localhost`             | Server hostname               |
| `PORT`              | `8080`                  | Server port                   |
| `SWAGGER_UI_TITLE`  | `FastEcho Service`      | Title for Swagger UI          |
| `SWAGGER_JSON_PATH` | `/swagger/swagger.json` | Path to swagger.json          |
| `LOG_LEVEL`         | `dev`                   | One of: `dev`, `test`, `prod` |

You can define additional env vars via `ExtraEnvs` in the config. These are merged with the defaults and loaded automatically when `Run()` or `Initialize()` is called:
```go
var envs = env.Map{
	"OTEL_ENABLED": {
		DefaultValue: "false",
		IsBoolean:    true,
	},
}

// load them
err := envs.SetEnv()
if err != nil {
    log.Fatalf("failed to set environment variables: %s", err)
}

config := fastecho.Config{
	ExtraEnvs: envs,
	// ...
}
```

### Observability (OpenTelemetry)

fastecho emits **traces + metrics** through OpenTelemetry from one shared
resource and one shutdown. Logs stay on zap (stdout JSON) and carry
`trace_id`/`span_id`/`request_id` so they correlate without a log exporter.

New here? The [observability guide](docs/observability.md) explains how traces,
metrics, and logs tie together (correlation IDs, pull vs push, exemplars) and
what happens under the hood.

Behaviour is configured by standard `OTEL_*` env vars (read by the OTel SDK and
`autoexport`); `fastecho.Opts` only carries on/off toggles.

#### Env vars

| Variable                       | Default                   | Effect                                                         |
|--------------------------------|---------------------------|----------------------------------------------------------------|
| `OTEL_SERVICE_NAME`            | (unset)                   | `service.name` on every span/metric; fastecho warns if unset while a signal is on |
| `OTEL_RESOURCE_ATTRIBUTES`     | (unset)                   | Extra resource attrs, e.g. `service.version=1.2.3,deployment.environment=prod` |
| `OTEL_TRACES_EXPORTER`         | `otlp`                    | `otlp` \| `console` \| `none`                                  |
| `OTEL_METRICS_EXPORTER`        | `prometheus`              | `prometheus` (served at `/metrics`) \| `otlp` (push) \| `none` |
| `OTEL_EXPORTER_OTLP_ENDPOINT`  | `localhost:4317`          | Collector endpoint. TLS by default; use an `http://` URL for a plaintext collector |
| `OTEL_EXPORTER_OTLP_PROTOCOL`  | `grpc` (fastecho default) | Set `http/protobuf` for a `:4318` collector                    |
| `OTEL_TRACES_SAMPLER`          | `parentbased_always_on`   | e.g. `parentbased_traceidratio`                                |
| `OTEL_TRACES_SAMPLER_ARG`      | `1.0`                     | Sample ratio for the ratio samplers                            |
| `OTEL_METRICS_EXEMPLAR_FILTER` | `trace_based`             | Which measurements carry trace exemplars                       |

> `/metrics` stays on the service's main port (default `prometheus` exporter)
> and keeps serving everything on `prometheus.DefaultGatherer` — your
> `prometheus.MustRegister` metrics and the `go_*`/`process_*` collectors —
> alongside the OTel output.
> The metric is `http_server_request_duration_seconds_*` (was `echo_http_*`).

#### Code toggles (`fastecho.Opts`)

| Field               | Effect                            |
|---------------------|-----------------------------------|
| `Opts.Tracing.Skip` | Disable tracing                   |
| `Opts.Metrics.Skip` | Disable metrics (and `/metrics`)  |
| `Opts.Logs.Skip`    | Disable the access-log middleware |

There are intentionally **no** code fields mirroring an env var (sampler ratio,
exporter, endpoint): one source of truth per setting.

#### Libraries

| Concern                     | Package                                                                                                        |
|-----------------------------|----------------------------------------------------------------------------------------------------------------|
| SDK + API                   | [`go.opentelemetry.io/otel`](https://pkg.go.dev/go.opentelemetry.io/otel)                                      |
| HTTP server spans + metrics | [`otelecho`](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho) |
| Env-driven exporters        | [`autoexport`](https://pkg.go.dev/go.opentelemetry.io/contrib/exporters/autoexport)                            |
| Go runtime metrics          | [`instrumentation/runtime`](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/runtime)            |
| Outbound client transport   | [`otelhttp`](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp)                 |
| `/metrics` exporter         | [`exporters/prometheus`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/prometheus)                     |
| Semantic conventions        | [`semconv/v1.41.0`](https://pkg.go.dev/go.opentelemetry.io/otel/semconv/v1.41.0)                               |

Full env reference: [OpenTelemetry SDK environment variables](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/).

#### Instrumentation coverage

| Layer                 | How                                   | Effort              |
|-----------------------|---------------------------------------|---------------------|
| HTTP requests         | Automatic via middleware              | Zero config         |
| Database (GORM)       | `gorm.io/plugin/opentelemetry`        | Add plugin yourself |
| Service functions     | `telemetry.StartSpan(ctx)`            | 2 lines             |
| Functions without ctx | `telemetry.SpanFunc(ctx, name, fn)`   | 3 lines             |
| Outbound HTTP         | `telemetry.WrapClient(c)`             | Wrap your client    |

Per-function tracing:

```go
import "github.com/ingka-group/fastecho/telemetry"

func (s *Service) Process(ctx context.Context, input Input) error {
    ctx, span := telemetry.StartSpan(ctx)
    defer span.End()
    // span name auto-discovered: "mypackage.Service.Process"
    return s.repo.Save(ctx, input)
}
```

Tracing a function without context:

```go
var result T
telemetry.SpanFunc(ctx, "heavy-algorithm", func() {
    result = computeHeavyAlgorithm(data)
})
```

Outbound HTTP calls:

```go
client := telemetry.WrapClient(&http.Client{Timeout: 5 * time.Second})
// client injects traceparent + X-Request-Id on every outbound request
resp, err := client.Do(req.WithContext(ctx))
```

For database tracing, add `gorm.io/plugin/opentelemetry` to your project:

```go
import "gorm.io/plugin/opentelemetry/tracing"

_ = db.Use(tracing.NewPlugin())
```

**Important:** Always pass context to GORM queries (`db.WithContext(ctx).Find(...)`) so DB spans attach to the request trace.

#### Upgrading

Upgrading from an earlier fastecho version? See [CHANGELOG.md](CHANGELOG.md) for
the full list of breaking changes and migration steps.

### Database (optional)

Fastecho has an optional postgres DB connection baked into it using `gorm`. We are using `goose` for migrations rather than gorm Automigrate. The migrations are expected to be under `db/migrations` in the root of your folder.

## Background workers

Services often need long-running background processes (e.g. a Pub/Sub listener) that share the server's lifecycle. Register them via `Workers` in the config and fastecho manages their full lifecycle: each worker runs in its own goroutine, receives a context that is cancelled on shutdown, and is waited for while the server drains. The drain shares the 10s graceful-shutdown budget with the HTTP server, so a worker is not guaranteed a full 10s to finish if HTTP draining is slow.

A `Worker` is a plain function that should block until its context is cancelled:

```go
config := fastecho.Config{
	Routes: configureRoutes,
	Workers: map[string]fastecho.Worker{
		"pubsub-listener": func(ctx context.Context) error {
			return pubsub.StartListener(ctx, "project", "subscription")
		},
	},
}
```

Workers are supervised: since a worker should block until its context is cancelled, any early return (error, `nil`, or a recovered panic) is treated as a fault and the worker is restarted with exponential backoff (1s, doubling, capped at 30s; reset after 1m stable). Restarts are logged at warning level; once a worker has restarted 10 times without a stable run the log escalates to error, so a persistent crash loop can be distinguished from an occasional restart by anything consuming the logs. Only a context-cancelled return is a clean stop — no restart, no error log. Other workers and the HTTP server are unaffected.

## Plugins

Plugins are a set of handlers and their bound components (validators, middlewares, etc) which can be reused across multiple services using fastecho.

```go
fastechoConfig.Use(<pluginConfig>)
```

## Access Echo instance

The underlying Echo instance can be accessed by passing the value for `EchoFn` in the config. This is useful for binding service level middlewares, enabling echo's debug mode or any other Echo features.

```go
config := fastecho.Config{
	ValidationRegistrar: func(*router.Validator) error {
		return nil
	},
	Routes: func(e *echo.Echo, r *router.Router) error {
		return configureRoutes(e, r, db)
	},
	EchoFn: func(e *echo.Echo) error {
		// Use a service level middleware
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"https://example.com"},
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
			MaxAge:           3600,
			AllowCredentials: true,
			ExposeHeaders:    []string{echo.HeaderContentLength},
		}))
		// Enable debug mode in echo
		e.Debug = true
		return nil
	},
	...
}
```

## Miscellaneous

Fastecho is fully compatible with [Echoprobe](https://github.com/ingka-group/echoprobe)

## Contributing

Please read [CONTRIBUTING](./CONTRIBUTING.md) for more details about making a contribution to this open source project and ensure that you follow our [CODE_OF_CONDUCT](./CODE_OF_CONDUCT.md).

## Contact

If you have any other issues or questions regarding this project, feel free to contact one of the [code owners/maintainers](.github/CODEOWNERS) for a more in-depth discussion.

## Licence

This open source project is licensed under the "Apache-2.0", read the [LICENCE](./LICENCE.md) terms for more details.
