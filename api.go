package middleware

import (
	"io"
	"net/http"
)


type MiddlewareResponse struct {
	isError      bool
	error        error
	responseSpec *responseSpec
}

type MiddlewareValue interface {
	Present() bool
	Value() interface{}
}

func NewErrorResult(err error) *MiddlewareResponse {
	return &MiddlewareResponse{isError: true, error: err}
}

type responseSpec struct {
	Header      http.Header
	StatusCode  int
	Body        io.Reader
}

func Response(
	statusCode int,
	body io.Reader,
	header http.Header,
) *MiddlewareResponse {
	return &MiddlewareResponse{
		responseSpec: &responseSpec{
			Header:      header,
			StatusCode:  statusCode,
			Body:        body,
		},
	}
}

func Respond(overide *MiddlewareResponse, res http.ResponseWriter) {
	if overide == nil {
		// programming error
	}
	if overide.isError {
		// handle error
	}
	// set headers, etc
}

// Idea: this could be used as a placeholder for the GoGenerate call
func GenerateAndRunStack(t interface{}) (interface{}, *MiddlewareResponse) {
	return nil, nil
}


