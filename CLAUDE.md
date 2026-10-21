# CLAUDE.md

Go service (standard library plus `golang-jwt/jwt/v5`) that serves the YiviTube
demo page and the two endpoints behind it. It replaced a PHP implementation.

## Build & test

- `go vet ./... && go test ./...`. The Go version is set in `go.mod`.
- Tests use `httptest` and a fake IRMA server; no network or IRMA server needed.
- `static/lib/` (Bootstrap, jQuery, Mustache, yivi-frontend) is gitignored and
  built from `package.json` with `npm install --ignore-scripts && npm run build:libs`.
  The Dockerfile does this. `go test` does not need it.

## Layout

- `server.go`: routes and handlers: `/api/session`, `/api/verifysession`, static files.
- `irma_client.go`: starts sessions on the IRMA server.
- `verifier.go`: checks a result JWT and the disclosed attributes.
- `static/` and `videos/` are embedded in the binary.

## Gotchas

- Only `static/` is served. Put nothing there that must stay private;
  `TestStaticFilesArePublic` fails on other file types.
- `videos/*.json` is both the allowlist of video ids and the source of the age
  limit and YouTube id. `static/content/movies.js` drives the UI and the age
  that is asked for, so keep both in sync.
- Do not loosen the token checks in `verifier.go`: RS256 only, only the
  configured key, `exp` required, 60 s leeway, `proofStatus` must be `VALID`.
  Each one has a test.

## CI

`.github/workflows/delivery.yml` builds the Docker image and scans it with
`anchore/scan-action` (`only-fixed: true`, `severity-cutoff: critical`). The job
fails when the image contains a critical vulnerability that has a fix available.
The scan also runs on a weekly schedule, so a pinned version drifts into
failure as new CVEs are published.

The runtime image is `distroless/static`, so findings are in the Go binary: the
standard library from the pinned `golang:<version>` tag in the `Dockerfile`, or a
module from `go.mod`. When the scan fails, bump the `golang` tag to the latest
patch release or update the module (`go get <module>@latest && go mod tidy`).

`.github/workflows/status-checks.yml` runs the shared Go checks, with the tests
in `Dockerfile.test`.
