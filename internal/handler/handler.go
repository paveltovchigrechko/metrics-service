package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/logger"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/paveltovchigrechko/metrics-service/internal/repository"
)

const databasePingTimeout = 3 * time.Second

var (
	errNoDatabase          = errors.New("no database configured")
	errDatabaseUnreachable = errors.New("cannot ping database")
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

type AppHandler struct {
	storage repository.Storage
	pinger  Pinger

	// A callback for synchronous writing metrics to a FileStorage.
	// Should be used by server.Server if config.ServerConfig.StoreInterval == 0.
	// For other cases, the function value must be nil.
	afterSuccessfulUpdate func() error
}

func NewHandler(s repository.Storage, p Pinger, updateFunc func() error) *AppHandler {
	h := &AppHandler{
		storage:               s,
		pinger:                p,
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
		case ErrInvalidMetricsValue, models.ErrUnknownMetricsType:
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

	t, err := template.New("mainpage").Funcs(templateFuncs).Parse(mainPageTpl)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	metrics, err := h.storage.GetAllMetrics(req.Context())
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, metrics); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}
}

func (h *AppHandler) MetricsValue(w http.ResponseWriter, req *http.Request) {
	err := validateMetricsType(req)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	metricsName := chi.URLParam(req, "metricsName")
	metricsType := chi.URLParam(req, "metricsType")
	metrics, err := h.storage.GetMetrics(req.Context(), metricsName, metricsType)
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

	// No need for parsing, just validate here.
	err = model.ValidateMetrics(parsedMetrics)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	err = h.storage.SaveMetrics(req.Context(), parsedMetrics)
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

func (h *AppHandler) UpdatesEndpoint(w http.ResponseWriter, req *http.Request) {
	if err := validateReqContentType(req, applicationJSON); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	parsedMetrics, err := decodeJSONBatch(req)
	if err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	if len(parsedMetrics) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.storage.SaveBatch(req.Context(), parsedMetrics)
	if err != nil {
		if errors.Is(err, models.ErrDeltaIsNil) ||
			errors.Is(err, models.ErrEmptyMetricsID) ||
			errors.Is(err, models.ErrUnknownMetricsType) ||
			errors.Is(err, models.ErrValueIsNil) ||
			errors.Is(err, models.ErrDeltaAndValuePresent) {
			WriteError(w, err, http.StatusBadRequest)
			return
		}
		WriteError(w, err, http.StatusInternalServerError)
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

	if err = model.ValidateName(decodedMetrics.ID); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	if err := model.ValidateType(decodedMetrics.MType); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	m, err := h.storage.GetMetrics(req.Context(), decodedMetrics.ID, decodedMetrics.MType)
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}

	if decodedMetrics.MType != m.MType {
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

func (h *AppHandler) PingEndpoint(w http.ResponseWriter, req *http.Request) {
	if h.pinger == nil {
		WriteError(w, errNoDatabase, http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), databasePingTimeout)
	defer cancel()

	if err := h.pinger.PingContext(ctx); err != nil {
		WriteError(w, errDatabaseUnreachable, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AppHandler) processMetrics(req *http.Request) error {
	// Parse metrics and its value. We know the metrics has a name, a correct type, and value.
	metrics, err := parseMetrics(req)
	if err != nil {
		return err
	}

	return h.storage.SaveMetrics(req.Context(), metrics)
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

func decodeJSONMetrics(req *http.Request) (*models.Metrics, error) {
	var jsonMetrics models.Metrics
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	// Reading request body and unmarhal it
	if err := decoder.Decode(&jsonMetrics); err != nil {
		return nil, err
	}

	return &jsonMetrics, nil
}

func decodeJSONBatch(req *http.Request) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&metrics); err != nil { // Trailing JSON issue as ([{"id":"A"}] {"unexpected":"second document"}): how should we treat it?
		return nil, err
	}

	return metrics, nil
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
