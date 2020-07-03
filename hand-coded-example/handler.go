//go:generate go run middleware HandlerMiddleware:Stack
package hand_coded_example

import (
	"fmt"
	"net/http"

	middleware2 "github.plaid.com/plaid/typedmiddleware"
	"github.plaid.com/plaid/typedmiddleware/hand-coded-example/appmiddleware"
)


// app programmer defines their needs
type HandlerMiddleware interface {
	// references each middleware they require here
	appmiddleware.ClientIDFromRequest
}

type getUserHandler struct {
	stack Stack
}

func NewUserHandler(
	authMiddleware appmiddleware.AuthenticationFromRequestMiddleware,
) getUserHandler {
	// This part will be dependency injected as per normal - currently via app context
	// Since clientID middleware has no constructor
	h := getUserHandler{}
	h.stack = NewStack(
		// currently push instantiation up here, leaving the generator
		// free of that complexity
		appmiddleware.ClientIDFromRequestMiddleware{},
		authMiddleware,
	)
	return h
}

func (h *getUserHandler) Handle(res http.ResponseWriter, req http.Request) {
	result, override := h.stack.Run(req)
	// the stack value could come from ctx for now, or be replaced by a mock
	if override != nil {
		// or can explicitly check out what's happened: an error, or a result spec
		middleware2.DefaultRespond(override, res)
		return
	}

	cid := result.ClientID()
	fmt.Fprintf(res, "Got client ID %d", cid)
}


