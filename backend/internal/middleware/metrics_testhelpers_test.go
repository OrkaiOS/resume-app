package middleware

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// gather collects all metrics from the default registry.
func gather() (map[string]*dto.MetricFamily, error) {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return nil, err
	}
	out := make(map[string]*dto.MetricFamily, len(families))
	for _, f := range families {
		out[f.GetName()] = f
	}
	return out, nil
}

// familyNames returns a comma-joined list of registered metric family names
// for helpful failure messages.
func familyNames(families map[string]*dto.MetricFamily) string {
	names := make([]string, 0, len(families))
	for n := range families {
		names = append(names, n)
	}
	return strings.Join(names, ", ")
}
