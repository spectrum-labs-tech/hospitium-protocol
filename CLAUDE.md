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

## Commands

`task check` (see `Taskfile.yaml`).
