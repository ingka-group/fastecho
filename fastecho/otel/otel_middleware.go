package otel

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	tracerKey = "ocp-go-utils-tracer"
	// ScopeName is the instrumentation scope name.
	ScopeName = "github.com/ingka-group-digital/ocp-go-utils/"
)

// Middleware returns echo middleware which will trace incoming requests.
func Middleware(options ...Option) echo.MiddlewareFunc {

	config := TracerConfig{}
	for _, opt := range options {
		opt.apply(&config)
	}

	setDefaultConfig(&config)

	tracer := config.TracerProvider.Tracer(
		ScopeName,
	)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			return handleRequest(c, tracer, &config, next)
		}
	}
}

func handleRequest(c echo.Context, tracer oteltrace.Tracer, config *TracerConfig, next echo.HandlerFunc) error {
	c.Set(tracerKey, tracer)
	request := c.Request()
	savedCtx := request.Context()
	defer func() {
		request = request.WithContext(savedCtx)
		c.SetRequest(request)
	}()

	// Extract the Traceparent header, if present, to support distributed tracing
	// https://www.w3.org/TR/trace-context/#traceparent-header-field-values
	ctx := config.Propagators.Extract(savedCtx, propagation.HeaderCarrier(request.Header))

	ctx, span := startSpan(c, tracer, ctx, config)
	defer span.End()

	c.SetRequest(request.WithContext(ctx))

	// Inject the trace information into the response header as Traceparent header
	config.Propagators.Inject(ctx, propagation.HeaderCarrier(c.Response().Header()))

	err := next(c)
	if err != nil {
		span.SetAttributes(attribute.String("echo.error", err.Error()))
		c.Error(err)
	}

	status := c.Response().Status
	setSpanStatus(span, status)

	return err
}

func setSpanStatus(span oteltrace.Span, status int) {
	span.SetStatus(HTTPServerStatus(status))
	if status > 0 {
		span.SetAttributes(semconv.HTTPStatusCode(status))
	}
}

func startSpan(c echo.Context, tracer oteltrace.Tracer, ctx context.Context, config *TracerConfig) (context.Context, oteltrace.Span) {
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(GetHttpRequestAttributes(c, c.Request(), config)...),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}
	if path := c.Path(); path != "" {
		rAttr := semconv.HTTPRoute(path)
		opts = append(opts, oteltrace.WithAttributes(rAttr))
	}

	spanName := c.Path()
	if spanName == "" {
		spanName = fmt.Sprintf("HTTP %s route not found", c.Request().Method)
	}

	return tracer.Start(ctx, spanName, opts...)
}

func setDefaultConfig(config *TracerConfig) {
	if config.Propagators == nil {
		config.Propagators = otel.GetTextMapPropagator()
	}
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	if config.TracerProvider == nil {
		config.TracerProvider = otel.GetTracerProvider()
	}
}

func GetHttpRequestAttributes(c echo.Context, req *http.Request, config *TracerConfig) []attribute.KeyValue {

	// http.method
	method := req.Method

	scheme := "http"
	if c.IsTLS() {
		scheme = "https"
	}

	// http.flavor
	flavor := req.Proto

	// http.target
	target := req.URL.RequestURI()

	// net.host.name
	hostName := req.Host

	// net.host.port
	host, port, _ := net.SplitHostPort(req.Host)

	// net.sock.peer.addr
	peerAddr := c.RealIP()

	// net.sock.peer.port
	peerPort := ""
	if remoteAddr := c.Request().RemoteAddr; remoteAddr != "" {
		_, peerPort, _ = net.SplitHostPort(remoteAddr)
	}

	// http.user_agent
	userAgent := req.UserAgent()

	// http.client_ip
	clientIP := c.RealIP()

	return []attribute.KeyValue{
		attribute.String("http.host", host),
		attribute.String("http.method", method),
		attribute.String("http.scheme", scheme),
		attribute.String("http.flavor", flavor),
		attribute.String("http.target", target),
		attribute.String("net.host.name", hostName),
		attribute.String("net.host.port", port),
		attribute.String("net.sock.peer.addr", peerAddr),
		attribute.String("net.sock.peer.port", peerPort),
		attribute.String("http.user_agent", userAgent),
		attribute.String("http.client_ip", clientIP),
		attribute.String("server.name", config.ServiceName),
		attribute.String("environment", config.Env),
	}
}

func HTTPServerStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}
