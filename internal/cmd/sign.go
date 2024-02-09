package cmd

import (
	"context"
	"fmt"

	"github.com/openshift-pipelines/catalog-cd/internal/attestation"
	"github.com/openshift-pipelines/catalog-cd/internal/config"
	"github.com/openshift-pipelines/catalog-cd/internal/contract"
	"github.com/spf13/cobra"
)

// signOptions subcommand "sign" to handles signing contract resources.
type signOptions struct {
	c *contract.Contract // catalog contract instance

	privateKey string // private key location
}

const signLongDescription = `# catalog-cd sign

Sign the catalog contract resources on the informed directory, or catalog file. By default it
assumes the current directory.

To sign the resources the subcommand requires a private-key ("--private-key" flag), and may
ask for the password when trying to interact with a encripted key.
`

func runSign(_ context.Context, cfg *config.Config, args []string, o signOptions) error {
	var err error
	o.c, err = LoadContractFromArgs(args)
	if err != nil {
		return err
	}
	helper, err := attestation.NewAttestation(o.privateKey)
	if err != nil {
		return err
	}
	if err = o.c.SignResources(func(payladPath, outputSignature string) error {
		fmt.Fprintf(cfg.Stream.Err, "# Signing resource %q on %q...\n", payladPath, outputSignature)
		return helper.Sign(payladPath, outputSignature)
	}); err != nil {
		return err
	}
	return o.c.Save()
}

// NewSignCmd instantiate the SignCmd and flags.
func NewSignCmd(cfg *config.Config) *cobra.Command {
	o := signOptions{}
	cmd := &cobra.Command{
		Use:          "sign [flags]",
		Short:        "Signs Tekton Pipelines resources",
		Long:         signLongDescription,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSign(cmd.Context(), cfg, args, o)
		},
	}

	cmd.PersistentFlags().StringVar(&o.privateKey, "private-key", "", "private key file location")

	if err := cmd.MarkPersistentFlagRequired("private-key"); err != nil {
		panic(err)
	}

	return cmd
}
