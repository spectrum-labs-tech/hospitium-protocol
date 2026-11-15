# hospitium-protocol: notes for AI assistants

Wire contracts (Go types) between a Hospitium control plane and its devices: the management agent and
the TV launcher. Read `README.md`.

## Ground rules

- **This repo is public.** Never reference private repositories, deployments, infrastructure, customers
  or secrets, and never name the operator's other products.
- **Additive only within `v1`.** Add fields; never remove, rename or re-type one. Consumers ignore unknown
  fields. A breaking change is a new `v2` package and path prefix.
- **The operation set is closed.** No shell or ADB passthrough, ever.
- **Guest data is minimal** (a display name and stay times) and always carries an expiry.
- No dependencies beyond the standard library.
- **Every Go file starts with** `// SPDX-License-Identifier: Apache-2.0`.

## Releases

- **Only through `task release BUMP=minor|patch`, and only when the maintainer asks.** Never
  `git tag` or `git push --tags` by hand. A tag is permanent: the Go module proxy caches it forever.
- The task tags the tip of `origin/main` once CI has passed on it, and runs `gorelease` against the
  previous tag (`task release:api`). Push first; local commits are never released.
- `task release:check` self-tests the version arithmetic in task's own shell. Never keep two copies
  of that arithmetic: `release:next` is the only implementation.

## Commands

`task check` (see `Taskfile.yaml`).
