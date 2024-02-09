package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/openshift-pipelines/catalog-cd/internal/attestation"
	"github.com/openshift-pipelines/catalog-cd/internal/config"
	"github.com/openshift-pipelines/catalog-cd/internal/contract"
	"github.com/spf13/cobra"
)

// verifyOptions represents the "verify" options to verify the signature of a resource file.
type verifyOptions struct {
	c         *contract.Contract
	publicKey string // path to the public key file
}

const verifyLongDescription = `# catalog-cd verify

Verifies the signature of all resources described on the contract. The subcommand takes
either a contract file as argument, or a directory containing the contract using default
name. By default it searches the current directory.

In order to verify the signature the public-key is required, it's specified either on the
catalog contract, or using the flag "--public-key".
`

func runVerify(ctx context.Context, cfg *config.Config, args []string, o verifyOptions) error {
	var err error
	o.c, err = LoadContractFromArgs(args)
	if err != nil {
		return err
	}
	if o.publicKey == "" {
		o.publicKey, err = o.c.GetPublicKey()
		if err != nil {
			return err
		}
	}
	cfg.Infof("# Public-Key: %q\n", o.publicKey)

	helper, err := attestation.NewAttestation(o.publicKey)
	if err != nil {
		return err
	}
	return o.c.VerifyResources(ctx, func(ctx context.Context, blobRef, sigRef string) error {
		fmt.Fprintf(os.Stderr, "# Verifying resource %q against signature %q...\n", blobRef, sigRef)
		return helper.Verify(ctx, blobRef, sigRef)
	})
}

// NewVerifyCmd instantiates the "verify" subcommand.
func NewVerifyCmd(cfg *config.Config) *cobra.Command {
	o := verifyOptions{}

	cmd := &cobra.Command{
		Use:          "verify",
		Args:         cobra.ExactArgs(1),
		Long:         verifyLongDescription,
		Short:        "Verifies the resource file signature",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd.Context(), cfg, args, o)
		},
	}
	cmd.PersistentFlags().StringVar(&o.publicKey, "public-key", "", "path to the public key file")
	return cmd
}
