# AGENTS.md

## Project Rules

- Follow the repository's existing style, package layout, and tooling choices.
- Keep changes scoped to the user's request.
- Do not revert user changes unless the user explicitly asks for that.
- Any code change must use the local TDD workflow skill and the local Go testing skill before implementation and verification.

## Required Skills

For any task that changes application code, generated code, tests, migrations, build tooling, or developer workflow:

1. Read and follow `.agents/skills/tdd/SKILL.md`.
2. Read and follow `.agents/skills/golang-testing/SKILL.md`.
3. Apply both skills throughout the work, not only at the end.

Use the TDD skill to drive the implementation order: understand the expected behavior, add or update a focused failing test where practical, implement the smallest useful change, then refactor while keeping tests green.

Use the Go testing skill for Go-specific test structure, table tests, test naming, fixtures, database boundaries, and verification expectations.

If a requested change is genuinely not testable or a test would add no useful signal, state that explicitly before making the change and keep the implementation small.

## Command Policy

- Only run project commands through `just`.
- Do not run `go`, `docker`, `sqlc`, `goose`, `golangci-lint`, shell scripts, or package-manager commands directly.
- If a needed workflow is not available as a `just` recipe, update the `justfile` first, then run the new or existing recipe through `just`.
- Prefer existing recipes over adding new ones.
- Do not bypass the `justfile` for one-off checks.

Examples:

- Use `just test`, not `go test ./...`.
- Use `just lint`, not `golangci-lint run`.
- Use `just sqlc`, not `sqlc generate`.
- Use `just migrate-up`, not `goose up`.
- Use `just db-up`, not `docker compose up -d db`.

## Go Workflow

- Run formatting, tests, generated-code updates, migrations, and linting through `just` recipes only.
- Keep sqlc-generated code in sync with query and schema changes.
- Keep goose migrations forward-compatible and reversible when practical.
- Prefer `context.Context` propagation over creating unnecessary background contexts in lower-level code.
- Return errors instead of panicking outside application startup or intentionally fatal setup paths.

## Verification

For code changes, use the smallest relevant `just` verification first, then broader checks as the change warrants.

Typical verification order:

1. `just test`
2. `just sqlc` when SQL queries or migrations changed
3. `just lint` when code changes are complete
4. `just check` when a full local verification pass is appropriate

If a verification command cannot be run, report the exact `just` recipe that was skipped and why.
