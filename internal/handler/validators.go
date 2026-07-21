package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

const (
	contentType     = "Content-Type"
	textPlain       = "text/plain"
	applicationJSON = "application/json"
)

var (
	ErrInvalidMetricsValue    = errors.New("Invalid metric value")
	ErrInvalidRequestMethod   = errors.New("Invalid request method")
	ErrUnsupportedContentType = errors.New("Unsupported content type")
)

func validateMetricsType(req *http.Request) error {
	metricsType := chi.URLParam(req, "metricsType")
	if metricsType != models.Counter && metricsType != models.Gauge {
		return models.ErrUnknownMetricsType
	}

	return nil
}

func validateMetricsValue(req *http.Request) error {
	metricsType := chi.URLParam(req, "metricsType")
	metricsValue := chi.URLParam(req, "metricsValue")

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

func validateReqPath(req *http.Request) error {
	if err := validateMetricsType(req); err != nil {
		return err
	}

	if err := validateMetricsValue(req); err != nil {
		return err
	}

	return nil
}

func validateReqContentType(req *http.Request, ctype string) error {
	ct := req.Header.Get(contentType)

	if ct == "" {
		return nil
	}

	if !strings.HasPrefix(ct, ctype) {
		return ErrUnsupportedContentType
	}

	return nil
}
