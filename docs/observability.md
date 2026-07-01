# Observability in fastecho

A guide to how fastecho's telemetry works: the three signals it emits, how they tie together, what happens under the hood (OpenTelemetry, Prometheus, Grafana), and the design choices behind it.

This is the **conceptual** companion to the README's Observability section — the README is the knob-by-knob reference (every `OTEL_*` env var and `Opts` toggle); this doc explains *what those knobs do and why it all fits together*. For the recommended way to *use* it day-to-day, see [do's and don'ts](observability-dos-and-donts.md).

---

## Overview

fastecho emits **three signals**, tied together by a single correlation ID.

```
                         ┌─────────────────────────────┐
   incoming request ───▶ │  fastecho service           │
                         │                             │
                         │   traces  ─── push ───▶  collector ─▶ Tempo/Jaeger
                         │   metrics ─── pull ◀───  Prometheus
                         │   logs    ─── stdout ─▶  your log pipeline
                         └─────────────────────────────┘
                                       │
                          every signal carries trace_id
```

- **Traces** — the path of a request through your service (and across service hops), as a tree of timed spans.
- **Metrics** — aggregate numbers: request rates, latencies, Go runtime stats.
- **Logs** — structured zap lines on stdout, now carrying correlation IDs.

**`trace_id`** appears in logs, on spans, and (via *exemplars*) on metrics, so you can pivot from any one signal to any other.

One bootstrap (`telemetry.Init`) builds one shared *resource* (the identity: `service.name`, `service.instance.id`), wires the exporters from environment variables, and returns one `Shutdown` that drains everything together. Both `Run` and `Initialize` go through it, so every entry point is instrumented the same way.

---

## Glossary

### SDK & pipeline

- **Provider** — the SDK object that produces tracers/meters; fastecho holds one of each behind `telemetry.Providers`.
- **Exporter** — the SDK component that sends a signal to a destination (OTLP, Prometheus, console, none); chosen per signal via `OTEL_*_EXPORTER`.
- **OTLP** — OpenTelemetry Protocol; the wire format for exporting telemetry (gRPC on `:4317`, HTTP/protobuf on `:4318`).
- **Collector** — a standalone OpenTelemetry process that receives, processes, and forwards telemetry; sits between the app and the backend.
- **Pull / push** — whether the backend scrapes the app (`/metrics`) or the app exports to the backend (OTLP).
- **semconv** — OpenTelemetry's standard attribute/metric names.

### Traces

