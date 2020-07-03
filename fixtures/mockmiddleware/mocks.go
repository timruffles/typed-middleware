package mockmiddleware

import (
	"net/http"
	"strings"

	middleware2 "github.plaid.com/plaid/typed-middleware"
)

type RequireContentType interface {
	RequireContentType()
}

type RequireContentTypeMiddleware struct {
}

var _ RequireContentType = RequireContentTypeMiddleware(nil)

func (g RequireContentTypeMiddleware) RequireContentType() {
}

func (g RequireContentTypeMiddleware) Run(req http.Request) (*middleware2.MiddlewareResponse, error) {
	if _, ok := req.Header["Content-Type"]; !ok {
		return middleware2.Response(
			400,
			strings.NewReader("Must supply a content type"),
			nil,
		), nil
	}
	return nil, nil
}

