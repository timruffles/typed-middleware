//go:generate go run ../../cmd/typed-middleware.go SimpleMiddlewareStack
package simple

import (
	"fmt"
	"net/http"
	"strings"

	middleware2 "github.plaid.com/plaid/typed-middleware"
	"github.plaid.com/plaid/typed-middleware/fixtures/mockmiddleware"
)

type SimpleMiddlewareStack interface {
	mockmiddleware.RequireContentType
}

type simpleHandler struct {
	stack SimpleStack
}

func NewSimpleHandler(
) simpleHandler {
	// This part will be dependency injected as per normal - currently via app context
	// Since clientID middleware has no constructor
	h := simpleHandler{}
	h.stack = NewSimpleStack(
	)
	return h
}

func (h *simpleHandler) Handle(res http.ResponseWriter, req http.Request) {
	_, override := h.stack.Run(req)
	// the stack value could come from ctx for now, or be replaced by a mock
	if override != nil {
		// or can explicitly check out what's happened: an error, or a result spec
		middleware2.Respond(override, res)
		return
	}

	fmt.Println(res, strings.NewReader("ok"))
}
