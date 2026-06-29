package handler

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

const (
	endpoint    = "update"
	contentType = "Content-Type"
	textPlain   = "text/plain"
)

var (
	ErrInvalidEndpoint        = errors.New("Invalid endpoint")
	ErrInvalidMetricsType     = errors.New("Invalid metrics type")
	ErrInvalidMetricsValue    = errors.New("Invalid metric value")
	ErrInvalidPath            = errors.New("Invalid path")
	ErrInvalidRequestMethod   = errors.New("Invalid request method")
	ErrMissingContentType     = errors.New("Missing Content-Type header")
	ErrMissingMetricsName     = errors.New("No metrics name")
	ErrUnsupportedContentType = errors.New("Unsupported content type")
)

func validateReqPath(req *http.Request) error {
	path := req.URL.RequestURI()

	pathParts := strings.Split(path, "/")
	// Expect /update/<metric_type>/<metric_name>/<metric_value> path in the request.
	if len(pathParts) != 5 {
		return ErrInvalidPath
	}

	if strings.Compare(pathParts[1], endpoint) != 0 {
		return ErrInvalidEndpoint
	}

	metricsType := pathParts[2]

	if strings.Compare(metricsType, models.Counter) != 0 && strings.Compare(metricsType, models.Gauge) != 0 {
		return ErrInvalidMetricsType
	}

	if strings.Compare(metricsType, models.Counter) == 0 { // Expect int64 value for this type.
		_, err := strconv.ParseInt(pathParts[4], 10, 64)
		if err != nil {
			return ErrInvalidMetricsValue
		}
	}

	if strings.Compare(metricsType, models.Gauge) == 0 { // Expect float64 value for this type.
		_, err := strconv.ParseFloat(pathParts[4], 64)
		if err != nil {
			return ErrInvalidMetricsValue
		}
	}

	metricName := pathParts[3]
	if len(metricName) == 0 { // Expect to have metric name.
		return ErrMissingMetricsName
	}

	return nil
}

func validateReqHeader(req *http.Request) error {
	for _, reqValidator := range reqHeaderValidators {
		if err := reqValidator(req); err != nil {
			return err
		}
	}

	return nil
}

type requestValidator func(*http.Request) error

var reqHeaderValidators = []requestValidator{
	validateReqMethod,
	validateReqContentType,
}

func validateReqMethod(req *http.Request) error {
	if req.Method != http.MethodPost {
		return ErrInvalidRequestMethod
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
