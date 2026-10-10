// Package handler implements the HTTP endpoints of the metrics server.
//
// The handlers read URL parameters with chi, so they must be registered on a
// chi router with the parameter names {metricsType}, {metricsName} and
// {metricsValue}.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paveltovchigrechko/metrics-service/internal/audit"
	"github.com/paveltovchigrechko/metrics-service/internal/logger"
	"github.com/paveltovchigrechko/metrics-service/internal/model"
	"github.com/paveltovchigrechko/metrics-service/internal/repository"
)

const databasePingTimeout = 3 * time.Second

var (
	errNoDatabase          = errors.New("no database configured")
	errDatabaseUnreachable = errors.New("cannot ping database")
)

// Pinger checks that the database is reachable. *sql.DB satisfies it.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// AppHandler serves the metrics endpoints on top of a repository.Storage.
// Create it with NewHandler.
type AppHandler struct {
	storage repository.Storage
	pinger  Pinger

	// A callback for synchronous writing metrics to a FileStorage.
	// Should be used by server.Server if config.ServerConfig.StoreInterval == 0.
	// For other cases, the function value must be nil.
	afterSuccessfulUpdate func() error
	notifier              audit.Notifier
}

// NewHandler returns an AppHandler that stores metrics in s.
//
// The other arguments are optional and may be nil:
//   - p is used by PingEndpoint; without it the endpoint always reports an error.
//   - updateFunc is called after every successful update, for example to save
//     the storage to a file synchronously. If it fails, the update endpoint
//     responds with 500 Internal Server Error.
//   - notifier receives an audit event after every successful update.
func NewHandler(s repository.Storage, p Pinger, updateFunc func() error, notifier audit.Notifier) *AppHandler {
	h := &AppHandler{
		storage:               s,
		pinger:                p,
		afterSuccessfulUpdate: updateFunc,
		notifier:              notifier,
	}

	return h
}

