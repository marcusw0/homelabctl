package runner

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

type checkerFunc func(
	context.Context,
) (check.ServiceResults, error)

func (f checkerFunc) Check(
	ctx context.Context,
) (check.ServiceResults, error) {
	return f(ctx)
}

func TestRunnerMaxConcurrent(t *testing.T) {
	const (
		maxConcurrent = 2
		jobCount      = 4
	)

	started := make(chan struct{}, jobCount)
	release := make(chan struct{})

	var mut sync.Mutex
	active := 0
	maxActive := 0

	checker := checkerFunc(func(
		ctx context.Context,
	) (check.ServiceResults, error) {
		mut.Lock()
		active++
		if active > maxActive {
			maxActive = active
		}
		mut.Unlock()

		defer func() {
			mut.Lock()
			active--
			mut.Unlock()
		}()

		started <- struct{}{}

		select {
		case <-release:
		case <-ctx.Done():
			return check.ServiceResults{}, ctx.Err()
		}

		return check.ServiceResults{}, nil
	})

	jobs := make([]Job, 0, jobCount)
	for i := range jobCount {
		jobs = append(jobs, Job{
			Name:    fmt.Sprintf("job-%d", i),
			Checker: checker,
		})
	}

	results := Runner{
		MaxConcurrent: maxConcurrent,
	}.Run(context.Background(), jobs)

	for range maxConcurrent {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("workers did not start")
		}
	}

	select {
	case <-started:
		t.Fatal("a third job started before a worker was released")
	case <-time.After(100 * time.Millisecond):
	}

	mut.Lock()
	gotActive := active
	mut.Unlock()

	if gotActive != maxConcurrent {
		t.Errorf("active workers: %d, want: %d", gotActive, maxConcurrent)
	}

	close(release)

	resultCount := 0
	for range results {
		resultCount++
	}

	if resultCount != jobCount {
		t.Errorf("got results: %d, want: %d", resultCount, jobCount)
	}

	mut.Lock()
	defer mut.Unlock()

	if maxActive > maxConcurrent {
		t.Errorf("max active workers: %d, limit: %d", maxActive, maxConcurrent)
	}
}
