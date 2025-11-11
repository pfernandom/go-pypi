package pipy

import (
	"fmt"
	"net/url"
)

// PyPI Simple API Response (PEP 503)
type Response struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
	Files    []File   `json:"files"`
}

// File Hashes
type FileHashes struct {
	SHA256 string `json:"sha256"`
}

// Python Package File
type File struct {
	Filename       string     `json:"filename"`
	URL            string     `json:"url"`
	Hashes         FileHashes `json:"hashes"`
	RequiresPython *string    `json:"requires-python,omitempty"`
	CoreMetadata   *any       `json:"core-metadata,omitempty"`
	Yanked         *any       `json:"yanked,omitempty"`
	Provenance     *string    `json:"provenance,omitempty"`
}

func (f *File) GetUrl() (*url.URL, error) {
	parsedUrl, err := url.Parse(f.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}
	return parsedUrl, nil
}

// Meta Data
type Meta struct {
	ApiVersion          string  `json:"api-version"`
	ProjectStatus       *string `json:"project-status,omitempty"`
	ProjectStatusReason *string `json:"project-status-reason,omitempty"`
}

// Index Response
type IndexResponse struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	Name string `json:"name"`
}
