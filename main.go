package main

import (
	"fmt"
	"net/http"

	"github.com/pfernandom/go-pypi/middleware"
)

var MAX_FILE_SIZE_MB int64 = 128

func main() {
	config := &middleware.PyPiConfig{
		MaxFileSizeMB: MAX_FILE_SIZE_MB,
	}
	mux := middleware.NewPyPiMux(config)
	rootMux := http.NewServeMux()

	rootMux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rootMux.Handle("/pypi/", http.StripPrefix("/pypi", mux))
	fmt.Println("Server listening on :4040")
	http.ListenAndServe(":4040", rootMux)
}
