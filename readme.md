# typedmiddleware

Middleware is convenient, but the cost is type safety (using `context.Context` to store and retrieve values), and opacity (it's hard to know which middleware ended the chain by responding, and it's often hard to step through with a debugger).

typedmiddleware uses code generation to avoid both of these issues. You define a stack of middleware as an interface, and use go generate to generate a runnable stack. It will return a result on which you can retrieve values set by the middleware via their individual interfaces - e.g the `User()` method of `UserForRequest`:


```go
// user_handler.go

// the following line configures the generation, which outputs NewMiddlewareStack and its implementation
//go:generate typedmiddleware Middleware

// this defines the stack of middleware you wish to use - order is significant, as middleware can
// return early
type Middleware interface {
	appmiddleware.UserForRequest
}

func GetUserHome(res http.ResponseWriter, req http.Request) {
    // after you 
	result, override := NewMiddlewareStack().Run(req)
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

## How does this work?

typedmiddleware defines a constract with compatible middleware, and uses this to generate explicit code that ensures they are called in order.

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

