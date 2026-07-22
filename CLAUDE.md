# CLAUDE.md

## CI

`.github/workflows/delivery.yml` builds the Docker image and scans it with
`anchore/scan-action` (`only-fixed: true`, `severity-cutoff: critical`). The job
fails when the image contains a critical vulnerability that has a fix available.

Most of these come from the `dunglas/frankenphp` base image in the `Dockerfile`
(the golang binary bundles e.g. `golang.org/x/crypto`, which `apt upgrade` cannot
patch). The scan also runs on a weekly schedule, so a pinned base tag drifts into
failure as new CVEs are published. When the scan fails, bump the pinned
`dunglas/frankenphp:<version>-php8.5` tag to the latest patch release and confirm
with `grype dunglas/frankenphp:<tag> --only-fixed` that no fixable critical remains.
