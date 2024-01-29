package catalog

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/openshift-pipelines/catalog-cd/internal/fetcher"
	"github.com/openshift-pipelines/catalog-cd/internal/fetcher/config"
)

func FetchFromExternals(e config.External, client *api.RESTClient) (Catalog, error) {
	c := Catalog{
		Resources: map[string]Resource{},
	}
	for _, r := range e.Repositories {
		if r.Name == "" {
			// Name is empty, take the last part of the URL
			r.Name = filepath.Base(r.URL)
		}
		c.Resources[r.Name] = Resource{}

		m, err := fetcher.FetchContractsFromRepository(r, client)
		if err != nil {
			return c, err
		}
		for _, v := range r.IgnoreVersions {
			// Remove ignored versions from map
			delete(m, v)
		}

		for version := range m {
			resourcesDownloaldURI := fmt.Sprintf("%s/releases/download/%s/%s", r.URL, version, r.ResourcesTarballName)
			version = strings.TrimPrefix(version, "v")
			c.Resources[r.Name][version] = resourcesDownloaldURI
		}
	}
	return c, nil
}


// Function to generate filesystem
func GenerateFilesystem(path string, c Catalog, resourceType string) error {
	for name, resource := range c.Resources {
			fmt.Fprintf(os.Stderr, "# Fetching resources from %s\n", name)
			for version, uri := range resource {
					fmt.Fprintf(os.Stderr, "## Fetching version %s\n", version)
					if err := fetchAndExtract(path, uri, version, resourceType); err != nil {
							fmt.Fprintf(os.Stderr, "Failed to fetch resource %s: %v, skipping\n", uri, err)
							continue
					}
					
					// Add source annotation to Task YAML file for each task
					taskDir := filepath.Join(path, "tasks", name, version)
					err := filepath.Walk(taskDir, func(file string, info os.FileInfo, err error) error {
							if err != nil {
									return err
							}
							if !info.IsDir() && filepath.Ext(file) == ".yaml" {
									fmt.Fprintf(os.Stderr, "Adding source annotation to Task YAML file: %s\n", file)
									if err := addSourceAnnotationToTask(file, uri, version); err != nil {
											fmt.Fprintf(os.Stderr, "Failed to add source annotation to Task YAML file %s: %v\n", file, err)
									}
							}
							return nil
					})
					if err != nil {
							fmt.Fprintf(os.Stderr, "Error traversing task directory %s: %v\n", taskDir, err)
					}
			}
	}
	return nil
}


// Function to add source annotation to Task YAML file
func addSourceAnnotationToTask(file, url, version string) error {
	// Read the Task YAML file
	data, err := ioutil.ReadFile(file)
	if err != nil {
			return err
	}

	// Parse JSON data
	var task map[string]interface{}
	if err := json.Unmarshal(data, &task); err != nil {
			return err
	}

	// Add source annotation to metadata
	metadata, ok := task["metadata"].(map[string]interface{})
	if !ok {
			return fmt.Errorf("metadata not found in Task YAML file: %s", file)
	}
	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
			annotations = make(map[string]interface{})
	}
	annotations["source"] = fmt.Sprintf("%s/releases/download/%s", url, version)
	metadata["annotations"] = annotations
	task["metadata"] = metadata

	// Marshal the updated data
	updatedData, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
			return err
	}

	// Rewrite the Task YAML file with updated metadata
	if err := ioutil.WriteFile(file, updatedData, 0644); err != nil {
			return err
	}

	return nil
}


func fetchAndExtract(path, url, version, resourceType string) error {
	resp, err := http.Get(url) // nolint:gosec,noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status error: %v", resp.StatusCode)
	}
	return untar(path, version, resourceType, resp.Body)
}

func untar(dst, version, resourceType string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil
		// return any other error
		case err != nil:
			return err
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		filename := filepath.Base(header.Name)
		targetFolder := filepath.Join(dst, filepath.Dir(header.Name), version)
		target := filepath.Join(targetFolder, filename)

		if resourceType != "" {
			if !strings.HasPrefix(header.Name, resourceType) {
				fmt.Fprintf(os.Stderr, "### Ignoring %s (type not %s)\n", header.Name, resourceType)
				continue
			}
		}

		if err := os.MkdirAll(targetFolder, os.ModePerm); err != nil {
			return err
		}
		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return err
				}
			}
		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(f, tr); err != nil { // nolint:gosec
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}

type Catalog struct {
	Resources map[string]Resource
}

type Resource map[string]string
