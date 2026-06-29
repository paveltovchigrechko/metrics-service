package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateReqPath(t *testing.T) {
	testCases := []struct {
		name string
		url  string
	}{
		{
			name: "positive test: valid path",
			url:  "/update/gauge/metric/1",
		},
	}

	for _, test := range testCases {
		req := httptest.NewRequest(http.MethodPost, test.url, nil)
		err := validateReqPath(req)
		assert.Nil(t, err)
	}
}

func TestValidateReqContentType(t *testing.T) {
	testCases := []struct {
		name        string
		method      string
		contentType string
		expectedErr error
	}{
		{
			name:        "positive test: POST method with text/plain content type",
			method:      http.MethodPost,
			contentType: textPlain,
			expectedErr: nil,
		},
		{
			name:        "positive test: GET method with text/plain content type",
			method:      http.MethodGet,
			contentType: textPlain,
			expectedErr: nil,
		},
		{
			name:        "negative test: POST method with application/json content type",
			method:      http.MethodGet,
			contentType: "application/json",
			expectedErr: ErrUnsupportedContentType,
		},
		{
			name:        "negative test: POST method with extended text/plain content type",
			method:      http.MethodGet,
			contentType: "text/plain; charset=utf-8",
			expectedErr: ErrUnsupportedContentType,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, "/", nil)
			req.Header.Add(contentType, test.contentType)

			err := validateReqContentType(req)

			if test.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			}
		})
	}
}
