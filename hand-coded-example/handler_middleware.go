package hand_coded_example

import (
	"net/http"

	middleware2 "github.plaid.com/plaid/typedmiddleware"
	"github.plaid.com/plaid/typedmiddleware/hand-coded-example/appmiddleware"
)

type HandlerDependenciesGenerated struct {
	// transitive dependencies
	appmiddleware.AuthenticationFromRequestMiddleware
	// direct dependencies
	appmiddleware.ClientIDFromRequestMiddleware
}

type Stack interface {
	Run(http.Request) (HandlerMiddleware, *middleware2.MiddlewareResponse)
}

func NewStack(
	clientId appmiddleware.ClientIDFromRequestMiddleware,
	authMiddleware appmiddleware.AuthenticationFromRequestMiddleware,
) Stack {
	return &HandlerDependenciesGenerated{
		ClientIDFromRequestMiddleware: clientId,
		AuthenticationFromRequestMiddleware: authMiddleware,
	}
}

func (s *HandlerDependenciesGenerated) Run(
	r http.Request,
) (HandlerMiddleware, *middleware2.MiddlewareResponse) {
	result, err := s.AuthenticationFromRequestMiddleware.Run(r)
	if result != nil {
		return nil, result
	}
	if err != nil {
		return nil, middleware2.NewErrorResult(err)
	}

	// if there are dependencies, we resolve them so stack will have the results
	result, err = s.ClientIDFromRequestMiddleware.Run(r, s)
	if result != nil {
		return nil, result
	}
	if err != nil {
		return nil, middleware2.NewErrorResult(err)
	}

	return s, nil
}
