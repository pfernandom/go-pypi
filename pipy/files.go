package pipy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var storagePath = "./uploads"
var metadataFileName = "request.json"
var responseMetadataFileName = "response.json"

func SetupStorage() {
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.MkdirAll(storagePath, 0755)
	}
}

func GetIndexResponse() (*IndexResponse, error) {
	repoPath := filepath.Join(storagePath)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, RepoNotFound
	}
	repos, err := os.ReadDir(repoPath)
	if err != nil {
		return nil, newError("failed to read repos: %v", err)
	}
	projects := []Project{}
	for _, repo := range repos {
		projects = append(projects, Project{
			Name: repo.Name(),
		})
	}
	return &IndexResponse{Projects: projects}, nil
}

// Gets the python package from the storage path
func GetPackageDescriptor(packageName string) (*Response, error) {
	repoPath := filepath.Join(storagePath, NormalizeProjectName(packageName))
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, RepoNotFound
	}
	versions, err := os.ReadDir(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, RepoNotFound
		}
		return nil, fmt.Errorf("failed to read versions: %v", err)
	}
	files := []File{}
	for _, version := range versions {
		versionPath := filepath.Join(repoPath, version.Name())
		repoFiles, err := os.ReadDir(versionPath)
		if err != nil {
			return nil, newError("failed to read files: %v", err)
		}

		var metadata UploadRequestForm
		var projectFiles []*os.DirEntry
		for _, file := range repoFiles {
			if file.Name() == metadataFileName {
				meta, err := os.ReadFile(filepath.Join(versionPath, file.Name()))
				if err != nil {
					return nil, newError("failed to read metadata file: %v", err)
				}
				err = json.Unmarshal(meta, &metadata)
				if err != nil {
					return nil, newError("failed to unmarshal metadata: %v", err)
				}
			} else {
				projectFiles = append(projectFiles, &file)
			}
		}

		for _, file := range projectFiles {
			file := *file
			if file.Name() == metadataFileName {
				continue
			}
			sha256, err := getFileSHA256(filepath.Join(versionPath, file.Name()))
			if err != nil {
				return nil, newError("failed to get file SHA256: %v", err)
			}
			urlPath, err := url.JoinPath("/files", packageName, version.Name(), file.Name())
			if err != nil {
				return nil, newError("failed to join URL path: %v", err)
			}
			files = append(files, File{
				Filename: file.Name(),
				URL:      urlPath,
				Hashes: FileHashes{
					SHA256: sha256,
				},
			})
		}
	}
	versionNumbers := []string{}
	for _, version := range versions {
		versionNumbers = append(versionNumbers, version.Name())
	}
	return &Response{
		Name:     packageName,
		Versions: versionNumbers,
		Files:    files,
	}, nil
}

// Saves the file from the multipart form
func SavePublishRequestFile(uploadRequest *UploadRequestForm, r *http.Request) error {
	file, header, err := r.FormFile("content")
	if err != nil {
		return newError("failed to get file from form: %v", err)
	}
	repoPath, err := getPackageVersionPath(uploadRequest.Name, uploadRequest.Version)
	if err != nil {
		return newError("failed to get package version path: %v", err)
	}
	filePath := filepath.Join(repoPath, header.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return newError("failed to create file: %v", err)
	}
	defer dst.Close()
	defer file.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return newError("failed to copy file: %v", err)
	}
	return nil
}

// Saves the upload request data to a file
func SaveUploadRequestData(request *UploadRequestForm) error {
	requestPath, err := getPackageVersionPath(request.Name, request.Version)
	if err != nil {
		return newError("failed to get package version path: %v", err)
	}
	requestFile := filepath.Join(requestPath, metadataFileName)
	requestData, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return newError("failed to marshal request: %v", err)
	}
	err = os.WriteFile(requestFile, requestData, 0644)
	if err != nil {
		return newError("failed to write request file: %v", err)
	}
	return nil
}

func getFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", newError("failed to open file: %v", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return "", newError("failed to read file: %v", err)
	}
	return CalculateSHA256(data), nil
}

func GetFile(repo string, version string, filename string) (io.Reader, error) {
	repoPath := filepath.Join(storagePath, NormalizeProjectName(repo), version, filename)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, newError("file not found: %v", err)
	}
	return os.Open(repoPath)
}

func SaveFileFromPyPI(url *url.URL, filename string, repoData *ProjectInfo) error {
	repoPath, err := getPackageVersionPath(repoData.Repo, repoData.Version)
	if err != nil {
		return newError("failed to get package version path: %v", err)
	}
	filePath := filepath.Join(repoPath, filename)
	// If the file already exists, don't download it again
	if _, err := os.Stat(filePath); err == nil {
		return nil
	}
	file, err := os.Create(filePath)
	if err != nil {
		return newError("failed to create file %s: %v", filePath, err)
	}
	defer file.Close()
	response, err := http.Get(url.String())
	if err != nil {
		return newError("failed to get file from PyPI: %v", err)
	}
	defer response.Body.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return newError("failed to copy file: %v", err)
	}

	return nil
}

func getPackageVersionPath(packageName string, version string) (string, error) {
	packagePath := filepath.Join(storagePath, NormalizeProjectName(packageName), version)
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		err := os.MkdirAll(packagePath, 0755)
		if err != nil {
			return "", newError("failed to create directory: %v", err)
		}
	}
	return packagePath, nil
}
