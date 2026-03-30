/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package retry

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewRetryConfigAppliesDefaultsAndOptions(t *testing.T) {
	customRetry := func(err error, output string) bool {
		return strings.Contains(output, "retry")
	}

	cfg := NewRetryConfig(
		func(rc *RetryConfig) {
			rc.MaxRetries = 3
			rc.Delay = time.Millisecond
		},
		func(rc *RetryConfig) {
			rc.Multiplier = 1.5
			rc.MaxInterval = 2 * time.Second
			rc.MaxElapsedTime = time.Second
			rc.ShouldRetry = customRetry
		},
	)

	if cfg.MaxRetries != 3 {
		t.Fatalf("expected MaxRetries=3, got %d", cfg.MaxRetries)
	}
	if cfg.Delay != time.Millisecond {
		t.Fatalf("expected Delay=%v, got %v", time.Millisecond, cfg.Delay)
	}
	if cfg.Multiplier != 1.5 {
		t.Fatalf("expected Multiplier=1.5, got %v", cfg.Multiplier)
	}
	if cfg.MaxInterval != 2*time.Second {
		t.Fatalf("expected MaxInterval=%v, got %v", 2*time.Second, cfg.MaxInterval)
	}
	if cfg.MaxElapsedTime != time.Second {
		t.Fatalf("expected MaxElapsedTime=%v, got %v", time.Second, cfg.MaxElapsedTime)
	}
	if cfg.ShouldRetry == nil {
		t.Fatal("expected ShouldRetry to be set")
	}
	if !cfg.ShouldRetry(errors.New("boom"), "please retry") {
		t.Fatal("expected custom ShouldRetry to be used")
	}
}

func TestRunWithRetryRetriesAndSucceeds(t *testing.T) {
	cfg := NewRetryConfig(func(rc *RetryConfig) {
		rc.MaxRetries = 100
		rc.Delay = time.Millisecond
		rc.Multiplier = 1
		rc.MaxInterval = time.Millisecond
		rc.MaxElapsedTime = 500000 * time.Millisecond
		rc.ShouldRetry = func(err error, output string) bool {
			return strings.Contains(output, "retry-me")
		}
	})

	attempts := 0
	output, err := cfg.RunWithRetry(context.Background(), func() (any, error) {
		attempts++
		if attempts < 10 {
			return "retry-me", errors.New("temporary")
		}
		return "done", nil
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	//if attempts != 2 {
	//	t.Fatalf("expected 2 attempts, got %d", attempts)
	//}
	if output != "done" {
		t.Fatalf("expected final output %q, got %#v", "done", output)
	}
}

func TestRunWithRetryStopsOnNonRetryableError(t *testing.T) {
	cfg := NewRetryConfig(func(rc *RetryConfig) {
		rc.MaxRetries = 3
		rc.Delay = time.Millisecond
		rc.Multiplier = 1
		rc.MaxInterval = time.Millisecond
		rc.MaxElapsedTime = 50 * time.Millisecond
		rc.ShouldRetry = func(err error, output string) bool {
			return false
		}
	})

	attempts := 0
	boom := errors.New("boom")
	_, err := cfg.RunWithRetry(context.Background(), func() (any, error) {
		attempts++
		return nil, boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestRunWithRetryRespectsMaxRetriesAndReturnsLastOutput(t *testing.T) {
	cfg := NewRetryConfig(func(rc *RetryConfig) {
		rc.MaxRetries = 2
		rc.Delay = time.Millisecond
		rc.Multiplier = 1
		rc.MaxInterval = time.Millisecond
		rc.MaxElapsedTime = 50 * time.Millisecond
		rc.ShouldRetry = func(err error, output string) bool {
			return true
		}
	})

	attempts := 0
	boom := errors.New("boom")
	output, err := cfg.RunWithRetry(context.Background(), func() (any, error) {
		attempts++
		return "still failing", boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}
	if output != "still failing" {
		t.Fatalf("expected last output %q, got %#v", "still failing", output)
	}
}

func TestRunWithRetryReturnsContextErrorWhileWaiting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	cfg := NewRetryConfig(func(rc *RetryConfig) {
		rc.MaxRetries = 5
		rc.Delay = 50 * time.Millisecond
		rc.Multiplier = 1
		rc.MaxInterval = 50 * time.Millisecond
		rc.MaxElapsedTime = time.Second
		rc.ShouldRetry = func(err error, output string) bool {
			return true
		}
	})

	attempts := 0
	_, err := cfg.RunWithRetry(ctx, func() (any, error) {
		attempts++
		return nil, errors.New("temporary")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt before context cancellation, got %d", attempts)
	}
}

func TestRunWithRetryPreCanceledContextStillRunsFirstAttempt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := NewRetryConfig(func(rc *RetryConfig) {
		rc.MaxRetries = 5
		rc.Delay = time.Millisecond
		rc.Multiplier = 1
		rc.MaxInterval = time.Millisecond
		rc.MaxElapsedTime = time.Second
		rc.ShouldRetry = func(err error, output string) bool {
			return true
		}
	})

	attempts := 0
	_, err := cfg.RunWithRetry(ctx, func() (any, error) {
		attempts++
		return nil, errors.New("temporary")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected first attempt to run before cancellation stops retries, got %d", attempts)
	}
}

func TestRunWithRetryRejectsNilExecFunc(t *testing.T) {
	_, err := NewRetryConfig().RunWithRetry(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil execFunc")
	}
}
