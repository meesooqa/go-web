package middleware

import "net/http"

// Middleware is a function that wraps one http.Handler with another.
// It uses the standard Go signature, directly compatible with net/http
// without any custom abstraction layers, any existing third-party middleware
// following this pattern can be used seamlessly without adapters
type Middleware func(http.Handler) http.Handler

// Chain wraps the given handler in a middleware chain.
// The first element in the list is the outermost middleware:
// it receives the request first and sees the response last.
//
//	Chain(handler, A, B) is equivalent to A(B(handler))
//	→ request flow: A → B → handler
//	→ response flow: handler → B → A
func Chain(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}
