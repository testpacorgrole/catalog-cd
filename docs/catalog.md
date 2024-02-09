# `catalog.yaml`

# Abstract

Describes what the file `catalog.yaml` contains and the use-cases indented for this **contract**. The file will be placed the release page of a repository containing Tekton Pipeline resources (Tasks and Pipelines).

The `catalog.yaml` goal is to serve as a blueprint to find the resources managed on the respective repository, as well to provide information for software supply chain attestation, and describe continuous integration test-cases. Usually, the `catalog.yaml` is created during a release on these repositories (using `catalog-cd release` or manually).
# Use-Cases

The file described on this document is meant to make possible the use-cases described below.

## Repository Root

The primary location for the `catalog.yaml` file is on a release page of the repository, describing all the elements, providing software supply chain attestation data and as well descring test cases, *for that release*.

### Release Artifacts

The `catalog.yaml` should be present on the repositories release payload, therefore when the maintainers decide to release a new version, the `catalog.yaml` is able to overwrite the entries on `.catalog.resources`.

This ability makes possible to template Tekton resources instead of the plain YAML files, and during the regular releases the resource are assembled.

## Continuous Integration

# `catalog.{yml,yaml}`

The file looks like the example below.

```yml
---
version: v1

catalog:
  repository:
    description: Tekton Task to interact with Git repositories
  attestation:
    publicKey: path/to/public.key
    annotations:
      team: tekton-ecosystem
  resources:
    tasks:
      - name: task-git
        version: "0.0.1"
        filename: path/to/resource.yaml
        checksum: resource-sha256-checksum
        signature: path/to/signature.sig
    pipelines: []
```

The support for the contract file is based on the `version` attribute, as this project moves forward we might change the attributes and the contract version marks breaking changes.

## Repository Metadata (`.catalog.repository`)

Attributes under `.catalog.repository` are meant to describe the repository containing Tekton resources, the `.description` should share a broad view of what the repository contains, what the user will find using the repository contents.

## Supply Chain Attestation (`.catalog.attestation`)

For the software supply chain security, the `.catalog.attestation` holds the elements needed to verify the authors signature. Initially it will contain the public key, either as a direct string or a file, and annotations for the verification processes.

## Tekton Pipeline Resources (`.catalog.resources`)

Under the `.catalog.resources` a inventory of all Tekton resources is recorded, all `.tasks` and `.pipelines` on the respective repository, or release payload, must be described here.

Each entry contains the following:

- `.name`: resource name, the Task's name or Pipeline's name
- `.version` (optional): the resource version, by default the repository's revision takes place
- `.filename`: relative path to the YAML resource file
- `.checksum`: sha256 sum, in order to validate the resource payload after network transfer.
- `.signature` (optional): relative path to the signature file, when empty it should search for the respective filename followed by the ".sig" extension, or the signature payload itself directly
