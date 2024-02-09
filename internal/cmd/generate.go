package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/openshift-pipelines/catalog-cd/internal/catalog"
	"github.com/openshift-pipelines/catalog-cd/internal/config"
	fc "github.com/openshift-pipelines/catalog-cd/internal/fetcher/config"
	"github.com/spf13/cobra"
)

// generateOptions represents the "generate" subcommand to generate the signature of a resource file.
type generateOptions struct {
	config string // path for the catalog configuration file
	target string // path to the folder where we want to generate the catalog
}

const generateLongDescription = `# catalog-cd generate

Generates a file-based catalog in the target folder, based of a configuration file.

  $ catalog-cd generate \
      --config="/path/to/external.yaml" \
      /path/to/catalog/target
`

func runGenerate(_ context.Context, cfg *config.Config, args []string, o generateOptions) error {
	if o.config == "" {
		return fmt.Errorf("flag --config is required")
	}

	if len(args) != 1 {
		return fmt.Errorf("you must specify a target to generate the catalog in")
	}
	o.target = args[0]
	required := []string{
		o.config,
	}
	for _, f := range required {
		if _, err := os.Stat(f); err != nil {
			return err
		}
	}
	cfg.Infof("Generating a catalog from %s in %s\n", o.config, o.target)
	ghclient, err := api.DefaultRESTClient()
	if err != nil {
		return err
	}

	e, err := fc.LoadExternal(o.config)
	if err != nil {
		return err
	}
	c, err := catalog.FetchFromExternals(e, ghclient)
	if err != nil {
		return err
	}

	return catalog.GenerateFilesystem(o.target, c, "")
}

// NewCatalogGenerateCmd instantiates the "generate" subcommand.
func NewCatalogGenerateCmd(cfg *config.Config) *cobra.Command {
	o := generateOptions{}
	cmd := &cobra.Command{
		Use:          "generate",
		Args:         cobra.ExactArgs(1),
		Long:         generateLongDescription,
		Short:        "Generates a file-based catalog in the target folder, based of a configuration file.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd.Context(), cfg, args, o)
		},
	}

	cmd.PersistentFlags().StringVar(&o.config, "config", "./externals.yaml", "path of the catalog configuration file")

	return cmd
}
