package metrics

import (
	"testing"

	"github.com/asjard/asjard/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func resetMetricsForTest(t *testing.T, conf Config) *MetricsManager {
	t.Helper()
	oldRegistry, oldManager := registry, metricsManager
	registry = prometheus.NewRegistry()
	m := &MetricsManager{conf: conf, collectors: make(map[string]prometheus.Collector)}
	metricsManager = m
	t.Cleanup(func() { registry, metricsManager = oldRegistry, oldManager })
	return m
}

func TestConfigComplete(t *testing.T) {
	conf := Config{
		BuiltInCollectors: utils.JSONStrings{"one", "two"},
		Collectors:        utils.JSONStrings{"two", "three"},
	}.complete()
	require.ElementsMatch(t, utils.JSONStrings{"one", "two", "three"}, conf.Collectors)
}

func TestRegisterCollectors(t *testing.T) {
	m := resetMetricsForTest(t, Config{Enabled: true, Collectors: utils.JSONStrings{"counter", "gauge", "histogram", "summary"}})
	counter := RegisterCounter("counter", "help", []string{"label"})
	require.NotNil(t, counter)
	counter.WithLabelValues("value").Inc()
	require.Same(t, counter, m.register("counter", prometheus.NewCounter(prometheus.CounterOpts{Name: "ignored"})))
	gauge := RegisterGauge("gauge", "help", nil)
	require.NotNil(t, gauge)
	gauge.WithLabelValues().Set(2)
	histogram := RegisterHistogram("histogram", "help", nil, []float64{1})
	require.NotNil(t, histogram)
	histogram.WithLabelValues().Observe(0.5)
	summary := RegisterSummaryVec("summary", "help", nil, map[float64]float64{0.5: 0.05})
	require.NotNil(t, summary)
	summary.WithLabelValues().Observe(1)
	families, err := registry.Gather()
	require.NoError(t, err)
	require.Len(t, families, 4)
}

func TestRegisterDisabledOrUnlisted(t *testing.T) {
	resetMetricsForTest(t, Config{})
	require.Nil(t, RegisterCounter("disabled", "help", nil))
	resetMetricsForTest(t, Config{Enabled: true})
	require.Nil(t, RegisterGauge("unlisted", "help", nil))
}
