# Observability: do's and don'ts

The recommended way to use fastecho's telemetry, in brief. For *why* any of these
hold, see [observability.md](observability.md).

## Spans

- ✅ Use `telemetry.StartSpan(ctx)` for your own spans, then `defer span.End()`.
  It auto-names from the caller and resolves the tracer from `ctx` — so it works
  in handlers *and* workers.
- ❌ Don't reach for `otel.Tracer(...)` inside a worker (or any seeded context) —
  it bypasses the ctx-scoped tracer and loses the `worker=<name>` label.
- ✅ Span the **unit of work** (one poll, one job, one iteration).
- ❌ Don't wrap a long-running loop or a whole worker run in a single span — it
  stays open for the process lifetime and never exports.
- ✅ Record failures on the span: `span.RecordError(err)` + `span.SetStatus(codes.Error, …)`.

## Goroutines

- ✅ Detach cancellation but keep the trace when you spawn work that outlives the
  request:
  ```go
  gctx := context.WithoutCancel(ctx)
  go func() { ctx, span := telemetry.StartSpan(gctx); defer span.End(); /* … */ }()
  ```
- ❌ Don't hand a request's `ctx` straight to a goroutine that outlives the
  handler — it's cancelled the moment the handler returns.

## Logging

- ✅ Use `fctx.Logger(ctx)` — it carries `trace_id`/`span_id`/`request_id`, so
  every line correlates to its trace.
- ❌ Don't use a package-level or global logger in request/worker code — you lose
  correlation.
- ✅ Set log levels with recognizable severities; the level field is lowercase
  (`info`/`warn`/`error`) so backends color-code it.
- ❌ Don't hand-log successful requests — the access log already covers 3xx+, and
  2xx is intentionally quiet.

## Metrics

- ✅ Create instruments **once** (service/package scope) via `otel.Meter("your-scope")`.
- ❌ Don't create counters/histograms per request.
- ✅ Pass the request `ctx` to `Add`/`Record` so measurements pick up exemplars.
- ❌ Don't put high-cardinality values (IDs, emails, raw paths) on metric
  attributes — that belongs on the span. Keep metric labels bounded.

## Outbound HTTP

- ✅ Wrap service-to-service clients with `telemetry.WrapClient(&http.Client{…})`
  so `traceparent` + `X-Request-Id` propagate.
- ❌ Don't make outbound calls with a bare `http.Client` — the trace stops at your
  service boundary.
- ✅ Build requests with `http.NewRequestWithContext(ctx, …)` so the context flows.

## Context

- ✅ Accept `context.Context` in service/handler functions and pass it down.
- ✅ Cross from Echo to `context.Context` once at the boundary: `ctx := fctx.From(ec)`.
- ❌ Don't thread `echo.Context` into business logic.
- ❌ Don't stash loggers/tracers/request-ids on structs — they live on the context.

## Configuration

- ✅ Configure behavior with `OTEL_*` env vars (exporter, endpoint, sampler) — one
  source of truth, set per deployment.
- ✅ Disable a signal with the code toggles: `Opts.Tracing.Skip` /
  `Opts.Metrics.Skip` / `Opts.Logs.Skip`.
- ❌ Don't look for `Config` fields mirroring an env var (sampler ratio, exporter,
  protocol) — by design they don't exist.
- ✅ On an OTel Collector that only speaks gRPC, keep the default; for an
  HTTP/protobuf collector set `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`.
