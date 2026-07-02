# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
follows [Semantic Versioning](https://semver.org/spec/v2.0.0/) — while on `0.x`,
minor releases may include breaking changes.

## v0.16.0

Observability moves fully onto OpenTelemetry: a single traces + metrics
bootstrap, env-driven exporters and sampling, stable semantic conventions via
`otelecho`, Go runtime metrics + exemplars, worker observability, and outbound
trace propagation.

### Added

- `telemetry.Init` — one OpenTelemetry bootstrap (shared resource, env-driven
  exporters/sampling, one shutdown) replacing the old `tracer.go`.
- `fctx.Fields` / `fctx.NewRequestID` / `fctx.WithNewRequestID` — a single
  correlation-ID emission point (`trace_id`/`span_id`/`request_id`).
- Go runtime metrics (`go.*`) and exemplars (`trace_based`).
- Worker observability: per-run seeded context tagged `worker=<name>` and a
  `fastecho.worker.failures{worker,kind}` counter.
- `telemetry.WrapClient` / `telemetry.WrapTransport` — outbound propagation of
  `traceparent` (via `otelhttp`) and `X-Request-Id`.
- `Opts.Logs.Skip` — disable the access-log middleware.
- Access-log lines now carry `trace_id`/`span_id`/`request_id`.
- Access log now records 2xx requests at `Debug`, replaces the `latency`
  duration string with a numeric `latency_ms` field, and adds `path` (the
  matched route template) as a low-cardinality grouping key.

### Breaking changes

- **Package renamed:** update import `github.com/ingka-group/fastecho/otel` →
  `…/telemetry` and call sites `otel.StartSpan` / `otel.SpanFunc` →
  `telemetry.StartSpan` / `telemetry.SpanFunc`.
- **Hand-rolled trace middleware removed:** drop any direct registration of
  `otel.Middleware`, `otel.WithSkipper`, or `otel.WithTracerProvider` — fastecho
  wires `otelecho` itself.
- **Metric renames:**

  | Old name                                    | New name                                                         |
  |---------------------------------------------|------------------------------------------------------------------|
  | `echo_http_requests_total`                  | derive from `http_server_request_duration_seconds_count`         |
  | `echo_http_request_duration_seconds_bucket` | `http_server_request_duration_seconds_bucket`                    |
  | label `code`                                | `http_response_status_code`                                      |
  | label `url`                                 | `http_route` (rename only — value was already the route pattern) |

- **Span attribute renames:** `http.*` / `net.*` → stable semconv
  (`http.request.method`, `url.path`, `http.route`, `http.response.status_code`).
  `fastecho.request_id` is also set on the server span.
- **`/metrics` survives — no action needed:** fastecho defaults
  `OTEL_METRICS_EXPORTER` to `prometheus`, so an existing app that sets nothing
  still gets `/metrics` on the main port. The endpoint serves the OTel registry
  and `prometheus.DefaultGatherer` together, so metrics registered via
  `prometheus.MustRegister` and the default `go_*`/`process_*` collector
  families stay served. Only the `echo_http_*` names change (above).
- **OTLP transport stays gRPC** (`:4317`) by default — existing collectors keep
  working. Set `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf` to opt into the OTel
  default (`:4318`).
- **Sampling now honors upstream decisions:** the hardcoded always-on sampler
  is replaced by the SDK default `parentbased_always_on`. Requests without a
  parent are still always sampled, but an inbound `traceparent` whose sampled
  flag is off now records nothing. Set `OTEL_TRACES_SAMPLER=always_on` to
  restore the old keep-everything behavior.
- **Response `Traceparent` header removed:** the old middleware injected the
  trace context into every response; `otelecho` follows the W3C spec, where
  `traceparent` is a request header. Correlate via the `X-Request-Id` response
  header instead.
- **`stringutils` package moved to `internal/`** — it was an implementation
  detail of fastecho, not a public API; copy the helpers if you imported them.
- **Access-log `latency` field removed** — the duration string is replaced by
  the numeric `latency_ms` (milliseconds).
- **Requests to fastecho's own endpoints are no longer measured:** paths
  containing `/health`, `/metrics`, or `/swagger/` are skipped by the trace,
  metrics, and access-log middlewares (the old `echo_http_*` middleware counted
  them). Dashboards summing total request rate no longer include probe traffic
  — and note the substring match also excludes user routes containing those
  segments (e.g. `/api/healthcheck`).
- **`prometheus.DefaultRegisterer` reset removed** — fastecho no longer mutates
  global Prometheus state.
- **Log level field is now lowercase** (`"level":"info"`, was `"INFO"`) — update
  any log filter or alert that matched uppercase `INFO`/`WARN`/`ERROR`.
- **`Config.Workers` is now `map[string]Worker`** (was `[]Worker`): the map key
  is the worker name. Migrate `Workers: []fastecho.Worker{fn}` →
  `Workers: map[string]fastecho.Worker{"my-worker": fn}`.
