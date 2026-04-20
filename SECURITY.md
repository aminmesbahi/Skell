# Security Policy

## Supported Versions

Only the latest minor release receives security fixes.

| Version | Supported |
| ------- | --------- |
| 0.1.x (latest) | ✅ |
| < 0.1.5 | ❌ |

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report vulnerabilities privately using one of the following methods:

- **GitHub Private Reporting:** [Report a vulnerability](https://github.com/aminmesbahi/skell/security/advisories/new) *(preferred)*
- **Email:** Contact the maintainer directly via GitHub profile

### What to include

Please provide as much of the following as possible:

- Description of the vulnerability and its potential impact
- Steps to reproduce or a proof-of-concept
- Affected version(s)
- Any suggested mitigations

### What to expect

| Timeline | Action |
| -------- | ------ |
| Within **3 business days** | Acknowledgement of your report |
| Within **14 days** | Initial assessment and severity classification |
| Within **90 days** | Patch release (or agreed disclosure date) |

We follow a **coordinated disclosure** model. Once a fix is released, a [GitHub Security Advisory](https://github.com/aminmesbahi/skell/security/advisories) will be published and a CVE requested if applicable.

## Security Scanning

This project uses the following automated security tooling on every commit and weekly:

| Tool | Purpose |
| ---- | ------- |
| [CodeQL](https://codeql.github.com) | Static analysis — detects CWEs and security anti-patterns |
| [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) | Known CVEs in Go dependencies (Go vulnerability database) |
| [nancy](https://github.com/sonatype-nexus-community/nancy) | OSS Index dependency audit |
| [Dependabot](https://docs.github.com/en/code-security/dependabot) | Automated dependency version updates |

## Scope

The following are **in scope** for vulnerability reports:

- The `skell` CLI binary
- The `internal/` packages
- Dependency vulnerabilities with a direct exploit path

The following are **out of scope**:

- Vulnerabilities in skill files (SKILL.md) fetched from third-party registries — those are the responsibility of the registry owner
- Issues requiring physical access to the machine
- Social engineering attacks
