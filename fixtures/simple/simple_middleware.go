package simple

import (
	typedmiddleware "github.plaid.com/plaid/typedmiddleware"
	mockmiddleware "github.plaid.com/plaid/typedmiddleware/fixtures/mockmiddleware"
	"net/http"
)

type SimpleMiddlewareStack interface {
	Run(req *http.Request) (SimpleMiddleware, *typedmiddleware.MiddlewareResponse)
}

func NewSimpleMiddlewareStack(requireContentTypeMiddleware mockmiddleware.RequireContentTypeMiddleware) *SimpleMiddlewareStackImpl {
	return &SimpleMiddlewareStackImpl{RequireContentTypeMiddleware: requireContentTypeMiddleware}
}

type SimpleMiddlewareStackImpl struct {
	mockmiddleware.RequireContentTypeMiddleware
}

func (s *SimpleMiddlewareStackImpl) Run(req *http.Request) (SimpleMiddleware, *typedmiddleware.MiddlewareResponse) {
	result, err := s.RequireContentTypeMiddleware.Run(req)
	if result != nil {
		return nil, result
	}
	if err != nil {
		return nil, typedmiddleware.NewErrorResult(err)
	}
	return s, nil
}
