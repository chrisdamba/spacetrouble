package health_test

import (
	"encoding/json"
	"github.com/chrisdamba/spacetrouble/pkg/health"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthGet(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		expectedCode  int
		checkResponse bool
	}{
		{
			name:          "Success GET request",
			method:        http.MethodGet,
			expectedCode:  http.StatusOK,
			checkResponse: true,
		},
		{
			name:          "Invalid POST request",
			method:        http.MethodPost,
			expectedCode:  http.StatusMethodNotAllowed,
			checkResponse: false,
		},
		{
			name:          "Invalid PUT request",
			method:        http.MethodPut,
			expectedCode:  http.StatusMethodNotAllowed,
			checkResponse: false,
		},
		{
			name:          "Invalid DELETE request",
			method:        http.MethodDelete,
			expectedCode:  http.StatusMethodNotAllowed,
			checkResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)

			rr := httptest.NewRecorder()

			handler := health.HealthGet()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.checkResponse {
				var response health.HealthResponse
				err := json.NewDecoder(rr.Body).Decode(&response)

				assert.NoError(t, err)
				assert.Equal(t, "healthy", response.Status)
				assert.NotEmpty(t, response.Timestamp)
				assert.NotEmpty(t, response.Version)
				assert.NotEmpty(t, response.Uptime)
				assert.NotEmpty(t, response.GoVersion)

				timestamp, err := time.Parse(time.RFC3339, response.Timestamp)
				assert.NoError(t, err)

				assert.True(t, time.Since(timestamp) < time.Minute)

				assert.Greater(t, response.Memory.Alloc, uint64(0))
				assert.Greater(t, response.Memory.TotalAlloc, uint64(0))
				assert.Greater(t, response.Memory.Sys, uint64(0))
				assert.GreaterOrEqual(t, response.Memory.NumGC, uint32(0))

				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			}
		})
	}
}
