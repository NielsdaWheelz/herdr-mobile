# codebase map

herdr owns terminal, workspace and process lifetimes. each gateway owns one host;
clients compose independent gateways. there is no application database.

| slice | owner | boundary |
| --- | --- | --- |
| startup and host configuration | `cmd/skidbladnir`, `internal/hostconfig`, `internal/platform` | compose one host from deployment-owned configuration; `validate-host-config` admits that configuration for the deployment owner |
| runtime and metadata | `internal/herdr`, `internal/sessions`, `internal/catalog`, `internal/reference` | public herdr socket/terminal control, inventory, lifetime claims, exact mutations and refs |
| agent identity and control | `internal/agentruntime`, `internal/process`, `internal/agenthook`, `internal/agentcontrol` | observe one foreground process; bind reads and controls to that lifetime |
| host resources | `internal/workdir`, `internal/pressure` | bounded directory browsing and native pressure observation |
| gateway transport and access | `internal/gateway`, `internal/auth`, `internal/pairing`, `internal/strictjson`, `internal/logging`, `internal/terminal` | authenticated http, strict messages, owned websocket/control-child lifetime |
| desktop clients | `internal/fleetclient`, `internal/agentcli`, `internal/terminalclient` | shared peer routing and references; cli and local tty |
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
