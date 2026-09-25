# herdr-mobile: deployment contract

2026-09-25. this checkout owns the herdr phone app and gateway. the shared
[separation spec](herdr-mobile-separation.md) owns cutover order and acceptance;
the [dev-server assignment](dev-server-separation-handoff.md) owns deployment.
source preparation is complete. no release or live activation is claimed.

## release inputs

repository id `1342599607`, verified canonical name
`NielsdaWheelz/herdr-mobile`. first separated release planned as `v0.9.0`,
android `0.9.0` / `9000`, package `dev.niels.herdr.mobile`, label `herdr`.
the new public signer pin is `android/app-signing-cert.sha256`; its private
configuration lives only in `~/.config/herdr-mobile/android-signing.properties`.
gateway installation must not read signing files.

published `release-pin.json`: **pending publication**. the old `v0.8.0` pin
was removed because it identifies the pre-separation skid-named artifacts.
its immutable bytes and original pin remain in git history at `d8bb9c4`.
do not synthesize a pin or activate a candidate as a published release.

the five release assets are `herdr-mobile-android.apk`,
`herdr-mobile-darwin-arm64.tar.gz`, `herdr-mobile-linux-amd64.tar.gz`,
`android-signing-cert.sha256`, and `SHA256SUMS`. each host archive contains
`herdr-mobile`, `characters.json`, and `release.json`; the manifest remains
`{platform,sourceSha,version}`. `herdr-mobile version` prints `TAG SOURCE_SHA`.

## executable and config contract

validate rendered config with:

```sh
/absolute/generation/herdr-mobile validate-host-config --host-config=/absolute/host-config.json
```

exit zero means structurally valid for the invoking platform, not runtime or
provider readiness. validation reads only the config. exact schema owner:
`internal/hostconfig/config.go`; profile validation: `internal/profile/profile.go`.
json members are case-sensitive and unknown/duplicate members are rejected.

- root: required `platform` (`Darwin` or `Linux`, matching the running binary),
  `herdr` object, and `profiles` array.
- herdr: required `path`, `socketPath` (clean absolute paths), and
  `testedVersion` (exact `herdr 0.9.1`). use the existing upstream binary and
  `~/.config/herdr/herdr.sock` after expanding the service user's home.
- profiles: empty, or exactly the four rows below in this order. each row has
  required `key`, unique nonempty `label`, `provider`, and `environment`.
- environment: array of `{name,value}` strings with unique names and exactly
  the appropriate absolute provider home; no other provider's home. the
  deployment supplies only the declared home variable below.

| key | provider | variable | value relative to the service user's home |
| --- | --- | --- | --- |
| `personal` | `Codex` | `CODEX_HOME` | `.local/share/herdr/providers/codex-personal` |
| `work` | `Codex` | `CODEX_HOME` | `.local/share/herdr/providers/codex-work` |
| `work2` | `Codex` | `CODEX_HOME` | `.local/share/herdr/providers/codex-work2` |
| `claude-work` | `Claude` | `CLAUDE_CONFIG_DIR` | `.local/share/herdr/providers/claude-work` |

no `nativeControlPath`, provider command/arguments, tmux fields, hooks, or
identity plugin are consumed here. native herdr receives bare `codex` or
`claude`; deployment-owned shell resolution selects the executable and existing
permission flags. manually typed bare/account commands must use the same
private homes. personal claude defaults to `claude-personal` under that root;
it is not a fifth phone profile. explicit profile homes take precedence.

start the gateway with explicit deployment paths:

```sh
/absolute/generation/herdr-mobile gateway \
  --listen=127.0.0.1:7342 \
  --bearer-file=/absolute/config/herdr-mobile/bearer \
  --machine-handle-file=/absolute/config/herdr-mobile/machine-handle \
  --host-config=/absolute/generation/host-config.json \
  --catalogue-path=/absolute/generation/characters.json
```

the config/data defaults use `herdr-mobile`; no old skid paths are consulted.
`machine init` creates or retains its handle; `bearer mint` rotates its bearer.
both print sensitive installation material: consume privately, never log it.
neither command nor gateway start creates or stops the herdr server.

## wire, receipts, and fleet tooling

serve only `https://HOST:8444/v1` to `http://127.0.0.1:7342/v1`.
headers are `Herdr-Mobile-Machine`, `Herdr-Mobile-Terminal-Takeover`, and
authorization scheme `Herdr-Mobile-Invite` for redemption. ordinary requests
retain `Bearer`; fresh credentials and product machine binding separate them.
fleet qr kind is `herdr-mobile.fleet-invite.v1`; only port `8444` is admitted.
metadata is `herdr_mobile_*` with source `user:herdr-mobile`; no legacy reader.

`scripts/fleet verify` uses `HERDR_MOBILE_DEV_SERVER_CHECKOUT` and
`assets/herdr-mobile/release-pin.json`. service identities are
`herdr-mobile.service` / `dev.niels.herdr-mobile`, launcher
`herdr-mobile-launch`, receipt stems `herdr-mobile.runtime` / `herdr-mobile.unit`.
the receipt byte formats remain the existing dev-server contract; only owned
paths and binary member change. fleet invite reads only
`~/.config/herdr-mobile/client.json`. upstream herdr's pin and identity remain
under `assets/herdr/` and the existing herdr roots.

## source qualification and pending delivery

the github name swap succeeded on 2026-09-25. original repository id
`1386409483` now owns `NielsdaWheelz/skidbladnir`; known publishing clones on
macbook, devbox, and arch use the corresponding canonical remotes. checkout
directory names remain stable while agent work is in progress.

`scripts/check verify` passed, including android lint/debug build, go vet/build,
shell checks, catalogue and generated-asset checks. a signed `0.9.0` / `9000`
candidate has the new package, label, and signer. the complete five-asset
candidate at source `50217ec89af8e1e01c700db65c9fb7d8d714a7e3` passed
`scripts/check-release`, including both host archives and native version.
these are unpublished candidate bytes, not deployment pins. a temporary probe against an
isolated native herdr `0.9.1` and this gateway passed product header/credential
rejection, single-use invitation redemption, terminal creation/inventory, and
retention of all renamed metadata keys/source. the same probe rejects the old
source at the expected new-header boundary. a temporary jvm probe against the
compiled android classes and exact pinned dependencies admitted only the new
invite kind and `8444` origins. neither probe exercises provider startup or
phone behavior; no behavioral harness is retained.

the root operator will append the published pin when available. fresh
provider login/trust, jarvis worker-home agreement,
host namespace handback, tailnet `8444` access, and phone coexistence remain
deployment dependencies. no other repo's running agent is authorized to change
this source contract merely by writing its own handoff.