// PostMetrics updates a single metric passed in the URL path:
//
//	POST /update/{metricsType}/{metricsName}/{metricsValue}
//
// The type is "counter" with an integer value or "gauge" with a float value.
// A counter value is added to the stored one, a gauge value replaces it.
// The request body is ignored; a Content-Type header, if present, must be
// text/plain.
//
// It responds with 200 OK on success and 400 Bad Request for an unknown type,
// an invalid value or a wrong Content-Type.
func (h *AppHandler) PostMetrics(w http.ResponseWriter, req *http.Request) {
	if err := validateReqContentType(req, textPlain); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	if err := validateReqPath(req); err != nil {
		switch err {
		case ErrInvalidMetricsValue, model.ErrUnknownMetricsType:
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

	if err := h.audit(req, model.Metrics{ID: chi.URLParam(req, "metricsName")}); err != nil {
		log.Printf("audit: %v", err) // Audit failure must not fail the request.
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

// MainPage renders all stored metrics as an HTML page:
//
//	GET /
//
// It responds with 200 OK, or 500 Internal Server Error if the storage fails.
func (h *AppHandler) MainPage(w http.ResponseWriter, req *http.Request) {
	metrics, err := h.storage.GetAllMetrics(req.Context())
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// Build the whole page in a reusable buffer and write it at once: small writes straight to the
	// connection are slow.
	bufPtr := pagePool.Get().(*[]byte)
	page := slices.Grow((*bufPtr)[:0], len(metrics)*mainPageRowSize+mainPageOverhead)
	page = appendMainPage(page, metrics)

	w.Header().Set(contentType, "text/html; charset=utf-8")
	w.Write(page)

	// The ResponseWriter must not keep the slice after Write returns, so the buffer can be reused.
	if cap(page) <= maxPooledPageSize {
		*bufPtr = page
		pagePool.Put(bufPtr)
	}
}

// MetricsValue writes the value of a single metric as plain text:
//
//	GET /value/{metricsType}/{metricsName}
//
// It responds with 200 OK and the value, 400 Bad Request for an unknown type,
// or 404 Not Found if there is no such metric.
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
	case model.Counter:
		w.Write([]byte(strconv.FormatInt(*metrics.Delta, 10)))
	case model.Gauge:
		w.Write([]byte(strconv.FormatFloat(*metrics.Value, 'f', -1, 64)))
	}
}

// UpdateEndpoint updates a single metric sent as a JSON model.Metrics:
//
//	POST /update
//	{"id":"PollCount","type":"counter","delta":5}
//	{"id":"Alloc","type":"gauge","value":123.45}
//
// A counter needs "delta", a gauge needs "value". A Content-Type header, if
// present, must be application/json.
//
// It responds with 200 OK on success and 400 Bad Request for malformed JSON,
// unknown fields, an invalid metric or a storage error.
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

	if err := h.audit(req, *parsedMetrics); err != nil {
		log.Printf("audit: %v", err) // Audit failure must not fail the request.
	}

	if h.afterSuccessfulUpdate != nil {
		if err := h.afterSuccessfulUpdate(); err != nil {
			WriteError(w, err, http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// UpdatesEndpoint updates several metrics sent as a JSON array of
// model.Metrics:
//
//	POST /updates
//	[{"id":"PollCount","type":"counter","delta":5},{"id":"Alloc","type":"gauge","value":123.45}]
//
// The metrics are saved as one batch. A Content-Type header, if present, must
// be application/json.
//
// It responds with 200 OK on success, including an empty array, 400 Bad Request
// for malformed JSON or an invalid metric, and 500 Internal Server Error if the
// storage fails.
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
		if errors.Is(err, model.ErrDeltaIsNil) ||
			errors.Is(err, model.ErrEmptyMetricsID) ||
			errors.Is(err, model.ErrUnknownMetricsType) ||
			errors.Is(err, model.ErrValueIsNil) ||
			errors.Is(err, model.ErrDeltaAndValuePresent) {
			WriteError(w, err, http.StatusBadRequest)
			return
		}
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	if err := h.audit(req, parsedMetrics...); err != nil {
		log.Printf("audit: %v", err) // Audit failure must not fail the request.
	}

	if h.afterSuccessfulUpdate != nil {
		if err := h.afterSuccessfulUpdate(); err != nil {
			WriteError(w, err, http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// ValueEndpoint returns a single metric as JSON. The request names the metric
// and its type:
//
//	POST /value
//	{"id":"PollCount","type":"counter"}
//
// The response is the stored model.Metrics with its current value. A
// Content-Type header, if present, must be application/json.
//
// It responds with 200 OK, 400 Bad Request for malformed JSON, an invalid name
// or type, or a type that does not match the stored metric, and 404 Not Found
// if there is no such metric.
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

// PingEndpoint checks the database connection:
//
//	GET /ping
//
// It responds with 200 OK if the database answers within 3 seconds, and 500
// Internal Server Error if it does not or if no Pinger was given to NewHandler.
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
func parseMetrics(req *http.Request) (*model.Metrics, error) {
	metricsType, metricsName, metricsValue := chi.URLParam(req, "metricsType"), chi.URLParam(req, "metricsName"), chi.URLParam(req, "metricsValue")
	// check for empty strings?
	m := model.Metrics{
		ID:    metricsName,
		MType: metricsType,
	}

	// TODO: Remove duplicate logic with validateReqPath().
	if metricsType == model.Counter {
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

func decodeJSONMetrics(req *http.Request) (*model.Metrics, error) {
	var jsonMetrics model.Metrics
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	// Reading request body and unmarhal it
	if err := decoder.Decode(&jsonMetrics); err != nil {
		return nil, err
	}

	return &jsonMetrics, nil
}

func decodeJSONBatch(req *http.Request) ([]model.Metrics, error) {
	metrics := make([]model.Metrics, 0)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&metrics); err != nil { // Trailing JSON issue as ([{"id":"A"}] {"unexpected":"second document"}): how should we treat it?
		return nil, err
	}

	return metrics, nil
}

// WriteError responds with status and the error text as a plain-text body.
// If w implements logger.ErrorRecorder, the error is also recorded there so the
// logging middleware can log it.
func WriteError(w http.ResponseWriter, err error, status int) {
	// Check if we use wrapper for the ResponseWriter. In logger.go we have loggingResponseWriter for that.
	// If so, use wrapper's method to catch the response error.
	if recorder, ok := w.(logger.ErrorRecorder); ok {
		recorder.SetError(err)
	}

	// Set response with error.
	http.Error(w, err.Error(), status)
}
