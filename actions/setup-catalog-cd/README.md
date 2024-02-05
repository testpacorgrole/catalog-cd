# `setup-catalog-cd`

This actions install and setup `catalog-cd` in a GitHub Workflow.

## Usage

```yaml
- uses: openshift-pipelines/catalog-cd/actions/setup-catalog-cd@main
  with:
    # Version of catalog-cd to install (tip, latest-release, v0.1.0, etc.)
    version: 'latest-release'
```

If you use `tip` it will try to do a `go install` (from the `main`
branch) ; in order for this to work, you will need to use
`actions/setup-go` prior to this action.  

For any other value (`latest-release`, or any version number),
`setup-go` is not required.

## Examples

```yaml
      - uses: openshift-pipelines/catalog-cd/actions/setup-catalog-cd@main
      - run: catalog-cd version

      - name: Install v0.1.2 release
        uses: openshift-pipelines/catalog-cd/actions/setup-catalog-cd@main
        with:
          version: v0.1.2
```
