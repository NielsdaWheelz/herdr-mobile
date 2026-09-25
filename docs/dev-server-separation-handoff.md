# dev-server: independent herdr-mobile and skidbladnir

2026-09-25 update: the github name transfer succeeded. repository id
`1342599607` is `NielsdaWheelz/herdr-mobile`; id `1386409483` is
`NielsdaWheelz/skidbladnir`. known clone remotes are retargeted. the baseline
names below describe the pre-transfer investigation. immutable releases are
enabled for both repositories. herdr-mobile `v0.9.0` is published and pinned;
original skid publication and live namespace handback remain pending.

implement the dev-server portion of the product split below. read this repo's
agent instructions, `SPEC.md`, deployment rules, and relevant issues first.
this prompt is self-contained; the cross-repo contract is recorded in
`herdr-mobile/docs/herdr-mobile-separation.md` after the repository rename.

## outcome and assignment

both products must run concurrently on macbook, devbox, and arch, with two
independent android apps. installing, updating, restarting, rolling back, or
removing either product must leave the other's files, credentials, services,
workers, and ingress intact. each product's deployment must work with the
other absent. shared host-managed tailscale and provider executables are
allowed; their upgrades are separate operations from gateway maintenance.

you own dev-server source, service/config templates, pins, scoped installation
and removal, provider integration provisioning, and the host cutover runbook.
edit only dev-server. coordinate app/helper requirements with the original
skid agent and worker-home changes with the jarvis agent. deliver reviewed
source and a concrete runbook; the root operator coordinates live namespace
handback. github renames, app code/releases, jarvis code, and phone installs
belong to their respective owners. do not activate unfinished app releases.

the owner permits existing herdr panes to die during coordinated cutover.
that does not authorize killing unrelated tmux sessions or disrupting jarvis
cognition. preserve applicable current-turn device/tmux/live-test rules;
unexecuted boundaries are `NOT_RUN`. logs/evidence contain no credentials,
account data, terminal bytes, prompts, objectives, histories, or qr contents.

## coordination and input locations

on this macbook, the current checkouts are:

- herdr-mobile owner/root operator: `/Users/nnandal/Documents/code/skidbladnir`.
  read `docs/herdr-mobile-separation.md` for the agreed contract,
  `docs/dev-server-handoff.md` for the implemented deployment interface, and
  `internal/hostconfig/config.go` for the current reduced config schema.
- original skid owner: `/Users/nnandal/Documents/code/skid-v1`.
  read its `internal/hostconfig/` and `internal/agentcontrol/native.go` as
  source evidence. its agent supplies the qualified deployment contract at
  `docs/dev-server-handoff.md` in that checkout.
- dev-server owner: `/Users/nnandal/Documents/code/dev-server`.

these are current paths, not instructions to rename directories. separate
agent sessions are not automatically reachable through another session's
subagent tools. exchange file paths through the user; read sibling checkouts
but write only your assigned repo. the handoff is pending until its owner
writes it; do not infer agreement from an unfinished working tree.

