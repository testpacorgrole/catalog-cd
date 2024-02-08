package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/openshift-pipelines/catalog-cd/internal/config"
	"github.com/openshift-pipelines/catalog-cd/internal/runner"
	"github.com/spf13/cobra"
	tkncli "github.com/tektoncd/cli/pkg/cli"
)

// Version is provided at compile-time.
var Version string

func NewRootCmd(stream *tkncli.Stream) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:  "catalog-cd",
		Long: `TODO`,
	}

	cfg := config.NewConfigWithFlags(stream, rootCmd.PersistentFlags())

	rootCmd.AddCommand(runner.NewRunner(cfg, NewRenderCmd()).Cmd())
	rootCmd.AddCommand(runner.NewRunner(cfg, NewVerifyCmd()).Cmd())
	rootCmd.AddCommand(runner.NewRunner(cfg, NewReleaseCmd()).Cmd())
	rootCmd.AddCommand(runner.NewRunner(cfg, NewSignCmd()).Cmd())

	rootCmd.AddCommand(CatalogCmd(cfg))

	rootCmd.AddCommand(versionCmd(cfg))

	return rootCmd
}

func versionCmd(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print catalog-cd version",
		Long:  "Print catalog-cd version",
		RunE: func(_ *cobra.Command, args []string) error {
			v := version()
			if v == "" {
				fmt.Fprintln(os.Stderr, "could not determine build information")
			} else {
				fmt.Fprintln(os.Stderr, v)
			}
			return nil
		},
	}

	return cmd
}

func version() string {
	if Version == "" {
		i, ok := debug.ReadBuildInfo()
		if !ok {
			return ""
		}
		Version = i.Main.Version
	}
	return Version
}
