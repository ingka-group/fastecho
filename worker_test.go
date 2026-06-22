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
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/ingka-group/fastecho/fctx"
)

// newObserverServer builds a bare server whose logs are captured for assertions.
func newObserverServer() (*server, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.InfoLevel)
	return &server{
		Echo:                       echo.New(),
		Logger:                     zap.New(core),
		workerInitialRestartDelay:  1 * time.Millisecond,
		workerMaxRestartDelay:      10 * time.Millisecond,
		workerStableResetThreshold: 1 * time.Second,
	}, logs
}

func TestServeCancelsWorkerContextOnShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)
	s := newTestServer()
	cancelled := make(chan struct{})
	s.Workers = []Worker{
		func(ctx context.Context) error {
			<-ctx.Done()
			close(cancelled)
			return ctx.Err()
		},
	}

	p := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())

	errc := make(chan error, 1)
	go func() {
		errc <- s.serve(ctx, "localhost", p)
	}()

	waitUntilServing(t, net.JoinHostPort("localhost", p))
	cancel()

	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("worker context was not cancelled on shutdown")
	}

	assert.NoError(t, waitForReturn(t, errc))
}

func TestServeWaitsForWorkersOnShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)
	s := newTestServer()
	var drained atomic.Bool
	s.Workers = []Worker{
		func(ctx context.Context) error {
			<-ctx.Done()
			time.Sleep(100 * time.Millisecond)
			drained.Store(true)
			return nil
		},
	}

	p := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())

	errc := make(chan error, 1)
	go func() {
		errc <- s.serve(ctx, "localhost", p)
	}()

	waitUntilServing(t, net.JoinHostPort("localhost", p))
	cancel()

	require.NoError(t, waitForReturn(t, errc))
	assert.True(t, drained.Load(), "serve returned before worker finished draining")
}

func TestServeSurvivesFailingWorkers(t *testing.T) {
	defer goleak.VerifyNone(t)
	s := newTestServer()
	healthy := make(chan struct{})
	s.Workers = []Worker{
		func(ctx context.Context) error {
			return errors.New("boom")
		},
		func(ctx context.Context) error {
			panic("kaboom")
		},
		func(ctx context.Context) error {
			close(healthy)
			<-ctx.Done()
			return nil
		},
	}

	p := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())

	errc := make(chan error, 1)
	go func() {
		errc <- s.serve(ctx, "localhost", p)
	}()

	waitUntilServing(t, net.JoinHostPort("localhost", p))

	select {
	case <-healthy:
	case <-time.After(2 * time.Second):
		t.Fatal("surviving worker did not run alongside failing workers")
	}

	cancel()
	assert.NoError(t, waitForReturn(t, errc))
}

func TestRunWorkerOnce(t *testing.T) {
	tests := []struct {
		name        string
		worker      Worker
		wantMessage string // empty means no logs expected
		wantError   string // expected "error" log field, if any
	}{
		{
			name:        "logs error",
			worker:      func(context.Context) error { return errors.New("boom") },
			wantMessage: "worker exited with error",
			wantError:   "boom",
		},
		{
			name:        "recovers panic",
			worker:      func(context.Context) error { panic("kaboom") },
			wantMessage: "worker panic recovered",
		},
		{
			name:        "silent on clean exit",
			worker:      func(context.Context) error { return nil },
			wantMessage: "",
		},
		{
			name:        "silent on context cancellation",
			worker:      func(ctx context.Context) error { return context.Canceled },
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, logs := newObserverServer()

			assert.NotPanics(t, func() {
				s.runWorkerOnce(context.Background(), tt.worker)
			})

			if tt.wantMessage == "" {
				assert.Equal(t, 0, logs.Len(), "clean worker exit should not log")
				return
			}

			entries := logs.FilterMessage(tt.wantMessage).All()
			require.Len(t, entries, 1)
			assert.Equal(t, zapcore.ErrorLevel, entries[0].Level)
			if tt.wantError != "" {
				assert.Equal(t, tt.wantError, entries[0].ContextMap()["error"])
			}
		})
	}
}

type stubTracer struct{ embedded.Tracer }

func (stubTracer) Start(ctx context.Context, _ string, _ ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, nil
}

func TestRunWorkerInjectsLogger(t *testing.T) {
	s, logs := newObserverServer()

	s.runWorkerOnce(context.Background(), func(ctx context.Context) error {
		fctx.Logger(ctx).Info("from worker")
		return nil
	})

	require.Equal(t, 1, logs.FilterMessage("from worker").Len(),
		"worker context should carry the server logger, not the no-op fallback")
}

func TestRunWorkerInjectsTracer(t *testing.T) {
	s, _ := newObserverServer()
	var tracer trace.Tracer = stubTracer{}
	s.Tracer = &tracer

	var got trace.Tracer
	s.runWorkerOnce(context.Background(), func(ctx context.Context) error {
		got = fctx.Tracer(ctx)
		return nil
	})

	assert.IsType(t, stubTracer{}, got,
		"worker context should carry the server tracer, not the no-op fallback")
}

