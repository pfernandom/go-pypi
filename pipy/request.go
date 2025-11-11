package pipy

import (
	"net/url"
	"regexp"
	"strings"
)

type UploadRequestForm struct {
	Name                   string `form:"name"`
	Version                string `form:"version"`
	Sha256Digest           string `form:"sha256_digest"`
	ProtocolVersion        string `form:"protocol_version"`
	MetadataVersion        string `form:"metadata_version"`
	FileType               string `form:"filetype"`
	Pyversion              string `form:"pyversion"`
	AuthorEmail            string `form:"author_email"`
	DescriptionContentType string `form:"description_content_type"`
	Summary                string `form:"summary"`
	RequiresPython         string `form:"requires_python"`
}

func ParseUploadRequestFrom(r *url.Values) (*UploadRequestForm, error) {
	return &UploadRequestForm{
		Name:            NormalizeProjectName(r.Get("name")),
		Version:         r.Get("version"),
		Sha256Digest:    r.Get("sha256_digest"),
		ProtocolVersion: r.Get("protocol_version"),
		MetadataVersion: r.Get("metadata_version"),
		FileType:        r.Get("filetype"),
		Pyversion:       r.Get("pyversion"),
	}, nil
}

var normalizeProjectNameRegex = regexp.MustCompile("[-_.]+")

// Normalizes the project name according to PEP 503
func NormalizeProjectName(projectName string) string {
	return strings.ToLower(normalizeProjectNameRegex.ReplaceAllString(projectName, "-"))
}
