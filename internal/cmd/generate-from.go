package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/openshift-pipelines/catalog-cd/internal/catalog"
	"github.com/openshift-pipelines/catalog-cd/internal/config"
	"github.com/openshift-pipelines/catalog-cd/internal/contract"
	fc "github.com/openshift-pipelines/catalog-cd/internal/fetcher/config"
	"github.com/spf13/cobra"
)

// generateFromExternalOptions represents the "generate" subcommand to generate the signature of a resource file.
type generateFromExternalOptions struct {
	name                string // name of the repository to pull (a bit useless)
	url                 string // url of the repository to pull
	resourceType        string // type of resource to pull
	ignoreVersions      string // versions to ignore while pulling
	target              string // path to the folder where we want to generate the catalog
	catalogName         string // name of the contract file to pull (default catalog.yaml)
	resourceTarballName string // name of the resources file to pull (default resources.tar.gz)
}

const generateLongFromExternalDescription = `# catalog-cd generate-partial

Generates a partial file-based catalog in the target folder, based of a set of flags.

  $ catalog-cd generate-from \
      --name="foo" --url="https://github.com/openshift-pipelines/task-containers" \
      --type="tasks" \
      /path/to/catalog/target
`

func runGenerateFromExternal(_ context.Context, cfg *config.Config, args []string, o generateFromExternalOptions) error {
	if o.url == "" {
		return fmt.Errorf("flag --config is required")
	}
	if o.resourceType == "" {
		return fmt.Errorf("flag --resourceType is required")
	}

	if len(args) != 1 {
		return fmt.Errorf("you must specify a target to generate the catalog in")
	}
	o.target = args[0]
	cfg.Infof("Generating a partial catalog from %s (type: %s)\n", o.url, o.resourceType)
	ghclient, err := api.DefaultRESTClient()
	if err != nil {
		return err
	}

	name := o.name
	if name == "" {
		name = path.Base(o.url)
	}
	ignoreVersions := []string{}
	if o.ignoreVersions != "" {
		ignoreVersions = strings.Split(o.ignoreVersions, ",")
	}

	e := fc.External{
		Repositories: []fc.Repository{{
			Name:                 name,
			URL:                  o.url,
			IgnoreVersions:       ignoreVersions,
			CatalogName:          o.catalogName,
			ResourcesTarballName: o.resourceTarballName,
		}},
	}
	c, err := catalog.FetchFromExternals(e, ghclient)
	if err != nil {
		return err
	}

	return catalog.GenerateFilesystem(o.target, c, o.resourceType)
}

// NewCatalogGenerateFromExternalCmd instantiates the "generate" subcommand.
func NewCatalogGenerateFromExternalCmd(cfg *config.Config) *cobra.Command {
	o := generateFromExternalOptions{}
	cmd := &cobra.Command{
		Use:          "generate-from",
		Args:         cobra.ExactArgs(1),
		Long:         generateLongFromExternalDescription,
		Short:        "Generates a partial file-based catalog in the target folder, based of a set of flags.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateFromExternal(cmd.Context(), cfg, args, o)
		},
	}

	cmd.PersistentFlags().StringVar(&o.name, "name", "", "name of the repository to pull")
	cmd.PersistentFlags().StringVar(&o.url, "url", "", "url of the repository to pull")
	cmd.PersistentFlags().StringVar(&o.resourceType, "type", "", "type of resource to pull")
	cmd.PersistentFlags().StringVar(&o.ignoreVersions, "ignore-versions", "", "versions to ignore while pulling")
	cmd.PersistentFlags().StringVar(&o.catalogName, "catalog-name", contract.Filename, "contract name to pull")
	cmd.PersistentFlags().StringVar(&o.resourceTarballName, "resource-tarball-name", contract.ResourcesName, "resource file to pull")

	return cmd
}
