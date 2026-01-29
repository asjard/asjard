package metrics

import (
	"sync"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server/handlers"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/push"
)

// MetricsManager coordinates the prometheus registry and collector lifecycle.
type MetricsManager struct {
	conf       Config
	collectors map[string]prometheus.Collector // Map of named collectors (e.g., "go_collector")
	cm         sync.RWMutex                    // Protects the collectors map for concurrent access
}

var (
	registry       *prometheus.Registry
	metricsManager *MetricsManager
)

func init() {
	// Initialize a custom registry instead of the DefaultRegistry to maintain isolation.
	registry = prometheus.NewRegistry()
	metricsManager = &MetricsManager{
		collectors: map[string]prometheus.Collector{
			// Standard Go runtime and process-level metrics.
			"go_collector":      collectors.NewGoCollector(),
			"process_collector": collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		},
		conf: defaultConfig,
	}
}

// Init initializes the monitoring system. It registers collectors,
// sets up the HTTP handler for scraping, and starts the background pusher.
func Init() error {
	conf := GetConfig()
	if conf.Enabled {
		for name, colletor := range metricsManager.collectors {
			for _, cname := range conf.Collectors {
				// Register only if the wildcard "*" is used or the name matches the config list.
				if conf.allCollectors || cname == name {
					registry.MustRegister(colletor)
					break
				}
			}
		}
		// Adds a default "/metrics" endpoint to the REST server.
		handlers.AddServerDefaultHandler("metrics", handlers.NewMetricsAPI(registry), rest.Protocol)
	}
	metricsManager.conf = conf
	// Start the background push service if configured.
	go metricsManager.push()
	return nil
}

// Registry returns the underlying prometheus registry.
func Registry() *prometheus.Registry {
	return registry
}

// register adds a new collector to the manager and registry if it's enabled in the config.
func (m *MetricsManager) register(name string, collector prometheus.Collector) prometheus.Collector {
	if !m.conf.Enabled {
		return nil
	}
	// Check if this specific collector is allowed by the configuration.
	exist := false
	for _, col := range m.conf.Collectors {
		if col == name {
			exist = true
			break
		}
	}
	if !exist {
		return nil
	}

	m.cm.RLock()
	col, ok := m.collectors[name]
	m.cm.RUnlock()
	if ok {
		return col // Already registered
	}

	registry.MustRegister(collector)
	m.cm.Lock()
	m.collectors[name] = collector
	m.cm.Unlock()
	return collector
}

// push periodically sends all registered metrics to a Prometheus PushGateway.
func (m *MetricsManager) push() {
	if !m.conf.Enabled || m.conf.PushGateway.Endpoint == "" {
		return
	}
	app := runtime.GetAPP()
	instanceId := utils.LocalIPv4() // Fallback to IP if Instance ID isn't set.
	if instanceId == "" {
		instanceId = app.Instance.ID
	}
	for {
		select {
		case <-time.After(m.conf.PushGateway.Interval.Duration):
			pusher := push.New(m.conf.PushGateway.Endpoint, app.App)
			m.cm.RLock()
			for _, collector := range m.collectors {
				pusher.Collector(collector)
			}
			m.cm.RUnlock()

			// Group metrics with metadata for better observability in Prometheus/Grafana.
			if err := pusher.Grouping("instance", instanceId).
				Grouping("region", app.Region).
				Grouping("az", app.AZ).
				Grouping("app", app.App).
				Grouping("env", app.Environment).
				Grouping("service_version", app.Instance.Version).
				Grouping("service", app.Instance.Name).
				Grouping("group", app.Instance.Group).
				Push(); err != nil {
				logger.Error("push metrics fail", "endpoint", m.conf.PushGateway.Endpoint, "err", err)
			}
		}
	}
}

// Global Registration Wrappers:
// These provide a clean API for developers to create new metrics.

func RegisterCollector(name string, collector prometheus.Collector) prometheus.Collector {
	return metricsManager.register(name, collector)
}

func RegisterCounter(name, help string, labelNames []string) *prometheus.CounterVec {
	if counter := metricsManager.register(name, prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name, Help: help,
	}, labelNames)); counter != nil {
		return counter.(*prometheus.CounterVec)
	}
	return nil
}

func RegisterGauge(name, help string, labelNames []string) *prometheus.GaugeVec {
	if gauge := metricsManager.register(name, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name, Help: help,
	}, labelNames)); gauge != nil {
		return gauge.(*prometheus.GaugeVec)
	}
	return nil
}

func RegisterHistogram(name, help string, labelNames []string, buckets []float64) *prometheus.HistogramVec {
	if histogram := metricsManager.register(name, prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: name, Help: help, Buckets: buckets,
	}, labelNames)); histogram != nil {
		return histogram.(*prometheus.HistogramVec)
	}
	return nil
}

func RegisterSummaryVec(name, help string, labelNames []string, objectives map[float64]float64) *prometheus.SummaryVec {
	if summary := metricsManager.register(name, prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: name, Help: help, Objectives: objectives,
	}, labelNames)); summary != nil {
		return summary.(*prometheus.SummaryVec)
	}
	return nil
}
