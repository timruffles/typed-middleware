# typed-middleware

Middleware is convenient, but the cost is type safety (using context.Context to store and retrieve values), and opacity (it's hard to see where in the stack issues arise, and it's very hard to step through with a debugger).

typed-middleware uses code generation to avoid both of these issues. You define a stack of middleware as an interface, and use go generate to generate a runnable stack:


```go
type GetUserHomeMiddleware interface {
	appmiddleware.UserForRequest
}

func GetUserHome(res http.ResponseWriter, req http.Request) {
	result, override := NewStack().Run(req)
	// if a middleware wishes to respond an override is returned,
    // and can be inspected or controlled by GetUserHome
	if override != nil {
		middleware.Respond(override, res)
		return
	}

    // the result has type-safe methods for accessing all middleware
    // methods, which are the methods of the GetUserHomeMiddleware interface
	cid := result.User()
	fmt.Fprintf(res, "Got client ID %d", cid)
}
```
