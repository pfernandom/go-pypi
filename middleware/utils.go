package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler
type MultiMiddleware []Middleware

func (m MultiMiddleware) WithMiddleware(next Middleware) MultiMiddleware {
	m = append(m, next)
	return m
}

func (m MultiMiddleware) WithHandlerFunc(handlerFunc func(w http.ResponseWriter, r *http.Request)) *MultiMiddleware {
	m = append(m, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerFunc(w, r)
			next.ServeHTTP(w, r)
		})
	})
	return &m
}

func (m MultiMiddleware) HandleFunc(handler http.HandlerFunc) http.Handler {
	root := func(next http.Handler) http.Handler {
		reversed := make([]Middleware, len(m))
		for i := range m {
			reversed[i] = m[len(m)-i-1]
		}
		for _, middleware := range reversed {
			next = middleware(next)
		}
		return next
	}
	return root(http.HandlerFunc(handler))
}

// ServeHTTP implements http.Handler.
func (m MultiMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	root := func(next http.Handler) http.Handler {
		reversed := make([]Middleware, len(m))
		for i := range m {
			reversed[i] = m[len(m)-i-1]
		}
		for _, middleware := range reversed {
			next = middleware(next)
		}
		return next
	}
	root(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do nothing
	})).ServeHTTP(w, r)
}

var _ http.Handler = MultiMiddleware{}
