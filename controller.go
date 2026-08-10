package web

import "net/http"

// Route describes a single controller route.
// Pattern is a standard net/http.ServeMux pattern (Go 1.22+),
// e.g., "GET /{$}", "POST /login", "GET /users/{id}".
// If no method is specified at the beginning, the route responds to any HTTP method.
type Route struct {
	Pattern string
	Handler http.HandlerFunc
}

// Controller represents a unit of web pages: a group of related routes
// (e.g., all pages in the "user profile" section). Registered via Server.Register.
//
// A typical implementation stores dependencies (templates, data access)
// as struct fields and returns its methods as handlers:
//
//	type HomeController struct {
//		tmpl *web.Templates
//	}
//
//	func (c *HomeController) Routes() []web.Route {
//		return []web.Route{
//			{Pattern: "GET /{$}", Handler: c.Index},
//			{Pattern: "GET /about", Handler: c.About},
//		}
//	}
type Controller interface {
	Routes() []Route
}
