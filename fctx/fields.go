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

package fctx

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Fields returns the correlation fields for ctx: trace_id and span_id (when a
// valid span context is present) and request_id (when set). Centralising them
// here keeps middleware, the recover handler, and worker seeds consistent.
// They are discrete, indexable log fields (not the traceparent wire header) so
// the log backend can query by trace_id and link a line to its trace.
func Fields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		fields = append(fields,
			zap.String("trace_id", sc.TraceID().String()),
			zap.String("span_id", sc.SpanID().String()),
		)
	}
	if id := RequestID(ctx); id != "" {
		fields = append(fields, zap.String("request_id", id))
	}
	return fields
}

// NewRequestID returns a fresh UUIDv4 request id. One format everywhere: HTTP
// uses it as Echo's request-id Generator, workers via WithNewRequestID.
func NewRequestID() string {
	return uuid.New().String()
}

// WithNewRequestID generates a new request id and stores it in ctx.
func WithNewRequestID(ctx context.Context) context.Context {
	return WithRequestID(ctx, NewRequestID())
}
