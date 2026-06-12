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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// newObserverServer builds a bare server whose logs are captured for assertions.
func newObserverServer() (*server, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.InfoLevel)
	return &server{
		Echo:   echo.New(),
		Logger: zap.New(core),
	}, logs
}

func TestServeCancelsWorkerContextOnShutdown(t *testing.T) {
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

func TestServeKeepsRunningWhenWorkerErrorsOrPanics(t *testing.T) {
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

func TestRunWorkerLogsError(t *testing.T) {
	s, logs := newObserverServer()

	s.runWorker(context.Background(), func(ctx context.Context) error {
		return errors.New("boom")
	})

	entries := logs.FilterMessage("worker exited with error").All()
	require.Len(t, entries, 1)
	assert.Equal(t, zapcore.ErrorLevel, entries[0].Level)
	assert.Equal(t, "boom", entries[0].ContextMap()["error"])
}

func TestRunWorkerRecoversPanic(t *testing.T) {
	s, logs := newObserverServer()

	assert.NotPanics(t, func() {
		s.runWorker(context.Background(), func(ctx context.Context) error {
			panic("kaboom")
		})
	})

	assert.Equal(t, 1, logs.FilterMessage("worker panic recovered").Len())
}

func TestRunWorkerSilentOnCleanExit(t *testing.T) {
	s, logs := newObserverServer()

	s.runWorker(context.Background(), func(ctx context.Context) error {
		return nil
	})

	assert.Equal(t, 0, logs.Len(), "clean worker exit should not log")
}

func TestDrainWorkersWaitsForCompletion(t *testing.T) {
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
}

func TestDrainWorkersTimesOut(t *testing.T) {
	s, logs := newObserverServer()

	var wg sync.WaitGroup
	wg.Add(1) // never released, simulating a worker ignoring cancellation

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	s.drainWorkers(ctx, &wg)
	assert.Equal(t, 1, logs.FilterMessage("workers did not drain within shutdown timeout").Len())
}
