package middleware

import (
	"net/http"

	"github.com/pfernandom/go-pypi/pipy"
)

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Recover from panics and handle errors
		defer func() {
			if r := recover(); r != nil {
				pipy.Logger.Error("Recovered from panic", "error", r)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
