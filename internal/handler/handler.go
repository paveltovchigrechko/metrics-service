package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

type AppHandler struct {
	storage *models.MemStorage
}

func NewHandler(s *models.MemStorage) *AppHandler {
	h := &AppHandler{
		storage: s,
	}

	return h
}

func (h *AppHandler) MainPage(w http.ResponseWriter, req *http.Request) {
	if err := validateReqHeader(req); err != nil {
		w.Write([]byte(err.Error())) // Delete
		return
	}

	if err := validateReqPath(req); err != nil {
		switch err.Error() {
		case "No metric", "Invalid path":
			w.WriteHeader(http.StatusNotFound)
		case "Invalid metric value", "Invalid metric type", "Invalid endpoint", "No metric name":
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write([]byte(err.Error())) // Delete
		return
	}

	w.WriteHeader(http.StatusOK)

	h.processMetric(req.URL)
}

func (h *AppHandler) processMetric(url *url.URL) {
	// Parse metric and its value. We know the metric has a name, a correct type, and value.
	metric, metricName := parseMetricAndName(url)

	// Check if the metric is new or already exists.
	if _, ok := h.storage.Metrics[metricName]; !ok {
		// Add new metric.
		h.storage.Metrics[metricName] = metric
	} else {
		// Update existing metric.
		h.storage.UpdateMetric(metric, metricName)
	}
}

func parseMetricAndName(url *url.URL) (*models.Metrics, string) {
	path := url.RequestURI()

	pathParts := strings.Split(path, "/")
	metricType, metricName, metricValue := pathParts[2], pathParts[3], pathParts[4]
	m := models.Metrics{
		MType: metricType,
	}

	// TODO: Remove duplicate logic with validateReqPath().
	if metricType == models.Counter {
		value, _ := strconv.ParseInt(metricValue, 10, 64) // Process error.
		m.Delta = &value
	} else {
		value, _ := strconv.ParseFloat(metricValue, 64) // Process error.
		m.Value = &value
	}

	return &m, metricName
}
