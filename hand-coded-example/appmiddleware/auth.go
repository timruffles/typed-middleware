package appmiddleware

import (
	"net/http"

	middleware2 "github.plaid.com/plaid/typedmiddleware"
)

type AuthenticationFromRequest interface {
	Token() string
}

type EnvString string

type AuthenticationFromRequestMiddleware struct {
	env EnvString
	tok string
}

func NewAuthenticationFromRequestMiddleware(env EnvString) (AuthenticationFromRequest, error) {
	return AuthenticationFromRequestMiddleware{env: env}, nil
}

func (m AuthenticationFromRequestMiddleware) Run(r http.Request) (*middleware2.MiddlewareResponse, error) {
	return nil, nil
}
func (m AuthenticationFromRequestMiddleware) Token() string {
	return m.tok
}