- **Span** — one timed operation; the unit of a trace.
- **Trace** — a tree of spans for one request, possibly spanning services.
- **Tracer** — the per-scope SDK object that creates spans; obtained from the tracer provider.
- **`traceparent`** — the W3C header that carries trace context between services.
- **Sampler** — decides which traces are recorded vs dropped. Configured by `OTEL_TRACES_SAMPLER` / `OTEL_TRACES_SAMPLER_ARG`; defaults to `parentbased_always_on` (keep all, honoring the caller's decision).

### Metrics

- **Meter** — the per-scope SDK object that creates metric instruments; obtained from the meter provider.
- **Instrument** — the handle you record a metric on: Counter (monotonic sum), Histogram (distribution, e.g. latency), Gauge (current value).
- **Exemplar** — an example measurement on a metric bucket, carrying a `trace_id` that links the aggregate to a specific trace.
- **Cardinality** — the number of distinct time series a metric produces; one per distinct attribute combination, so bounded attribute values keep it in check.

---

## Design decisions

The key choices behind fastecho's telemetry:

- **One correlation-ID source.** `trace_id`/`span_id`/`request_id` are assembled in a single place — `fctx.Fields(ctx)` — and reused by the request middleware, the panic recover handler, the access log, and worker seeds. One emission point means every signal correlates identically, with no per-call-site drift.
- **Env is the single source of truth.** Sampling ratio, exporter, and endpoint live only in `OTEL_*` env, never duplicated as struct fields — so the same binary is configured per deployment and there's no second place to disagree.
- **Global registration is internal, not a public toggle.** Registering the providers + propagator as the OTel globals is always on for `Run`/`Initialize` and isn't exposed on `fastecho.Config`. There's no legitimate production reason to turn it off, and doing so would silently break `traceparent` propagation (otelecho/otelhttp read the global propagator). Tests get isolation via injected providers instead.
- **Don't hand-roll what the SDK does.** No custom sampler (the SDK reads the env and handles every standard value); no custom server-span middleware (`otelecho`). Less code, fewer ways to be subtly wrong.
- **Span helpers read the tracer from context.** `telemetry.StartSpan`/`SpanFunc` resolve `fctx.Tracer(ctx)`, not the OTel global. This keeps them consistent with the rest of `fctx` (logger/tracer/request-id all flow through the context), makes `fctx.WithTracer` meaningful, and lets tests and multiple instances use isolated providers without mutating global state. In production the middleware seeds the context with the global's tracer, so behavior is identical; with none set, it falls back to a no-op.
- **Metrics default to pull.** fastecho builds its own Prometheus exporter and registry so `/metrics` stays on the service's main port — plain `autoexport` would otherwise start a separate server on `:9464` and leave `/metrics` empty. `OTEL_METRICS_EXPORTER=otlp` switches to push; [Pull vs push](#pull-vs-push) covers why pull is the default.
- **OTLP transport defaults to gRPC** (`:4317`), preserving fastecho's previous behavior so existing deployments upgrade untouched. It's no longer *hardcoded* — set `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf` to use the OTel default (`:4318`). fastecho sets this default only when the var is unset, so operators keep full control. (Same reasoning as the metrics-exporter default: override an SDK default to preserve existing behavior, but stay overridable.)
- **Outbound client wraps, doesn't replace.** `telemetry.WrapClient`/ `WrapTransport` decorate a normal `*http.Client`/`RoundTripper` with `traceparent` injection and `X-Request-Id` forwarding — no factory, options, or typed verbs. The caller keeps full control of timeout, redirects, and headers; fastecho only adds the propagation that's specific to it.
- **Logs stay on zap.** Correlation comes from structured fields, not a log exporter — your platform already ships stdout.

---

## Correlation IDs

Three IDs travel with every request:

| ID           | What it is                                                | Where it comes from                                                    |
|--------------|-----------------------------------------------------------|------------------------------------------------------------------------|
| `trace_id`   | Identifies the whole request, end to end, across services | The active span (created by `otelecho`)                                |
| `span_id`    | Identifies one operation within the trace                 | The active span                                                        |
| `request_id` | A fastecho-level request identifier (UUIDv4)              | The request-id middleware, or forwarded from an inbound `X-Request-Id` |

They are built in **one place** — `fctx.Fields(ctx)` — so middleware, the panic recover handler, the access log, and background workers all derive them identically. `Fields` reads `trace_id`/`span_id` from the active span context and `request_id` from the request context, and returns them as structured log fields.

Discrete fields (not the raw `traceparent` header) let the log backend **query by `trace_id`** and link a line to its trace.

On the trace side, the request id also rides along as a span attribute (`fastecho.request_id`), so you can find a trace from a request id and vice versa.

---

## Traces

A **trace** is a tree of **spans**. Each span is a timed operation with a name, start/end, attributes, and a parent. The root span for an HTTP request is created by **`otelecho`**, the upstream Echo instrumentation — fastecho no longer hand-rolls this.

**Server spans.** `otelecho` starts a span per request and records it on **stable semantic conventions** (semconv): `http.request.method`, `http.route`, `http.response.status_code`, `url.path`, etc. These names are an ecosystem standard, so any OTel-aware backend or dashboard understands them out of the box.

**Spans in your own code.** Inside a handler or service, use the helpers:

```go
ctx, span := telemetry.StartSpan(ctx) // auto-named after the calling function
defer span.End()
```

`StartSpan` resolves the tracer from `ctx` (`fctx.Tracer(ctx)`), so it uses whichever provider the request is running under — and falls back to a no-op tracer (zero spans, no panic) if tracing is off.

**Propagation.** Trace context crosses service hops via the W3C `traceparent` propagator, registered globally:

- **Inbound:** `otelecho` extracts `traceparent` and continues the upstream trace.
- **Outbound:** `telemetry.WrapClient` injects `traceparent` (and forwards `X-Request-Id`) onto every request — see [Outbound calls](#outbound-calls).

**Sampling.** How many traces are kept is **env-driven**, read directly by the OTel SDK from `OTEL_TRACES_SAMPLER` / `OTEL_TRACES_SAMPLER_ARG`. fastecho does *not* hand-roll a sampler — the SDK already handles every standard value (`always_on`, `always_off`, ratio-based, and the `parentbased_*` variants) and defaults to `parentbased_always_on` (keep everything, respecting the caller's decision). "ParentBased" means: if an incoming request was already sampled upstream, honor that; only the root of a new trace consults the ratio.

---

## Metrics

A **metric** is an aggregate number over time. fastecho produces four families:

1. **HTTP server metrics** (from `otelecho`) — request duration histograms and counts, on stable semconv names like `http_server_request_duration_seconds_*` (labelled by route pattern — `http_route` — so the series stay bounded).
2. **Go runtime metrics** (`go.*`) — GC, goroutines, memory, from the OTel runtime instrumentation.
3. **Worker failures** — `fastecho.worker.failures{worker,kind}` (`kind` = `panic`|`error`), a framework counter for background-worker failures. Workers aren't on the auto-traced HTTP path, so this is what makes them *alertable*: `rate(fastecho_worker_failures_total{worker="…"}[5m]) > 0`. Normal/cancelled shutdowns don't count.
4. **Whatever you record** — `telemetry.Init` registers the meter provider as the OTel **global**, so record custom metrics with plain OTel via `otel.Meter(...)` (see the [OpenTelemetry Go metrics docs](https://opentelemetry.io/docs/languages/go/instrumentation/#metrics)). Pass the request `ctx` to `Add`/`Record` so measurements pick up exemplars, and keep attribute values low-cardinality (high-cardinality detail belongs on the span, not the metric).

### The `/metrics` endpoint

By default fastecho serves metrics at **`/metrics` on the service's main port**, in Prometheus format, for a Prometheus server to scrape. (More on that delivery model in [Pull vs push](#pull-vs-push).) The metric names changed from the old `echo_http_*` to the semconv `http_server_request_duration_seconds_*` — see the README's Breaking changes section.

---

## Logs

Logs **stay on zap** — structured JSON on stdout. There is intentionally **no OTLP log bridge**: shipping logs is your platform's job (stdout → your log pipeline), and `fctx.Logger(ctx)` remains a plain `*zap.Logger`.

Logs now **carry the correlation fields** (`trace_id`/`span_id`/`request_id`) via `fctx.Fields`, so a log line links to its trace without any log exporter. The access log, the panic recover handler, and worker logs all use the same enrichment.

The access log records every request: 5xx at Error, 4xx at Warn, 3xx at Info, and 2xx at Debug (so successful traffic is queryable when Debug is enabled but stays out of the way at higher levels). Lines carry a numeric `latency_ms`, the matched route template as `path`, and keep the error string on 5xx lines, so a failing request leaves both a log entry *and* an error on its span.

---

## Pull vs push

**Delivery model (pull/push) and wire transport (gRPC/HTTP) are two different axes.**

**Metrics** can be delivered either way:

|                  | **Pull** (default)                   | **Push** (opt-in)                                |
|------------------|--------------------------------------|--------------------------------------------------|
| How              | Prometheus scrapes `/metrics`        | App pushes via OTLP to a collector               |
| App does         | Just expose current state            | Run an export loop, buffer, retry                |
| Liveness         | Free — a failed scrape = target down | Ambiguous — no data could mean crashed *or* idle |
| Short-lived jobs | Weak (may exit before scrape)        | Strong (push before exit)                        |
| Set via          | `OTEL_METRICS_EXPORTER=prometheus`   | `OTEL_METRICS_EXPORTER=otlp`                     |

fastecho **defaults to pull**: its services are long-lived HTTP servers with stable network identity (ideal scrape targets), pull gives a free up/down signal, and the framework only has to expose a registry. Push suits short-lived jobs or a single unified OTLP pipeline.

**Traces are always push** — there's no pull model for spans:

|         | Readable current *state*?   | Volume         | Model      |
|---------|-----------------------------|----------------|------------|
| Metrics | Yes (a counter/gauge value) | Small, bounded | Pull works |
| Traces  | No (completed *events*)     | High, bursty   | Push only  |

A metric is a value that exists *now* and can be read on demand. A span is a past *event* — by the time you'd pull it, it's already over and would have to be buffered waiting for you. Spans are also a firehose. So every tracing protocol (OTLP, Jaeger, Zipkin) is push; there is no `/metrics`-equivalent for spans.

> In the **default** config, fastecho runs two delivery models at once: traces **push** (OTLP), metrics **pull** (scrape). Setting `OTEL_METRICS_EXPORTER=otlp` unifies them onto one OTLP push pipeline — simpler to operate (one endpoint, one collector) at the cost of the free scrape-based liveness signal.

---

## Exemplars: jumping from a metric to a trace

Exemplars link an aggregated metric back to an individual trace.

Metrics are aggregates, and aggregation throws away individual identity. A histogram bucket says:

```
http_server_request_duration_seconds_bucket{le="0.25"} 47
```

"47 requests finished under 250ms" — but not *which* 47. When p99 latency spikes, the metric shows something is slow, not which request, leaving you to search traces by time range.

**An exemplar** is a concrete example measurement attached to a metric bucket, carrying the `trace_id` of a request that landed there:

```
http_server_request_duration_seconds_bucket{le="0.25"} 47 # {trace_id="4bf92f...",span_id="00f0..."} 0.237 1700000000
```

That trailing part is one specific request that took 0.237s, and the exact trace you can open. In Grafana these render as **clickable dots** on the latency graph: click the dot on the p99 spike to jump to *that* trace's waterfall in Tempo/Jaeger.

**Why it's tied to sampling** (`OTEL_METRICS_EXEMPLAR_FILTER=trace_based`, the default): an exemplar is only recorded when the measurement happened inside a **sampled** span context. Its entire value is the `trace_id` it points at — if that trace wasn't kept, the link would dangle. `trace_based` means "only attach exemplars whose trace actually got stored."

**Visibility is a backend concern.** fastecho records exemplars at the SDK. To see and click them, the backend must be wired up: Prometheus with exemplar storage enabled, and a trace backend (Tempo) for Grafana to jump into.

---

## Backends and dashboards

```
  fastecho service
    │
    │  traces  ──OTLP push──▶  OpenTelemetry Collector ──▶  Tempo / Jaeger
    │
    │  /metrics  ◀──scrape──   Prometheus
    │
    └─ logs ──stdout──▶  (your log pipeline)

                      Grafana
                        ├─ Prometheus data source  → dashboards, alerts
                        ├─ Tempo data source       → trace waterfalls
                        └─ exemplars link Prometheus panels → Tempo traces
```

- **Prometheus** scrapes `/metrics` and stores the time series; you build dashboards and alerts on `http_server_request_duration_seconds_*`, `go.*`, etc.
- **The Collector** receives pushed traces (and optionally pushed metrics if you flip to OTLP) and forwards them to **Tempo** (or Jaeger).
- **Grafana** reads both, and — with exemplar storage + a Tempo data source — lets you click an exemplar on a metrics panel straight through to the trace.

The correlation IDs make the manual pivots work too: a `trace_id` in a log line pastes into Tempo; a `fastecho.request_id` span attribute matches the `request_id` in your logs.

---

## Deploying behind an OTel Collector

In production you almost never point fastecho straight at a backend. The recommended topology is **app SDK → a local OpenTelemetry Collector → backend**. The collector owns batching, retry, and back-pressure (the SDK deliberately doesn't), and it's where org-wide resource attributes get injected. **The app→collector hop and the collector→backend hop are independent** — the collector can receive plain HTTP from your app and forward gRPC+TLS to the backend, so the backend wanting gRPC never forces gRPC on your app.

The three signals reach the collector by **three different paths**:

```
  fastecho pod
    │  traces   ──OTLP push (HTTP :4318 or gRPC :4317)──▶  collector ─▶ backend
    │  /metrics ◀──scrape (prometheus.io/scrape)──────────  collector ─▶ backend
    │  stdout   ──tailed by a filelog receiver───────────▶  collector ─▶ backend
```

**Traces — push OTLP to the collector.** Point `OTEL_EXPORTER_OTLP_ENDPOINT` at the collector. fastecho defaults the protocol to gRPC (`:4317`); for an HTTP/protobuf receiver (`:4318`) set `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`.

**Metrics — pick *one* path, not both** (see [Pull vs push](#pull-vs-push) for the trade-off):

- **Scrape (default):** keep `OTEL_METRICS_EXPORTER=prometheus` and let a collector's `prometheus` receiver scrape `/metrics`. Annotate the pod (`prometheus.io/scrape: "true"`, `prometheus.io/port`, `prometheus.io/path: "/metrics"`). Zero exporter config; the metric *names* are the semconv ones.
- **Push:** set `OTEL_METRICS_EXPORTER=otlp` and metrics push over the same OTLP endpoint. This carries **exemplars natively** (the cleanest path for metric→trace links). If you switch to push, **remove the scrape annotation** so the collector doesn't also scrape and double-count.

**Logs — nothing to configure.** A node/DaemonSet collector with a `filelog` receiver tails every pod's stdout and forwards it. fastecho writes JSON to stdout, so it's picked up automatically — no OTLP log exporter. Two app-side details make the logs *useful* once they land:

- **Lowercase `level`.** fastecho emits `"level":"info"` so the collector can map it to a log-level label and the UI can color-code by severity. (Uppercase `INFO` is often not recognised by that mapping.)
- **`trace_id` / `span_id` fields.** A collector's trace parser reads exactly these JSON keys to attach trace context to each log line — which is what makes a log line link to its trace in the UI. fastecho emits them via `fctx.Fields`.

**Mandatory resource attributes.** Many platforms require certain resource attributes (team, system, environment) and *silently drop* data without them. Inject them at the collector (a `resource` processor) or set them on the app via `OTEL_RESOURCE_ATTRIBUTES` / `OTEL_SERVICE_NAME` — wherever your org standard lives. The collector path is convenient because it applies them uniformly.

> **With the Grafana stack**, the three signals land in **Mimir** (metrics), **Loki** (logs), and **Tempo** (traces), fed by the Collector above: metrics scraped from `/metrics` (keep the default exporter + a `prometheus.io/scrape` annotation), logs tailed from stdout into Loki, and traces pushed to the Collector (gRPC `:4317` by default).

---

## Outbound calls

When your service calls another service, the trace should continue across the hop. Wrap your HTTP client:

```go
client := telemetry.WrapClient(&http.Client{Timeout: 5 * time.Second})
```

`WrapClient` augments the client's transport so every request **injects `traceparent`** (via `otelhttp`, continuing the current trace) and **forwards `fctx.RequestID(ctx)` as `X-Request-Id`**. You keep a completely normal `*http.Client` — your timeout, redirect policy, and headers are yours; fastecho only decorates the transport. (`WrapTransport(base)` does the same for a bare `RoundTripper`, e.g. a generated client.)

The result: a downstream service running this same stack continues *your* trace and logs *your* request id — one unbroken story across services.

---

## Under the hood

- **OpenTelemetry SDK** is the engine. fastecho's `telemetry` package is a thin bootstrap over it.
- **`telemetry.Init`** builds one **resource** (shared identity for every signal), the tracer and meter providers, env-driven exporters, and a single `Shutdown` closer that drains traces and metrics together. It always returns non-nil providers — when a signal is disabled it installs a **no-op** provider, so nothing downstream has to nil-check.
- **Exporters are chosen by environment**, via the OTel `autoexport` helper (`OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`). The same binary is configured per deployment; fastecho adds only two `OTEL_*` defaults beyond the SDK's own, both to preserve prior behavior and both overridable: metrics to `prometheus` (so `/metrics` keeps serving on the main port) and OTLP transport to `grpc`.
- **Service identity** comes from `OTEL_SERVICE_NAME` (+ `OTEL_RESOURCE_ATTRIBUTES`); a `service.instance.id` is generated if you don't set one.
- **Startup banner.** On boot, fastecho prints plain key/value sections (`Database configuration`, `Telemetry configuration`, `Log configuration`) showing the resolved exporters, OTLP protocol/endpoint, and metrics delivery model. It's deliberately plain text (no ANSI) so it reads cleanly in a dev console and stays free of escape codes in a log aggregator, making an upgrade or misconfiguration visible at a glance.

### Configuration

The control surface is two layers:

- **`OTEL_*` environment variables** — *behaviour* values (which exporter, which endpoint, sampling ratio), read by the OTel SDK. One source of truth per setting; nothing is mirrored as a Go struct field.
- **`fastecho.Opts`** — on/off *toggles* only: `Opts.Tracing.Skip`, `Opts.Metrics.Skip`, `Opts.Logs.Skip`.

See the README's Observability section for the full env-var and toggle tables.
