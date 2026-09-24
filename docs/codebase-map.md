# codebase map

herdr owns terminal, workspace, agent and process lifetimes. each gateway serves
the phone over one host's herdr; the phone composes independent gateways. there
is no application database.

| slice | owner | boundary |
| --- | --- | --- |
| startup and host configuration | `cmd/skidbladnir`, `internal/hostconfig`, `internal/profile`, `internal/platform` | compose one host from deployment-owned configuration; `validate-host-config` admits that configuration for the deployment owner |
| runtime and metadata | `internal/herdr`, `internal/sessions`, `internal/catalog`, `internal/reference` | public herdr socket/terminal control; inventory, launch, and every pane write (mutations, agent controls, stream input) re-read before dispatch; refs as thin herdr ids |
| host resources | `internal/workdir`, `internal/pressure` | bounded directory browsing and native pressure observation |
| gateway transport and access | `internal/gateway`, `internal/auth`, `internal/pairing`, `internal/strictjson`, `internal/logging`, `internal/terminal` | authenticated http, strict messages, owned websocket/control-child lifetime |
| phone fleet and dashboard | `MachineStore.kt`, `FleetPersistence.kt`, `FleetInvite.kt`, `GatewayClient.kt`, `SkidbladnirController.kt`, dashboard/forge/space/chooser files | encrypted pairings, reconciliation, selection and mutations |
| phone polling and ordering | `Polling.kt` | coalesced reads, per-machine mutation fences and awaited inventory reads; the controller owns lane lifetimes |
| phone machine pressure | `Pressure.kt`, `PressurePresentation.kt`, `MachinePressureRail.kt` | strict pressure contract and state, dashboard visibility and content, rendered rail and details |
| phone terminal | `TerminalConnection.kt`, `LockedTerminalWebView.kt`, terminal composables, `assets/terminal` | transport, page protocol, input, selection and rendering |
| visual assets | theme/chrome/seal/ornament files, `catalog`, `scripts/gen-ornament` | shared presentation and generated artwork |
| build and operations | `scripts`, `.github/workflows`, android build files | engineering checks, release artifacts, installation and fleet operations; `scripts/fleet verify` probes each host's gateway and configured herdr; host installation belongs to `dev-server` |

phone source paths are relative to
`android/app/src/main/java/dev/niels/skidbladnir`; terminal assets are under
`android/app/src/main`.

cleanup proceeds one finding and one merged pr at a time: establish the current
contract and callers, demonstrate the finding, characterize important behavior
through its real boundary, simplify its owner, and independently review the
result. remove temporary tests before committing, as requested for this cleanup.
[testing policy](rules/testing.md) distinguishes that evidence from retained
engineering checks. unresolved findings belong in `docs/issues`, one per issue.
