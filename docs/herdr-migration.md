# herdr migration plan

2026-09-22: accepted direction and delivery plan. the original pr 1 `reconsider`
result prompted an [approved scope amendment](herdr-pr1.md#accepted-scope-amendment):
native herdr closure effects and single-terminal phone interaction without remote
click/drag parity. the user subsequently authorized the pr 2 source candidate
with unsupported keys visibly disabled or locally rejected. this branch contains
that candidate; its live acceptance and fleet deployment remain incomplete.
[the architecture](architecture.md) describes the candidate. [pr 1's contract](herdr-pr1.md#candidate-pr-2-contract)
owns exact wire shapes, and [pr 2](herdr-pr2.md) owns the implementation and
remaining gates. interactive provider readiness and the user-waived human phone
actions remain `NOT_RUN`; source completion is not a `proceed` or deployment
decision.

2026-09-23: [the accepted non-blocking follow-ups](herdr-pr2.md#accepted-non-blocking-follow-ups)
supersede blocking language below for the remaining phone/mac proofs, optional
identity carryover and dns residue. retain their evidence and revisit only on
actual-use failure or user request; do not move them to a later mandatory gate.

## outcome and scope

use herdr for terminal ownership and the desktop ui. retain skid's native android
experience, direct fleet gateways, and common cli for jarvis and other agents.
pause further work on skid's custom desktop terminal embedding while this
direction is qualified.

the purpose is to delegate terminal compatibility and agent interaction mechanics
to a maintained upstream. jarvis retains responsibility for assignments and result
acceptance. terminal status and successful input delivery never prove task success.

| owner | responsibility after cutover |
| --- | --- |
| herdr | host terminals, pane processes, layouts, desktop navigation, upstream observations and interaction primitives |
| skid gateway | one host's authenticated interface, product projection, configured profile launches, bounded controls and terminal bridges |
| skid android | fleet composition, dwarves, focused terminal, native input and navigation |
| skid cli | common fleet routing and explicit controls through the gateways |
| jarvis | intent, authority, assignments, result inspection and existing write uncertainty |
| dev-server | pinned installation, host configuration and independent runtime service lifecycle |
| providers | execution semantics, credentials and native conversation history |

candidate defaults:

- one herdr server per host; desktop and gateway address that same instance.
- one dwarf per live terminal; its foreground agent is a separate observation.
  manually created herdr terminals are discoverable alongside skid-created ones.
- phone grouping follows herdr workspaces; tabs and splits remain herdr layout.
  the old independent space labels are retired.
- automatic provider resume is disabled initially. cold restart creates new
  unnamed shell lifetimes without prior profile, objective or dwarf metadata.
  per the user's 2026-09-22 decision, a surviving native pane label is shown
  only as a hint; the same rule applies to newly discovered manual panes.
  actions require a fresh reference until named through skid. no persistent
  dwarf registry.
- phone entry acquires terminal control and detach/backgrounding releases it.
  takeover is explicit; geometry handback still requires product live proof.

retain machine identity, private ingress, authentication, encrypted pairings,
profile selection, directory browsing, pressure, and the native visual language.
preserve the provider permission policy unless a separately reviewed change says
otherwise. credentials and provider history stay on their hosts.
the accepted amendment retains keys, scrolling and local selection/copy, and
requires explicit closure-scope disclosure to phone users and cli/jarvis callers.
the user also accepted weaker terminal-only claude status/history/interrupt on
2026-09-22: sampled herdr status and bounded reads replace native claude history,
status and confirmed halt. jarvis must treat delivery and stop as unconfirmed
until it observes the actual work or exit.

excluded: a new terminal emulator, generic backend framework, permanent tmux/herdr
dual support, copied transcripts, chat ui, generalized hook runtime, worker
database, durable task scheduler, completion callbacks, automatic write retry,
automatic worktree policy, and migration of live ptys. using upstream worktrees or plugins
does not implicitly add those capabilities to skid.

## delivery order

these numbers name migration stages, not existing github pr numbers or the older
spaces/shells delivery sequence.

| pr | repository | dependency | status |
| --- | --- | --- | --- |
| 1. feasibility and contract | skid | isolated proof resources and applicable live/device approval | candidate contract accepted for implementation; remaining live proofs open |
| 2. complete skid cutover | skid | user-authorized candidate scope | merged as pr 134 at `9999033`; named remaining gaps accepted as non-blocking; not deployed |
| 3. jarvis alignment | jarvis | pr 2's final cli contract and runnable candidate | merged at `fb5e4a9`; boundary proofs recorded; unmodified live journey remains `NOT_RUN` |
| 4. deployment and fleet acceptance | dev-server, skid, jarvis | merged candidates, exact artifacts and applicable execution approval | source prerequisites implemented on `herdr-pr4` branches in all three repositories with deleted boundary proofs; merges, nonproduction qualification, publication, installation and activation pending approval ([activation record](issues/herdr-fleet-activation.md)) |

pr 2 changes gateway, cli and android together. use reviewable commits inside that
pr; do not release an intermediate schema that its consumers cannot read. pr 3
may be prepared against the candidate, but its deployment waits for the matching
cli. pr 4 can prepare installer changes earlier; activation follows qualification
and release verification. implementation, merge, publication, installation and
behavioral acceptance are separate outcomes.

## pr 1: feasibility and implementation contract

[the pr 1 execution spec](herdr-pr1.md) owns the bounded investigation, work split,
content design, contract outputs and adversarial temporary-proof workflow.

### scope

build disposable proof adapters around an isolated herdr server and test-owned
workers. exercise the android terminal boundary with the smallest temporary
client change. inspect the independent
[mobile relay](https://github.com/benkraus/herdr-plugin-mobile-relay) as a reference;
adopting its dependencies or broader product is not required.

commit the resulting contract, source-attributed findings and issue updates only.
remove experiment code and temporary tests before commit. no production adapter,
deployment change or later-pr scaffolding belongs in this pr.

### requirements

- select and record an exact upstream release/source. the research baseline is
  herdr v0.9.1, commit `065ef9d6a531c49fb8bee7e818ef837065b21ee9`; this is not an
  installed-version claim. distinguish release behavior from later upstream code.
- define terminal and agent references, including server restart, rename, pane
  movement, provider exit/replacement, and reused names. prove controls target
  the intended lifetime under pre-dispatch revalidation, with no implicit target
  substitution. accepted 2026-09-22: replacement between check and write remains
  possible, as in current skid; no atomic expected-worker guarantee is claimed.
  close dispatch revalidates the original terminal; effect scope follows herdr's
  native closure semantics, not an exact-terminal-only guarantee.
- define discovery, naming, dwarf metadata lifetime and assignment, workspace
  grouping, cross-host equal labels, unassigned terminals, and regrouping. specify
  create, new-terminal-here, rename, interrupt, stop, kill and detach individually.
  close/stop may close the linked git-worktree group. qualify and disclose that
  effect, preserve upstream refusal and report uncertainty under
  [the closure contract](herdr-pr1.md#accepted-scope-amendment); resolve
  [the native closure qualification](herdr-pr1.md#2026-09-22-reopened-qualification-at-the-accepted-scope).
- launch `personal`, `work`, `work2` and `claude-work` with the intended account,
  cwd and permission arguments. distinguish launch profile from proven current
  runtime profile. retain ordinary shells and hosts with zero agent profiles.
- specify status source, unknown/unavailable states, upstream idle fallback,
  `done`, readiness, and read coverage. any loss of current native claude
  status/history/stop capability must be named and accepted, never hidden in an
  unchanged field. do not implement a second detector to mimic upstream.
- define creation versus readiness, startup dialogs, bounded timeouts and partial
  creation. herdr cli startup waits must not silently exceed jarvis's existing
  operation budget. prompts are submitted once; lost replies remain uncertain.
- establish rendered-frame, terminal-reply, input, scrolling, selection, geometry,
  control acquisition/release and reconnect ownership. preserve literal multiline
  and composed input. attach one terminal, not the herdr ui; remote clicks/drags
  are excluded, while public mode-aware scrolling and local selection/copy remain.
  direct-controller exclusivity is not a global writer lock.
- publish exact gateway/cli/phone schema changes and an operation mapping, plus
  the intended no-argument `skid` and `skid enter` behavior when herdr owns desktop
  navigation. do not invent tmux fields or counts to preserve old schemas.

### completion criteria

1. a headless coordinator controls two independent workers without relying on ui
   focus; it reads, sends a follow-up and interrupts the intended worker.
2. profile launches and restart behavior resolve
   [the profile issue](issues/herdr-profile-restore.md), including the decision to
   disable automatic resume rather than promise account-correct restoration.
3. replacement, move, rename and restart experiments resolve
   [the targeting issue](issues/herdr-agent-targeting.md) without silently
   weakening the accepted revalidation contract. disclosed native closure and
   refusal qualify [native linked-workspace closure](herdr-pr1.md#2026-09-22-reopened-qualification-at-the-accepted-scope).
4. codex, claude and a shell work through the phone bridge, including shared
   desktop use and lifecycle cleanup, resolving
   [the terminal issue](issues/herdr-terminal-acceptance.md). qualify the host
   bridge on linux and darwin; a build alone does not establish that boundary.
5. a lost mutation reply is reported as uncertain without automatic resubmission;
   gateway/client loss leaves workers alive and later discovery truthful.
6. update affected architecture and feature contracts as the explicit migration
   target, keeping delivered tmux behavior distinguishable until pr 2 lands.
   record a proceed/reconsider conclusion and the remaining accepted limitations.

documentation can be reviewed before these proofs run. unperformed proof does
not authorize pr 2. a negative result is a valid completed investigation: record
the blocking evidence and revise the plan. if a small upstream change is needed,
scope it separately and qualify a release containing it before proceeding.

[pr 1's original evidence and decision](herdr-pr1.md#investigation-result-at-the-pinned-baseline)
remain recorded: the old terminal-only assertion failed, and no public remote
click path exists. the approved amendment changes the requirements, not those
results. the pr 2 source candidate has now replaced the tmux runtime in this
branch; installed hosts retain their prior release until pr 4 activation.

## pr 2: coherent skid cutover

[the pr 2 spec](herdr-pr2.md) owns the bounded implementation design, exclusive
file slices, content responsibilities, entry gates and red/green/refactor plan.
the user authorized candidate implementation with unsupported keys visibly
disabled or locally rejected; unresolved proofs still block acceptance.

### scope

replace the runtime integration across host, cli and android in one product pr.
retire skid's desktop browser in favor of herdr. preserve the useful surrounding
services and native phone interaction under the pr 1 contract.

### requirements and implementation ownership

| slice | intended implementation |
| --- | --- |
| upstream boundary | one concrete `internal/herdr` owner for json socket operations, bounded decoding/errors and terminal-session subprocesses |
| product operations | adapt `internal/sessions`, `internal/agentcontrol` and necessary profile types to the accepted terminal/agent model; remove duplicated observation when superseded |
| host composition | update `internal/hostconfig`, `cmd/skidbladnir`, gateway routes/dtos and terminal lifecycle; no request may restart the herdr server |
| fleet and cli | update `internal/fleetclient` and `internal/agentcli`; retain ordinary commands, opaque references, stdin prompts and truthful outcome envelopes |
| phone | update `ProductModel.kt`, `GatewayClient.kt`, `AgentControl.kt`, controller/dashboard projections, `TerminalConnection.kt` and affected terminal assets/composables |
| desktop retirement | remove `internal/sessionui` and obsolete attachment paths after their replacement is connected; implement the pr 1 desktop entry decision |
| contract and packaging | update feature docs, codebase map, applicable repository instructions, help, dependencies and release contents together |

the root integrator owns shared composition and specifications. if work is
delegated, assign exact non-overlapping files before edits.

- use socket requests for controls and prompt content. use the installed herdr
  terminal-session command for frame/input streams, without implementing its
  private binary protocol. sanitize upstream errors before logging.
- the gateway owns only its connections and bridge subprocesses. cancellation,
  detach and gateway restart must not own worker or herdr-server lifetime.
- retain independent host routing and outage handling. all actionable references
  stay bound to the selected machine and accepted runtime lifetime.
- input delivery, interruption and terminal closure have separate reported
  effects. preserve partial/unknown results; reconnect never replays input.
- reuse pairings and rendering/input components where the proven contract permits.
  adapt scrolling and terminal replies deliberately; rendered frames are not a
  raw provider byte stream. do not replace the phone with the desktop layout.
- remove tmux-specific schemas, adapters, hooks, native helpers and dependencies
  only after checking their consumers. preserve unrelated provider services and
  shared account wrappers. remove installed assets through pr 4's deployment owner.
- no legacy protocol decoder, backend switch or fake compatibility fields remain.

pr 1 caller audit adds explicit pr 2 review targets before deletion:
`internal/terminalclient` and `internal/terminal/protocol.go` own existing
attachment and failure handling; `internal/space` owns cosmetic-label validation
and grouping; `internal/logging` names tmux routes/events;
`cmd/skidbladnir/terminal_exec.go` and the main command own tmux version,
configuration and agent-hook entry. on android, review `Spaces.kt`,
`DashboardEntryState.kt`, `SkidbladnirController.kt`, `DashboardScreen.kt`,
`WorkingDirectoryPicker.kt`, `FleetPersistence.kt`, `TerminalScreen.kt` and
`TerminalSelection.kt` for old identity, grouping, scroll and attachment
assumptions. this is a caller audit, not a commitment to edit every file.
also inspect `SessionRename.kt`, `TerminalKeyDeck.kt` and
`LockedTerminalWebView.kt` before replacing their terminal contracts. the
`internal/agenthook` tmux binding is a deletion target, but its process-bound
identity role must first be judged against the pr 1 runtime-profile contract;
deleting a caller cannot manufacture proof of a current account profile.
pr 4 also audits `scripts/fleet` and installed hook/config references before
retirement. no production deletion belongs to pr 1.

### completion criteria

1. the changed candidate passes `scripts/check verify`; record that this covers
   engineering checks/builds only.
2. host and cli demonstrate inventory, all configured launch choices, bounded
   reads, literal sends, keys, interrupt, stop, disclosed native close/refusal, rename,
   grouping and new-terminal-here under the accepted contract.
3. android demonstrates fleet inventory, creation and focused attachment;
   automatable typing, synthetic paste, supported keys, scrolling, backgrounding,
   sizing and return navigation work on the physical phone with test-owned
   terminals. user-waived dictation, gboard paste, local copy and rotation stay
   `NOT_RUN`, without an inferred pass.
4. repeat the important pr 1 lifecycle/targeting proofs through the actual gateway,
   cli and phone paths on linux and darwin. direct upstream success is insufficient.
5. native herdr desktop sees and operates the same workers; selecting another
   pane does not hide an existing worker from the fleet projection.
6. engineering/source checks confirm the retired runtime/browser paths and unused
   dependencies are gone, the declared schemas agree, and the docs describe the
   new implementation. temporary proof code is removed before commit.

completion supplies a release candidate, not fleet deployment acceptance.
the [dated waivers](herdr-pr2.md#accepted-non-blocking-follow-ups) apply to the
criteria above and later rollout criteria; waived checks remain unrun, not passed.

## pr 3: jarvis consumer alignment

[the pr 3 spec](herdr-pr3.md) owns the exact consumer delta, non-overlapping
file slices, content contract, temporary proofs and coordinated cutover handoff.

### scope

change only jarvis's agent tools, strict result schemas, cli integration and
instructions that the new contract actually affects. preserve its cognition,
authority, durable actions, budgets and immutable history.

### requirements

- continue using `skid --json` and opaque refs. jarvis gains no independent herdr,
  gateway, ssh or provider client. literal prompt text continues through stdin.
- preserve one write attempt and existing partial/unknown outcome recording.
  metadata reads must not replace the original submitted target implicitly.
- align readiness, read coverage and status wording with the actual cli. a
  successful herdr wait is not an assignment receipt or task-success record.
- tool descriptions and schemas disclose that kill/stop closure may end other
  terminals in the linked workspace group. never imply single-worker effect or
  native halt of every affected provider; preserve refusal and unknown outcomes.
- initial delegation uses one outstanding assignment per worker and explicit
  response/work-product inspection. human intervention requires reassessing the
  observation; this convention is not an enforced exclusive-writer mechanism.
- preserve decoding/rendering needed for existing durable history. retire old
  execution grammar without deleting history or adding a worker ledger.

### completion criteria

1. jarvis discovers and starts codex and claude workers, observes readiness, sends
   literal multiline instructions, inspects results, sends follow-up input and
   interrupts the intended worker through the candidate cli.
2. a partial fleet result remains partial; a lost write reply or replaced target
   cannot cause an implicit retry or target substitution.
3. the changed repository's required checks pass, and the actual cli integration
   is qualified separately from fixtures. deployment stays coordinated with pr 4.
4. align the strict consumer with `terminals`, separate agent refs, removed
   client counts, new status/readiness and dispatch/partial outcomes. the current
   candidate makes this pr required; do not claim unchanged compatibility.

## pr 4: deployment and fleet acceptance

[the pr 4 spec](herdr-pr4.md) owns exact prerequisites, installation/lifetime
contracts, non-overlapping repository slices, durable-work handling, rollback
and acceptance. it carries the unmodified isolated-host proof, truthful stop
partials and removal of the unemitted start stage. it supersedes any broader
phone journey language below under the existing waivers.

### scope

dev-server owns herdr installation, service/configuration changes, matching skid
release pins, and removal of superseded installed assets. skid remains the owner
of host/apk packaging and release integrity. jarvis deployment consumes its
qualified candidate through its existing owner.

### requirements

- pin and verify the qualified upstream binaries and matching skid artifacts.
  configure the selected shared herdr instance, profiles and disabled automatic
  resume. update existing wrappers only where necessary; preserve account data.
- supervise herdr independently of the gateway on linux and darwin. gateway
  install/restart must preserve workers; upstream runtime restart is a distinct,
  explicit operation with its actual process-loss semantics.
- stage all artifacts and matching consumers before activation. first qualify on
  one isolated host instance, then switch the three-host fleet in a bounded window.
  use explicit per-host cli probes during staging; add no smaller-fleet phone mode.
- handle all nonterminal jarvis actions and incompatible unfinished model/read
  scopes under the old release before replacing their execution contract,
  including approvals and future reminders. preserve history and charged work;
  no restamping, silent cancellation or reinterpretation of old refs.
- preserve machine handles, bearers, phone pairings, signing identity and unrelated
  private ingress. remove only inventoried owned assets whose callers are gone.
- leave existing user tmux sessions running and accessible through tmux while they
  drain deliberately. do not migrate their ptys, kill them or uninstall tmux as an
  incidental cleanup. new herdr workers remain accessible through herdr on rollback.
- prepare matching host/cli/jarvis/apk rollback inputs. prove android's permitted
  version-code path and pairing preservation; do not assume a signed apk can be
  downgraded or require clearing app data. rollback does not transfer workers
  between runtimes or guarantee their display in the previous skid version.

### completion criteria

1. installer checks and release-integrity checks pass against exact artifacts;
   unchanged apply and gateway restart preserve runtime state and credentials.
2. arch, devbox and macbook advertise the intended runtime/profile configuration
   and successfully launch their configured profiles in approved test resources.
3. the signed phone installs in place with pairing continuity; the accepted phone
   journey waivers remain closed. native desktop and gateway share workers;
   jarvis's unmodified linux provider journey and installed fleet routing pass.
4. one unavailable machine does not block observation/control of available peers.
   no mixed-version fleet is declared complete.
5. rollback is exercised on isolated resources, including a higher-version
   emergency apk at the package/data boundary. new jarvis receipts/positions
   require a qualified reader or stopped forward repair, never history loss.
   record installation, live behavior and unperformed checks separately.
6. remove superseded installed integration assets only after verifying no remaining
   caller. reconcile issue records and roadmap status with actual evidence.

the deployment pr can be merged before activation. the migration is complete only
after these fleet acceptance criteria pass; a merged pr or published apk alone
does not complete this stage.

## verification and issue handling

follow [testing policy](rules/testing.md) and [repository guardrails](../AGENTS.md).
use temporary integration/live proofs, review their sensitivity, then remove them
before commit. retain engineering checks; do not recreate retired suites, test
harnesses or proof-ledger infrastructure. documentation-only changes use diff and
link checks.

use isolated test-owned herdr instances and synthetic work. any tmux operations
use only test-owned resources on an isolated `-L` socket. tmux/live and phone/adb
operations require the applicable explicit current-turn approval; this plan does
not supply it. missing boundaries remain `NOT_RUN`, never a pass.

keep results content-free: record versions, platform, operation, outcome and
limitations, not prompts, terminal bytes, account data or credentials. keep one
file per unresolved issue under `docs/issues`; update it as evidence changes and
delete it when its resolution criteria are met. accepted exclusions belong in
the final contract rather than permanent unresolved bug records.

research references: [agent automation](https://herdr.dev/docs/agent-automation/),
[status authority](https://herdr.dev/docs/agents/),
[terminal bridge](https://herdr.dev/docs/persistence-remote/#direct-terminal-attach),
[restore semantics](https://herdr.dev/docs/session-state/), and
[release source](https://github.com/herdrdev/herdr/tree/065ef9d6a531c49fb8bee7e818ef837065b21ee9).
documentation describes capabilities; only exercised boundaries establish acceptance.
