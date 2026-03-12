# Security Policy

## Scope

This repository contains the application code, container builds, and release
automation for the `netzbremse` measurement and dashboard services.

This policy covers security issues in:

- the Go services under `cmd/` and `internal/`
- the Puppeteer-based browser runner under `scripts/`
- the Docker build definitions
- the release automation and published container images for this repository

If a report is primarily about third-party infrastructure, hosting, or a
deployment outside this repository, it may need to be handled elsewhere. If you
are unsure, please use the private reporting channel first and we will triage
it.

## Supported Versions

Security fixes are applied only to the actively maintained code line. Older
releases are not guaranteed to receive backports, and maintainers may ask users
to upgrade to a newer release to receive a fix.

| Version | Supported |
| ------- | --------- |
| `main` | Yes |
| latest published release | Yes |
| older releases and superseded tags | No |

## Reporting a Vulnerability

Please **do not** open public GitHub issues, pull requests, or discussions for
security vulnerabilities.

Use GitHub's private vulnerability reporting for this repository:

- **Private report**: <https://github.com/sm-moshi/netzbremse/security/advisories/new>
- **Security advisories**: <https://github.com/sm-moshi/netzbremse/security/advisories>

If your finding involves leaked credentials, tokens, or other active secrets,
please say so clearly in the report so containment and rotation can begin
quickly.

## What to Include

To help triage quickly, include as much of the following as you can:

- affected component, endpoint, image, or workflow step
- affected version, tag, commit SHA, or container digest
- deployment context, including whether the issue affects `measurement`,
  `dashboard`, or both
- reproduction steps or a proof of concept
- expected impact, including confidentiality, integrity, or availability risk
- any suggested remediation or mitigating controls

Please avoid sending plaintext credentials unless there is no safer way to
demonstrate the issue. Redacted examples are preferred where possible.

## Disclosure Policy

We follow coordinated vulnerability disclosure.

Our usual process is:

1. acknowledge receipt of the report
2. validate the issue and assess impact
3. develop and test a fix or mitigation
4. prepare a release and, where appropriate, a GitHub Security Advisory
5. publish details after a fix is available and affected users have had a
   reasonable opportunity to update

We aim to acknowledge valid reports within 5 business days and will try to keep
reporters informed as triage and remediation progress.

## Release and Supply-Chain Notes

This project publishes container images to GHCR and mirrors them to Harbor.
Release pipelines are intended to scan and sign published images and attach
software supply-chain metadata.

Those controls reduce risk, but they do not replace responsible disclosure. If
you find a vulnerability in application code, dependencies, images, build
steps, signatures, SBOM generation, or attestation flow, please report it
privately through the security advisory flow above.

## Bug Bounties and Rewards

This project does not currently operate a paid bug bounty programme. We still
greatly appreciate responsible reports and will credit reporters in public
advisories where appropriate and where they would like to be named.

## Security Best Practices for Contributors

When contributing to this repository:

- never commit secrets, tokens, passwords, or private keys
- avoid adding plaintext test credentials to fixtures or examples
- keep Go, Node.js, and container dependencies current
- prefer least-privilege defaults in application and CI changes
- treat browser automation inputs and external command execution as
  security-sensitive surfaces

## Non-Security Reports

For ordinary defects, feature requests, usage questions, and operational
problems that do not have a plausible security impact, please use the normal
public issue tracker or discussion channels instead of the private
vulnerability reporting flow.
