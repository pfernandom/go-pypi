package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MiddlewareFactory struct {
	logs []string
}

func (m *MiddlewareFactory) NewMiddleware(log string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.logs = append(m.logs, log)
			next.ServeHTTP(w, r)
		})
	}
}
func TestMultiMiddleware(t *testing.T) {
	mf := &MiddlewareFactory{}
	var reachedHandler bool
	multiMiddleware := MultiMiddleware{}.
		WithMiddleware(mf.NewMiddleware("Middleware 1")).
		WithMiddleware(mf.NewMiddleware("Middleware 2")).
		WithMiddleware(mf.NewMiddleware("Middleware 3")).
		WithHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reachedHandler = true
			mf.logs = append(mf.logs, "Handler")
			w.WriteHeader(http.StatusSeeOther)
		}).
		WithMiddleware(mf.NewMiddleware("Middleware 4"))

	server := httptest.NewServer(multiMiddleware)

	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("Failed to get: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, resp.StatusCode)
		return
	}

	assert.Equal(t, 5, len(mf.logs))
	assert.Equal(t, "Middleware 1", mf.logs[0])
	assert.Equal(t, "Middleware 2", mf.logs[1])
	assert.Equal(t, "Middleware 3", mf.logs[2])
	assert.Equal(t, "Handler", mf.logs[3])
	assert.Equal(t, "Middleware 4", mf.logs[4])
	assert.True(t, reachedHandler)
}

func TestNestedMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("1"))
	})
	mux.HandleFunc("/2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("2"))
	})

	rootMux := http.NewServeMux()
	rootMux.Handle("/numbers/", http.StripPrefix("/numbers", mux))
	server := httptest.NewServer(rootMux)
	defer server.Close()
	resp, err := http.Get(server.URL + "/numbers/1")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "1", string(body))
	resp, err = http.Get(server.URL + "/numbers/2")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "2", string(body))
}
