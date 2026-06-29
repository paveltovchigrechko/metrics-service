package handler

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

const (
	contentType = "Content-Type"
	textPlain   = "text/plain"
)

var (
	ErrInvalidMetricsType     = errors.New("Invalid metrics type")
	ErrInvalidMetricsValue    = errors.New("Invalid metric value")
	ErrInvalidRequestMethod   = errors.New("Invalid request method")
	ErrMissingContentType     = errors.New("Missing Content-Type header")
	ErrUnsupportedContentType = errors.New("Unsupported content type")
)

func validateReqPath(req *http.Request) error {
	metricsType := chi.URLParam(req, "metricsType")
	metricsValue := chi.URLParam(req, "metricsValue")

	if metricsType != models.Counter && metricsType != models.Gauge {
		return ErrInvalidMetricsType
	}

	if metricsType == models.Counter { // Expect int64 value for this type.
		_, err := strconv.ParseInt(metricsValue, 10, 64)
		if err != nil {
			return ErrInvalidMetricsValue
		}
	}

	if metricsType == models.Gauge { // Expect float64 value for this type.
		_, err := strconv.ParseFloat(metricsValue, 64)
		if err != nil {
			return ErrInvalidMetricsValue
		}
	}

	return nil
}

func validateReqContentType(req *http.Request) error {
	contentTypes := req.Header.Values(contentType)

	if len(contentTypes) == 0 {
		return ErrMissingContentType
	}

	if !slices.Contains(contentTypes, textPlain) { // Do we need other types here?
		return ErrUnsupportedContentType
	}

	return nil
}
