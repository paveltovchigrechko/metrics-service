package handler

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	models "github.com/paveltovchigrechko/metrics-service/internal/model"
)

func validateReqPath(req *http.Request) error {
	path := req.URL.RequestURI()

	pathParts := strings.Split(path, "/")
	// Expect /update/<metric_type>/<metric_name>/<metric_value> path in the request.
	if len(pathParts) != 5 {
		return errors.New("Invalid path: update/<metric_type>/<metric_name>/<metric_value> is expected")
	}

	if strings.Compare(pathParts[1], "update") != 0 {
		return errors.New("Invalid route: expected `update`")
	}

	metricType := pathParts[2]

	if strings.Compare(metricType, models.Counter) != 0 && strings.Compare(metricType, models.Gauge) != 0 {
		return errors.New("Invalid metric name")
	}

	if strings.Compare(metricType, models.Counter) == 0 { // Expect int64 value for this type.
		_, err := strconv.ParseInt(pathParts[4], 10, 64)
		if err != nil {
			return errors.New("Invalid metric value")
		}
	}

	if strings.Compare(metricType, models.Gauge) == 0 { // Expect float64 value for this type.
		_, err := strconv.ParseFloat(pathParts[4], 64)
		if err != nil {
			return errors.New("Invalid metric value")
		}
	}

	metricName := pathParts[3]
	if len(metricName) == 0 { // Expect to have metric name.
		return errors.New("No metric")
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
		return errors.New("Only POST requests are allowed")
	}

	return nil
}

func validateReqContentType(req *http.Request) error {
	contentTypes := req.Header.Values("Content-Type")

	if len(contentTypes) == 0 {
		return errors.New("Missing Content-Type header")
	}

	if !slices.Contains(contentTypes, "text/plain") { // Do we need other types here?
		return errors.New("Unsupported content type, only text/plain is supported")
	}

	return nil
}
