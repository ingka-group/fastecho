// Copyright © 2026 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fastecho

import (
	"context"
	"errors"
	"runtime/debug"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/ingka-group/fastecho/fctx"
	"github.com/ingka-group/fastecho/telemetry"
)

// Worker is a long-running background process managed by fastecho's lifecycle.
// It should block until ctx is canceled, returning ctx.Err() (or nil) on shutdown.
type Worker func(ctx context.Context) error

// runWorker supervises a worker for the lifetime of ctx, restarting it with
// capped exponential backoff on any early return so a crashed worker is never
// left silently dead until shutdown. It exits only when ctx is canceled.
func (s *server) runWorker(ctx context.Context, name string, w Worker) {
	delay := s.workerInitialRestartDelay
	failures := 0
	for {
		start := time.Now()
		s.runWorkerOnce(ctx, name, w)
		if ctx.Err() != nil {
			return
		}

		if time.Since(start) >= s.workerStableResetThreshold {
			delay = s.workerInitialRestartDelay
			failures = 0
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		failures++
		logRestart := s.Logger.Warn
		if failures >= s.workerCrashLoopThreshold {
			logRestart = s.Logger.Error
		}
		logRestart("restarting worker",
			zap.String("worker", name),
			zap.Duration("backoff", delay),
			zap.Int("failures", failures),
		)
		delay = min(delay*2, s.workerMaxRestartDelay)
	}
}

// runWorkerOnce runs a worker exactly once, recovering panics and logging any
// error so that one worker cannot crash the process or affect the others.
func (s *server) runWorkerOnce(ctx context.Context, name string, w Worker) {
	// Fresh request id per run, so each restart's logs/spans correlate to that run.
	ctx = fctx.WithNewRequestID(ctx)
	// base carries only the worker tag; correlation fields are appended per
	// seed/span start so they never stack up as duplicates.
	base := s.Logger.With(zap.String("worker", name))
	ctx = fctx.WithLogger(ctx, base.With(fctx.Fields(ctx)...))

	// Seed a tracer that stamps worker=<name> and the run's request id on every
	// span; worker spans are roots (no inbound request), so these labels are how
	// you find all runs of a worker and mirror the fastecho.request_id attribute
	// HTTP spans get. Providers is always non-nil (noop tracer when tracing is
	// skipped).
	ctx = fctx.WithTracer(ctx, workerTracer{
		Tracer: s.Providers.TracerProvider.Tracer(telemetry.ScopeName),
		opt: trace.WithAttributes(
			attribute.String("worker", name),
			attribute.String("fastecho.request_id", fctx.RequestID(ctx)),
		),
		base: base,
	})

	defer func() {
		if r := recover(); r != nil {
			s.recordWorkerFailure(ctx, name, "panic")
			fctx.Logger(ctx).Error("worker panic recovered",
				zap.Any("panic", r),
				zap.ByteString("stack", debug.Stack()),
			)
		}
	}()

	// A canceled context is the expected shutdown signal, not an error.
	if err := w(ctx); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		s.recordWorkerFailure(ctx, name, "error")
		fctx.Logger(ctx).Error("worker exited with error", zap.Error(err))
	}
}

// recordWorkerFailure increments the worker-failure counter, labeled by worker
// and kind (panic|error), so failures are alertable per worker.
func (s *server) recordWorkerFailure(ctx context.Context, name, kind string) {
	s.workerFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String("worker", name),
		attribute.String("kind", kind),
	))
}

// workerTracer stamps the worker's identity on every span it starts and
// re-seeds the context logger with the span's correlation fields, so worker log
// lines written inside a span carry trace_id/span_id like request logs do. We
// deliberately do not span the whole run: workers block until shutdown, so a
// run-spanning span would stay open for the service lifetime and never export.
// The worker body spans each unit of work; this tracer labels those spans.
type workerTracer struct {
	trace.Tracer
	opt  trace.SpanStartOption
	base *zap.Logger
}

func (t workerTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// Copied, not appended in place: appending to the variadic slice can write
	// into the caller's backing array when it has spare capacity.
	merged := make([]trace.SpanStartOption, 0, len(opts)+1)
	merged = append(append(merged, opts...), t.opt)
	ctx, span := t.Tracer.Start(ctx, name, merged...)
	// Rebuilt from base (not the stored logger) so nested spans don't stack
	// duplicate correlation fields.
	return fctx.WithLogger(ctx, t.base.With(fctx.Fields(ctx)...)), span
}

// drainWorkers waits for all workers to return, bounded by ctx. If they do not
// drain in time it logs a warning and returns anyway.
func (s *server) drainWorkers(ctx context.Context, workers *sync.WaitGroup) {
	done := make(chan struct{})

	// A worker that ignores cancellation leaks this goroutine, which is
	// acceptable: drainWorkers runs only as the process exits.
	go func() {
		workers.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		select {
		case <-done:
		default:
			s.Logger.Warn("workers did not drain within shutdown timeout")
		}
	}
}
