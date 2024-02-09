package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/openshift-pipelines/catalog-cd/internal/config"
	fc "github.com/openshift-pipelines/catalog-cd/internal/fetcher/config"
	"github.com/spf13/cobra"
)

// externalsOptions represents the "externals" subcommand to externals the signature of a resource file.
type externalsOptions struct {
	config string // path for the catalog configuration file
}

const externalsLongDescription = `# catalog-cd externals

Generate a GitHub matrix strategy-compatible json from an externals.yaml file.

  $ catalog-cd catalog externals --config=./externals.yaml
`

type GitHubRunObject struct {
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	Type                 string `json:"type"`
	IgnoreVersions       string `json:"ignoreVersions"`
	CatalogName          string `json:"catalog-name"`
	ResourcesTarballName string `json:"resources-tarball-name"`
}

type GitHubMatrixObject struct {
	Include []GitHubRunObject `json:"include"`
}

func runCatalogExternals(_ context.Context, cfg *config.Config, args []string, o externalsOptions) error {
	required := []string{
		o.config,
	}
	for _, f := range required {
		if _, err := os.Stat(f); err != nil {
			return err
		}
	}
	if o.config == "" {
		return fmt.Errorf("flag --config is required")
	}

	if len(args) != 0 {
		return fmt.Errorf("externals takes no argument")
	}
	e, err := fc.LoadExternal(o.config)
	if err != nil {
		return err
	}
	m := GitHubMatrixObject{}
	for _, repository := range e.Repositories {
		types := repository.Types
		if len(types) == 0 {
			types = []string{"tasks", "pipelines"}
		}
		ignoreVersions := ""
		if len(repository.IgnoreVersions) > 0 {
			ignoreVersions = strings.Join(repository.IgnoreVersions, ",")
		}
		for _, t := range types {
			name := repository.Name
			if name == "" {
				name = path.Base(repository.URL)
			}
			o := GitHubRunObject{
				Name:                 name,
				URL:                  repository.URL,
				Type:                 t,
				IgnoreVersions:       ignoreVersions,
				CatalogName:          repository.CatalogName,
				ResourcesTarballName: repository.ResourcesTarballName,
			}
			m.Include = append(m.Include, o)
		}
	}
	j, err := json.Marshal(m)
	if err != nil {
		return err
	}
	fmt.Fprintf(cfg.Stream.Out, "%s\n", j)
	return nil
}

// NewCatalogExternalsCmd instantiates the "externals" subcommand.
func NewCatalogExternalsCmd(cfg *config.Config) *cobra.Command {
	o := externalsOptions{}
	cmd := &cobra.Command{
		Use:          "externals",
		Args:         cobra.ExactArgs(0),
		Long:         externalsLongDescription,
		Short:        "Generate a GitHub matrix strategy-compatible json from an externals.yaml file.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCatalogExternals(cmd.Context(), cfg, args, o)
		},
	}

	cmd.PersistentFlags().StringVar(&o.config, "config", "./externals.yaml", "path of the catalog configuration file")

	return cmd
}
