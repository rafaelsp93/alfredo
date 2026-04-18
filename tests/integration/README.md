# Integration Flow Checklist

Every new route or user-facing pet-care flow should add integration coverage in
`tests/integration/http_integration_test.go` unless a reviewer accepts a narrower
unit-test-only case.

Prefer covering at least one of these behaviors:

- A create/list/get/update/delete happy path for the route or aggregate.
- Auth, path validation, or payload validation failures.
- Calendar, Telegram, or agent side effects when the flow owns them.
- Transaction or best-effort behavior when external adapters fail.

When a route is intentionally not covered here, leave a short note in the
story or pull request explaining which lower-level test protects it instead.
