// Package metrics exposes the numbers needed to run this in production.
//
// The set is deliberately small and chosen from the cost model rather than from
// whatever is easy to instrument: storage growth, encode cost per source hour, cache
// hit ratio and JIT hit rate are the four that decide whether the economics hold.
package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Registry is a minimal Prometheus text-format exporter.
//
// Hand-rolled rather than pulling the client library: the whole surface is a handful
// of counters and histograms, and a scrape endpoint is a few lines. Swap in the real
// client when the exposition format stops being enough.
type Registry struct {
	mu         sync.RWMutex
	counters   map[string]*counter
	histograms map[string]*histogram
}

type counter struct {
	help   string
	values map[string]float64 // label-set -> value
}

type histogram struct {
	help    string
	buckets []float64
	series  map[string]*histSeries
}

type histSeries struct {
	counts []uint64
	sum    float64
	total  uint64
}

func New() *Registry {
	return &Registry{
		counters:   map[string]*counter{},
		histograms: map[string]*histogram{},
	}
}

// Inc adds to a counter. Labels are passed as alternating key/value pairs.
func (r *Registry) Inc(name, help string, delta float64, labels ...string) {
	key := labelKey(labels)
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.counters[name]
	if !ok {
		c = &counter{help: help, values: map[string]float64{}}
		r.counters[name] = c
	}
	c.values[key] += delta
}

// Observe records a value into a histogram.
func (r *Registry) Observe(name, help string, buckets []float64, v float64, labels ...string) {
	key := labelKey(labels)
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.histograms[name]
	if !ok {
		h = &histogram{help: help, buckets: buckets, series: map[string]*histSeries{}}
		r.histograms[name] = h
	}
	s, ok := h.series[key]
	if !ok {
		s = &histSeries{counts: make([]uint64, len(h.buckets))}
		h.series[key] = s
	}
	for i, b := range h.buckets {
		if v <= b {
			s.counts[i]++
		}
	}
	s.sum += v
	s.total++
}

// Timer returns a function that records elapsed seconds when called.
func (r *Registry) Timer(name, help string, buckets []float64, labels ...string) func() {
	start := time.Now()
	return func() {
		r.Observe(name, help, buckets, time.Since(start).Seconds(), labels...)
	}
}

// Handler serves the Prometheus scrape endpoint.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		for name, c := range r.counters {
			writeHelp(w, name, c.help, "counter")
			for labels, v := range c.values {
				writeSample(w, name, labels, "", v)
			}
		}
		for name, h := range r.histograms {
			writeHelp(w, name, h.help, "histogram")
			for labels, s := range h.series {
				for i, b := range h.buckets {
					writeBucket(w, name, labels, formatFloat(b), float64(s.counts[i]))
				}
				writeBucket(w, name, labels, "+Inf", float64(s.total))
				writeSample(w, name+"_sum", labels, "", s.sum)
				writeSample(w, name+"_count", labels, "", float64(s.total))
			}
		}
	})
}

func writeHelp(w http.ResponseWriter, name, help, kind string) {
	if help != "" {
		w.Write([]byte("# HELP " + name + " " + help + "\n"))
	}
	w.Write([]byte("# TYPE " + name + " " + kind + "\n"))
}

func writeSample(w http.ResponseWriter, name, labels, extra string, v float64) {
	w.Write([]byte(name + renderLabels(labels, extra) + " " + formatFloat(v) + "\n"))
}

func writeBucket(w http.ResponseWriter, name, labels, le string, v float64) {
	w.Write([]byte(name + "_bucket" + renderLabels(labels, `le="`+le+`"`) + " " +
		formatFloat(v) + "\n"))
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}
