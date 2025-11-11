package pipy

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParseRepoData(t *testing.T) {
	tests := []struct {
		fileName string
		want     ProjectInfo
	}{
		{
			fileName: "numpy-1.26.0.tar.gz",
			want: ProjectInfo{
				Repo:     "numpy",
				Version:  "1.26.0",
				Filename: "numpy-1.26.0.tar.gz",
			},
		},
		{
			fileName: "numpy-1.26.0.exe",
			want: ProjectInfo{
				Repo:     "numpy",
				Version:  "1.26.0",
				Filename: "numpy-1.26.0.exe",
			},
		},
		{
			fileName: "numpy-1.26.0.zip",
			want: ProjectInfo{
				Repo:     "numpy",
				Version:  "1.26.0",
				Filename: "numpy-1.26.0.zip",
			},
		},
		{
			fileName: "numpy-1.26.0.whl",
			want: ProjectInfo{
				Repo:     "numpy",
				Version:  "1.26.0",
				Filename: "numpy-1.26.0.whl",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.fileName, func(t *testing.T) {
			got, err := ParseProjectData(test.fileName)
			if err != nil {
				t.Errorf("parseRepoData(%s) = %v", test.fileName, err)
			}
			if got != test.want {
				t.Errorf("parseRepoData(%s) = %v, want %v", test.fileName, got, test.want)
			}
		})
	}
}

func TestEncodeUrlAsUrlSafeBase64(t *testing.T) {

	tests := []struct {
		originalUrl string
		filename    string
		path        string
		want        string
	}{
		{
			originalUrl: "https://pypi.org/simple/numpy/1.26.0/numpy-1.26.0.tar.gz",
			filename:    "numpy-1.26.0.tar.gz",
			path:        "/proxy",
		},
		{
			originalUrl: "https://pypi.org/packages/b5/f4/098d2270d52b41f1bd7db9fc288aaa0400cb48c2a3e2af6fa365d9720947/numpy-2.3.4.tar.gz",
			filename:    "numpy-2.3.4.tar.gz",
			path:        "/proxy",
		},
	}

	for _, test := range tests {
		t.Run(test.originalUrl, func(t *testing.T) {
			parsedUrl, err := url.Parse(test.originalUrl)
			if err != nil {
				t.Errorf("parse URL: %v", err)
				return
			}
			got, err := EncodeUrlAsUrlSafeBase64(test.path, parsedUrl)
			if err != nil {
				t.Errorf("encodeUrlAsUrlSafeBase64(%s, %s, %s) = %v", test.path, test.originalUrl, test.filename, err)
			}

			fmt.Println("got", got)
			decodedUrl, err := DecodeUrlFromUrlSafeBase64(test.path, *got)
			if err != nil {
				t.Errorf("decodeUrlFromUrlSafeBase64(%s, %s) = %v", test.path, got, err)
				return
			}
			if decodedUrl.String() != test.originalUrl {
				t.Errorf("got %s, want %s", decodedUrl, test.originalUrl)
			}
		})
	}
}

//
