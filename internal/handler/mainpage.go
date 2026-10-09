package handler

import (
	"html/template"
	"strconv"

	"github.com/paveltovchigrechko/metrics-service/internal/model"
)

// The template is parsed once at startup. It uses html/template, so metric names are HTML-escaped.
const mainPageTpl = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics Service</title>
</head>
<body>
    <h2>Welcome to the Metrics service</h2>
    {{if .}}
        <p>ID&nbsp;&nbsp;&nbsp;&nbsp;Value</p>
        {{- range .}}
        <p>{{.ID}}&nbsp;&nbsp;&nbsp;&nbsp;{{.Value}}</p>
        {{- end}}
    {{else}}
        <p>Currently there are no metrics to display</p>
    {{end}}
</body>
</html>
`

var mainPageTmpl = template.Must(template.New("mainpage").Parse(mainPageTpl))

// metricRow is a metric prepared for the template: the value is already formatted,
// so the template needs no function calls (which are slow, as they use reflection).
type metricRow struct {
	ID    string
	Value string
}

func newMetricRows(metrics []model.Metrics) []metricRow {
	rows := make([]metricRow, len(metrics))
	for i := range metrics {
		rows[i] = metricRow{
			ID:    metrics[i].ID,
			Value: formatValue(&metrics[i]),
		}
	}

	return rows
}

// formatValue formats the counter delta as an integer and the gauge value with two decimals.
// A missing delta or value is shown as zero.
func formatValue(m *model.Metrics) string {
	switch m.MType {
	case model.Counter:
		var delta int64
		if m.Delta != nil {
			delta = *m.Delta
		}
		return strconv.FormatInt(delta, 10)
	case model.Gauge:
		var value float64
		if m.Value != nil {
			value = *m.Value
		}
		return strconv.FormatFloat(value, 'f', 2, 64)
	}

	return ""
}