func TestRunWorkerRestartsAfterFailuresThenStabilizes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, logs := newObserverServer()
		var calls atomic.Int32
		healthy := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())

		go s.runWorker(ctx, func(ctx context.Context) error {
			if calls.Add(1) < 3 {
				return errors.New("boom")
			}
			close(healthy)
			<-ctx.Done()
			return ctx.Err()
		})

		<-healthy
		synctest.Wait()
		assert.Equal(t, int32(3), calls.Load())
		assert.Equal(t, 2, logs.FilterMessage("restarting worker").Len())

		cancel()
		synctest.Wait()
	})
}

func TestRunWorkerRestartsAfterPanic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, logs := newObserverServer()
		var calls atomic.Int32
		healthy := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())

		go s.runWorker(ctx, func(ctx context.Context) error {
			if calls.Add(1) == 1 {
				panic("kaboom")
			}
			close(healthy)
			<-ctx.Done()
			return ctx.Err()
		})

		<-healthy
		synctest.Wait()
		assert.Equal(t, 1, logs.FilterMessage("worker panic recovered").Len())
		assert.GreaterOrEqual(t, logs.FilterMessage("restarting worker").Len(), 1)

		cancel()
		synctest.Wait()
	})
}

func TestRunWorkerRestartsAfterCleanReturn(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, _ := newObserverServer()
		var calls atomic.Int32
		healthy := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())

		go s.runWorker(ctx, func(ctx context.Context) error {
			if calls.Add(1) == 1 {
				return nil
			}
			close(healthy)
			<-ctx.Done()
			return ctx.Err()
		})

		<-healthy
		synctest.Wait()
		assert.GreaterOrEqual(t, calls.Load(), int32(2),
			"a clean return while ctx is live should be treated as a fault and restarted")

		cancel()
		synctest.Wait()
	})
}

func TestRunWorkerStopsOnContextCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, logs := newObserverServer()
		running := make(chan struct{})
		done := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			s.runWorker(ctx, func(ctx context.Context) error {
				close(running)
				<-ctx.Done()
				return ctx.Err()
			})
			close(done)
		}()

		<-running
		cancel()
		<-done
		synctest.Wait()
		assert.Equal(t, 0, logs.FilterMessage("restarting worker").Len(),
			"a worker stopped by shutdown must not be reported as restarting")
		assert.Equal(t, 0, logs.FilterMessage("worker exited with error").Len())
	})
}

func TestRunWorkerCancelDuringBackoffReturnsPromptly(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, logs := newObserverServer()
		s.workerInitialRestartDelay = time.Hour
		s.workerMaxRestartDelay = time.Hour
		var calls atomic.Int32
		failed := make(chan struct{})
		done := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			s.runWorker(ctx, func(ctx context.Context) error {
				calls.Add(1)
				close(failed)
				return errors.New("boom")
			})
			close(done)
		}()

		<-failed
		cancel()
		<-done
		synctest.Wait()
		assert.Equal(t, int32(1), calls.Load(),
			"cancelling during backoff must not wait out the (1h) delay or restart")
		assert.Equal(t, 0, logs.FilterMessage("restarting worker").Len())
	})
}

func TestRunWorkerResetsBackoffAfterStableRun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, logs := newObserverServer()
		s.workerInitialRestartDelay = time.Millisecond
		s.workerMaxRestartDelay = time.Second
		s.workerStableResetThreshold = time.Second
		var calls atomic.Int32
		healthy := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())

		go s.runWorker(ctx, func(ctx context.Context) error {
			switch calls.Add(1) {
			case 1:
				return errors.New("boom")
			case 2:
				time.Sleep(2 * time.Second) // outlasts the stable-reset threshold
				return errors.New("boom again")
			default:
				close(healthy)
				<-ctx.Done()
				return ctx.Err()
			}
		})

		<-healthy
		synctest.Wait()
		entries := logs.FilterMessage("restarting worker").All()
		require.Len(t, entries, 2)
		assert.Equal(t, time.Millisecond, entries[0].ContextMap()["backoff"])
		assert.Equal(t, time.Millisecond, entries[1].ContextMap()["backoff"],
			"backoff must reset to the initial delay after a stable run")

		cancel()
		synctest.Wait()
	})
}

func TestDrainWorkersWaitsForCompletion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, logs := newObserverServer()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			time.Sleep(20 * time.Millisecond)
			wg.Done()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s.drainWorkers(ctx, &wg)
		assert.Equal(t, 0, logs.Len(), "drain that completes in time should not log")
	})
}

func TestDrainWorkersTimesOut(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, logs := newObserverServer()

		var wg sync.WaitGroup
		wg.Add(1) // not released until after the timeout, simulating a worker ignoring cancellation

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		s.drainWorkers(ctx, &wg)
		assert.Equal(t, 1, logs.FilterMessage("workers did not drain within shutdown timeout").Len())

		// Release so drainWorkers' helper goroutine can exit and the bubble drains.
		wg.Done()
		synctest.Wait()
	})
}
