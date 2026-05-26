// Copyright © 2026 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fctx provides accessors for request-scoped values carried on
// context.Context — logger, tracer, request ID.
//
// Use [From] at the handler boundary to extract context.Context from
// echo.Context, then pass it to services and repositories.
package fctx

import (
	"context"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

// From returns the request-scoped context.Context from an echo.Context.
//
// Precondition: [Middleware] (or the deprecated ServiceContextMiddleware)
// must have run before this handler. If unsure, call [MustLogger] on the
// returned context to fail fast when the middleware chain is misconfigured.
//
// Call this once at the handler boundary; downstream code should accept
// context.Context, not echo.Context.
func From(ec echo.Context) context.Context {
	return ec.Request().Context()
}

// Logger returns the request-scoped *zap.Logger from ctx.
// Returns a no-op logger if none was registered.
func Logger(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok && l != nil {
		return l
	}
	return zap.NewNop()
}

// MustLogger returns the request-scoped *zap.Logger from ctx.
// Panics if no logger is present — use in code that must fail fast
// when context is misconfigured.
func MustLogger(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(loggerKey{}).(*zap.Logger)
	if !ok || l == nil {
		panic("fctx: no logger in context — ensure fctx.Middleware is registered")
	}
	return l
}

// Tracer returns the OTel tracer from ctx.
// Returns a no-op tracer if none was registered — no spans are produced.
func Tracer(ctx context.Context) trace.Tracer {
	if t, ok := ctx.Value(tracerKey{}).(trace.Tracer); ok && t != nil {
		return t
	}
	return noop.NewTracerProvider().Tracer("")
}

// RequestID returns the request ID from ctx.
// Returns "" if no request ID is present.
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
