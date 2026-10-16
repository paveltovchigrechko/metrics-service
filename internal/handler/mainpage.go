package handler

import (
	"strconv"
	"sync"

	"github.com/paveltovchigrechko/metrics-service/internal/model"
)

const (
	mainPageHead = `<!DOCTYPE html>
<html>
<head>
    <title>Metrics Service</title>
</head>
<body>
    <h2>Welcome to the Metrics service</h2>
`
	mainPageColumns = `    <p>ID&nbsp;&nbsp;&nbsp;&nbsp;Value</p>
`
	mainPageEmpty = `    <p>Currently there are no metrics to display</p>
`
	mainPageTail = `</body>
</html>
`

	rowStart     = "    <p>"
	rowSeparator = "&nbsp;&nbsp;&nbsp;&nbsp;"
	rowEnd       = "</p>\n"

	mainPageRowSize  = 64
	mainPageOverhead = 512

	maxPooledPageSize = 4 << 20
)

var pagePool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 64<<10)
		return &buf
	},
}

func appendMainPage(dst []byte, metrics []model.Metrics) []byte {
	dst = append(dst, mainPageHead...)

	if len(metrics) == 0 {
		dst = append(dst, mainPageEmpty...)
	} else {
		dst = append(dst, mainPageColumns...)
		for i := range metrics {
			dst = append(dst, rowStart...)
			dst = appendEscaped(dst, metrics[i].ID)
			dst = append(dst, rowSeparator...)
			dst = appendValue(dst, &metrics[i])
			dst = append(dst, rowEnd...)
		}
	}

	return append(dst, mainPageTail...)
}

func appendValue(dst []byte, m *model.Metrics) []byte {
	switch m.MType {
	case model.Counter:
		var delta int64
		if m.Delta != nil {
			delta = *m.Delta
		}
		return strconv.AppendInt(dst, delta, 10)
	case model.Gauge:
		var value float64
		if m.Value != nil {
			value = *m.Value
		}
		return strconv.AppendFloat(dst, value, 'f', 2, 64)
	}

	return dst
}

func appendEscaped(dst []byte, s string) []byte {
	start := 0
	for i := 0; i < len(s); i++ {
		var escaped string
		switch s[i] {
		case '<':
			escaped = "&lt;"
		case '>':
			escaped = "&gt;"
		case '&':
			escaped = "&amp;"
		case '"':
			escaped = "&#34;"
		case '\'':
			escaped = "&#39;"
		default:
			continue
		}

		dst = append(dst, s[start:i]...)
		dst = append(dst, escaped...)
		start = i + 1
	}

	return append(dst, s[start:]...)
}
