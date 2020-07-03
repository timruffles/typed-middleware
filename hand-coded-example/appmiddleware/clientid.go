package appmiddleware

import (
	"net/http"

	middleware2 "github.plaid.com/plaid/typedmiddleware"
)

type GetUserResult struct {
	Value int
}

type ClientIDFromRequest interface {
	ClientID() int
}

type ClientIDFromRequestMiddleware struct {
	id int
}

type dependencies interface {
	AuthenticationFromRequest
}

func (m ClientIDFromRequestMiddleware) Run(r http.Request, deps dependencies) (*middleware2.MiddlewareResponse, error) {
	return nil, nil
}
func (m ClientIDFromRequestMiddleware) ClientID() int {
	return m.id
}

