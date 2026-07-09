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

package telemetry

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ingka-group/fastecho/fctx"
)

// WrapClient decorates c's transport so every request propagates the caller's
// trace context (via otelhttp) and forwards fctx.RequestID as X-Request-Id. It
// mutates and returns c; a nil client is allocated. Only the transport changes.
func WrapClient(c *http.Client) *http.Client {
	if c == nil {
		c = &http.Client{}
	}
	c.Transport = WrapTransport(c.Transport)
	return c
}

// WrapTransport wraps base (nil → http.DefaultTransport) with trace-context
// propagation and X-Request-Id forwarding. Use this when you hold a
// RoundTripper directly (e.g. instrumenting a generated client). Wrapping an
// already-wrapped transport is a no-op, so double-wrapping cannot stack
// duplicate spans.
func WrapTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if _, ok := base.(requestIDTransport); ok {
		return base
	}
	return requestIDTransport{base: otelhttp.NewTransport(base), original: base}
}

// requestIDTransport injects the fctx request id and delegates to the otelhttp
// transport in base; original is kept because otelhttp forwards nothing beyond
// RoundTrip, and net/http reaches CloseIdleConnections via type assertion.
type requestIDTransport struct {
	base     http.RoundTripper
	original http.RoundTripper
}

func (t requestIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// The RoundTripper contract forbids mutating the caller's request, so clone
	// before injecting. otelhttp clones again for its own header injection, so
	// the no-injection path can pass req through untouched.
	if id := fctx.RequestID(req.Context()); id != "" && req.Header.Get(echo.HeaderXRequestID) == "" {
		req = req.Clone(req.Context())
		req.Header.Set(echo.HeaderXRequestID, id)
	}
	return t.base.RoundTrip(req)
}

// CloseIdleConnections forwards to the original transport so connection
// rotation and shutdown hygiene keep working on wrapped clients.
func (t requestIDTransport) CloseIdleConnections() {
	if ci, ok := t.original.(interface{ CloseIdleConnections() }); ok {
		ci.CloseIdleConnections()
	}
}
