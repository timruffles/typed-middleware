//go:generate go run ../../cmd/typedmiddleware.go SimpleMiddleware
package simple

import (
	"fmt"
	"net/http"

	middleware2 "github.plaid.com/plaid/typedmiddleware"
	"github.plaid.com/plaid/typedmiddleware/fixtures/mockmiddleware"
)

type SimpleMiddleware interface {
	mockmiddleware.RequireContentType
}

type simpleHandler struct {
	stack SimpleMiddlewareStack
}

func NewSimpleHandler(
	stack SimpleMiddlewareStack,
) *simpleHandler {
	// This part will be dependency injected as per normal - currently via app context
	// Since clientID middleware has no constructor
	return &simpleHandler{
		stack: stack,
	}
}

func (h *simpleHandler) Handle(res http.ResponseWriter, req *http.Request) {
	result, override := h.stack.Run(req)
	// the stack value could come from ctx for now, or be replaced by a mock
	if override != nil {
		// or can explicitly check out what's happened: an error, or a result spec
		middleware2.DefaultRespond(override, res)
		return
	}

	fmt.Fprintf(res, "Content type from middleware: %s", result.ContentType())
}
