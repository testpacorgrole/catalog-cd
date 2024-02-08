package contract

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

const (
	// Version current contract version.
	Version = "v1"
	// Filename default contract file name.
	Filename = "catalog.yaml"
	// Resources default file name.
	ResourcesName = "resources.tar.gz"
	// SignatureExtension.
	SignatureExtension = "sig"
)

// Repository contains the general repository information, including metadata to categorize
// and describe the repository contents, objective, ecosystem, etc.
type Repository struct {
	// Description long description text.
	Description string `json:"description"`
}

// Catalog describes the contents of a repository part of a "catalog" of Tekton resources,
// including repository metadata, inventory of Tekton resources, test-cases and more.
type Catalog struct {
	Repository  *Repository  `json:"repository"`  // repository long description
	Attestation *Attestation `json:"attestation"` // software supply provenance
	Resources   *Resources   `json:"resources"`   // inventory of Tekton resources
}

// Contract contains a versioned catalog.
type Contract struct {
	file    string  // contract file full path
	Version string  `json:"version"` // contract version
	Catalog Catalog `json:"catalog"` // tekton resources catalog
}

// Print renders the YAML representation of the current contract.
func (c *Contract) Print() ([]byte, error) {
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Save saves the contract on the original file.
func (c *Contract) Save() error {
	if c.file == "" {
		return fmt.Errorf("contract file location is not set")
	}
	return c.SaveAs(c.file)
}

// SaveAs writes itself on the informed file path.
func (c *Contract) SaveAs(file string) error {
	payload, err := c.Print()
	if err != nil {
		return err
	}
	return os.WriteFile(file, payload, 0o644) // nolint: gosec
}

// NewContractEmpty instantiates a new Contract{} with empty attributes.
func NewContractEmpty() *Contract {
	return &Contract{
		Version: Version,
		Catalog: Catalog{
			Repository:  &Repository{},
			Attestation: &Attestation{},
			Resources: &Resources{
				Tasks:     []*TektonResource{},
				Pipelines: []*TektonResource{},
			},
		},
	}
}

// NewContractFromFile instantiates a new Contract{} from a YAML file.
func NewContractFromFile(location string) (*Contract, error) {
	// contract yaml file location
	var file string

	// when the location is a directory, it assumes the directory contains a default catalog
	// file name inside, otherwise the location is assumed to be the actual file
	info, _ := os.Stat(location)
	if info.IsDir() {
		file = path.Join(location, Filename)
	} else {
		file = location
	}

	payload, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return NewContractFromData(payload)
}

// NewContractFromURL instantiates a new Contract{} from a URL.
func NewContractFromURL(url string) (*Contract, error) {
	resp, err := http.Get(url) // nolint:gosec,noctx
	if err != nil {
		return nil, fmt.Errorf("could not load contract from %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %v", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not load contract from %s: %w", url, err)
	}
	return NewContractFromData(data)
}

// NewContractFromData instantiates a new Contract{} from a YAML payload.
func NewContractFromData(payload []byte) (*Contract, error) {
	c := Contract{}
	if err := yaml.Unmarshal(payload, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
