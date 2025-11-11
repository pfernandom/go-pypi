package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorHandler(t *testing.T) {
	mf := &MiddlewareFactory{}
	multiMiddleware := MultiMiddleware{}.
		WithMiddleware(mf.NewMiddleware("Middleware 1")).
		WithMiddleware(mf.NewMiddleware("Middleware 2")).
		WithMiddleware(mf.NewMiddleware("Middleware 3")).
		WithMiddleware(ErrorHandler).
		WithHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test error")
		})
	server := httptest.NewServer(multiMiddleware)
	defer server.Close()
	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
