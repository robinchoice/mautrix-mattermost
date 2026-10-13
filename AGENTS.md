# mautrix-mattermost

## Purpose & links

Matrix-Mattermost bridge on mautrix-go bridgev2, public, AGPL-3.0. Meant to move upstream (mautrix) or wherever [linagora/twalk#241](https://github.com/linagora/twalk/issues/241) decision D1 puts it. Code and comments stay neutral: no personal or deployment details, others read along.

Image: `ghcr.io/robinchoice/mautrix-mattermost`. It is deployed as one more bridge in an existing Matrix homeserver stack maintained elsewhere, not from this repo.

## Checks

Go isn't installed on the dev laptop, the Dockerfile has a test stage (gofmt, vet, tests with race detector):

```sh
docker build --target test .
```

## Deploy

A push to `main` runs CI and publishes `latest` and `sha-<commit>` images for amd64 and arm64. Check with `gh run watch` and the package page on GitHub.

## Pitfalls

- Commits need a DCO sign-off: `git commit -s`.
- Mattermost keeps a websocket open after its session is revoked and just stops sending events. A revoked token on reconnect is also accepted and then closed without a 401. Only `GET /users/me` tells, which is why `Listen` checks it.
- Without a scheme, `NormalizeServerURL` assumes `https://`.
- Full end-to-end testing needs Mattermost Team Edition plus a Synapse. Never test against the enterprise image (licence terms, see twalk#241 D2).
