// Package httproute adds middleware to the routes of an http.ServeMux.
package httproute

import (
	"net/http"
	"slices"
)

type Middleware func(http.Handler) http.Handler

// Routes applies its middleware to each route and not to the mux, so a request that matches no route skips it.
type Routes struct {
	mux        *http.ServeMux
	middleware []Middleware
	sealed     bool
}

func New(mux *http.ServeMux) *Routes {
	return &Routes{mux: mux}
}

// Use panics after a call to Handle or With, because the middleware would not apply to those routes.
func (r *Routes) Use(middleware ...Middleware) {
	if r.sealed {
		panic("httproute: Use after Handle or With")
	}

	r.middleware = append(r.middleware, middleware...)
}

// With gives a copy of r that runs its middleware after the middleware of Use, each in the order of its list.
func (r *Routes) With(middleware ...Middleware) *Routes {
	r.sealed = true

	return &Routes{mux: r.mux, middleware: slices.Concat(r.middleware, middleware)}
}

func (r *Routes) Handle(pattern string, h http.Handler) {
	r.sealed = true

	for _, m := range slices.Backward(r.middleware) {
		h = m(h)
	}

	r.mux.Handle(pattern, h)
}
