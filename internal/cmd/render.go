package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/openshift-pipelines/catalog-cd/internal/config"
	"github.com/openshift-pipelines/catalog-cd/internal/render"
	"github.com/spf13/cobra"
)

const renderLongDescription = `# catalog-cd render

Renders the informed Tekton resource file as markdown, focusing on the most important attributes
which should always be part of the Task documentation.

The markdown generated contains the Workspaces, Params and Results formated as a mardown tables.
`

func runRender(_ context.Context, cfg *config.Config, args []string) error {
	var resource string
	if len(args) != 1 {
		return fmt.Errorf("you must inform a single argument (%d)", len(args))
	}
	resource = args[0]
	if _, err := os.Stat(resource); err != nil {
		return err
	}
	md, err := render.NewMarkdown(cfg, resource)
	if err != nil {
		return err
	}
	return md.Render()
}

// NewRenderCmd instantiate the "render" subcommand.
func NewRenderCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Renders the informed Tekton resource file as markdown",
		Long:  renderLongDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRender(cmd.Context(), cfg, args)
		},
	}
	return cmd
}
