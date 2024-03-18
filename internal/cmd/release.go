package cmd

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift-pipelines/catalog-cd/internal/config"
	"github.com/openshift-pipelines/catalog-cd/internal/contract"
	"github.com/openshift-pipelines/catalog-cd/internal/resource"
	"github.com/spf13/cobra"
)

// releaseOptions creates a contract (".catalog.yaml") based on Tekton resources files.
type releaseOptions struct {
	version       string   // release version
	paths         []string // tekton resource paths
	output        string   // output path, where the contract and tarball will be written
	catalogName   string   // name for the catalog.yaml
	resourcesName string   // name for the resources tarball containing names
}

const releaseLongDescription = `# catalog-cd release

Creates a contract file (".catalog.yaml") for the Tekton resources specified on
the last argument(s), the contract is stored on the "--output" location, or by
default ".catalog.yaml" on the current directory.

The following examples will store the ".catalog.yaml" on the current directory, in
order to change its location see "--output" flag.

  # release all "*.yaml" files on the subdirectory
  $ catalog-cd release --version="0.0.1" path/to/tekton/files/*.yaml

  # release all "*.{yml|yaml}" files on the current directory
  $ catalog-cd release --version="0.0.1" *.yml *.yaml

  # release all "*.yml" and "*.yaml" files from the current directory
  $ catalog-cd release --version="0.0.1"

It always require the "--version" flag specifying the common revision for all
resources in scope.
`

func runRelease(_ context.Context, cfg *config.Config, args []string, o releaseOptions) error {
	// making sure the output flag is informed before attempt to search files
	if o.output == "" {
		return fmt.Errorf("--output flag is not informed")
	}
	o.paths = args
	if len(o.paths) == 0 {
		return fmt.Errorf("no tekton resource paths have been found")
	}
	fmt.Fprintf(cfg.Stream.Err, "# Found %d path to inspect!\n", len(o.paths))
	c := contract.NewContractEmpty()
	// going through the pattern slice collected before to select the tekton resource files
	// to be part of the current release, in other words, release scope
	fmt.Fprintf(cfg.Stream.Err, "# Scan Tekton resources on: %s\n", strings.Join(o.paths, ", "))
	for _, p := range o.paths {
		files, err := resource.Scanner(p)
		if err != nil {
			return err
		}

		for _, f := range files {
			fmt.Fprintf(cfg.Stream.Err, "# Loading resource file: %q\n", f)
			taskname := filepath.Base(filepath.Dir(f))
			resourceType, err := resource.GetResourceType(f)
			if err != nil {
				return err
			}
			resourceFolder := filepath.Join(o.output, strings.ToLower(resourceType)+"s", taskname)
			if err := os.MkdirAll(resourceFolder, os.ModePerm); err != nil {
				return err
			}
			if err := c.AddResourceFile(f, o.version); err != nil {
				if errors.Is(err, contract.ErrTektonResourceUnsupported) {
					return err
				}
				fmt.Fprintf(cfg.Stream.Err, "# WARNING: Skipping file %q!\n", f)
			}
			// Copy it to output
			if err := copyFile(f, filepath.Join(resourceFolder, filepath.Base(f))); err != nil {
				return err
			}
			readmeFile := filepath.Join(filepath.Dir(f), "README.md")
			if _, err := os.Stat(readmeFile); err == nil {
				// This is the README, copy it to output
				if err := copyFile(readmeFile, filepath.Join(resourceFolder, "README.md")); err != nil {
					return err
				}
				continue
			}
		}
	}

	catalogPath := filepath.Join(o.output, o.catalogName)
	fmt.Fprintf(cfg.Stream.Err, "# Saving release contract at %q\n", catalogPath)
	if err := c.SaveAs(catalogPath); err != nil {
		return err
	}

	// Create a tarball (without catalog.yaml
	tarball := filepath.Join(o.output, o.resourcesName)
	fmt.Fprintf(cfg.Stream.Err, "# Creating tarball at %q\n", tarball)
	return createTektonResourceArchive(tarball, o.catalogName, o.resourcesName, o.output)
}

// NewReleaseCmd instantiates the NewReleaseCmd subcommand and flags.
func NewReleaseCmd(cfg *config.Config) *cobra.Command {
	o := releaseOptions{}
	cmd := &cobra.Command{
		Use:          "release [flags] [glob|directory]",
		Short:        "Creates a contract for Tekton resource files",
		Long:         releaseLongDescription,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRelease(cmd.Context(), cfg, args, o)
		},
	}

	cmd.PersistentFlags().StringVar(&o.version, "version", "", "release version")
	cmd.PersistentFlags().StringVar(&o.output, "output", ".", "path to the release files (to attach to a given release)")
	cmd.PersistentFlags().StringVar(&o.catalogName, "catalog-name", contract.Filename, "name for the catalog.yaml file")
	cmd.PersistentFlags().StringVar(&o.resourcesName, "resources-tarball-name", contract.ResourcesName, "name for the catalog.yaml file")

	if err := cmd.MarkPersistentFlagRequired("version"); err != nil {
		panic(err)
	}

	return cmd
}

func createTektonResourceArchive(archiveFile, catalogFileName, resourcesFileName, output string) error {
	// Create output file
	out, err := os.Create(archiveFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create the archive
	return createArchive(output, catalogFileName, resourcesFileName, out)
}

func createArchive(output, catalogFileName, resourcesFileName string, buf io.Writer) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	return filepath.Walk(output, func(file string, fi os.FileInfo, err error) error {
		// return on any error
		if err != nil {
			return err
		}
		if filepath.Base(file) == catalogFileName || filepath.Base(file) == resourcesFileName {
			return nil
		}
		if fi.IsDir() || !fi.Mode().IsRegular() {
			return nil
		}
		return addToArchive(tw, file, output)
	})
}

func addToArchive(tw *tar.Writer, filename, output string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = strings.TrimPrefix(filename, filepath.Join(output, "/")+"/")

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	// Open the source file for reading
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Flush the destination file to ensure all data is written
	err = dstFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
