package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

const (
	introText = "Welcome to the Metrics service\n"
)

type AppHandler struct {
	storage models.Storage
}

func NewHandler(s models.Storage) *AppHandler {
	h := &AppHandler{
		storage: s,
	}

	return h
}

func (h *AppHandler) PostMetrics(w http.ResponseWriter, req *http.Request) {
	// log.Printf("[DEBUG] Received request: %s", req.URL.Path) // Delete
	if err := validateReqContentType(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		// log.Printf("[DEBUG] Header validation failed: %v", err) // Delete
		return
	}

	if err := validateReqPath(req); err != nil {
		switch err {
		case ErrInvalidMetricsValue, ErrInvalidMetricsType:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
		// log.Printf("[DEBUG] Path validation failed: %v", err) // Delete
		return
	}

	err := h.processMetrics(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AppHandler) MainPage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set(contentType, "text/plain; charset=utf-8")

	w.Write([]byte(introText))
	h.storage.ListMetrics(w)
}

func (h *AppHandler) MetricsValue(w http.ResponseWriter, req *http.Request) {
	err := validateMetricsType(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricsName := chi.URLParam(req, "metricsName")
	metricsType := chi.URLParam(req, "metricsType")
	metrics, err := h.storage.GetMetrics(metricsName, metricsType)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	switch metrics.MType {
	case models.Counter:
		w.Write([]byte(strconv.FormatInt(*metrics.Delta, 10)))
	case models.Gauge:
		w.Write([]byte(strconv.FormatFloat(*metrics.Value, 'f', -1, 64)))
	}
}

func (h *AppHandler) processMetrics(req *http.Request) error {
	// Parse metric and its value. We know the metric has a name, a correct type, and value.
	metric, err := parseMetrics(req)
	if err != nil {
		return err
	}

	return h.storage.SaveMetrics(metric)
}

// This function assumes the input url.Url passed validateReqPath().
func parseMetrics(req *http.Request) (*models.Metrics, error) {
	metricType, metricName, metricValue := chi.URLParam(req, "metricsType"), chi.URLParam(req, "metricsName"), chi.URLParam(req, "metricsValue")
	metricID := metricType + ":" + metricName
	// check for empty strings?
	m := models.Metrics{
		ID:    metricID,
		MType: metricType,
	}

	// TODO: Remove duplicate logic with validateReqPath().
	if metricType == models.Counter {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return nil, err
		}
		m.Delta = &value
	} else {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return nil, err
		}
		m.Value = &value
	}

	return &m, nil
}
