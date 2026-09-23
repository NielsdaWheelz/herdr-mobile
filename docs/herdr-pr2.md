# herdr pr 2: one runtime, one product contract

2026-09-22 · implementation specification. the user has since authorized work
on this pr, conditional on its entry gate.
one skid pr; reviewable commits, no intermediate release. [the migration plan](herdr-migration.md)
owns delivery. [pr 1's candidate](herdr-pr1.md#candidate-pr-2-contract-proposed-not-implemented)
owns exact wire shapes and upstream mappings; this document owns implementation
structure and acceptance. do not maintain a second schema here.

## entry gate

pr 1 remains incomplete. v0.9.1 cannot receive `proceed`. before implementation:

- qualify a public mode-aware key operation covering the retained phone and
  desktop input vocabulary; record its exact release/source and operation map.
  no raw-csi workaround, hidden key removal or private upstream protocol.
- qualify the desktop input decoder path for literal, bounded paste and
  cancellation; [the recorded defects](issues/herdr-terminal-acceptance.md#pr-2-input-contract-review)
  prevent using the pinned ultraviolet `TerminalReader`.
- close the recorded readiness, process-bound claude identity, phone typing,
  provider scrolling, background release and linux stream/desktop checks.
  [pr 1's evidence](herdr-pr1.md#2026-09-22-reopened-qualification-at-the-accepted-scope)
  distinguishes failures from missing proof. none becomes a pass here.
  [claude submission](issues/herdr-claude-submission.md) is separately open.

the user waived the human dictation, gboard paste, local copy and rotation
checks for this candidate. record each as `NOT_RUN`; record later failures as
live-repair issues. do not infer that these behaviors work.

then freeze the candidate and close its issues. the user's task already
authorizes implementation once these prerequisites hold.
pr 2 repeats important proofs through product code. a disposable adapter cannot
qualify the implementation that replaces it.

accepted 2026-09-22: retain the full phone and desktop key vocabulary and defer
cutover until a qualified released herdr public operation supports it. do not
remove controls, guess raw csi, ship a partial runtime, or ask herdr to conform
to skid's unpublished contract. no external request or patch was submitted.

accepted 2026-09-22: ordinary `skid send` requires recognized idle and rejects
working/unconfirmed states before dispatch; it does not wait or queue.
`send --terminal` remains the explicit readiness override. this closes the policy
decision, not the outstanding readiness proof.

## target and limits

one dwarf represents one live herdr terminal, including manual panes. android
attaches that terminal, not herdr chrome. the native desktop and headless controls
address the same workers, independently of desktop focus.

```text
jarvis / workers -> skid cli -> selected gateway -> herdr -> provider
android --------------------> selected gateway -----^
local skid -> native herdr desktop ------------------^
```

herdr owns terminals, processes, modes, layout and observations. skid owns
authenticated projection, exact-reference controls and native phone presentation.
providers own execution/history; jarvis owns assignments and result acceptance.
dev-server independently supervises herdr. no request starts/restarts its server.

explicit costs: native group closure can end linked workers; check/write retains
its race; desktop/api writers can coexist with phone control; claude loses native
status/history/halt confirmation; cold restart loses worker/metadata lifetimes;
unmarked native labels are hints, not names; workspace moves change real layout;
bare `skid` becomes local desktop, not a fleet browser. remote clicks/drags, live
pty migration, automatic resume/retry/replay, runtime fallbacks, persistent worker
registries, new orchestration engines and visual redesign are excluded.

pr 3 is required: jarvis's strict consumer cannot decode the new contract.
pr 4 owns installation, coordinated activation and rollback. pr 2 produces a
coherent runnable candidate; it neither deploys nor upgrades the installed phone.

## contracts and composition

reuse existing owners; no backend interface or new service framework.

| owner | contract and lifetime |
| --- | --- |
| `internal/herdr` | concrete public socket requests and terminal-control child; bounded codecs and sanitized errors. returns an owning stream handle whose close cancels io, releases control and joins the child; never signals herdr or its workers |
| `internal/sessions` | terminal/workspace discovery, original-ref resolution, metadata, name/cwd validation, creation/move/close. reuse its mutation lock for metadata reread-if-absent/write/readback; inventory and create share it |
| `internal/agentcontrol` | enrich observations and enforce foreground identity, readiness, bounded reads and ordered send/interrupt/stop policy. depends on sessions, never the reverse; no local screen detector |
| `internal/agentruntime`, `internal/process`, `internal/agenthook` | retain launch/profile validation and platform process observations; adapt only the content-free process-bound identity registration. launch metadata never proves the current account |
| `internal/gateway` | auth, machine fence, routes/dtos and attachment lifetime; opens/closes the upstream handle. owns websocket heartbeat, cancellation and shutdown, not a second child supervisor |
| `internal/fleetclient`, `internal/agentcli` | one opaque-reference codec, peer routing, strict results, parsing/help and literal stdin. names/labels select observations; never substitute a refreshed action target |
| `internal/terminal`, `internal/terminalclient` | shared typed stream codec/bounded queue; local tty acquisition, decoded input, resize, detach and restoration. retire old protocol, not these retained responsibilities |

`sessions` projects provider identity using upstream classification plus host
process evidence, including on zero-profile hosts; configured launch signatures
alone cannot discover those workers. `agentcontrol` enriches that projection with
upstream status/readiness. gateway composes both; do not introduce an import cycle.

refs identify terminal, agent and workspace lifetimes separately. preserve rename,
move and gateway restart; reject old terminal refs after runtime restart and old
agent refs after observed replacement. locally serialize lifetime-token claims;
upstream metadata keys are globally shared, not protected by `source`. reserve
`skid_*` by convention; conflicting/exhausted metadata makes a resource
unaddressable and inventory partial. no fabricated ref or empty-host projection.

the candidate's command fingerprint is a privacy-preserving equality check, not
reference authentication. auth owns its key; control requests a fingerprint,
never loads credentials itself. bearer replacement invalidates agent refs. at
cutover, make this narrow distinction explicit in [identity rules](rules/keys-and-identities.md);
do not add signed refs, authorization claims or a second secret/identity store.

### local configuration

`hostconfig` replaces `tmux` and `nativeControlPath` with required
`herdr:{path,socketPath,testedVersion}`; retain `platform` and the existing
`profiles` schema, including zero profiles. paths are canonical absolute paths;
`testedVersion` is the exact qualified version, checked against the configured
executable at entry. each new upstream connection also validates public `ping`
against the qualified server version/protocol pair; a new binary does not prove
which server owns the socket. mismatch fails without restart. reject retired
fields. no version negotiation or autostart.

bare tty `skid` reads `host-config.json` beside its resolved executable, matching
the existing installed generation layout. `skid --host-config /absolute/path`
explicitly overrides that file for local use/qualification; it is not a fleet
flag. gateway and identity hook retain their explicit host-config argument.
missing/invalid configuration fails, with no search or fallback. `client.json`
and `--config` remain fleet-only; remote commands need no local runtime.
dev-server supplies installed paths and upstream resume configuration in pr 4.

the gateway dials `socketPath`; every herdr child receives that value as
`HERDR_SOCKET_PATH`, overrides inherited routing and passes no `--session`.
bare `skid` execs `herdr client` after loading the same configuration, without
opening a gateway attachment. herdr derives its client socket. no default-socket
fallback. `skid enter` remains an exact fleet attachment through the gateway.

### api and mutations

implement [the operation table](herdr-pr1.md#retained-gateway-operations) literally:

- `/v1/terminals` owns inventory/create/info/rename/move/shell/kill/stream;
  `/v1/agents` owns read/send/keys/interrupt/stop. retire `/v1/sessions`.
- `terminal.ref`, `agent.ref` and `workspaceRef` remain distinct. clients preserve
  original action targets across refresh. `nativeLabel` never participates in
  name selection; duplicate names/destination labels remain ambiguous.
- create validates first, reuses the upstream root pane, establishes metadata,
  then submits exactly one configured launch. return creation/partial facts, not
  readiness. no objective-as-prompt, compensating close or second launch.
- ordinary send requires `readiness:ready`; `--terminal` bypasses readiness only.
  both recheck the original worker immediately before dispatch. stop interrupts
  once, then revalidates before
  native close; unknown interrupt means close not attempted. close refusal stays
  refused; never change configuration or switch to a broader close primitive.
- preserve `not_sent | sent | unknown`, partial stages, ten-second host budget
  and fifteen-second client deadline. `sent` is not effect or task completion.
- pairing, machine/auth, directory and pressure routes retain their schemas.
  one unavailable peer does not block explicitly selected available peers.

`skid move` replaces cosmetic space assignment. create destination is an exact
workspace ref or explicit new workspace; omitted means new. no `unassigned`,
clear-membership command, fake client count, alternate decoder or migration shim.

## attached terminal

one websocket attempt owns one upstream control handle. acquire without takeover;
explicit takeover retries acquisition against the original terminal ref, never
replays input. reentry after loss starts with discovery and a new attempt.
wait for valid fitted initial `Resize` before opening upstream control with that
geometry. the gateway allows ten seconds from websocket admission for geometry
and the first full frame together; timeout releases the child and reports
`control_unavailable`. client application of that frame must complete within its
fifteen-second acquisition deadline or detach. input remains disabled throughout.

the gateway validates frames and withholds input until it emits a valid first
full frame. the phone applies that frame's reset, geometry and ansi before its
page acknowledges application to the current connection attempt, which then
enables input. no websocket acknowledgement is added. native code checks
contiguous sequence numbers; decimal strings preserve uint64 through javascript.
retain acknowledged rendering/backpressure. gaps, malformed data and overflow
end only the attachment. qualify the candidate's frame/queue caps against the
largest intended geometry/output; exceeding a cap disconnects rather than grows
memory. tiny pr 1 samples are not that qualification.

input is semantic `Text | Paste | Key | Scroll | Resize | Detach`, never a merged
xterm `onData` stream. typed text, paste and keys use one ordered public control
queue; herdr performs mode-aware encoding. swipes/page scrolling and resize use
the public terminal stream; they have no total order with socket or desktop input.
text/paste limits and error codes follow the candidate. no emulated-terminal
reply is sent back to the provider.

phone page owns composition, modifier toggles, gestures and viewport-local
selection/copy. keep the immutable copy snapshot through incremental output;
clear it on full redraw/resize, attachment loss and explicit cancellation.
native code owns credentials and transport. ctrl/alt apply only to proven discrete
keys/deck actions; ime/dictation, paste and uncertain text consume them unchanged.
retain hardware keys and enter/backspace as well as every deck key.

desktop input frames escapes with the already pinned `x/ansi` parser, decodes
complete named/modified keys and terminal replies with ultraviolet's public
`EventDecoder`, and handles utf-8 text and bounded literal bracketed paste in
`terminalclient`. reuse its cancellable tty reader; do not use the defective
`ultraviolet.TerminalReader` or bubbletea. map events to the shared wire, enable
local bracketed-paste reporting, consume replies locally and never submit a
paste. `ctrl-] d` is local detach only outside paste; unknown keyboard and
terminfo-only sequences end attachment with an explanation, not raw forwarding.
preserve ctrl/alt on kitty key events carrying text; resolve or explicitly
reject ambiguous key/reply events. feed bounded chunks before paste/escape
accumulation. release tty modes/readers on every exit. no copied parser or private upstream
protocol.

`SkidbladnirController` remains the sole owner of selected target, foreground
state, attempt generation and navigation. `TerminalConnection` owns one socket,
not reconnection policy. background invalidates input admission before cancelling
resize/render callbacks and releasing once; late callbacks cannot reopen input.
gateway ping every two seconds/six-second unanswered-pong deadline handles abrupt
loss; okhttp already answers pings. it does NOT repair a backgrounded live socket.
graceful release must finish without waiting indefinitely for `onClosed`.
qualify geometry handback within two seconds after detach/background and eight
seconds after abrupt loss. these are acceptance targets, not current results.
retain the gateway's existing two-second bearer revalidation separately from
ping/pong. unreadable/replaced credentials stop input and release only the
attachment; liveness is not continuing authorization.

keep native fleet/forge/directory/pressure/landmarks and existing polling lanes.
reset only obsolete dashboard restoration data; keep encrypted pairings untouched.
after native close, invalidate and refresh the affected host inventory: linked
closure can remove several cards; do not invent its affected set from selection.

## non-overlapping implementation slices

root freezes types, signatures and wire mappings before parallel edits. one owner
per file, including both sides of any rename; extend a slice explicitly, never by
convenience. android paths below are under its existing java package unless stated.

| slice | exclusive paths |
| --- | --- |
| runtime/control | new `internal/herdr/`; `internal/{sessions,agentcontrol,agentruntime,agenthook,process}/` |
| cli/desktop | `internal/{fleetclient,agentcli,terminalclient,terminal}/` |
| phone domain | `ProductModel.kt`, `GatewayClient.kt`, `AgentControl.kt`, `SkidbladnirController.kt`, `MainActivity.kt`, `DashboardEntryState.kt`, `DashboardScreen.kt`, `ForgeSheet.kt`, `SessionCard.kt`, `SessionRename.kt`, `Spaces.kt`, `SpaceSheet.kt` |
| phone terminal | `TerminalConnection.kt`, `TerminalScreen.kt`, `LockedTerminalWebView.kt`, `TerminalKeyDeck.kt`, `TerminalSelection.kt`; `android/app/src/main/assets/terminal/terminal.js` |
| root integration | `cmd/skidbladnir/`, `internal/{hostconfig,gateway,auth,logging}/`, `go.mod`, `go.sum`, `scripts/`, `docs/`, `AGENTS.md`; android gradle files, `app/src/main/res/`, `terminal.lock`, `xterm-6.0.0-skidbladnir.patch`; final deletion of `internal/{tmux,sessionui,space}/` |
| independent reviewers | read/run only; no test, production or content edits |

before phone parallelism, coordinate the existing
[terminal-state ownership repair](issues/terminal-ui-state-owner.md): domain owner
removes declarations, terminal owner houses them in `TerminalScreen.kt`; agree
typed frame/input callbacks. no new lifecycle module. move reusable workspace
label validation out of `internal/space` before removing cosmetic grouping.
pairing/store/scanner, polling, directory, pressure, catalogue, art, themes and
font-sizing owners are reuse targets, not additional cleanup scope.
root assigns a separate `mktemp -d` proof directory per builder; no shared probe
writer or permanent fixture package. reviewers write nothing.

each builder has a content designer; before code, that designer supplies a small
state → wording → action → accessibility table using existing surfaces. another
reviewer checks that each statement is justified by a schema fact.

| feature / designer | good content and acceptance |
| --- | --- |
| runtime/control | separate created, ready, delivered and completed. `terminal created; startup not confirmed`; `delivery unknown; inspect before sending again`. no native-halt or task-success claim |
| cli/desktop | examples use separate terminal/agent refs and literal stdin; help discloses group closure, local desktop entry, ambiguity and unknown outcomes before use |
| phone domain | distinguish name from `herdr label` hint; unnamed cards show short ids; duplicate workspace labels show distinguishing ids. confirmations: `close <terminal> on <machine>? linked workspaces and their running terminals may also close` |
| phone terminal | `another connection may be active` offers explicit takeover/back; loss says input is paused and nothing will be resent. detach says `leave terminal running`; sizing help says desktop input remains active |

no new art or lore in operational errors. idle-only ordinary-send wording must
distinguish working from readiness, not describe a busy worker as broken.

## sequence, proof and completion

1. **contract:** close entry gates; freeze exact schemas, local configuration,
   public primitives, input mapping and slice signatures. independently review
   authority, partial outcomes, content and deletion reachability before coding.
2. **red:** write a small temporary end-to-end set below, through real boundaries.
   retained behavior gets an old-system baseline, which may pass; defect fixes
   must reproduce the genuine failure first. new herdr routes first execute on
   the candidate: an old-protocol rejection is not behavioral red. sensitivity
   controls are separate evidence; unavailable boundaries remain `NOT_RUN`.
   do not manufacture red or recreate retired harnesses.
3. **green:** implement runtime/composition, then cli and phone against the frozen
   contract. review each slice's implementation and assertion sensitivity; rerun
   unchanged assertions. upstream deficiencies return to their owner, not a skid hack.
4. **refactor:** remove duplicated/obsolete paths, simplify under [the rules](rules/index.md),
   review independently and rerun affected proofs. then delete temporary tests,
   probes and owned resources. this leaves no retained behavioral regression suite.
5. **integration:** reconcile docs/help/env references and dependencies; run
   `scripts/check verify` on final source and obtain native darwin build evidence.
   final review covers complete diff, decoder agreement, cleanup and honest results.

| actual boundary | completion criterion |
| --- | --- |
| gateway + cli, linux and darwin | inventory/manual discovery, all four launches, shell/zero-profile, recognized readiness/working/blocked/unknown, bounded reads, literal input, rename/move/new-shell; ordinary send rejects busy/default-idle without dispatch, explicit override sends once, both reject stale workers; one root pane, no profile inference |
| targeting + orchestration | original-ref matrix across replacement/rename/move/name reuse/runtime restart; an unfocused second worker unaffected by read/send/interrupt. a real provider worker controls another through the same cli |
| uncertainty + ownership | suppress one actual reply after dispatch: one effect, truthful unknown/partial, no replay. client/gateway loss preserves workers; runtime restart creates new lifetimes with resume disabled |
| native closure | ordinary close, linked cascade and confirmation refusal; outside-group terminals survive. stop reports its separate steps, rejects observed replacement, and stops after unknown interrupt |
| desktop | bare skid attaches existing local server and refuses absent server; `enter` qualifies typing/keys/paste, resize, detach and tty restoration through the real gateway on both platforms |
| physical phone | shell, codex and claude: exact automatable typing, synthetic multiline paste, full key vocabulary in relevant modes, actual history scroll, sizing, takeover and return navigation; linux and darwin host stream boundaries. human dictation, gboard paste, local copy and rotation remain `NOT_RUN` under the user waiver above |
| phone lifetime + bounds | background during acquisition, first frame and steady state; abrupt loss including retained tcp relay, and bearer revocation. desktop regains geometry within the stated deadlines, worker survives, stale callbacks/input never resume; stress intended maximum frames and overflow closure. rotation remains `NOT_RUN` under the user waiver |
| retained services + retirement | pairing/machine rejection, directory and pressure smoke through actual routes; old config/routes/commands reject. source/import/build checks show no tmux/browser/local detector/native-control execution path or unused dependency |

root removes `terminal_exec.go` and old tmux hook/config/logging fields; retire
bubbletea/pty only after final caller checks. retain shared process/profile/auth,
local tty cleanup, renderer/input/selection and provider account wrappers. update
architecture, agent-control, spaces/shells/desktop/terminal specs, codebase map,
roadmap and `AGENTS.md` to the implemented contract, without retaining alternative
runtime instructions. installed helpers/services are pr 4's separate inventory.

live/device/tmux work requires explicit current-turn approval under [AGENTS.md](../AGENTS.md).
use isolated test-owned servers and a separate proof-package build of the actual
candidate, with unchanged production transport; no pairing store bypass in product
code. never clear device-wide logs. compare synthetic contents in memory; record
only source/build, platform, scenario, `PASS | FAIL | NOT_RUN`, outcome and limit.
keep unresolved [issue records](issues) current; delete resolved ones, not evidence
of outstanding failures. no `proceed`, merge or engineering check implies deployment.
