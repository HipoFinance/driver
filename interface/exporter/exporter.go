package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	METRIC_ERROR_COUNT = "error_count"
)

var (
	counters map[string]prometheus.Counter
)

func Init() {

	// --- Static Metrics: the metrics which are not depended on running configuration

	// Create metric spaces
	counters = make(map[string]prometheus.Counter)

	// Register metrics
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "hipo",
		Subsystem: "driver",
		Name:      METRIC_ERROR_COUNT,
		Help:      "Counts the number of successful requests",
	})
	prometheus.MustRegister(counter)
	counters[METRIC_ERROR_COUNT] = counter
}

func GetCounter(name string) prometheus.Counter {
	return counters[name]
}

func IncErrorCount() {
	counters[METRIC_ERROR_COUNT].Inc()
}
