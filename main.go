package main

import (
	"os"

	"github.com/openshift-pipelines/catalog-cd/internal/cmd"
	tkncli "github.com/tektoncd/cli/pkg/cli"
)

func main() {
	rootCmd := cmd.NewRootCmd(&tkncli.Stream{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
