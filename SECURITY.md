# Security Policy

## Supported Versions

Only the latest release on `main` is supported with security updates. Alpha releases are not covered by security support.

| Version | Supported |
|---------|-----------|
| Latest release (main) | Yes |
| Alpha releases | No |
| Older releases | No |

## Reporting a Vulnerability

**Do NOT open a public GitHub issue for security vulnerabilities.**

Please use one of the following methods:

1. **GitHub Private Vulnerability Reporting** (preferred): Go to the [Security tab](https://github.com/arcavenae/ThreeDoors/security) and select "Report a vulnerability"
2. **GitHub Issue** (for non-sensitive security questions): Open an issue with the `security` label

We will acknowledge your report within **7 days** and work with you to coordinate disclosure timing.

## Scope

ThreeDoors is a **local-first** application. This shapes its security profile:

- **No network services by default** — ThreeDoors does not listen on any ports or make network requests unless an integration adapter is configured
- **Local data storage** — Task data is stored as YAML files and JSONL session logs on the user's machine
- **Integration tokens** — Adapters for external services (Jira, GitHub, Apple Notes, etc.) use user-provided tokens stored locally on disk. ThreeDoors never transmits these tokens to third parties
- **No telemetry** — ThreeDoors does not collect usage data, analytics, or error reports

## Security Design Principles

- **Local-first:** Data never leaves the user's machine unless the user explicitly configures an integration
- **No cloud sync:** There is no built-in cloud synchronization
- **No telemetry:** No analytics, error reporting, or phone-home behavior
- **Atomic writes:** File persistence uses atomic write operations (write to temp file, fsync, rename) to prevent data corruption
- **Minimal dependencies:** The project prefers the Go standard library over third-party packages to minimize supply chain risk
