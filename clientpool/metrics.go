package clientpool

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rfratto/ckit/internal/metricsutil"
)

type metrics struct {
	container metricsutil.Container

	currentConns prometheus.Gauge
	gcActive     prometheus.Gauge
	gcTotal      prometheus.Histogram
	eventsTotal  *prometheus.CounterVec
	lookupsTotal *prometheus.CounterVec

	maxConns  prometheus.Gauge
	autoClose prometheus.Gauge
}

var _ prometheus.Collector = (*metrics)(nil)

func newMetrics(o Options) *metrics {
	var m metrics

	m.currentConns = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clientpool_conns",
		Help: "Current number of open gRPC connections",
	})
	m.gcActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clientpool_gc_active",
		Help: "1 if the clientpool GC is running",
	})
	m.gcTotal = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "clientpool_gc_duration_seconds",
		Help:    "Histogram of the latency for GCs",
		Buckets: prometheus.DefBuckets,
	})
	m.eventsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clientpool_events_total",
		Help: "Total number of times connections were opened or closed.",
	}, []string{"event"})
	m.lookupsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clientpool_lookups_total",
		Help: "Total number of lookups for a connection. result will be one of: success, error_dial, error_max_conns, or error_other.",
	}, []string{"result"})

	m.maxConns = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clientpool_max_conns",
		Help: "Maximum number of connections the clientpool can accept. 0 = unlimited",
	})

	m.autoClose = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clientpool_auto_close",
		Help: "When 1, the least-recently-used connection will closed when opening a new connection and the connection limit is reached.",
	})

	// Set constants
	m.maxConns.Set(float64(o.MaxClients))
	m.autoClose.Set(boolToFloat64(o.CleanupLRU))

	m.container.Add(
		m.currentConns,
		m.gcActive,
		m.gcTotal,
		m.eventsTotal,
		m.lookupsTotal,
		m.maxConns,
		m.autoClose,
	)

	return &m
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func (m *metrics) Describe(ch chan<- *prometheus.Desc) {
	m.container.Describe(ch)
}

func (m *metrics) Collect(ch chan<- prometheus.Metric) {
	m.container.Collect(ch)
}
