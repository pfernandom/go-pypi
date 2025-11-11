package pipy

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

type ProjectInfo struct {
	Repo     string
	Version  string
	Filename string
}

var repoRegex = regexp.MustCompile(`([\w_\-]+)-(.*?)(?:\.tar\.gz|\.exe|\.zip|\.whl|\.egg|\.rpm)$`)

func ParseProjectData(fileName string) (ProjectInfo, error) {
	if strings.HasPrefix(fileName, "http") {
		fileName = filepath.Base(fileName)
	}

	match := repoRegex.FindStringSubmatch(fileName)
	if len(match) == 0 {
		return ProjectInfo{}, fmt.Errorf("failed to parse repo data: %v", fileName)
	}
	return ProjectInfo{
		Repo:     match[1],
		Version:  match[2],
		Filename: fileName,
	}, nil
}

func EncodeUrlAsUrlSafeBase64(path string, parsedUrl *url.URL) (*url.URL, error) {
	originalHost := parsedUrl.Host
	originalScheme := parsedUrl.Scheme
	parsedUrl.Host = "localhost:4040"
	parsedUrl.Scheme = "http"
	newPath, err := url.JoinPath(path, parsedUrl.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to join path: %v", err)
	}
	parsedUrl.Path = newPath
	queryParams := parsedUrl.Query()
	queryParams.Set("originalHost", originalHost)
	queryParams.Set("originalScheme", originalScheme)
	parsedUrl.RawQuery = queryParams.Encode()
	return parsedUrl, nil
}

func DecodeUrlFromUrlSafeBase64(path string, encodedUrl url.URL) (*url.URL, error) {
	queryParams := encodedUrl.Query()
	originalHost := queryParams.Get("originalHost")
	if originalHost == "" {
		return nil, fmt.Errorf("original host not found")
	}
	originalScheme := queryParams.Get("originalScheme")
	if originalScheme == "" {
		originalScheme = "https"
	}
	queryParams.Del("originalHost")
	queryParams.Del("originalScheme")
	encodedUrl.Host = originalHost
	encodedUrl.Scheme = originalScheme
	encodedUrl.Path = strings.TrimPrefix(encodedUrl.Path, path)
	encodedUrl.RawQuery = queryParams.Encode()
	return &encodedUrl, nil
}

func TryPrettyPrintJson(body []byte) string {
	var jsonBody map[string]interface{}
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		return string(body)
	}
	jsonString, err := json.MarshalIndent(jsonBody, "", "  ")
	if err != nil {
		return string(body)
	}
	return string(jsonString)
}
