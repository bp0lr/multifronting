# multifronting

A small Go CLI for studying domain fronting in infrastructure you own and control.
It reports whether a configured response marker is present. A match alone is not
proof of a security vulnerability; routing, caching, and shared content can affect
the result.

## Requirements

- [Go 1.27.1 or newer](https://go.dev/dl/), matching `go.mod`.
- A controlled test environment with valid HTTPS certificates.
- Windows, Linux, or macOS. CI builds and tests on all three platforms.

## Build and install

Build this checkout to use the unreleased changes documented here.

**Bash (Linux/macOS):**

```bash
go build -trimpath -o multifronting .
./multifronting --help
```

**PowerShell (Windows):**

```powershell
go build -trimpath -o multifronting.exe .
.\multifronting.exe --help
```

To install the current checkout into Go's binary directory:

```text
go install .
```

Once a release containing these changes is published, install or update from the
module path with:

```text
go install github.com/bp0lr/multifronting@latest
```

`@latest` resolves a published version; it does not install local edits. To find
the installation directory, run `go env GOBIN GOPATH`: Go uses `GOBIN` when set,
otherwise the `bin` directory under `GOPATH`. Add that directory to your `PATH`.

## Options

| Flag | Default | Meaning |
| --- | --- | --- |
| `-f`, `--fronturl` | Required | Your configured host, without a scheme or path. |
| `-n`, `--needle` | Required | Expected response text; empty or whitespace-only values are rejected. |
| `-u`, `--testurl` | Standard input | A single host from your test environment. |
| `-w`, `--workers` | `50` | Worker count, from `1` through `99`. |
| `-p`, `--proxy` | None | Explicit HTTP or HTTPS proxy URL. |
| `-o`, `--output` | None | Append matching hosts to a file. |
| `--ca-cert` | System roots only | PEM CA certificates to add to the system trust store for this process. |
| `-v`, `--verbose` | `false` | Write per-host diagnostics to standard error. |
| `--use-pb` | `false` | Write progress to standard error; suppress per-host diagnostics. |
| `-h`, `--help` | | Display usage without making network requests. |

Host arguments accept an optional port. Examples of the input format are
`service.test`, `localhost:8443`, and `[::1]:8443`. Schemes, paths, credentials,
queries, and fragments are rejected in host arguments.

When `--testurl` is omitted, standard input supplies one host per line and must
reach EOF before processing starts. LF and CRLF are supported; surrounding
whitespace and blank lines are ignored. Invalid lines and scanner errors stop
the run before the output file is opened. Lines must fit Go's default 64 KiB
scanner buffer. Inputs are held in memory, and duplicate lines are retained.

The [sample input](domains/example.test.txt) contains reserved `.test` names for
format illustration. It does not configure a working network environment. The
repository does not distribute live third-party target lists.

## Output and exit status

- Standard output contains matching hosts, one per line, without a `[+]` prefix.
- Standard error contains diagnostics, progress, and the final error summary.
- Progress never hides matching hosts from standard output.
- `--output` creates the file if needed and appends results. It does not truncate
  existing data or remove duplicates. Concurrent completion means result order
  is unspecified.
- Input validation happens before opening the output file. Failures during a
  run can leave partial output; consult the exit status before using the results.

| Exit code | Meaning |
| --- | --- |
| `0` | Completed without reported errors, including runs with no matches; also used for help. |
| `1` | Input, certificate, request, or output failure. Results may be incomplete. |
| `2` | Invalid command-line arguments. |
| `130` | Interrupted with Ctrl+C. |

## TLS and resource limits

Certificate-chain and hostname verification are enabled, with TLS 1.2 as the
minimum version. For an internal CA, use `--ca-cert` with its PEM certificate
file. This adds trust for the current client only and does not change the OS
certificate store. Hostname validation remains enabled. There is no insecure
verification bypass.

Requests have a five-second total timeout. Redirects are returned without being
followed. Response bodies are limited to 1 MiB after any automatic decompression;
larger responses produce an error. Ctrl+C cancels in-flight requests and releases
a pending standard-input read.

## Local development

The project uses a single package with separate files for configuration, input,
HTTP transport, the existing response check, output, and CLI lifecycle.

```text
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
```

Tests use temporary files and local `httptest` servers. They do not query public
hosts. The race detector requires a supported C toolchain; CI runs it on Linux.
Dependency downloads and `govulncheck` require internet access. The latter checks
the dependency graph against Go's public vulnerability database.

CI also checks formatting and module consistency and builds on Linux, macOS, and
Windows. Dependency updates are proposed weekly through Dependabot.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| `x509: certificate signed by unknown authority` | Supply the internal CA with `--ca-cert`, or correct the server's certificate chain. |
| Certificate hostname mismatch | Correct the certificate or the host configuration. Adding a CA does not bypass hostname checks. |
| Invalid host argument | Remove the scheme and path; use brackets around IPv6 addresses. |
| No input hosts / waiting for input | Provide input and EOF, or use the single-host option. |
| Output cannot be written | Check the parent directory, permissions, and available disk space. Parent directories are not created automatically. |
| Response exceeds the byte limit | Use a small, controlled test response. |

## Changes and licensing

See [CHANGELOG.md](CHANGELOG.md) for compatibility changes and
[CONTRIBUTING.md](CONTRIBUTING.md) for the development and release workflow.
No project license has been selected yet.
