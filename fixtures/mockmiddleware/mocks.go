package mockmiddleware

import (
	"net/http"
	"strings"

	middleware2 "github.plaid.com/plaid/typedmiddleware"
)

type RequireContentType interface {
	ContentType() string
}

type RequireContentTypeMiddleware struct {
	ct string
}

var _ RequireContentType = (*RequireContentTypeMiddleware)(nil)

func (g *RequireContentTypeMiddleware) ContentType() string {
	return g.ct
}

func (g *RequireContentTypeMiddleware) Run(req *http.Request) (*middleware2.MiddlewareResponse, error) {
	ct, ok := req.Header["Content-Type"]
	if !ok || len(ct) == 0 || ct[0] == "" {
		return middleware2.Response(
			400,
			strings.NewReader("Must supply a content type"),
			nil,
		), nil
	}
	g.ct = ct[0]
	return nil, nil
}

