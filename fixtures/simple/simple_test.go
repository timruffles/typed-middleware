package simple

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.plaid.com/plaid/typedmiddleware/fixtures/mockmiddleware"
)

func TestMiddlewareApplied(t *testing.T) {
	handler := NewSimpleHandler(
		NewSimpleMiddlewareStack(mockmiddleware.RequireContentTypeMiddleware{}),
	)

	t.Run("middleware value accessible", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Add("Content-Type", "test-type")
		recorder := httptest.NewRecorder()
		handler.Handle(recorder, req)
		assert.Equal(t, "Content type from middleware: test-type", recorder.Body.String())
	})

	t.Run("handles early response from middleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()
		handler.Handle(recorder, req)
		assert.Equal(t, 400, recorder.Code)
	})
}


