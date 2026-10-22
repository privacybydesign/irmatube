# YiviTube

YiviTube is a demo attribute issuance and verifier. It offers membership attributes to users, who then need to disclose these in order to watch a movie.

It showcases the following features:

* Issuance and verification of attributes to users.
* Selective disclosure of attributes: although an identifying membership number is issued to the user, this number is not disclosed during verification, so that YiviTube knows that a member is watching a movie, but not which member.
* Granting access to a resource only after successful Yivi attribute verification.

Note that the movies themselves are not included in this repository and only a Youtube embedded video will display a trailer for the chosen movie. 

## Installation

### Without Docker

You need Go (see `go.mod` for the version), Node.js and an [IRMA server](https://github.com/privacybydesign/irmago).

* Fetch the front-end libraries into `static/lib`: `npm ci --ignore-scripts && npm run build:libs`
* Start YiviTube with `IRMA_SERVER_PUBLICKEY=data/demo-publickey.pem go run .` and open http://localhost:8080.
* Note that demo keys provided are only for local testing. YiviTube needs [IRMA  server's](https://github.com/privacybydesign/irmago) public key and irma server will need it's corresponding private key, which can be provided using `--jwt-privkey-file` flag.

### With Docker

You can use the `docker compose up --build` to build and test the project in one go. 

## Configuration

YiviTube is configured with environment variables.

| Variable | Default | Description |
| --- | --- | --- |
| `IRMA_SERVER_PUBLICKEY` | none, required | Path to the PEM file with the public key of the IRMA server. Result JWTs are only accepted when signed with RS256 and this key. |
| `IRMA_SERVER_URL` | `http://localhost:8088` | URL of the IRMA server. |
| `IRMA_SERVER_API_TOKEN` | empty | Token for the IRMA server, sent in the `Authorization` header when set. |
| `IRMATUBE_CREDENTIAL_ID` | `irma-demo.IRMATube.member` | Credential that is issued and verified, for example `pbdf.pbdf.yivitube` or `pbdf-staging.pbdf.yivitube`. |

The server listens on port 8080.

## Movies

The movies and their covers are compiled into the binary. To add or change one:

* Add a cover to `static/content/covers`, named after the movie id.
* Add the movie to `static/content/movies.js`, with its age limit.
* Add `videos/<id>.json` with the `ageRestriction` and the `youtubeId` to embed. A movie can only be watched when it has such a file.

## Tests

`go test ./...`
