package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPyPiMux(t *testing.T) {
	mux := NewPyPiMux(&PyPiConfig{
		MaxFileSizeMB: 128,
	})
	rootMux := http.NewServeMux()
	rootMux.Handle("/pypi/", http.StripPrefix("/pypi", mux))
	server := httptest.NewServer(rootMux)
	defer server.Close()

	for _, url := range []string{
		server.URL + "/pypi/simple/",
		server.URL + "/pypi/proxy/packages/b5/f4/098d2270d52b41f1bd7db9fc288aaa0400cb48c2a3e2af6fa365d9720947/numpy-2.3.4.tar.gz?originalHost=pypi.org&originalScheme=https",
	} {
		t.Run(url, func(t *testing.T) {
			_ = requestAndAssertOk(t, url)
		})
	}
}

func requestAndAssertOk(t *testing.T, url string) []byte {
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	return body
}
