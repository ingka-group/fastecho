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

// Fields returns the correlation fields for ctx, shared by middleware, the
// recover handler, and worker seeds. They are discrete, indexable log fields
// (not the traceparent wire header) so the backend can query by trace_id.
func Fields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		fields = append(fields,
			// trace_id: correlates logs and spans across services for one request.
			zap.String("trace_id", sc.TraceID().String()),
			// span_id: the specific operation within that trace.
			zap.String("span_id", sc.SpanID().String()),
		)
	}
	if id := RequestID(ctx); id != "" {
		// request_id: our own per-request id, for correlation without a sampled trace.
		fields = append(fields, zap.String("request_id", id))
	}
	return fields
}

// NewRequestID returns a fresh UUIDv4 request id
func NewRequestID() string {
	return uuid.New().String()
}
