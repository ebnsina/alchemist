package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func scrape(t *testing.T, r *Registry) string {
	t.Helper()
	w := httptest.NewRecorder()
	r.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	return w.Body.String()
}

func TestCounterAndLabels(t *testing.T) {
	r := New()
	r.Inc("alchemist_jobs_total", "jobs processed", 1, "kind", "transcode", "result", "ok")
	r.Inc("alchemist_jobs_total", "jobs processed", 2, "kind", "transcode", "result", "ok")
	r.Inc("alchemist_jobs_total", "jobs processed", 1, "kind", "transcode", "result", "failed")

	out := scrape(t, r)
	if !strings.Contains(out, `alchemist_jobs_total{kind="transcode",result="ok"} 3`) {
		t.Errorf("counter did not accumulate:\n%s", out)
	}
	if !strings.Contains(out, `alchemist_jobs_total{kind="transcode",result="failed"} 1`) {
		t.Errorf("label sets were not kept apart:\n%s", out)
	}
	if !strings.Contains(out, "# TYPE alchemist_jobs_total counter") {
		t.Errorf("missing TYPE line, scrapers reject this:\n%s", out)
	}
}

func TestHistogramBuckets(t *testing.T) {
	r := New()
	buckets := []float64{1, 10, 60}
	for _, v := range []float64{0.5, 5, 30, 300} {
		r.Observe("alchemist_encode_seconds", "encode duration", buckets, v)
	}

	out := scrape(t, r)
	for _, want := range []string{
		`alchemist_encode_seconds_bucket{le="1"} 1`,
		`alchemist_encode_seconds_bucket{le="10"} 2`,
		`alchemist_encode_seconds_bucket{le="60"} 3`,
		`alchemist_encode_seconds_bucket{le="+Inf"} 4`,
		`alchemist_encode_seconds_count 4`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// Label values can carry customer-controlled text; a stray quote would produce a
// scrape body Prometheus rejects, silently losing every metric on the endpoint.
func TestLabelValuesAreEscaped(t *testing.T) {
	r := New()
	r.Inc("alchemist_errors_total", "errors", 1, "code", `we"ird\value`)
	out := scrape(t, r)
	if !strings.Contains(out, `code="we\"ird\\value"`) {
		t.Errorf("label value not escaped:\n%s", out)
	}
}

// A metric call must never be able to take the process down.
func TestOddLabelsDoNotPanic(t *testing.T) {
	r := New()
	r.Inc("alchemist_odd_total", "odd", 1, "only_a_key")
	if out := scrape(t, r); !strings.Contains(out, "alchemist_odd_total") {
		t.Errorf("odd label count lost the metric entirely:\n%s", out)
	}
}

func TestTimerRecords(t *testing.T) {
	r := New()
	done := r.Timer("alchemist_op_seconds", "op", []float64{1000}, "op", "test")
	done()
	if out := scrape(t, r); !strings.Contains(out, `alchemist_op_seconds_count{op="test"} 1`) {
		t.Errorf("timer did not record:\n%s", out)
	}
}
