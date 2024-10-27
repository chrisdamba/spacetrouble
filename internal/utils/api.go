package utils

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"time"
)

type ApiError struct {
	StatusCode int    `json:"-"`
	Msg        string `json:"error,omitempty"`
}
type ContentType string

type XMLResponse struct {
	XMLName xml.Name    `xml:"response"`
	Data    interface{} `xml:"data,omitempty"`
	Error   string      `xml:"error,omitempty"`
}

const (
	ContentTypeJSON ContentType = "application/json"
	ContentTypeXML  ContentType = "application/xml"
)

func (o *ApiError) Error() string {
	return fmt.Sprintf("%d: %s", o.StatusCode, o.Msg)
}

func JsonDecodeBody(r *http.Request, dst interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}

func NewInternalServerError(msg string) ApiError {
	return ApiError{http.StatusInternalServerError, msg}
}

func NewBadRequest(msg string) ApiError {
	return ApiError{http.StatusBadRequest, msg}
}

func RenderResponse(r *http.Request, w http.ResponseWriter, statusCode int, res interface{}) {
	contentType := getResponseContentType(r)
	switch contentType {
	case ContentTypeJSON:
		renderJson(w, statusCode, res)
	case ContentTypeXML:
		renderXML(w, statusCode, res)
	default:
		renderJson(w, http.StatusUnsupportedMediaType, nil)
	}
}

func AllowedMethods(next http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		found := existsInSlice(methods, r.Method)
		if found {
			next(w, r)
		} else {
			RenderResponse(r, w, http.StatusMethodNotAllowed, nil)
		}
	}
}

func AllowedContentTypes(next http.HandlerFunc, mediaTypes ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		found := existsInSlice(mediaTypes, r.Header.Get("content-type"))
		if found {
			next(w, r)
		} else {
			RenderResponse(r, w, http.StatusUnsupportedMediaType, nil)
		}
	}
}

func EncodeCursor(t time.Time, id uuid.UUID) string {
	cursor := fmt.Sprintf("%s,%s", t.Format(time.RFC3339Nano), id.String())
	return base64.StdEncoding.EncodeToString([]byte(cursor))
}

func DecodeCursor(encoded string) (time.Time, uuid.UUID, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	parts := strings.Split(string(decodedBytes), ",")
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor format")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	return t, id, nil
}

func existsInSlice(list []string, needle string) bool {
	for i := range list {
		if list[i] == needle {
			return true
		}
	}
	return false
}

func getResponseContentType(r *http.Request) ContentType {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return ContentTypeJSON // default to JSON if no Accept header
	}

	// parse Accept header and get the highest priority content type we support
	types := strings.Split(accept, ",")
	for _, t := range types {
		mt := strings.TrimSpace(strings.Split(t, ";")[0]) // rmove quality values
		switch mt {
		case string(ContentTypeJSON):
			return ContentTypeJSON
		case string(ContentTypeXML):
			return ContentTypeXML
		}
	}
	return ContentTypeJSON // default to JSON if no supported type found
}

func renderJson(w http.ResponseWriter, statusCode int, res interface{}) {
	w.Header().Set("Content-Type", "application/json")
	var body []byte
	if res != nil {
		var err error
		body, err = json.Marshal(res)
		if err != nil {
			ae := NewInternalServerError(err.Error())
			statusCode = ae.StatusCode
			body, err = json.Marshal(&ae)
			if err != nil {
				body = []byte(`{"msg": "` + err.Error() + `"}`)
			}
		}
	}
	w.WriteHeader(statusCode)
	if len(body) > 0 {
		w.Write(body)
	}
}

func renderXML(w http.ResponseWriter, statusCode int, res interface{}) {
	w.Header().Set("Content-Type", "application/xml")

	var body []byte
	var err error

	if res != nil {
		switch v := res.(type) {
		case ApiError:
			xmlRes := XMLResponse{Error: v.Msg}
			body, err = xml.Marshal(xmlRes)
		case error:
			xmlRes := XMLResponse{Error: v.Error()}
			body, err = xml.Marshal(xmlRes)
		default:
			xmlRes := XMLResponse{Data: res}
			body, err = xml.Marshal(xmlRes)
		}

		if err != nil {
			ae := NewInternalServerError(err.Error())
			xmlRes := XMLResponse{Error: ae.Msg}
			statusCode = ae.StatusCode
			body, err = xml.Marshal(xmlRes)
			if err != nil {
				body = []byte(`<?xml version="1.0" encoding="UTF-8"?>
                    <response><error>Internal Server Error</error></response>`)
			}
		}
	}

	w.WriteHeader(statusCode)
	if len(body) > 0 {
		w.Write(body)
	}
}
