# netzbremse

[![push-status](https://ci.m0sh1.cc/api/badges/9/status.svg)](https://ci.m0sh1.cc/repos/9)
[![tag-status](https://ci.m0sh1.cc/api/badges/9/status.svg?event=tag)](https://ci.m0sh1.cc/repos/9)

Forked `netzbremse` application for `m0sh1.cc`.

This repository replaces the original shared-filesystem contract with a
CNPG-backed PostgreSQL contract so the measurement and dashboard pods can run
independently in the `apps` namespace.

Current scope:

- `cmd/measurement`: measurement worker skeleton with JSON import support
- `cmd/dashboard`: simple HTTP dashboard backed by PostgreSQL
- `db/migrations`: schema used by both services
- `Dockerfile.measurement`: image for the measurement worker
- `Dockerfile.dashboard`: image for the dashboard service

CI and release flow:

- Woodpecker runs in-cluster directly against `sm-moshi/netzbremse` on GitHub.
- Pushes to `main` run validation and `semantic-release`.
- `semantic-release` creates GitHub tags and releases directly on
  `https://github.com/sm-moshi/netzbremse`.
- Version tags (`v*`) trigger Woodpecker image publishing to:
  - `ghcr.io/sm-moshi/netzbremse-dashboard`
  - `ghcr.io/sm-moshi/netzbremse-measurement`
- The image builds are pinned to Docker Hardened Images:
  - `dhi.io/golang:1.26.1-debian13-dev` for the Go builder stages
  - `dhi.io/debian-base:trixie-debian13` for the dashboard runtime
  - `dhi.io/node:24.14.0-debian13-dev` for the measurement runtime
- The measurement runtime intentionally stays on the hardened Node `dev`
  variant because Puppeteer still needs Debian package installation in the
  final image for browser system libraries.
- The default Woodpecker badge tracks push pipelines on repo `9`.
- The tag badge tracks tag-publish pipelines on repo `9`.

Required Woodpecker secrets:

- `github_username`
- `github_token`
- `dhi_username`
- `dhi_token`

The `dhi_*` secrets are the Docker Hardened Images pull credentials used for:

- `dhi.io/golang:1.26.1-debian13-dev`
- `dhi.io/node:24.14.0-debian13-dev`
- `dhi.io/debian-base:trixie-debian13`

The production deployment is owned by:

- `/Users/smeya/git/m0sh1.cc/infra/apps/user/netzbremse`
- `/Users/smeya/git/m0sh1.cc/infra/apps/cluster/cloudnative-pg`
- `/Users/smeya/git/m0sh1.cc/infra/apps/user/cilium-policies`
