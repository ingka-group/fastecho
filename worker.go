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
	"runtime/debug"
	"sync"

	"go.uber.org/zap"
)

// Worker is a long-running background process managed by fastecho's lifecycle.
// It should block until ctx is cancelled, returning ctx.Err() (or nil) on shutdown.
type Worker func(ctx context.Context) error

// runWorker runs a single worker, recovering panics and logging any error so
// that one worker cannot crash the process or affect the others.
func (s *server) runWorker(ctx context.Context, w Worker) {
	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("worker panic recovered",
				zap.Any("panic", r),
				zap.ByteString("stack", debug.Stack()),
			)
		}
	}()

	if err := w(ctx); err != nil {
		s.Logger.Error("worker exited with error", zap.Error(err))
	}
}

// drainWorkers waits for all workers to return, bounded by ctx. If they do not
// drain in time it logs a warning and returns anyway.
func (s *server) drainWorkers(ctx context.Context, workers *sync.WaitGroup) {
	done := make(chan struct{})
	// If a worker ignores cancellation this goroutine stays blocked on Wait
	// forever; that is acceptable because drainWorkers only runs once as the
	// process exits, and a WaitGroup offers no way to abandon the wait.
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
