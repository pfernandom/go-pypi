package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/pfernandom/go-pipy/pipy"
)

var logger *slog.Logger

func init() {
	logger = pipy.Logger
}

func PyPiCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("CacheMiddleware", "path", r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/proxy") {
			err := pipy.HandleProxyFileDownload(w, r, next)
			if err != nil {
				pipy.Logger.Error("Failed to handle proxy file download", "error", err)
				http.Error(w, fmt.Sprintf("Failed to handle proxy file download: %v", err), http.StatusInternalServerError)
				next.ServeHTTP(w, r)
				return
			}
		}

	})
}

type PyPiConfig struct {
	MaxFileSizeMB int64
}

func NewPyPiMux(config *PyPiConfig) *http.ServeMux {
	pipy.SetupStorage()
	mux := http.NewServeMux()

	mid := MultiMiddleware{}.
		WithMiddleware(ErrorHandler)

	mux.Handle("GET /simple/", mid.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		response, err := pipy.GetIndexResponse()
		if err != nil {
			logger.Error("Failed to get index response", "error", err)
			http.Error(w, fmt.Sprintf("Failed to get index response: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
		json.NewEncoder(w).Encode(response)
	}))

	mux.Handle("POST /simple/", mid.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form with a maximum memory limit (e.g., 32MB)
		err := r.ParseMultipartForm(config.MaxFileSizeMB << 20) // 32 MB
		if err != nil {
			logger.Error("Failed to parse multipart form", "error", err)
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		parsedRequest, err := pipy.ParseUploadRequestFrom(&r.Form)
		if err != nil {
			logger.Error("Failed to parse upload request", "error", err)
			http.Error(w, "Failed to parse upload request", http.StatusBadRequest)
			return
		}

		// Handle file uploads
		err = pipy.SavePublishRequestFile(parsedRequest, r)
		if err != nil {
			logger.Error("Failed to save file", "error", err)
			http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
			return
		}

		// Save the upload request data
		err = pipy.SaveUploadRequestData(parsedRequest)
		if err != nil {
			logger.Error("Failed to save upload request data", "error", err)
			http.Error(w, fmt.Sprintf("Failed to save upload request data: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	mux.Handle("GET /simple/{package}/", mid.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := r.PathValue("package")
		logger.Debug("Getting repo", "repo", repo)

		files, err := pipy.GetPackageDescriptor(repo)
		if err != nil {
			if err == pipy.RepoNotFound {
				pipy.HandleProxyGetDescriptor(repo, w, r)
				return
			} else {
				logger.Error("Failed to get repo", "error", err)
				http.Error(w, fmt.Sprintf("Failed to get repo: %v", err), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
		json.NewEncoder(w).Encode(files)
	}))
	mux.Handle("GET /simple/{package}/{version}/{filename}", mid.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Handling get filename", "path", r.URL.Path)
		repo, version, filename := r.PathValue("package"), r.PathValue("version"), r.PathValue("filename")

		file, err := pipy.GetFile(repo, version, filename)
		if err != nil {
			logger.Error("Failed to get file in Handling get filename", "error", err)
			http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(w, file)
		if err != nil {
			logger.Error("Failed to copy file", "error", err)
			http.Error(w, fmt.Sprintf("Failed to copy file: %v", err), http.StatusInternalServerError)
			return
		}
	}))

	mux.Handle("/proxy/", PyPiCacheMiddleware(
		http.FileServer(http.Dir("uploads")),
	))

	return mux
}
