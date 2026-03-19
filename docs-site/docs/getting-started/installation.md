# Installation

## Prerequisites

- **Terminal emulator** with 256-color support (most modern terminals)
- For source builds: **Go 1.25.4+**

## Install Methods

=== "Homebrew (recommended)"

    ```bash
    brew install arcaven/tap/threedoors
    ```

    This installs the latest stable release. Runs as `threedoors`.

=== "Alpha Channel"

    Latest development builds from `main`:

    ```bash
    brew install arcaven/tap/threedoors-a
    ```

    Runs as `threedoors-a`. Both stable and alpha can be installed side-by-side.

=== "Pre-built Binary"

    Download from [GitHub Releases](https://github.com/arcaven/ThreeDoors/releases). Binaries are available for macOS (Apple Silicon) and Linux (x86_64). macOS binaries are code-signed and Apple-notarized.

    ```bash
    chmod +x threedoors-*
    mv threedoors-darwin-arm64 /usr/local/bin/threedoors   # adjust for your platform
    ```

=== "Go Install"

    ```bash
    go install github.com/arcaven/ThreeDoors/cmd/threedoors@latest
    ```

    Requires Go 1.25.4 or later.

=== "Build from Source"

    ```bash
    git clone https://github.com/arcaven/ThreeDoors.git
    cd ThreeDoors
    just build
    # Binary at bin/threedoors
    ```

    Requires Go 1.25.4+ and [`just`](https://github.com/casey/just).

## Verify Installation

```bash
threedoors --version
```

## Next Steps

Head to the [Quick Start](quickstart.md) to launch ThreeDoors and complete your first task.
