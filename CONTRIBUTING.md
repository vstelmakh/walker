# Contributing

Thank you for the interest in **Walker**. If you find a problem or want to discuss new features
you are welcome to open an [issue](https://github.com/vstelmakh/walker/issues/new)
and/or a [pull request](https://github.com/vstelmakh/walker/compare).

For a large or breaking change, open an issue first. Agreeing on the direction is easier before the code is written.

When submitting your pull request, consider writing a description which explains the changes.
Please note that not all pull requests may be merged. Merging a PR is at the discretion of the maintainer
and depends on factors such as: relevance to the project's goals, impact on maintainability and code quality.

By contributing to this project, you agree that your contributions will be licensed under
[project's MIT License](./LICENSE).

## Workflow

1. Fork the repository.
2. Start your branch from `main`.
3. Implement your change and add tests for it.
4. Follow the contributing [rules](#rules).
5. Ensure all the checks pass, run `make fmt-check vet test`. See [tests and tools](#tests-and-tools) for details.
6. Publish pull request and wait for review.

## Rules

Here are a few rules to follow when making changes to this project:
- Follow the project code style. Run `make fmt` to apply it.
- Add tests for new behavior and bug fixes.
- Keep documentation up to date with new features and changes.
- Be consistent with existing code in the project.
- Code should build for every platform that the [CI workflow](./.github/workflows/build-and-test.yml) builds.

## Tests and Tools

All the checks run via [Makefile](./Makefile) targets. Run `make` to see the list:

| Command                 | Description                                               |
|-------------------------|-----------------------------------------------------------|
| `make test`             | Run tests.                                                |
| `make vet`              | Report suspicious Go code.                                |
| `make fmt`              | Format Go files.                                          |
| `make fmt-check`        | Fail if any Go file is not formatted.                     |
| `make get PKG=<module>` | Add or update a Go module.                                |
| `make tidy`             | Add missing and remove unused Go modules.                 |
| `make build-<platform>` | Build one platform into `dist/`, see `make` for the list. |
| `make clean`            | Remove the `dist/` directory.                             |

Go runs in Docker by default, so the only requirement is Docker. To use the Go installed on your machine, set
`GO_USE_DOCKER=0`, for example `make test GO_USE_DOCKER=0`.

The [CI workflow](./.github/workflows/build-and-test.yml) runs the same checks and builds the binaries for all
platforms on every pull request.
