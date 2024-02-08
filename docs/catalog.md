`catalog.yaml`

# Abstract

Describes what the file `catalog.yaml` contains and the use-cases indented for this **contract**. The file will be placed on the root of a CVS repository containing Tekton Pipeline resources (Tasks and Pipelines).

The `catalog.yaml` goal is to serve as a blueprint to find the resources managed on the respective repository, as well to provide information for software supply chain attestation, and describe continuous integration test-cases.

During the release of these repositories the `catalog.yaml` is added to the payload in order to describe the Tekton resource artifacts.

# Use-Cases

The file described on this document is meant to make possible the use-cases described below.

## Repository Root

The primary location for the `catalog.yaml` file is on the root of the (Git?) repository, describing all the elements, providing software supply chain attestation data and as well descring test cases.

For repositories containing the direct YAML payload of Tekton resource files stored, the file will also contain `.catalog.resources` entries, reflecting the location of the data.

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
        sha256sum: resource-sha256-checksum
        tri: git://openshift-pipelines/task-git@0.0.1
        bundle: ghcr.io/openshift-pipelines/task-git:0.0.1-bundle
        signature: path/to/signature.sig
    pipelines: []
```

The support for the contract file is based on the `version` attribute, as this project moves forward we might change the attributes and the contract version marks breaking changes.

## Repository Metadata (`.catalog.repository`)

Attributes under `.catalog.repository` are meant to describe the repository containing Tekton resources, the `.description` should share a broad view of what the repository contains, what the user will find using the repository contents.

## Supply Chain Attestation (`.catalog.attestation`)

For the software supply chain security, the `.catalog.attestation` hols the elements needed to verify the authors signature. Initially it will contain the public key, either as a direct string or a file, and annotations for the verification processes.

## Tekton Pipeline Resources (`.catalog.resources`)

Under the `.catalog.resources` a inventory of all Tekton resources is recorded, all `.tasks` and `.pipelines` on the respective repository, or release payload, must be described here.

Each entry contains the following:

- `.name`: resource name, the Task's name or Pipeline's name
- `.version` (optional): the resource version, by default the repository's revision takes place
- `.filename`: relative path to the YAML resource file
- `.bundle` (optional): shows the respective OCI bundle image name, where the current resource is stored
- `.signature` (optional): relative path to the signature file, when empty it should search for the respective filename followed by the ".sig" extension, or the signature payload itself directly
