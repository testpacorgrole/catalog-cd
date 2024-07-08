# Git Clone Step Action

This step action clones a git repository into a specified directory.

## Parameters

- `output-path`: The git repo will be cloned onto this path.
- `ssh-directory-path`: A .ssh directory with private key, known_hosts, config, etc.
- `basic-auth-path`: A directory path containing a .gitconfig and .git-credentials file.
- `ssl-ca-directory-path`: A directory containing CA certificates.
- `url`: Repository URL to clone from.
- `revision`: Revision to checkout. (branch, tag, sha, ref, etc...)
- `refspec`: Refspec to fetch before checking out revision.
- `submodules`: Initialize and fetch git submodules.
- `depth`: Perform a shallow clone, fetching only the most recent N commits.
- `sslVerify`: Set the `http.sslVerify` global git config.
- `crtFileName`: File name of mounted crt using ssl-ca-directory workspace.
- `subdirectory`: Subdirectory inside the `output` Workspace to clone the repo into.
- `sparseCheckoutDirectories`: Define the directory patterns to match or exclude when performing a sparse checkout.
- `deleteExisting`: Clean out the contents of the destination directory if it already exists before cloning.
- `httpProxy`: HTTP proxy server for non-SSL requests.
- `httpsProxy`: HTTPS proxy server for SSL requests.
- `noProxy`: Opt out of proxying HTTP/HTTPS requests.
- `verbose`: Log the commands that are executed during `git-clone`'s operation.
- `gitInitImage`: The image providing the git-init binary that this StepAction runs.
- `userHome`: Absolute path to the user's home directory.

## Results

- `commit`: The precise commit SHA that was fetched by this StepAction.
- `url`: The precise URL that was fetched by this StepAction.
- `committer-date`: The epoch timestamp of the commit that was fetched by this StepAction.
