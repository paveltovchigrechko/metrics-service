package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestValidateReqPath(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		urlParams   map[string]string
		expectedErr error
	}{
		{
			name: "positive test: valid gauge",
			url:  "/update/gauge/metric/1.23",
			urlParams: map[string]string{
				"metricsType":  "gauge",
				"metricsValue": "1.23",
			},
			expectedErr: nil,
		},
		{
			name: "positive test: valid counter",
			url:  "/update/counter/metric/42",
			urlParams: map[string]string{
				"metricsType":  "counter",
				"metricsValue": "42",
			},
			expectedErr: nil,
		},
		{
			name: "negative test: invalid metrics type",
			url:  "/update/histogram/metric/100",
			urlParams: map[string]string{
				"metricsType":  "histogram",
				"metricsValue": "100",
			},
			expectedErr: ErrInvalidMetricsType,
		},
		{
			name: "negative test: invalid gauge float",
			url:  "/update/gauge/metric/not-a-float",
			urlParams: map[string]string{
				"metricsType":  "gauge",
				"metricsValue": "not-a-float",
			},
			expectedErr: ErrInvalidMetricsValue,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, test.url, nil)

			chiCtx := chi.NewRouteContext()
			for key, val := range test.urlParams {
				chiCtx.URLParams.Add(key, val)
			}

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx)
			req = req.WithContext(ctx)

			err := validateReqPath(req)

			if test.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, test.expectedErr)
			}
		})
	}
}

// func TestValidateReqContentType(t *testing.T) {
// 	testCases := []struct {
// 		name        string
// 		method      string
// 		contentType string
// 		hasHeader   bool
// 		expectedErr error
// 	}{
// 		{
// 			name:        "positive test: POST method with text/plain content type",
// 			method:      http.MethodPost,
// 			contentType: "text/plain",
// 			hasHeader:   true,
// 			expectedErr: nil,
// 		},
// 		{
// 			name:        "positive test: GET method with text/plain content type",
// 			method:      http.MethodGet,
// 			contentType: "text/plain",
// 			hasHeader:   true,
// 			expectedErr: nil,
// 		},
// 		{
// 			name:        "positive test: missing content type header is allowed",
// 			method:      http.MethodPost,
// 			contentType: "",
// 			hasHeader:   false,
// 			expectedErr: nil,
// 		},
// 		{
// 			name:        "positive test: extended text/plain content type is now valid via prefix match",
// 			method:      http.MethodPost,
// 			contentType: "text/plain; charset=utf-8",
// 			hasHeader:   true,
// 			expectedErr: nil,
// 		},
// 		{
// 			name:        "negative test: POST method with application/json content type",
// 			method:      http.MethodPost,
// 			contentType: "application/json",
// 			hasHeader:   true,
// 			expectedErr: ErrUnsupportedContentType,
// 		},
// 	}

// 	for _, test := range testCases {
// 		t.Run(test.name, func(t *testing.T) {
// 			req := httptest.NewRequest(test.method, "/", nil)

// 			if test.hasHeader {
// 				req.Header.Set("Content-Type", test.contentType)
// 			}

// 			err := validateReqContentType(req)

// 			if test.expectedErr == nil {
// 				assert.NoError(t, err)
// 			} else {
// 				assert.ErrorIs(t, err, test.expectedErr)
// 			}
// 		})
// 	}
// }
