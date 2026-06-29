package handler

import (
	"log"
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

func (h *AppHandler) PostMetrics(w http.ResponseWriter, req *http.Request) {
	log.Printf("[DEBUG] Received request: %s", req.URL.Path) // Delete
	if err := validateReqContentType(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("[DEBUG] Header validation failed: %v", err) // Delete
		return
	}

	if err := validateReqPath(req); err != nil {
		switch err {
		case ErrInvalidMetricsValue, ErrInvalidMetricsType:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
		log.Printf("[DEBUG] Path validation failed: %v", err) // Delete
		return
	}

	w.WriteHeader(http.StatusOK)
	h.processMetrics(req.URL)
}

func (h *AppHandler) MainPage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.Write([]byte("Welcome to the Metrics service\n"))
	h.storage.ListMetrics(w)
}

func (h *AppHandler) processMetrics(url *url.URL) {
	// Parse metric and its value. We know the metric has a name, a correct type, and value.
	metric := parseMetrics(url)

	// Check if the metric is new or already exists.
	if _, ok := h.storage.Metrics[metric.ID]; !ok {
		// Add new metric.
		h.storage.Metrics[metric.ID] = metric
	} else {
		// Update existing metric.
		h.storage.UpdateMetrics(metric)
	}
}

// This function assumes the input url.Url passed validateReqPath().
func parseMetrics(url *url.URL) *models.Metrics {
	path := url.RequestURI()

	pathParts := strings.Split(path, "/")
	metricType, metricName, metricValue := pathParts[2], pathParts[3], pathParts[4]
	m := models.Metrics{
		ID:    metricName,
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

	return &m
}
