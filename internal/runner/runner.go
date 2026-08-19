package runner

import (
	"context"
	"sync"

	"github.com/marcusw0/homelabctl/internal/check"
)

type ServiceChecker interface {
	Check(context.Context) (check.ServiceResults, error)
}

type Job struct {
	Name    string
	Checker ServiceChecker
}

type Result struct {
	Name   string
	Checks check.ServiceResults
	Err    error
}

type Runner struct {
	MaxConcurrent int
}

func (r Runner) Run(ctx context.Context, jobs []Job) <-chan Result {
	results := make(chan Result)

	go func() {
		defer close(results)

		if len(jobs) == 0 {
			return
		}

		workerCount := r.MaxConcurrent
		if workerCount < 1 {
			workerCount = 1
		}

		if workerCount > len(jobs) {
			workerCount = len(jobs)
		}

		jobCh := make(chan Job)
		var wg sync.WaitGroup

		wg.Add(workerCount)
		for range workerCount {
			go func() {
				defer wg.Done()

				for job := range jobCh {
					checks, err := job.Checker.Check(ctx)

					select {
					case results <- Result{
						Name:   job.Name,
						Checks: checks,
						Err:    err,
					}:
					case <-ctx.Done():
						return
					}
				}
			}()
		}

		for _, job := range jobs {
			select {
			case jobCh <- job:
			case <-ctx.Done():
				close(jobCh)
				wg.Wait()
				return
			}
		}

		close(jobCh)
		wg.Wait()
	}()

	return results
}
