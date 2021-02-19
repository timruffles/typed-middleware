# typed-middleware

Middleware is convenient but it comes at a cost to type safety and explicit code:
- using `context.Context` to store and retrieve values is not type-safe, and relies on implicit temporal coupling
- it's hard to know which middleware ended the chain by responding, and the code that runs the middleware can be confusing to to step through with a debugger

typedmiddleware uses code generation to avoid both of these issues without sacrificing convenience. You define a stack of middleware as an interface, and use go generate to generate a runnable stack.


```go
//go:generate typedmiddleware Middleware

// 1️⃣ this defines the stack of middleware you wish to use - order is significant, as middleware can
// return early
type Middleware interface {
	appmiddleware.UserForRequest
}

func GetUserHome(res http.ResponseWriter, req http.Request) {
	result, override := NewMiddlewareStack().Run(req)
	// 2️⃣ if a middleware wishes to respond an override is returned,
	if override != nil {
		// you can define your own response handlers that can inspect the response struct
		middleware.DefaultRespond(override, res)
		return
	}

	// 3️⃣ result is a GetUserHomeMiddleware, and as no override was 
	// returned all methods are now safe to use.
	user := result.User()
	fmt.Fprintf(res, "User ID %d", user.ID)
}
```
The result returned by `Run()` is a `Middleware` value, on which you can retrieve values set by the middleware via their individual interfaces - e.g the `User()` method of `UserForRequest`:

```go
type UserForRequest interface {
    User() models.User
}
```

The `go:generate` line configures typedmiddleware - here specifying that `NewMiddlewareStack` should be generated.

## Getting started

### With existing middleware

First write your handler. This can be in any style - `HandleFn` or a handler. Then define an interface value which references each of the middleware your handler needs in the order they should be called:

```diff
+ type HandlerMiddleware interface {
+     appmiddleware.MustAuthenticate
+     appmiddleware.AdminOnly
+     appmiddleware.RetrieveUser
+ }
  
  func YourHandler(res http.Response, req *http.Request) {
  }
```

then add the `go:generate` line to your file:

```diff
+ //go:generate typedmiddleware HandlerMiddleware
```

run `go generate path/to/your/file.go`. You should see `.../file_middleware.go` was generated. You won't edit this - instead you can change the file containing the `go:generate` file and re-generate it as your middleware changes.

You can now use the `NewHandlerMiddlewareStack()` method to construct a runnable implementation of your middleware stack. You will need to pass in instances of dependent middleware - which may include dependencies of the middleware you specified. If this is an existing application you'll likely have helpers to construct them.

```diff
  func YourHandler(res http.Response, req *http.Request) {
+    result, override := NewHandlerMiddlewareStack(/* pass dependencies */).Run(req)
+    if override != nil {
+       return middleware.DefaultRespond(override, res)
+    }
  }
```

This is using the `DefaultRespond` method. Your application will likely have its own functions that decide how to respond to given a `MiddlewareResponse` - e.g formatting an API error for your app.

If there was no override, you can now access any method on the middleware interfaces you specified in your handler.

### Writing your own middleware


Middleware is comprised of the interface that defines methods dependent code can access, and its implementing type. Normally this will be a struct.

As an example, let's write a middleware that requires a content-type is set. First we'll define the public interface

```go
type RequireContentType interface {
	ContentType() string
}
```

A middleware needs to follow a contract: if it doesn't indicate an error or early response, handlers and other middleware that depend on it should be able to safely use its interface.

Handlers or middleware that depend on our RequireContentType middleware will be able to call `ContentType()` to access the non-empty value supplied in the user request. If one was not supplied, we'll end the chain before our dependencies are called by indicating a 400 response should be returned. This fulfils the contract: either we find a content type and let the control flow continue to our dependencies, or indicate we must end the middleware chain at this point.

To implement that we'll need somewhere to define our `Run()` method, which will need to store a content type for future calls to `ContentType()`:

```go
type RequireContentTypeMiddleware struct {
	ct string
}

func (g *RequireContentTypeMiddleware) ContentType() string {
	return g.ct
}

func (g *RequireContentTypeMiddleware) Run(req *http.Request) (*middleware2.MiddlewareResponse, error) {
	ct, ok := req.Header["Content-Type"]
	if !ok || len(ct) == 0 || ct[0] == "" {
		return middleware2.Response(
			400,
			strings.NewReader("Must supply a content type"),
			nil,
		), nil
	}
	g.ct = ct[0]
	return nil, nil
}
```

It's useful to typecheck our implementation fulfils our public interface via this go idiom:

```
var _ RequireContentType = (*RequireContentTypeMiddleware)(nil)
```

Handlers and middleware can now specify a dependency on `RequireContentType`. This will ensure the `RequireContentTypeMiddleware.Run()` method is called before they are, and they can be written with the knowledge that a content type will always be present.

## How does this work?

typedmiddleware defines a contract with compatible middleware, and uses this to generate explicit code that ensures they are called in order.

The contract for middleware is:
1. use `req` to ensure it is ready to respond to its interface methods being called, by returning (nil,nil)
2. stop the chain, by either
    - returning a non-nil MiddlewareResponse
    - returning an error
    
Using the semantics of this contract, typedmiddleware can generate the implementation of the Run method that will call each middleware in order, ensuring that all dependant middleware are called before middleware that depend on them are. e.g if you specify a dependency on `UserForRequest`, and `UserForRequest` requires the `Authenticated` middleware, the following will happen when `Run()` is called:

1. `Authenticated` middleware's `Run(req)` method is called. 
    a. If run returns a response or error, we return to the caller
2. `UserForRequest` middleware's `Run(req, auth)`, with the second argument being an interface through which it can access the values it needs from the `Authenticated` middleware
3. With no more middleware we're done, and return to the caller, who can safely use methods from `UserForRequest`

