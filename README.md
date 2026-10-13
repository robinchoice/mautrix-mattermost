# mautrix-mattermost

An **unofficial** Matrix-Mattermost puppeting bridge built on [mautrix-go](https://github.com/mautrix/go)'s bridgev2 framework. It is not affiliated with mautrix or Mattermost. The goal is to contribute it upstream. Until then it lives here, coordinated with [linagora/twalk#241](https://github.com/linagora/twalk/issues/241).

## Status

Early first slice: Mattermost → Matrix only.

Works:
- Login with server URL, username or email, and password. No admin rights are needed, so it works on servers where personal access tokens are disabled.
- New posts in public and private channels arrive in Matrix as they're sent, one portal room per channel. Markdown is converted, and thread replies become Matrix threads.
- An expired or revoked session is reported as bad credentials, so the user knows to log in again.

Not yet:
- Sending from Matrix to Mattermost
- Direct and group messages
- Edits, deletions, reactions, attachments, backfill
- Personal access token login, MFA
- Syncing channel members and teams as spaces

## Running

Images for amd64 and arm64 are published to `ghcr.io/robinchoice/mautrix-mattermost` on every push to `main`, tagged `latest` and `sha-<commit>`.

Setup follows the [mautrix bridgev2 docs](https://docs.mau.fi/bridges/go/setup.html):

1. Start the container with an empty `/data` volume. It writes `config.yaml` and exits.
2. Edit the config (homeserver address and domain, database, permissions) and start it again. It writes `registration.yaml` and exits.
3. Register `registration.yaml` with your homeserver, then start the bridge.
4. Open a DM with the bridge bot (`@mattermostbot:your.domain` by default) and send `login password`.

The bridge stores the Mattermost session token in its database, not your password. Logging out from the bridge ends that session on the Mattermost server.

## Development

Building needs cgo (SQLite). Encryption uses goolm:

```sh
go vet -tags goolm ./...
go test -race -tags goolm ./...
go build -tags goolm ./cmd/mautrix-mattermost
```

Contributions are signed off under the [Developer Certificate of Origin](https://developercertificate.org/): `git commit -s`.

## License

[AGPL-3.0](LICENSE), like the other mautrix bridges.
