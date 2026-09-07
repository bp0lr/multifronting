# Changelog

## Unreleased

### Changed

- Require Go 1.27.1 and use `github.com/bp0lr/multifronting` as the module path.
- Update direct and indirect dependencies.
- Split the single-file implementation into focused files in the same package.
- Remove unused user-agent generation and obsolete `io/ioutil` usage.
- Replace the old bundled third-party domain lists with reserved sample names.

### Security and reliability

- Verify TLS certificates and hostnames by default; require TLS 1.2 or newer.
- Add `--ca-cert` for additional PEM trust roots without changing OS settings.
- Limit response bodies to 1 MiB and propagate read failures.
- Check request construction, standard-input errors, and file-write/close errors.
- Cancel active requests and unblock standard-input reads on Ctrl+C.
- Add local tests, CI for three operating systems, and dependency checks.

### CLI compatibility

- Invalid arguments now exit with status 2, runtime failures with 1, and Ctrl+C
  with 130. Previously many failures returned success silently.
- Worker counts outside 1-99 are errors instead of falling back silently to 50.
- A nonempty response marker is required.
- Input is validated before the output file is opened. Malformed input aborts
  the run rather than being silently skipped. Blank lines are ignored.
- Standard output is now plain matching hosts; diagnostics and progress use
  standard error. Progress no longer suppresses matching results.
- Output files retain append behavior and may contain partial results on errors.
- Existing invalid or untrusted TLS configurations now fail verification.
