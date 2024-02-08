package catalog

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/openshift-pipelines/catalog-cd/internal/contract"
	"github.com/openshift-pipelines/catalog-cd/internal/fetcher"
	"github.com/openshift-pipelines/catalog-cd/internal/fetcher/config"
)

// Catalog represent the list of repositories from which we fetch informations
type Catalog struct {
	Repositories map[string]Repository
}

// Repository holds a map of version + "fetch information" useful to generate a catalog
type Repository map[string]Release

// Release holds information per release per repository
// It mainly holds the pre-loaded catalog information (containing the list of object published, their hash),
// as well as the URI to download the tarball containing those resources.
type Release struct {
	ResourcesURI string
	Catalog      contract.Catalog
}

func FetchFromExternals(e config.External, client *api.RESTClient) (Catalog, error) {
	c := Catalog{
		Repositories: map[string]Repository{},
	}
	for _, r := range e.Repositories {
		if r.Name == "" {
			// Name is empty, take the last part of the URL
			r.Name = filepath.Base(r.URL)
		}
		c.Repositories[r.Name] = Repository{}

		m, err := fetcher.FetchContractsFromRepository(r, client)
		if err != nil {
			return c, err
		}
		for _, v := range r.IgnoreVersions {
			// Remove ignored versions from map
			delete(m, v)
		}

		for version, contract := range m {
			resourcesDownloaldURI := fmt.Sprintf("%s/releases/download/%s/%s", r.URL, version, r.ResourcesTarballName)
			version = strings.TrimPrefix(version, "v")
			c.Repositories[r.Name][version] = Release{
				ResourcesURI: resourcesDownloaldURI,
				Catalog:      contract.Catalog,
			}
		}
	}
	return c, nil
}

func GenerateFilesystem(path string, c Catalog, resourceType string) error {
	for name, repository := range c.Repositories {
		fmt.Fprintf(os.Stderr, "# Fetching resources from %s\n", name)
		for version, release := range repository {
			fmt.Fprintf(os.Stderr, "## Fetching version %s\n", version)
			if err := fetchAndExtract(path, release, version, resourceType); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to fetch resource %s: %v, skipping\n", release.ResourcesURI, err)
				continue
			}
		}
	}
	return nil
}

func fetchAndExtract(path string, release Release, version, resourceType string) error {
	resp, err := http.Get(release.ResourcesURI) // nolint:gosec,noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status error: %v", resp.StatusCode)
	}
	// Let's get the file we want to fetch from the release object
	tektonResources := getResourcesFromType(release, resourceType)
	return untar(path, version, tektonResources, resp.Body)
}

// untar, filter and validate content
func untar(dst, version string, tektonResources map[string]contract.TektonResource, r io.Reader) error {
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
		versionnedFolder := filepath.Join(filepath.Dir(header.Name), version)
		targetFolder := filepath.Join(dst, versionnedFolder)
		target := filepath.Join(targetFolder, filename)

		tektonResource, ok := tektonResources[header.Name]
		if !ok && filename != "README.md" {
			fmt.Fprintf(os.Stderr, "### Ignoring %s (file not present in the catalog file)\n", header.Name)
			continue
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
		// FIXME: we probably can ignore this.
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

			h := sha256.New()
			r := io.TeeReader(tr, h)
			// copy over contents
			if _, err := io.Copy(f, r); err != nil { // nolint:gosec
				return err
			}
			sum := hex.EncodeToString(h.Sum(nil))
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()

			if filename != "README.md" {
				if tektonResource.Checksum != sum {
					fmt.Fprintf(os.Stderr, "%s checksum is different than the specified checksum in the catalog file: %s", sum, tektonResource.Checksum)
					// FIXME: maybe handle *all* file before erroring out ?
					return fmt.Errorf("invalid checksum for %s: %s != %s", filename, sum, tektonResource.Checksum)
				} else {
					fmt.Fprintf(os.Stderr, "✅ %s\n", tektonResource.Filename)
				}
			}
		}
	}
}

func getResourcesFromType(release Release, resourceType string) map[string]contract.TektonResource {
	m := map[string]contract.TektonResource{}
	switch resourceType {
	case "tasks":
		for _, r := range release.Catalog.Resources.Tasks {
			m[r.Filename] = *r
		}
	case "pipelines":
		for _, r := range release.Catalog.Resources.Tasks {
			m[r.Filename] = *r
		}
	case "":
		for _, r := range release.Catalog.Resources.Tasks {
			m[r.Filename] = *r
		}
		for _, r := range release.Catalog.Resources.Tasks {
			m[r.Filename] = *r
		}
	}
	return m
}
