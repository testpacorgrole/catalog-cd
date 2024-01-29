// Package config holds the "configuration" that is used for fetching information from external repositories.
package config

import (
	"fmt"
	"os"

	"github.com/openshift-pipelines/catalog-cd/internal/contract"
	"sigs.k8s.io/yaml"
)

// External is a representation of the configuration for specifying repositories we have to pull from
type External struct {
	// Repositories defines the repositories to pull from
	Repositories []Repository
}

// Repository represent a git repository
type Repository struct {
	Name string
	URL  string
	// Type defines the type to fetch (Task, Pipeline, …)
	Types                []string
	IgnoreVersions       []string `json:"ignore-versions"`
	CatalogName          string   `json:"catalog-name"`
	ResourcesTarballName string   `json:"resources-tarball-name"`
}

// setDefaults sets the default values for the configuration
func setDefaults(e External) External {
	for i, r := range e.Repositories {
		if r.CatalogName == "" {
			r.CatalogName = contract.Filename
		}
		if r.ResourcesTarballName == "" {
			r.ResourcesTarballName = contract.ResourcesName
		}
		e.Repositories[i] = r
	}
	return e
}

func LoadExternal(filename string) (External, error) {
	var c External
	data, err := os.ReadFile(filename)
	if err != nil {
		return External{}, fmt.Errorf("Could not load external configuration from %s: %w", filename, err)
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return External{}, fmt.Errorf("Could not load external configuration from %s: %w", filename, err)
	}
	c = setDefaults(c)
	return c, nil
}
