# Contributing

Thanks for helping. Issues and pull requests are welcome.

## Before you open a pull request

```bash
task check   # gofmt, vet, golangci-lint, tests, go mod tidy, SPDX headers
```

CI runs the same checks. Tests must pass with no network and no hardware.

## Rules

- **Context flows end to end.** Every function that does I/O or calls something that accepts a `context.Context` takes one and passes it on; never `context.Background()` below `main`. Log with the `slog` `*Context` variants. `contextcheck`, `noctx` and `sloglint` enforce this.
- **Every Go file starts with** `// SPDX-License-Identifier: Apache-2.0`.
- **Standard library first.** Justify any new module in the pull request.
- Table-driven tests and `t.Parallel()` where safe.
- Sign off your commits under the [Developer Certificate of Origin](https://developercertificate.org/) (`git commit -s`).

## License

By contributing you agree that your contributions are licensed under the Apache License 2.0.
