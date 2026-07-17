package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/common"
	"github.com/paveltovchigrechko/metrics-service/internal/logger"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

const (
	introText = "<h2>Welcome to the Metrics service</h2>"
)

type AppHandler struct {
	storage models.Storage

	// A callback for synchronous writing metrics to a FileStorage.
	// Should be used by server.Server if config.ServerConfig.StoreInterval == 0.
	// For other cases, the function value must be nil.
	afterSuccessfulUpdate func() error
}

func NewHandler(s models.Storage, updateFunc func() error) *AppHandler {
	h := &AppHandler{
		storage:               s,
		afterSuccessfulUpdate: updateFunc,
	}

	return h
}

func (h *AppHandler) PostMetrics(w http.ResponseWriter, req *http.Request) {
	if err := validateReqContentType(req, textPlain); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	if err := validateReqPath(req); err != nil {
		switch err {
		case ErrInvalidMetricsValue, ErrInvalidMetricsType:
			WriteError(w, err, http.StatusBadRequest)
		default:
			WriteError(w, err, http.StatusBadRequest)
		}
		return
	}

	err := h.processMetrics(req)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	if h.afterSuccessfulUpdate != nil {
		// The think about: if we proccessed metrics earlier, but fail on writing to the file, the client receives Internal Server Error.
		// Should we handle this case with 200 response?
		if err := h.afterSuccessfulUpdate(); err != nil {
			WriteError(w, err, http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AppHandler) MainPage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set(contentType, "text/html; charset=utf-8")

	w.Write([]byte(introText))
	h.storage.ListMetrics(w)
}

func (h *AppHandler) MetricsValue(w http.ResponseWriter, req *http.Request) {
	err := validateMetricsType(req)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
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

func (h *AppHandler) UpdateEndpoint(w http.ResponseWriter, req *http.Request) {
	if err := validateReqContentType(req, applicationJSON); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	parsedMetrics, err := decodeJSONMetrics(req)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	metrics, err := parseUpdateMetrics(parsedMetrics)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	err = h.storage.SaveMetrics(metrics)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	if h.afterSuccessfulUpdate != nil {
		if err := h.afterSuccessfulUpdate(); err != nil {
			WriteError(w, err, http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AppHandler) ValueEndpoint(w http.ResponseWriter, req *http.Request) {
	if err := validateReqContentType(req, applicationJSON); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	decodedMetrics, err := decodeJSONMetrics(req)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	name, mType, err := parseValueMetrics(decodedMetrics)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	m, err := h.storage.GetMetrics(name, mType)
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}

	if mType != m.MType {
		WriteError(w, errors.New("metrics type does not match"), http.StatusBadRequest)
		return
	}

	encodedMetrics, err := json.Marshal(m)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(encodedMetrics); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
	}
}

func (h *AppHandler) processMetrics(req *http.Request) error {
	// Parse metrics and its value. We know the metrics has a name, a correct type, and value.
	metrics, err := parseMetrics(req)
	if err != nil {
		return err
	}

	return h.storage.SaveMetrics(metrics)
}

// This function assumes the input url.Url passed validateReqPath().
func parseMetrics(req *http.Request) (*models.Metrics, error) {
	metricsType, metricsName, metricsValue := chi.URLParam(req, "metricsType"), chi.URLParam(req, "metricsName"), chi.URLParam(req, "metricsValue")
	// check for empty strings?
	m := models.Metrics{
		ID:    metricsName,
		MType: metricsType,
	}

	// TODO: Remove duplicate logic with validateReqPath().
	if metricsType == models.Counter {
		value, err := strconv.ParseInt(metricsValue, 10, 64)
		if err != nil {
			return nil, err
		}
		m.Delta = &value
	} else {
		value, err := strconv.ParseFloat(metricsValue, 64)
		if err != nil {
			return nil, err
		}
		m.Value = &value
	}

	return &m, nil
}

func parseUpdateMetrics(jsonMetrics *common.Metrics) (*models.Metrics, error) {
	if jsonMetrics.ID == "" {
		return nil, models.ErrEmptyMetricsID
	}

	var m *models.Metrics
	var err error

	switch jsonMetrics.MType {
	case models.Counter:
		if jsonMetrics.Delta == nil {
			return nil, fmt.Errorf("missing delta for counter metric %s", jsonMetrics.ID)
		}
		m, err = models.CreateMetrics(jsonMetrics.ID, jsonMetrics.MType, *jsonMetrics.Delta, 0)
	case models.Gauge:
		if jsonMetrics.Value == nil {
			return nil, fmt.Errorf("missing value for gauge metric %s", jsonMetrics.ID)
		}
		m, err = models.CreateMetrics(jsonMetrics.ID, jsonMetrics.MType, 0, *jsonMetrics.Value)
	default:
		return nil, models.ErrUnknownMetricsType
	}

	if err != nil {
		return nil, err
	}
	return m, nil
}

func parseValueMetrics(jsonMetrics *common.Metrics) (string, string, error) {
	if jsonMetrics.ID == "" {
		return "", "", models.ErrEmptyMetricsID
	}

	if jsonMetrics.MType != models.Counter && jsonMetrics.MType != models.Gauge {
		return "", "", models.ErrUnknownMetricsType
	}

	return jsonMetrics.ID, jsonMetrics.MType, nil
}

func decodeJSONMetrics(req *http.Request) (*common.Metrics, error) {
	var jsonMetrics common.Metrics
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	// Reading request body and unmarhal it
	if err := decoder.Decode(&jsonMetrics); err != nil {
		return nil, err
	}

	return &jsonMetrics, nil
}

func WriteError(w http.ResponseWriter, err error, status int) {
	// Check if we use wrapper for the ResponseWriter. In logger.go we have loggingResponseWriter for that.
	// If so, use wrapper's method to catch the response error.
	if recorder, ok := w.(logger.ErrorRecorder); ok {
		recorder.SetError(err)
	}

	// Set response with error.
	http.Error(w, err.Error(), status)
}
