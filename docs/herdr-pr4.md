# herdr pr 4: installation and coordinated activation

2026-09-23: implementation specification; no deployment authorized or claimed.
baseline: skid `056d491` (includes merged pr 2), jarvis `fb5e4a9` (pr 3),
dev-server `bb8e218`. recheck exact revisions before implementation.
[the migration plan](herdr-migration.md) owns sequencing;
[pr 2 waivers](herdr-pr2.md#accepted-non-blocking-follow-ups) remain non-blocking.
this spec supersedes the plan's narrower agent-only drain and repeated phone
journey requirements. no product decision is needed before implementation;
host access, provider spend and the maintenance window need separate approval.

## 1. outcome and boundaries

one independently supervised, pinned herdr per host; matching skid gateways,
desktop/cli, signed android app and jarvis consumer. existing credentials,
pairings, canonical history and unrelated services survive. clients cannot
silently resume an old worker, retarget a write or retry an unknown mutation.

```text
dev-server installs: herdr service + skid gateway/config + installed fleet cli
desktop skid ────────────────────────────> same local herdr
phone / fleet cli / jarvis tools ─HTTPS─> skid gateway ─socket─> herdr
jarvis deploy owns: release + service + durable-work preparation
```

pr 4 is one milestone with three repository prs, not one cross-repo transaction:
skid prerequisites/release support; jarvis deployment safety/copy; dev-server
installation and exact pins. merge prerequisites before publishing skid and
finalizing installer pins. merging, publication, installation and acceptance are
separate outcomes. source prs may merge before approved live work.

non-goals: new orchestration, transport or gateway routes; provider/library
upgrades; upstream patches; blue/green fleet, distributed coordinator, permanent
test infrastructure, general backup/restore, account migration, new phone ui,
or the unrelated jarvis rememberer material-window issue. no old execution
decoder, backend switch or automatic rollback/replay path. immutable historical
receipt readers and inactive rollback artifacts are not legacy execution.

## 2. bounded product prerequisites

### truthful stop partials

fix `internal/agentcontrol/actions.go:stopFailure` at the composition boundary.
retain the close error's dispatch before promoting whole-stop dispatch: an
acknowledged interrupt already makes the whole operation `sent`.

| close evidence | `partial.terminal` | whole-stop dispatch |
| --- | --- | --- |
| close not entered, or close-local `not_sent` | `not_attempted` | `sent` |
| acknowledged close rejection, including confirmation refusal | `refused` | `sent` |
| close-local `unknown` | `unconfirmed` | `unknown` |

entering `closePane` alone proves no dispatch: its ping can fail first. preserve
`partial.agent`, the error code, no-close-after-unknown-interrupt behavior and
no replay. `refused` means acknowledged rejection, not necessarily a confirmation
dialog. reuse existing enums and jarvis settlement; add no new partial variant.

### exact start contract

remove `launch_submitted` from [pr 1's wire contract](herdr-pr1.md#retained-gateway-operations).
only `resource_created` and `identified` are failure prefixes; only the latter
carries a terminal. successful launch still has `launch:submitted`. do not add
a producer stage to justify dead documentation. jarvis's existing closed model
already matches. narrow its misleading current-tool stop copy and retire both
issue records only after evidence agrees. keep jarvis's v4 handler revision for
this copy-only change; the tool/plan fingerprint already changes. never reinterpret
old partials as stronger historical evidence. an actual handler revision change
would also require retaining its prior receipt reader, not an old executor.

### configuration validation at its owner

add administrative `skidbladnir validate-host-config --host-config=ABSOLUTE_PATH`
through the existing command dispatcher. call `hostconfig.Load` for the native
platform; exit `0` valid, `1` invalid/unreadable, `64` usage. success prints
`host config valid`; failure is bounded and content-free. no socket, provider,
file mutation, credential read or subprocess; this validates configuration, not
runtime readiness. replace dev-server's duplicate product-schema parser with
this candidate-binary command. keep deployment-owned path/account checks there.

## 3. installation and lifetime contract

use herdr v0.9.1, source `065ef9d6a531c49fb8bee7e818ef837065b21ee9`,
`--version` exactly `herdr 0.9.1`, public ping version `0.9.1`/protocol `22`.
add one dev-server-owned `assets/herdr/release-pin.json`:

```text
{schemaVersion:1, version:"v0.9.1", sourceSha:40-hex,
 artifacts:{"linux-amd64":{url,sha256},"darwin-arm64":{url,sha256}}}
```

pin official raw executables `herdr-linux-x86_64` and `herdr-macos-aarch64` under
`https://github.com/herdrdev/herdr/releases/download/v0.9.1/`, with sha256 values
before merge; no `latest`, curl-to-shell, substitute or unverified cache. validate
platform, digest and executable identity before atomic promotion. source/tag identity and
artifact checksums are not a reproducible-build claim.

- one concrete `lib/herdr.sh` owns installation, service/config identity and
  readiness. reuse `lib/common.sh` atomic files, hashes and result vocabulary;
  no generic installer framework. stage immutable bytes; record active identity
  only after the selected service and pinned public ping agree.
- native user systemd service on arch/devbox, launchagent on macbook, under the
  development user. foreground command is `herdr server`, not an invented
  `--foreground` flag. set explicit `HERDR_CONFIG_PATH` and `HERDR_SOCKET_PATH`;
  omit `--session`, clear inherited `HERDR_SESSION`, and fix the intended xdg
  namespace. directories/socket are `0700`/`0600`; never adopt a foreign socket.
- deployment fixes one absolute socket path per host. host config's `herdr`
  entry is `{path,socketPath,testedVersion}`; existing `skid` desktop and gateway
  use that same instance. a missing runtime fails without request-time autostart.
  do not alter upstream's unrelated manual-session behavior.
- managed config sets `[session] resume_agents_on_restore = false` and
  `[update] version_check = false, manifest_check = false`. no updater runs.
  config-path override does not relocate snapshots or detection caches: inspect
  the selected `agent-detection` config and `agent-detection/remote` state paths;
  require bundled codex/claude sources via `herdr server agent-manifests --json`.
  conflicting overrides/caches/config require action, not deletion or reset.
- herdr and gateway have independent service lifetimes. no `PartOf`, stop hook,
  process-group ownership or launcher chain from gateway to runtime. gateway
  stop/restart may end bridge children, never herdr or workers. runtime stop
  deliberately ends worker lifetimes; cold restore creates new shells, not resumes.
  macbook's launchagent is login-scoped; do not promise logout survival.
- unchanged apply mutates/restarts nothing. missing managed runtime starts;
  changed operational inputs on a running runtime return `ACTION`/exit `2`
  without switching those inputs. operator explicitly stops that runtime under
  separate approval, then reapplies. no new restart flag or live-handoff machinery.
- keep skid's artifact cache, immutable generations, `current`/`previous`,
  nonblocking lock and verified per-host rollback. a failed gateway activation
  restores the previous gateway/config/unit only; it does not stop herdr.
- replace `tmux`/`nativeControlPath` host fields with `herdr`; retain the exact
  four ordered profiles, account homes, foreground signatures and permission
  flags. retain content-free process-bound identity hooks and terminal-local BEL.
  no native status/history helper or upstream status-hook integration.

the pinned [server entry](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/main.rs)
and [release](https://github.com/herdrdev/herdr/releases/tag/v0.9.1) are reference
inputs; installer/service behavior still requires qualification.

## 4. prepare, qualify, activate

1. freeze a release tuple in the runbook: full skid/jarvis/dev-server commits,
   herdr pin, skid host/apk hashes, signer, android candidate/emergency/next-forward
   version codes, previous installed artifacts/config/units and current journal
   compatibility. use existing pins/manifests; no new runtime release registry.
2. qualify the unmodified candidate on a nonproduction linux host with trusted
   `https://HOST:8443`, isolated gateway/herdr, test-owned workers and isolated
   jarvis postgres/runtime state. do not take production's serve mapping, relax
   origin/tls checks, copy production databases or run a second production bot.
   exercise real jarvis tools, dispatcher, gate and recorder through the actual
   cli/gateway, with real codex and claude. no scripted gateway counts as this proof.
3. publish via existing exact-main draft/release-integrity flow; publication
   remains separately authorized. record final artifact identities; if qualification
   bytes change, rerun only the affected boundary. stage validated artifacts and
   configs on all hosts before the window, using existing prepare/cache owners
   under their lock (`skidbladnir_prepare_artifact`, never `skidbladnir_apply`).
   staging changes no active pointers, units, hooks, ingress or cli copies.
   document exact invocation; no new public staging mode or fleet coordinator.
4. under the old jarvis release, inventory **every** `queued`, `awaiting_approval`
   and `executing` action, including mail/calendar approvals and future wakes.
   exhaust the status query; recovery selectors omit receipt-backed reminders.
   finish/reconcile or obtain explicit owner denial/cancellation through existing
   operations. explain cancelled reminders; replacement reminders require fresh
   owner authority. never restamp contracts, reset attempts, fabricate outcomes
   or cancel an uncertain external effect by sql. unresolved work holds activation.
5. finish or safely stop old main/isolated-role scopes before switching. inspect
   unprocessed input lineage and its model/read positions, including completed
   model decisions whose enclosing turn never settled. do not require all
   historical unknown rows to disappear. incompatible work remains fail-closed
   for operator repair, never replayed under the new plan.
6. close phone/interactive fleet callers; owner-pause jarvis under the old release,
   verify durable pause and settled active work, then stop it cooperatively. require
   confirmed inactive service and no main pid before replacing its unit, cli or
   release. repeat the inventory in a stopped one-shot under existing
   `deployment_ownership`; release that connection's lock before service startup.
   no concurrent dream/rebuild/operator process during the window; no maintenance
   daemon. untouched ingress/outbox and settled historical
   records remain intact. the current swallowed stop error must be removed.
   use the existing stopped admission-journal preparation if needed; preserve
   reservations/charges. never initialize empty state for an existing deployment.
7. within the owner-selected window, apply devbox, arch, macbook from the pinned
   installer closure; verify each before advancing. devbox apply also replaces
   `/usr/local/libexec/skidbladnir`, so jarvis MUST already be stopped. preserve
   handles, bearers, origins, signing material, provider accounts and unrelated
   private ingress. cli peer configuration remains mode `0600`, with three peers.
8. install the signed apk in place; verify stored pairing continuity separately
   from terminal behavior. activate matching jarvis only after fleet and cli
   identity checks. activation installs the selected release's own unit and
   reloads systemd, including on rollback; inactive release installation must
   no longer overwrite the global active unit. require the prepared paused state:
   startup recovery happens before ingress, so starting paused is not a read-only
   audit substitute. resume owner use only after recovery/containment checks pass.

stop on a failed host, expired window or mismatched tuple. keep mismatched callers
closed; use exact per-host probes to repair or roll back. no global atomicity or
zero downtime is promised. health is not model-turn acceptance.

## 5. rollback and retirement

- retain previous healthy host generations, unit/config inputs and installed
  cli bytes until coordinated acceptance; do not let a second apply prune the
  rollback baseline. do not downgrade herdr or stop its workers to undo a gateway.
  rollback restores required prior hook/helper inputs before starting the old
  gateway; a binary alone is not the prior operational configuration.
- before new jarvis durable work, a qualified same-schema previous release may
  return with its matching cli and unit. after new receipts/positions exist,
  use only a reader qualified for them; otherwise keep jarvis stopped and repair
  forward. check stored contract revisions, successful receipts as well as staged
  `agent_control_v2`, pending plans/scopes and admission compatibility. the zero
  incompatible-obligation rule is for this cutover, not every future deployment.
  no old runtime interpreter, database rewind, row deletion or budget reset.
- prepare a local emergency apk from the clean previous source with the same
  package/signer and a reserved **higher** version code, using existing
  `scripts/build release-assets` and current `scripts/check-release --source`.
  keep it local; do not relax exact-main publication or deploy its generated host
  bundles. reserve the next forward code above it if used. `install-android` stays
  in-place: no downgrade flag, uninstall, data clear, new invite or keystore reset.
  prove the package/data roundtrip on isolated android resources; production
  install checks pairing continuity, not the waived terminal journey.
- inventory installed legacy integration paths and callers before removal:
  provider-runtime-control's skid-owned installation, its pin/environment,
  obsolete tmux hook/config fragments. remove only those exact unreferenced
  owned assets, after rollback inputs are retained. old rollback code stays in
  its immutable artifact, not the new installer. no recursive account cleanup.
- preserve user tmux sessions, binaries, dotfiles and unrelated shared codex
  services. old sessions drain manually through tmux; herdr workers remain
  accessible through herdr during rollback. no pty transfer or cross-runtime refs.

## 6. exclusive implementation and content slices

assign one builder and one content designer per row; designers may be reused.
designers define the copy and its quality examples before code. root alone owns
shared specifications/composition; reviewers are read-only.

| owner | files / deliverable | what good content means |
| --- | --- | --- |
| skid boundary | `internal/agentcontrol/actions.go`, `cmd/skidbladnir/main.go`; reuse `internal/hostconfig`; narrow relevant contract/help docs | distinguish interrupt, close rejection, no close dispatch and unknown; validation never implies live readiness |
| deployment | dev-server `lib/herdr.sh` and `assets/herdr/{release-pin.json,config.toml,herdr.service,dev.niels.herdr.plist}` (new), `lib/skidbladnir.sh`, `assets/skidbladnir/` except `release-pin.json`, `workstation`, `devbox`, `ansible/roles/skidbladnir/tasks/main.yml`, staged-input inventory | existing `ACTION`/`DEFERRED` vocabulary; name affected consumer, worker-loss consequence and exact recovery action |
| jarvis activation | jarvis `deploy/install-release`, `deploy/activate-release`, `docs/operations.md`, narrow stop descriptions in `src/jarvis/agent_tools.py` | pending work and discarded approvals are explicit; unknown is not retryable; service active is not live qualification |
| root release/integration | skid `scripts/fleet`, build/release/install helpers only where needed, skid `release-pin.json`, dev-server `assets/skidbladnir/release-pin.json`, migration/roadmap/architecture; jarvis qualification/issues; dev-server `SPEC.md`/`README.md`/issues | separate staged, published, active and accepted; record per-boundary limits without printing content or credentials |

no android production change is planned. preserve encrypted pairing format and
keystore identity; losing old dashboard/scroll restoration is an accepted cost,
not permission to lose pairings. no visual assets or new ui copy system.

operator evidence is one dated markdown table, not a proof ledger or api:
`target | boundary | exact revision/artifact | operation | observed outcome |
PASS/FAIL/NOT_RUN | limitation/next action`. keep bodies, refs, prompts, account
data and secrets out. content review rejects “installed” as “working”, “sent”
as “completed”, or a waiver as a pass.

## 7. proof and completion

for each slice: adversarially review invariants → temporary failing boundary
proof → minimal implementation → passing proof → refactor and rerun → independent
review of behavior, copy, ownership and cleanup → delete tests before commit.
documentation deletions use source/caller/link checks, not artificial red tests.
no permanent harness or test-only production seams. this is a scoped temporary
exception to jarvis's testing reset, not a testing redesign.

| required boundary | acceptance |
| --- | --- |
| contract corrections | through real gateway/cli decoding, non-confirmation close rejection produces `refused`; pre-close failure stays `not_attempted`; lost close reply stays unknown with no replay; jarvis stages/settles each correctly. deterministic error injection is allowed for this narrow temporary proof, not the live journey |
| config/install | candidate validator agrees with gateway admission; invalid config/digest fails before activation; second apply is unchanged; failed gateway activation restores exact prior bytes/unit and leaves herdr worker alive |
| linux + darwin lifecycle | native supervision, shared desktop/gateway socket, gateway restart preserves exact worker/ref; isolated cold herdr restart invalidates old refs without provider auto-resume; no unrelated process touched |
| unmodified linux journey | both providers: list/start → info/current agent → inspect/readiness → one literal submission → distinct observed response → follow-up → interrupt. also ordinary-send refusal at unconfirmed readiness, deliberate terminal override after inspection, original-target replacement rejection, known refusal and lost-reply non-replay through jarvis |
| jarvis activation | failed stop or unresolved incompatible work prevents switch; selected unit/cli/release agree; journal charges/history persist; tested stopped recovery never redispatches a paid decision/read/write; rollback obeys receipt compatibility |
| android package/data | signed candidate installs in place; encrypted pairings remain usable without reenrollment. isolated emergency-apk roundtrip preserves them. waived terminal/lifecycle/manual journeys remain `NOT_RUN` |
| fleet | all three hosts advertise exact intended versions/config/profiles; configured launches use intended executable/environment/cwd/flags. one unavailable peer does not block control of another; installed jarvis reads each peer and controls approved test workers. do not repeat every paid journey on every host or reopen the waived mac claude model turn |

run skid `scripts/check verify`, jarvis `scripts/verify`, relevant installer
syntax/native checks and existing release-integrity checks; report their narrow
scope. tmux/live, adb/phone and provider spending require applicable current-turn
approval. missing boundaries are `NOT_RUN`, not passes. record unavoidable
follow-ups individually under the owning repo's `docs/issues`; resolve/delete
the two contract issues and live-acceptance issue only on their actual evidence.

done means source prerequisites merged, exact releases published/staged,
coordinated fleet activated, non-waived criteria passed, temporary resources
removed and docs reconciled. existing waivers require no new campaign. source
completion alone may be reported, but never as completed fleet migration.

costs accepted by this design: coordinated downtime; explicit disposition of old
pending work; no guaranteed jarvis binary rollback after new history; a prepared
higher-version emergency apk; transient view-state loss; frozen detection rules
may need explicit future repair as providers change; no retained regression
suite. prior accepted runtime/key/closure/check-write limits remain unchanged.
