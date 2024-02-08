package catalog_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/openshift-pipelines/catalog-cd/internal/catalog"
	"github.com/openshift-pipelines/catalog-cd/internal/contract"
	"github.com/openshift-pipelines/catalog-cd/internal/fetcher/config"
	"gopkg.in/h2non/gock.v1"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/golden"
)

func TestFetchFromExternal(t *testing.T) {
	t.Cleanup(gock.Off)

	repo := config.Repository{
		Name: "golang-task",
		URL:  "https://github.com/shortbrain/golang-tasks",
	}
	r := strings.TrimPrefix(repo.URL, "https://github.com/")

	gock.New("https://api.github.com").
		Get(fmt.Sprintf("repos/%s/releases", r)).
		Reply(200).
		File("testdata/releases.yaml")
	gock.New("https://github.com").
		Get(fmt.Sprintf("%s/releases/download/v1.0.0/catalog.yaml", r)).
		Reply(200).
		File("testdata/catalog.simple.yaml")

	client, err := api.DefaultRESTClient()
	if err != nil {
		t.Fatal(err)
	}
	e := config.External{
		Repositories: []config.Repository{{
			Name:                 "sbr-golang",
			URL:                  "https://github.com/shortbrain/golang-tasks",
			Types:                []string{"tasks"},
			CatalogName:          "catalog.yaml",
			ResourcesTarballName: "resources.tar.gz",
		}},
	}
	c, err := catalog.FetchFromExternals(e, client)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Repositories) != 1 {
		t.Fatalf("Should have created a catalog with only 1 repository, got %d: %v", len(c.Repositories), c.Repositories)
	}
}

func TestGenerateFilesystem(t *testing.T) {
	t.Cleanup(gock.Off)

	gock.New("https://fake.host").
		Get("resources.tar.gz").
		Reply(200).
		File("testdata/resources.tar.gz")

	dir := fs.NewDir(t, "catalog")
	defer dir.Remove()

	c := catalog.Catalog{
		Repositories: map[string]catalog.Repository{
			"sbr-golang": map[string]catalog.Release{
				"0.5.0": {
					ResourcesURI: "https://fake.host/resources.tar.gz",
					Catalog: contract.Catalog{
						Resources: &contract.Resources{
							Tasks: []*contract.TektonResource{{
								Name:     "go-crane-image",
								Version:  "0.5.0",
								Filename: "tasks/go-crane-image/go-crane-image.yaml",
								Checksum: "9b1f8e2ecbb5795727de93a6b95bbed2a4f44871f0f0ded6a2d8a04b2283a2b9",
							}, {
								Name:     "go-ko-image",
								Version:  "0.5.0",
								Filename: "tasks/go-ko-image/go-ko-image.yaml",
								Checksum: "e84e01f61a25aee509a4e3513b19f8f33a865eed60fd17647b56df8b716edfde",
							}},
							Pipelines: []*contract.TektonResource{},
						},
					},
				},
			},
		},
	}
	err := catalog.GenerateFilesystem(dir.Path(), c, "")
	if err != nil {
		t.Fatal(err)
	}
	expected := fs.Expected(t, fs.WithDir("tasks",
		fs.WithDir("go-crane-image",
			fs.WithDir("0.5.0",
				fs.WithFile("go-crane-image.yaml", "", fs.WithBytes(golden.Get(t, "tasks/go-crane-image/go-crane-image.yaml"))),
				fs.WithFile("README.md", "", fs.WithBytes(golden.Get(t, "tasks/go-crane-image/README.md"))),
			),
		),
		fs.WithDir("go-ko-image",
			fs.WithDir("0.5.0",
				fs.WithFile("go-ko-image.yaml", "", fs.WithBytes(golden.Get(t, "tasks/go-ko-image/go-ko-image.yaml"))),
				fs.WithFile("README.md", "", fs.WithBytes(golden.Get(t, "tasks/go-ko-image/README.md"))),
			),
		),
	))

	assert.Assert(t, fs.Equal(dir.Path(), expected))
}
