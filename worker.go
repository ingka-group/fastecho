// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
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

	"go.uber.org/zap"

	"github.com/ingka-group/fastecho/fctx"
)

// Worker is a long-running background process managed by fastecho's lifecycle.
// It should block until ctx is cancelled, returning ctx.Err() (or nil) on shutdown.
type Worker func(ctx context.Context) error

// runWorker supervises a worker for the lifetime of ctx, restarting it with
// capped exponential backoff on any early return so a crashed worker is never
// left silently dead until shutdown. It exits only when ctx is cancelled.
func (s *server) runWorker(ctx context.Context, w Worker) {
	delay := s.workerInitialRestartDelay
	for {
		start := time.Now()
		s.runWorkerOnce(ctx, w)
		if ctx.Err() != nil {
			return
		}

		if time.Since(start) >= s.workerStableResetThreshold {
			delay = s.workerInitialRestartDelay
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		// After the wait, so cancelling during backoff exits without a misleading log.
		s.Logger.Warn("restarting worker", zap.Duration("backoff", delay))
		delay = min(delay*2, s.workerMaxRestartDelay)
	}
}

// runWorkerOnce runs a worker exactly once, recovering panics and logging any
// error so that one worker cannot crash the process or affect the others.
func (s *server) runWorkerOnce(ctx context.Context, w Worker) {
	ctx = fctx.WithLogger(ctx, s.Logger)
	if s.Tracer != nil {
		ctx = fctx.WithTracer(ctx, *s.Tracer)
	}

	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("worker panic recovered",
				zap.Any("panic", r),
				zap.ByteString("stack", debug.Stack()),
			)
		}
	}()

	// A cancelled context is the expected shutdown signal, not an error.
	if err := w(ctx); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		s.Logger.Error("worker exited with error", zap.Error(err))
	}
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
