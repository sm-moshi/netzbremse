# netzbremse

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

- Woodpecker runs in-cluster from this repository.
- Pushes to `main` run validation and `semantic-release`.
- `semantic-release` creates Git tags and GitHub releases.
- Version tags (`v*`) trigger Woodpecker image publishing to:
  - `ghcr.io/sm-moshi/netzbremse-dashboard`
  - `ghcr.io/sm-moshi/netzbremse-measurement`

Required Woodpecker secrets:

- `github_username`
- `github_token`

The production deployment is owned by:

- `/Users/smeya/git/m0sh1.cc/infra/apps/user/netzbremse`
- `/Users/smeya/git/m0sh1.cc/infra/apps/cluster/cloudnative-pg`
- `/Users/smeya/git/m0sh1.cc/infra/apps/user/cilium-policies`