each app's `release-pin.json` is the published artifact authority only after
its owner publishes and updates it for the separated product. herdr-mobile
`v0.9.0` is published from `68a652d7ccbeaaf472ef1c5f3a4ea6949808bca4`;
derive `assets/herdr-mobile/release-pin.json` from this checkout's
[`release-pin.json`](../release-pin.json) using the exact schema conversion in
[its handoff](dev-server-handoff.md#release-inputs). the upstream five-asset pin
is not the deployment's host-only pin. that handoff also records config,
assets and qualification. its removed `v0.8.0` pin
belongs to the old shared identity. original skid's old `v0.6.0` pin is also
stale; await that owner's new pin. its source contract now exists at
`/Users/nnandal/Documents/code/skid-v1/docs/dev-server-handoff.md`.

jarvis's matching worker map is `bada736` in
[draft pr 42](https://github.com/NielsdaWheelz/jarvis/pull/42) and
`/Users/nnandal/Documents/code/jarvis`; static/build checks passed. activate it
only after settling pending incompatible worker actions. private provider
homes, authentication/trust, host handback and phone proof are still required;
publication alone does not authorize skipping that sequence.

2026-09-25 coordination follow-up: original skid's handoff and `scripts/fleet`
now implement the ten-file receipt order, encoding and modes. replace your
stale six-file/pending-adoption notes after inspecting that source. the original
owner's `docs/issues/restoration-receipt-validation.md` identifies the remaining
deployment defect: `gateway_generation_owned` does not require directory mode
`0700` or equality between the computed runtime digest and basename suffix.
fix this at generation admission for both products, retaining the existing
digest function and explicit product file lists. qualify intact admission,
wrong directory mode, wrong digest suffix and prior-generation rollback with
disposable fixtures; report native lifecycle proof separately.

after that correction, acknowledge the original app's contract in the runbook
and commit the reviewed deployment slice, including the real herdr-mobile pin.
keep original skid's pin pending its own publication. this does not authorize
live activation or old namespace reclamation.

## fixed identities

| concern | herdr-mobile | original skidbladnir |
| --- | --- | --- |
| final github repo | `NielsdaWheelz/herdr-mobile` | `NielsdaWheelz/skidbladnir` |
| github repository id | `1342599607` | `1386409483` |
| current github name, before swap | `NielsdaWheelz/skidbladnir` | `NielsdaWheelz/skid-v1` |
| command / launcher | `herdr-mobile` / `herdr-mobile-launch` | `skidbladnir` / `skidbladnir-launch` |
| linux unit | `herdr-mobile.service` | `skidbladnir.service` |
| mac label | `dev.niels.herdr-mobile` | `dev.niels.skidbladnir` |
| config / data / state leaf | `herdr-mobile` | `skidbladnir` |
| loopback gateway | `127.0.0.1:7342` | `127.0.0.1:7341` |
| tailscale serve | `:8444/v1` to `127.0.0.1:7342/v1` | `:8443/v1` to `127.0.0.1:7341/v1` |
| archive / archive binary | `herdr-mobile-<platform>.tar.gz` / `herdr-mobile` | `skidbladnir-<platform>.tar.gz` / `skidbladnir` |
| active receipt stems | `herdr-mobile.runtime`, `herdr-mobile.unit` | `skid.runtime`, `skid.unit` |
| machine header | `Herdr-Mobile-Machine` | `Skidbladnir-Machine` |
| invitation scheme / kind | `Herdr-Mobile-Invite` / `herdr-mobile.fleet-invite.v1` | `Skidbladnir-Invite` / `skidbladnir.fleet-invite.v1` |

config, data, and state roots are respectively `~/.config/<leaf>`,
`~/.local/share/<leaf>`, and `~/.local/state/<leaf>`. command links are in
`~/.local/bin`. give each product independent `current`, `previous`, release,
artifact, and unit generations, bearer, random `mh-` installation handle,
host config, and private operator `client.json`. no credential/state symlinks.

upstream `herdr`, `herdr.service`, `dev.niels.herdr`,
`~/.config/herdr/herdr.sock`, and its runtime/config identity stay intact.
herdr-mobile is its phone gateway, with no hook implementation of its own.
ordinary gateway stop/remove/rollback must not stop upstream herdr or tmux.

## investigated starting points

dev-server was inspected at `8498933` on 2026-09-25; refresh before editing.
all hosts had the old skid-named gateway and herdr active, ports `7341`/`8443`
occupied, and no listener on `7342`/`8444`. this does not prove new-port access
through tailnet policy. the old gateway is herdr-backed `v0.8.0`, not original
skid. the phone holds that build under `dev.niels.skidbladnir`.

inspect `lib/skidbladnir.sh`, `assets/skidbladnir/`, `lib/herdr.sh`,
`assets/herdr/`, `assets/codex/codex-shared.py`, `assets/codex/profiles.json`,
`assets/routers/ai-profile`, ansible roles `skidbladnir` and `workspace_assets`,
and all platform/command dispatch entrypoints. search callers and deletion,
retention, rollback, protected-path, and receipt logic before moving files.

known traps: linux hardcodes/refuses a gateway port other than `7341`; release
validation hardcodes the old repo and archive names; ansible removes
`/usr/local/libexec/skidbladnir` and retired skid integration assets; gateway
validation unnecessarily inspects android signing files. the latter is tracked
in `docs/issues/skid-signing-boundary.md` and must be removed from both gateway
paths while leaving signing files untouched.

## implementation requirements

1. provide two independently selectable deployment operations, with distinct
   pins, units, launchers, config renderers, and owned state. one operation
   must not validate, install, prune, or remove the other's resources. preserve
   the existing rollback guarantees within each product's namespace. use
   direct owners and a small finite configuration; add no general product
   deployment framework. shared host-tool provisioning remains separate.
2. coordinate executable/config validation with each app's actual interface.
   herdr-mobile retains the reduced herdr/profile schema and its renamed
   `validate-host-config` command. original skid needs the tmux-era schema,
   explicit launch arguments/signatures, and `nativeControlPath`; obtain its
   exact validation contract from its agent rather than assuming v0.8 behavior.
3. configure only each owned tailscale port/handler. preserve unrelated serve
   state and forbid public funnel for these origins. never reset all serve
   state. update health/fleet checks, action messages, and cleanup recipes so
   an operation for `8444` cannot disable `8443`, or vice versa.
4. remove obsolete unconditional skid retirement before original skid
   returns. retention must never select herdr-backed v0.8 as original skid's
   rollback, or prune provider homes/signing material. do not restore an old
   whole dev-server revision to recover removed skid machinery.
5. use fresh credentials for both products after the split. receive exact
   published version/source/digests from each app owner; never fabricate pins
   or copy the original repo's stale v0.6 pin. github name reclamation is
   complete and the old-name redirect no longer selects herdr-mobile: verify
   repository ids before selecting sources. publication belongs to the root
   operator; live namespace transfer is a separate later step.

## provider homes, commands, and helper

both runtimes get five private interactive homes:

| runtime | root | children |
| --- | --- | --- |
| herdr | `~/.local/share/herdr/providers/` | `codex-personal`, `codex-work`, `codex-work2`, `claude-personal`, `claude-work` |
| original skid | `~/.local/share/skidbladnir/providers/` | the same five child names |

the phone still offers four existing forge profiles; personal claude is a
manual/native default. provision fresh homes and required instructions/settings;
authenticate and trust hooks through normal provider setup. do not copy whole
account trees, credentials, plugin caches, trust records, history, or discovery
sockets. missing authentication/trust is an explicit prerequisite.

install herdr's own pinned native integrations only in herdr homes; restore
skid integrations only in skid homes. once ready, remove precisely inventoried
old product entries/scripts from ordinary `~/.codex*`/`~/.claude*`, preserving
user hooks/settings and credentials. no recurring product installer owns those
ordinary homes or project settings. codex can load `~/.codex/hooks.json` as
project config at `cwd=$HOME`; audit inline/plugin sources as well.

keep jarvis COGNITION services, existing account-home declarations, discovery,
and unrelated settings intact. jarvis WORKERS move to herdr's new homes. update
mobile host profiles, herdr integration targets, scoped shell/account command
resolution, and `assets/herdr/herdr-gate`'s environment allowlist together.
give the jarvis agent the exact matching map for `src/jarvis/agent_tools.py`,
including explicit personal claude default. do not widen the gate's commands.

forge and manually typed commands in new terminals must select the same
product-owned homes after login-shell startup. current work wrappers force
old homes and cannot be reused unchanged. preserve explicit profile selection
and each product's existing permission policy; original skid profiles use
absolute native providers and their own arguments/plugin. use minimal scoped
wrappers only where needed, and coordinate shell environment creation with
the original skid agent. scrub foreign `HERDR_*` context from skid creation
paths and tmux context from herdr; never rewrite unrelated live tmux sessions
or default-server environment. keep upstream hook guards intact.

restore original skid's native helper as its separately pinned dependency.
historical inputs are `llm-calling@ec97adeb9ddd0f91b141f89cc42cff7cc7efdb8f`,
uv `0.11.28`, python `3.12.13`, claude sdk `0.2.130`. the original agent owns
current compatibility qualification; you own reproducible installation and
config wiring. use a skid-owned command such as
`skidbladnir-provider-runtime-control`. verify its claude subprocess resolves
the intended executable and private home. herdr-mobile never consumes it.

## staged delivery and recovery

prepare source and a host-by-host runbook without prematurely claiming live
completion. app release inputs and the jarvis worker mapping are dependencies;
source preparation can proceed while they are pending.

1. stage herdr-mobile in its new namespace, with fresh credentials and `8444`
   ingress, preserving the old deployment until the new inputs are ready.
2. the root operator coordinates jarvis quiescence and the herdr reset from
   outside the discarded panes. settling pending actions remains required.
   stop the intended runtime before discarding its old session snapshot;
   restarting alone can restore old shells. activate matching provider-home,
   gate, worker-map, and gateway changes. preserve herdr configuration/runtime
   identity and cognition. report exact results for each host.
3. after all new gateways and the new phone app work, retire the old
   herdr-backed skid services, ingress, credentials, launchers, receipts, and
   obsolete generations. inventory removal precisely; preserve original
   android signing files and unrelated files in `~/.config/skidbladnir`.
   the root operator declares the skid namespace handed back.
4. only then activate original skid with its own release, credentials,
   helper/config, service, and `8443` ingress. the phone operator handles the
   old package's data reset, increasing-version installation, and fresh pairing.

before handback, saved old inputs may recover the former gateway in the old
namespace. after handback, herdr-mobile recovery stays entirely within its
new namespace; never restore it over skid's service, port, or state. the first
separated release may require stop-and-repair rather than a previous separated
version. subsequent rollback chains remain product-specific. discarded herdr
panes need not be recovered.

## done when

deliver reviewed dev-server changes, exact configuration maps/interfaces for
both app agents and jarvis, a runbook naming staged inputs and rollback limits,
engineering-check results, and unresolved prerequisites in `docs/issues/`.

qualify each platform: both gateways active concurrently; independent apply
with the other absent; repeat apply makes no cross-product changes; each
gateway restart/reinstall/rollback leaves the other's worker and attachment
usable; scoped removal preserves the other's files and ingress. exercise
removal/rollback on disposable installations before live use. test all forge
profiles and manually typed provider commands from home and a shared project,
including startup from the opposite runtime. confirm hook/runtime/home
isolation and cognition continuity. coordinate new-port reachability and
wrong-product auth rejection with the app/phone owners.

follow the repository's testing policy. temporary probes must not recreate
retired harnesses. tmux mutations target only probe-created resources on an
isolated `-L` socket. builds or free ports are not coexistence acceptance;
report unavailable/unexecuted behavior as `NOT_RUN`, with its owner and blocker.
