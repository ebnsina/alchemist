package pipeline

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/ebnsina/alchemist/internal/platform/metrics"
)

// JobMetrics counts every job outcome, on every queue, in one place.
//
// It replaces a method on the transcode worker that nothing ever called, so
// alchemist_jobs_total was never emitted and an alert written against it would have
// stayed silent -- indistinguishable from a healthy fleet. Middleware rather than a
// call in each Work: a worker added later is counted without anyone remembering to.
func JobMetrics(reg *metrics.Registry) rivertype.WorkerMiddleware {
	return river.WorkerMiddlewareFunc(func(ctx context.Context, job *rivertype.JobRow,
		doInner func(ctx context.Context) error) error {
		err := doInner(ctx)
		result := "success"
		if err != nil {
			result = "error"
		}
		reg.Inc("alchemist_jobs_total", "jobs processed by kind and result", 1,
			"kind", job.Kind, "result", result)
		return err
	})
}
