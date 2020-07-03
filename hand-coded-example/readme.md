# typedmiddleware

Middleware is convenient, but the cost is type safety (using context.Context to store and retrieve values), and opacity (it's hard to know which middleware ended the chain by responding, and it's often hard to step through with a debugger).

typedmiddleware uses code generation to avoid both of these issues. You define a stack of middleware as an interface, and use go generate to generate a runnable stack:


```go
type GetUserHomeMiddleware interface {
	appmiddleware.UserForRequest
}

func GetUserHome(res http.ResponseWriter, req http.Request) {
	result, override := NewStack().Run(req)
	// if a middleware wishes to respond an override is returned,
	if override != nil {
        // you can define your own response handlers that can inspect the response struct
		middleware.DefaultRespond(override, res)
		return
	}

    // result is a GetUserHomeMiddleware, and by the middleware contract (see below), if 
    // no override was returned all methods are now safe to use.
	cid := result.User()
	fmt.Fprintf(res, "Got client ID %d", cid)
}
```

The contract for middleware is:
1. use `req` to ensure it is ready to respond to its interface methods being called, by returning (nil,nil)
2. stop the chain, by either
    - returning a non-nil MiddlewareResponse
    - returning an error
    
Using this contract, typedmiddleware can generate the implementation of the Run method that will call each middleware in order, ensuring that all dependant middleware are called before middleware that depend on them are. e.g if you specify a dependency on `UserForRequest`, and `UserForRequest` requires the `Authenticated` middleware, the following will happen when `Run()` is called:

1. `Authenticated` middleware's `Run(req)` method is called. 
    a. If run returns a response or error, we return to the caller
2. `UserForRequest` middleware's `Run(req, auth)`, with the second argument being an interface through which it can access the values it needs from the `Authenticated` middleware
3. With no more middleware we're done, and return to the caller, who can safely use methods from `UserForRequest`

