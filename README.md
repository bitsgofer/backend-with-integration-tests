An example backend project with integration tests.

# Goals

- A simplified setup for a Go backend project.
- Geared for local development
  - Dependecies: `Docker CE` and `docker-compose`.
  - Plus an "escape hatch" to work with Go directly on local machine.
- Have a "packaged" development environment.
- Must be able to run the same way on Jenkins (with docker)
- Have some boilderplate code / guideline for running tests against real inra

# Usage

## Running all tests (local)

To run tests: `make test`. This will:

- Copy current folder (all the code) to a docker volume volume
- Spins up infra containers (Postgres)
- Spins up test container (ci.test) and wait for infra to be ready
- Run normal tests and tests-on-infra (files with `// +build integration`) in a "packaged" environment

For tests against infra, we usually need a clean space per tests.
Creating this is pretty tedious, so there is a helper package: `internal/integration`.

- Use `integration.New(t)` to create a new, clean "space"
- In this case, we will create a new user and database in PG to test.
  Configurations is loaded from a template, located in `config.d/`.

## Running infra locally + use local Go

- `make local.infra`. This spins up the infrastructure only.
- Update `/etc/hosts` so that infra's hostname (e.g. `postgres.test`) points to `127.0.0.1` - Run `go test` (self-contained) or `go test -tags integration` (against infra).

## Running in CI (Jenkins with docker)

- `make test`
- We should not expose any infra ports and use different `--project` flag with docker-compose, so multiple build can run on the same CI machines.

# Thoughts

## Good

- Self-contained, package dev environment (as docker image) => easy to distribute.
- Easier to onboard existing developers: It's easier to convince them to install `docker` and `docker-compose` vs asking them to run work on a team-maintained VM image.
- Still have an "escape hatch" for local dev (i.e. `make local.infra`), for those who are well-versed with docker.
- New, clean "space" for tests against infra every time, no "DB.Clean()" magic.

## Bad

- Very tight coupling between docker-compose setup, build/test image and the `internal/integration` package.
- No good way to clean up infra right after "integration" tests finish (though we could do solve it with a deferred call).
- Use a new, randomized "space" for test against infra every time => might be hard to track.
