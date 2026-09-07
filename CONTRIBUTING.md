# Contributing

Use the Go version declared in `go.mod`. Keep the CLI small and its responsibilities
separate within the existing package. Preserve documented flags unless a change
is explicitly recorded in the changelog.

## Verification

Run the format, vet, test, race, and vulnerability checks in the
[README](README.md#local-development). Tests should cover observable behavior
and failure cases using temporary files and local servers. Do not make tests
depend on public hosts, external credentials, or live domain lists.

When changing CLI behavior, update the options, exit codes, and migration notes
together with the code. Keep examples limited to reserved names and controlled
local environments.

## Dependencies

Review dependency release notes before accepting updates. Run `go mod tidy`,
commit `go.mod` and `go.sum` together, and rerun verification. Go modules and
GitHub Actions have weekly Dependabot checks. The pinned `govulncheck` version in
CI and the README is maintained explicitly.

## Releases

Use `vX.Y.Z` tags and summarize user-visible changes in `CHANGELOG.md`. Review CLI
compatibility before choosing the version. Before a release, all CI jobs must
pass and the project owner must select the project license. No release is
published automatically by the current workflows.
