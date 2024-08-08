/*
Package metrics 监控维护，根据配置过滤需要上报的指标
*/
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

type MetricsManager struct {
	conf       Config
	collectors map[string]prometheus.Collector
	cm         sync.RWMutex
}

var (
	registry       = prometheus.NewRegistry()
	metricsManager *MetricsManager
)

func init() {
	metricsManager = &MetricsManager{
		collectors: map[string]prometheus.Collector{
			"go_collector":      collectors.NewGoCollector(),
			"process_collector": collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		},
		conf: defaultConfig,
	}
}

// Init 监控初始化
func Init() error {
	conf := GetConfig()
	if conf.Enabled {
		for name, colletor := range metricsManager.collectors {
			for _, cname := range conf.Collectors {
				if conf.allCollectors || cname == name {
					registry.MustRegister(colletor)
					break
				}
			}
		}
		handlers.AddServerDefaultHandler("metrics", handlers.NewMetricsAPI(registry), rest.Protocol)
	}
	metricsManager.conf = conf
	go metricsManager.push()
	return nil
}

// 注册成功返回collector,否则返回nil
func (m *MetricsManager) register(name string, collector prometheus.Collector) prometheus.Collector {
	if !m.conf.Enabled {
		return nil
	}
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
		registry.Unregister(col)
	}
	registry.MustRegister(collector)
	m.cm.Lock()
	m.collectors[name] = collector
	m.cm.Unlock()
	return collector
}

func (m *MetricsManager) push() {
	if !m.conf.Enabled || m.conf.PushGateway.Endpoint == "" {
		return
	}
	app := runtime.GetAPP()
	instanceId := utils.LocalIPv4()
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
			// TODO 此处instance是个无法辨认的字符串, 重启后会更新
			// 是否可以生成一个可辨认的字符串
			if err := pusher.Grouping("instance", instanceId).
				Grouping("region", app.Region).
				Grouping("az", app.AZ).
				Grouping("app", app.App).
				Grouping("env", app.Environment).
				Grouping("service_version", app.Instance.Version).
				Grouping("service", app.Instance.Name).
				Push(); err != nil {
				logger.Error("push metrics fail", "endpoint", m.conf.PushGateway.Endpoint, "err", err)
			}
		}
	}
}

func RegisterCollector(name string, collector prometheus.Collector) prometheus.Collector {
	return metricsManager.register(name, collector)
}

// RegisterCounter 注册一个新的counter指标，如果注册成功则返回true，否则返回false
// 如果没有开启监控或者收集指标不在配置范围内则返回true
func RegisterCounter(name, help string, labelNames []string) *prometheus.CounterVec {
	if counter := metricsManager.register(name, prometheus.NewCounterVec(prometheus.CounterOpts{
		// Namespace: runtime.APP,
		// Subsystem: runtime.Name,
		Name: name,
		Help: help,
	}, labelNames)); counter != nil {
		return counter.(*prometheus.CounterVec)
	}
	return nil
}

func RegisterGauge(name, help string, labelNames []string) *prometheus.GaugeVec {
	if gauge := metricsManager.register(name, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		// Namespace: runtime.APP,
		Name: name,
		Help: help,
	}, labelNames)); gauge != nil {
		return gauge.(*prometheus.GaugeVec)
	}
	return nil
}

func RegisterHistogram(name, help string, labelNames []string, buckets []float64) *prometheus.HistogramVec {
	if histogram := metricsManager.register(name, prometheus.NewHistogramVec(prometheus.HistogramOpts{
		// Namespace: runtime.APP,
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labelNames)); histogram != nil {
		return histogram.(*prometheus.HistogramVec)
	}
	return nil
}

func RegisterSummaryVec(name, help string, labelNames []string, objectives map[float64]float64) *prometheus.SummaryVec {
	if summary := metricsManager.register(name, prometheus.NewSummaryVec(prometheus.SummaryOpts{
		// Namespace:  runtime.APP,
		Name:       name,
		Help:       help,
		Objectives: objectives,
	}, labelNames)); summary != nil {
		return summary.(*prometheus.SummaryVec)
	}
	return nil
}
