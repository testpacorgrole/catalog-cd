package cmd

import (
	"github.com/openshift-pipelines/catalog-cd/internal/config"
	"github.com/spf13/cobra"
)

const catalogLongDescription = `# catalog-cd catalog

Group of commands to manage a catalog, from generating a full or partial catalog to generate a
GitHub Action matrix strategy compatible json.
`

func CatalogCmd(cfg *config.Config) *cobra.Command {
	catalogCmd := &cobra.Command{
		Use:   "catalog",
		Short: `Catalog management commands.`,
		Long:  catalogLongDescription,
	}

	catalogCmd.AddCommand(NewCatalogGenerateCmd(cfg))
	catalogCmd.AddCommand(NewCatalogGenerateFromExternalCmd(cfg))
	catalogCmd.AddCommand(NewCatalogExternalsCmd(cfg))

	return catalogCmd
}
