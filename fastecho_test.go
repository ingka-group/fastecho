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
	gocontext "context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newTestServer returns a minimal server backed by a no-op logger and a fresh
// Echo, suitable for exercising the shutdown lifecycle without tracing or
// metrics wiring.
func newTestServer() *server {
	return &server{
		Echo:   echo.New(),
		Logger: zap.NewNop(),
	}
}

// waitForReturn fails the test if serve/run does not return within the timeout.
func waitForReturn(t *testing.T, errc <-chan error) error {
	t.Helper()
	select {
	case err := <-errc:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within timeout")
		return nil
	}
}

func TestServeShutsDownWhenContextCancelled(t *testing.T) {
	s := newTestServer()
	ctx, cancel := gocontext.WithCancel(gocontext.Background())

	errc := make(chan error, 1)
	go func() {
		errc <- s.serve(ctx, "localhost", "0")
	}()

	// Give the listener a moment to bind before triggering shutdown.
	time.Sleep(50 * time.Millisecond)
	cancel()

	assert.NoError(t, waitForReturn(t, errc))
}

func TestRunShutsDownOnSIGTERM(t *testing.T) {
	s := newTestServer()

	errc := make(chan error, 1)
	go func() {
		errc <- s.run("localhost", "0")
	}()

	// Allow run() to register the signal handler and the listener to bind.
	time.Sleep(100 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(syscall.SIGTERM))

	assert.NoError(t, waitForReturn(t, errc))
}

func TestRunShutsDownOnSIGINT(t *testing.T) {
	s := newTestServer()

	errc := make(chan error, 1)
	go func() {
		errc <- s.run("localhost", "0")
	}()

	time.Sleep(100 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(syscall.SIGINT))

	assert.NoError(t, waitForReturn(t, errc))
}
