package pipy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var piPyUrl = "https://pypi.org/simple"

func HandleProxyGetDescriptor(filename string, w http.ResponseWriter, r *http.Request) {
	Logger.Debug("Proxying file from PyPI", "filename", filename)
	filename = strings.TrimPrefix(filename, "/")

	Logger.Debug("Proxying file from PyPI", "url", fmt.Sprintf("%s/%s/", piPyUrl, filename))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/", piPyUrl, filename), nil)
	if err != nil {
		Logger.Error("Failed to get file from PyPI", "error", err)
		http.Error(w, fmt.Sprintf("Failed to get file from PyPI: %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", "application/vnd.pypi.simple.v1+json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		Logger.Error("Failed to get file from PyPI", "error", err)
		http.Error(w, fmt.Sprintf("Failed to get file from PyPI: %v", err), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		Logger.Error("Failed to read response body", "error", err)
		http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), http.StatusInternalServerError)
		return
	}

	var responseData Response
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		Logger.Error("Failed to unmarshal response", "error", err)
		jsonBody := TryPrettyPrintJson(body)
		Logger.Debug("Response body", "body", jsonBody)
		http.Error(w, fmt.Sprintf("Failed to unmarshal response: %v", err), http.StatusInternalServerError)
		return
	}

	updatedFiles := []File{}
	for _, file := range responseData.Files {

		fileUrl, err := file.GetUrl()
		if err != nil {
			Logger.Error("Proxied response has invalid URL", "error", err)
			continue
		}
		newUrl, err := EncodeUrlAsUrlSafeBase64("/proxy", fileUrl)
		if err != nil {
			Logger.Warn("Failed to encode URL", "error", err)
			continue
		}
		file.URL = newUrl.String()

		updatedFiles = append(updatedFiles, file)
	}
	responseData.Files = updatedFiles
	w.Header().Set("Content-Type", "application/vnd.pypi.simple.v1+json")
	json.NewEncoder(w).Encode(responseData)
}

// Requests the file from PyPI and saves it to the local storage
func HandleProxyFileDownload(w http.ResponseWriter, r *http.Request, next http.Handler) error {
	decodedUrl, err := DecodeUrlFromUrlSafeBase64("/proxy", *r.URL)
	if err != nil {
		Logger.Error("Failed to decode URL", "error", err)
		return fmt.Errorf("failed to decode URL: %v", err)
	}
	Logger.Debug("Decoded URL", "url", decodedUrl.String())
	repoData, err := ParseProjectData(decodedUrl.String())
	if err != nil {
		Logger.Error("Failed to parse repo data", "error", err)
		return fmt.Errorf("failed to parse repo data: %v", err)
	}
	err = SaveFileFromPyPI(decodedUrl, repoData.Filename, &repoData)
	if err != nil {
		Logger.Error("Failed to save file", "error", err)
		return fmt.Errorf("failed to save file: %v", err)
	}
	r.URL.Path = fmt.Sprintf("/%s/%s/%s", repoData.Repo, repoData.Version, repoData.Filename)
	next.ServeHTTP(w, r)
	return nil
}
