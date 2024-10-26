package utils_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/chrisdamba/spacetrouble/internal/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResponse struct {
	Name  string `json:"name" xml:"name"`
	Value int    `json:"value" xml:"value"`
}

type TestXMLResponse struct {
	Name      string    `xml:"name"`
	Age       int       `xml:"age"`
	Timestamp time.Time `xml:"timestamp"`
}

type ComplexXMLResponse struct {
	ID      int               `xml:"id"`
	Items   []TestXMLResponse `xml:"items>item"`
	Details struct {
		Description string `xml:"description"`
		Status      bool   `xml:"status"`
	} `xml:"details"`
}

func TestJsonDecodeBody(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		want        testResponse
		wantErr     bool
		errContains string
	}{
		{
			name: "valid json",
			input: map[string]interface{}{
				"name":  "test",
				"value": 123,
			},
			want: testResponse{
				Name:  "test",
				Value: 123,
			},
			wantErr: false,
		},
		{
			name:        "invalid json",
			input:       "{invalid json}",
			wantErr:     true,
			errContains: "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			var body []byte
			var err error
			if str, ok := tt.input.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.input)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
			var result testResponse
			err = utils.JsonDecodeBody(req, &result)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRenderResponse(t *testing.T) {
	tests := []struct {
		name         string
		acceptHeader string
		statusCode   int
		response     interface{}
		wantStatus   int
		wantContent  string
		isXML        bool
	}{
		{
			name:         "json response",
			acceptHeader: "application/json",
			statusCode:   http.StatusOK,
			response: testResponse{
				Name:  "test",
				Value: 123,
			},
			wantStatus:  http.StatusOK,
			wantContent: `{"name":"test","value":123}`,
			isXML:       false,
		},
		{
			name:         "xml response",
			acceptHeader: "application/xml",
			statusCode:   http.StatusOK,
			response: testResponse{
				Name:  "test",
				Value: 123,
			},
			wantStatus:  http.StatusOK,
			wantContent: `<response><data><name>test</name><value>123</value></data></response>`,
			isXML:       true,
		},
		{
			name:         "error response json",
			acceptHeader: "application/json",
			statusCode:   http.StatusBadRequest,
			response:     utils.ApiError{StatusCode: http.StatusBadRequest, Msg: "test error"},
			wantStatus:   http.StatusBadRequest,
			wantContent:  `{"error":"test error"}`,
			isXML:        false,
		},
		{
			name:         "error response xml",
			acceptHeader: "application/xml",
			statusCode:   http.StatusBadRequest,
			response:     utils.ApiError{StatusCode: http.StatusBadRequest, Msg: "test error"},
			wantStatus:   http.StatusBadRequest,
			wantContent:  `<response><error>test error</error></response>`,
			isXML:        true,
		},
		{
			name:         "nil response",
			acceptHeader: "application/json",
			statusCode:   http.StatusNoContent,
			response:     nil,
			wantStatus:   http.StatusNoContent,
			wantContent:  "",
			isXML:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			w := httptest.NewRecorder()
			utils.RenderResponse(req, w, tt.statusCode, tt.response)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantContent != "" {
				if tt.isXML {
					// For XML responses, compare normalized XML
					assert.Equal(t,
						normalizeXML(tt.wantContent),
						normalizeXML(string(body)),
						"XML content mismatch",
					)
				} else {
					// For JSON responses, use JSONEq
					assert.JSONEq(t, tt.wantContent, string(body))
				}
			} else {
				assert.Empty(t, string(body))
			}

			// Verify content type header
			expectedContentType := tt.acceptHeader
			assert.Equal(t, expectedContentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestAllowedMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		allowedMethods []string
		wantStatus     int
	}{
		{
			name:           "allowed method",
			method:         "GET",
			allowedMethods: []string{"GET", "POST"},
			wantStatus:     http.StatusOK,
		},
		{
			name:           "not allowed method",
			method:         "DELETE",
			allowedMethods: []string{"GET", "POST"},
			wantStatus:     http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			utils.AllowedMethods(handler, tt.allowedMethods...)(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAllowedContentTypes(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		allowedTypes []string
		wantStatus   int
	}{
		{
			name:         "allowed content type",
			contentType:  "application/json",
			allowedTypes: []string{"application/json"},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "not allowed content type",
			contentType:  "text/plain",
			allowedTypes: []string{"application/json"},
			wantStatus:   http.StatusUnsupportedMediaType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}

			req := httptest.NewRequest("POST", "/test", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			utils.AllowedContentTypes(handler, tt.allowedTypes...)(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func normalizeXML(xmlStr string) string {
	xmlStr = strings.ReplaceAll(xmlStr, "\n", "")
	xmlStr = strings.ReplaceAll(xmlStr, "\t", "")
	xmlStr = strings.TrimSpace(xmlStr)

	var temp interface{}
	err := xml.Unmarshal([]byte(xmlStr), &temp)
	if err != nil {
		return xmlStr
	}

	normalized, err := xml.Marshal(temp)
	if err != nil {
		return xmlStr
	}

	return string(normalized)
}
