# herdr pr 1: feasibility and contract closure

2026-09-22 · execution specification with a negative pinned-baseline result.
[the migration plan](herdr-migration.md) owns delivery order. this pr owns the
evidence and precise contract needed to authorize pr 2. current production
contracts remain unchanged. an isolated darwin closure proof now fails; other
runtime/device acceptance remains `NOT_RUN` unless recorded below.

## goal, scope and final state

prove that herdr can own terminals and desktop interaction while skid retains
its native phone and common fleet cli. test the difficult boundaries first;
do not build a miniature replacement product to demonstrate feasibility.

pr 1 may build disposable integration probes and a minimal phone adapter.
its committed result is documentation only: a qualified upstream pin, closed
capability/wire contract, affected-contract deltas, evidence, and a
`proceed | reconsider` conclusion. a negative finding is a valid investigation;
missing evidence is not completed qualification. no production code, installed
configuration, release, deployment or existing user terminal changes.

## composition and ownership

```text
jarvis / agents -> skid cli -> one selected host gateway -> herdr -> provider pty
android --------------------> same gateway -------------> same herdr terminal
native herdr desktop -----------------------------------> same herdr server
```

herdr owns process/terminal lifetime, layout, terminal emulation and its agent
observations. the gateway translates one host's capabilities and owns only its
socket connections and terminal-stream subprocesses. clients own presentation;
jarvis owns assignments and result acceptance. dev-server owns installation and
the independently supervised herdr service. no gateway request starts, replaces
or restarts that service. providers retain credentials, execution and history.

use the json socket for control; qualify the installed
`herdr terminal session control` command for terminal streams. no private binary
protocol implementation, mobile relay dependency, new emulator or backend interface.
inspect the [existing mobile relay](https://github.com/benkraus/herdr-plugin-mobile-relay)
for lessons, not as a required component.

## decisions and their costs

| decision | target and explicit cost |
| --- | --- |
| worker targeting — accepted by user | revalidate the original terminal and foreground-worker lifetime immediately before dispatch; reject observed replacement and never substitute a refreshed target. replacement between check and terminal write remains possible. no atomic compare-and-write or exclusive-writer claim. |
| terminal granularity | one dwarf per live terminal, including manually created terminals, independent of selected pane. rename preserves identity; process replacement preserves the dwarf but invalidates its old agent reference. more cards than today's one-active-pane-per-session view. |
| runtime restart | one shared herdr server per host; disable automatic provider resume. client/gateway loss preserves workers; cold runtime restart ends their old lifetimes. restoration is not live process migration. |
| grouping — proposed replacement | use actual host-bound herdr workspaces. equal labels group visually across hosts, never identify mutations. moving a terminal can change desktop layout; retire cosmetic space assignment and clearing, not conceal them behind the old command. |
| provider capabilities — proposed replacement | herdr observations and bounded terminal reads for both providers; no second skid detector. this loses native claude history/status/halt confirmation. document and obtain explicit acceptance of that loss before proceeding; otherwise reconsider adoption. |
| phone control | opening acquires direct control without takeover; conflict offers explicit takeover or back. no observe-first mode or fallback. phone sets geometry while connected; desktop/api input can still occur. |
| desktop entry — proposed replacement | no-argument `skid` opens the configured local herdr desktop; `skid enter` retains exact fleet-target attachment through the gateway stream. no inference from gateway origins to ssh hosts; native herdr remote navigation remains upstream-owned. qualify both paths. |

the existing [targeting issue](issues/herdr-agent-targeting.md) tracks whether the
accepted revalidation boundary is implementable. current skid also checks then
writes; its mutex does not prevent independent process exit. terminal destruction
must still address the confirmed terminal lifetime, never a name-reused replacement.

## capability and schema contract to close

these are semantic requirements, not assertions that upstream exposes matching
fields. pr 1 must supply the exact encodings and upstream source for every field
before `proceed`. an opaque wrapper cannot create a missing lifetime guarantee.

```text
terminal target = machine handle + upstream terminal lifetime
agent target    = terminal target + observed foreground-worker lifetime
workspace      = host-bound upstream workspace identity + display label
terminal       = target, name, character, workspace, cwd?, launchProfile?,
                 objective?, agent?
agent          = target extension, provider, provenRuntimeProfile?, status, methods
status         = state, source, reason?
read result    = text, source, scope, truncated
write result   = method, outcome: written | unknown
stop result    = separately observed agent outcome + terminal closure outcome
```

- retain machine/bearer binding, peer-oriented fleet outcomes, observation time,
  strict handwritten codecs and omitted optional fields. reuse opaque-reference
  handling; clients never assemble targets or receive account paths. remove tmux
  ids, pid-shaped assumptions and client counts unless real evidence justifies
  their replacement. no placeholder values or legacy reader.
- references survive rename and gateway restart while their targets survive.
  cold restart and terminal replacement invalidate old references; foreground
  replacement invalidates only the agent portion. qualify movement: preserve the
  reference if upstream identity permits it, otherwise return stale and require
  explicit reselection. never resolve old references through current names.
- distinguish launch profile from proven current runtime profile. preserve the
  four configured launch rows, exact executable/arguments/environment/cwd,
  permission policy, ordinary shells and zero-profile hosts. choose one proven
  launch route; canonical `agent.start` is not automatically equivalent to the
  configured account wrapper. launch sends no objective or initial prompt.
- define where name, character, launch profile and objective live, including
  manual creation and gateway restart. reuse upstream metadata where supported;
  do not add a registry. if an existing metadata capability cannot survive,
  record the proposed loss for explicit acceptance rather than silently omit it.
- name observation source and coverage truthfully. upstream `done` is not task
  success; map it to idle without adding unread tracking. distinguish recognized
  readiness from upstream default-idle inference. unfamiliar/unproven readiness
  cannot authorize automatic prompt submission. no local substitute detector.
- create returns a terminal, not readiness. retain the ten-second host operation
  budget and fifteen-second client deadline; do not inherit a longer upstream
  startup wait. later failure may leave a visible terminal. report uncertainty,
  never retry creation or compensate with an unproven kill.
- retain bounded input/read/error envelopes and logical keys from
  [agent control](agent-control.md#capability-and-api-contract). `written` means
  delivery only; interrupt does not prove cancellation; closure does not prove
  detached descendants stopped. no automatic write retry or reconnect replay.

publish one compact wire table covering the retained `/v1` operations and cli
commands. each row specifies **request, success, typed failure, target predicate,
effect/dispatch boundary, bounds, uncertainty and upstream primitive**. include
inventory/info, create, rename, workspace move, new-terminal-here, read, send,
keys, interrupt, stop, kill, enter, acquire/release and reconnect. spell out every
gateway/cli/phone field or command removed or renamed; do not preserve `space`
semantics accidentally. new-terminal-here samples the exact source's cwd and
workspace, creates an independent shell, and never splits or replaces its source.
forge offers an exact existing workspace or explicit new workspace. qualify
workspace inventory/creation and root-pane reuse: creating one dwarf must not
leave an extra shell. reject ambiguous equal-label destinations; retire
`unassigned` rather than invent a workspace. allow removal of an emptied
tab/workspace, but reject any close that would destroy another terminal,
including a linked worktree. see [the closure issue](issues/herdr-terminal-closure.md).

## phone boundary and content design

retain native fleet, forge, directory chooser, pressure, dwarf landmarks,
key deck, composition, selection and font sizing. no visual redesign or new art.
the designer for each slice defines content alongside its schema, before coding:

| feature / designer owner | what good content must communicate |
| --- | --- |
| runtime/control — runtime slice | separate creation, readiness, delivery and completion. e.g. `terminal created; startup not confirmed`, `delivery unknown; inspect before sending again`, `interrupt sent; cancellation unconfirmed`. never label inferred status native. |
| grouping/creation — phone slice | name the machine, exact terminal and actual workspace; say when a new workspace will be created. use `move to workspace` and disclose layout effects. unnamed workspace labels are presentation, not fabricated resources. no dwarf lore in operational errors. |
| attachment/input — phone slice | `another terminal connection is active`, `take over`, `back`; loss pauses input on this phone, not remote work. disclose that desktop input remains active. destructive copy names target and effect; detach never implies stop. |
| acceptance — root | concise observed facts, not assurances. state platform/version, boundary, result and limitation without terminal content or account data. |

qualify and write down one owner for each stream behavior:

- herdr emulates the provider terminal; xterm renders frames. distinguish user
  text/keys/paste from emulator-generated replies. do not blindly forward today's
  merged `onData` stream or strip ime input to make the experiment pass.
- define full-frame admission, subsequent-frame ordering, reconnect reset,
  decoded-byte/frame/queue limits and overflow closure. input stays disabled until
  control and initial geometry/frame are established. measure against existing
  bounds; any necessary bound change is an explicit contract delta.
- herdr owns attached scroll position and modes; prove the route for touch,
  logical keys, mouse reporting and alternate-screen history. reuse gestures,
  not assumptions that local xterm scrollback is authoritative.
- keep selection/copy phone-local; define redraw/resize invalidation. prove
  geometry handback after detach, takeover, backgrounding and abrupt loss.
  reconnect needs fresh discovery and attachment; never input replay.

## work split and files

root assigns exact temporary paths before execution. temporary directories are
created with `mktemp -d`; no shared probe file has two writers. every builder also
owns its slice's content definitions above. a different reviewer challenges each
step; the reviewer writes neither tests nor production files.

| exclusive owner | files and boundary |
| --- | --- |
| runtime builder | one temporary runtime-probe directory only: upstream socket/launch/reference/status experiments and a minimal gateway-shaped transport shim; supplies findings, not product changes |
| phone builder | isolated source copy in its temporary directory; changes only to `android/app/build.gradle.kts`, `android/app/src/main/java/dev/niels/skidbladnir/{MainActivity,TerminalConnection,LockedTerminalWebView}.kt` and `android/app/src/main/assets/terminal/terminal.js`; the page decoder was added by root scope review for public scroll testing, with no repository source edit |
| root integrator | this spec, `docs/herdr-migration.md`, `docs/roadmap.md`, four `docs/issues/herdr-*.md` records; at closure, explicit target amendments in affected architecture/feature docs, never false delivered-state claims |
| independent reviewer | read/run only; reviews target semantics, test sensitivity, content, conclusions and final diff; no edits |

reuse/read first: `internal/{agentruntime,process,sessions}` for profile/cwd and
lifetime checks; `internal/{strictjson,auth,gateway}` for boundary contracts;
`internal/fleetclient/{request,response}.go` for references/outcomes;
`TerminalConnection.kt` and `terminal.js` for transport/input. inspect callers
before proposing consolidation. pr 1 lists pr 2's replacements in
`internal/{tmux,sessionui,agentcontrol,agenthook}`, terminal attachment owners,
gateway/fleet codecs and phone product models; it deletes none of them.

the disposable phone app uses application id `dev.niels.skidbladnir.herdrproof`
and a fixture entry into the existing renderer/native input components. use only
synthetic credentials and approved isolated endpoints, never installed pairings.
this bypasses onboarding for the experiment, not transport correctness; it proves
neither production pairing nor fleet integration. install/remove only that proof
package under current-turn approval. preserve the installed skid app and data.
any additional file or test-ingress configuration needs an explicit scope review.

## execution, adversarial review and acceptance

1. pin the upstream release and source; start from v0.9.1,
   `065ef9d6a531c49fb8bee7e818ef837065b21ee9`, not moving documentation or head.
   review the ownership/contract and smallest probe design before running it.
2. establish the unmodified baseline through real boundaries. write temporary
   end-to-end assertions first. a genuine failed requirement is red; a missing
   adapter, unavailable device or intentionally broken mock is not. a capability
   already working upstream need not be broken to manufacture red.
3. for each failure, review its cause and assertion sensitivity, implement the
   smallest correct disposable adaptation, rerun unchanged assertions, refactor
   under the [rules](rules/index.md), review again, then rerun affected assertions.
   negative controls establish sensitivity, not product acceptance. an upstream
   defect needs a separately scoped fix and qualified revision, not a skid hack.
4. independently challenge the evidence and closed contract; delete all temporary
   code/tests/resources owned by this investigation before commit. retain only
   content-free findings and unresolved [issues](issues). pr 2 repeats the important
   proofs through its actual implementation.

| proof | completion criterion |
| --- | --- |
| targeting and orchestration | a headless coordinator lists, reads, sends a literal multiline follow-up and interrupts worker a while desktop focus is on worker b; b is unaffected. an approved provider worker also invokes the same probe surface to control another worker; no special coordinator role. |
| lifetime | exercise each retained mutation after rename, move, foreground replacement, terminal/name reuse and server restart. already-replaced targets reject. examine and deliberately schedule replacement between check/write; record the accepted race, not a fictitious atomic pass. close only the confirmed terminal; other terminals survive. include last-pane closure of a parent worktree with a live linked workspace under both confirmation settings: refuse a cascading operation. |
| launch/readiness | all four profiles preserve configured executable/environment/cwd/flags; verify internally and report equality only. shell and zero-profile host work; startup dialog/timeout differs from failed creation. restart the isolated server with resume disabled and prove no unintended provider relaunch. |
| observation/read | codex and claude working, input wait, idle, unfamiliar screen and unavailable observation have honest source/readiness; reads declare bounded visible/history coverage. no inferred task success or hidden native-capability loss. |
| uncertainty/lifetime owners | suppress a reply after one real synthetic mutation: caller reports unknown within budget, dispatch count remains one after reconnect. kill the probe client/bridge and gateway separately; workers survive and rediscover. cold server restart is tested separately. |
| phone/desktop | shell, codex and claude through the physical phone renderer: typing, ime/dictation, multiline paste, keys, touch scroll/history, copy, keyboard resize, rotation, control conflict/takeover, detach, background and abrupt loss. desktop shares the worker and regains sizing. qualify host stream/lifetime behavior on linux and darwin, without multiplying every input case across every profile. |

no live operations are authorized by this document. obtain explicit current-turn
approval for live/tmux/device work under [AGENTS.md](../AGENTS.md); use isolated
test-owned herdr instances and, if needed, exact test-created tmux resources on
an isolated `-L` socket. no default server or installed phone data mutation.
missing approved host/device boundaries remain `NOT_RUN`.

record one short result per boundary: exact source/build, platform, scenario,
`PASS | FAIL | NOT_RUN`, observed outcome, limitation. compare synthetic payloads
in memory; retain no prompts, terminal bytes, screenshots, hashes of content,
credentials or account data. source review and static checks are separate evidence.
follow [testing policy](rules/testing.md); do not recreate retired harnesses.

## exit decision

`proceed` requires all necessary real-boundary proofs, the exact wire/operation
table without placeholders, explicit acceptance of capability losses, and
resolution of the [targeting](issues/herdr-agent-targeting.md),
[profile](issues/herdr-profile-restore.md) and
[terminal interaction](issues/herdr-terminal-acceptance.md), plus
[closure](issues/herdr-terminal-closure.md) issues. accepted limitations belong
in the contract; delete resolved issue records. negative findings mean
`reconsider`, with the responsible blocker and smallest next decision stated.

hard cutover occurs in pr 2, not here: no legacy runtime/protocol paths, fallback
detectors, compatibility negotiation or dual backend. no persistent registry,
task scheduler, history copy, new provider protocol, generalized test machinery,
live pty migration or unrelated cleanup. deleting temporary tests leaves no
retained behavioral regression suite; that is the explicit verification cost.

## investigation result at the pinned baseline

**reconsider.** the required terminal-only closure contract is impossible through
v0.9.1's public commands. its `pane.close` can remove other terminals when the
target is the final pane of a parent worktree workspace and `confirm_close=false`.
`workspace.close(close_group=false)` prevents that group cascade but closes every
terminal in its workspace. neither command accepts an expected terminal lifetime
or a predicate that the layout still contains only the confirmed terminal. a
gateway check before either command cannot constrain a later desktop layout
change. this is an upstream capability gap, not a request for a skid-side
preflight. [the closure issue](issues/herdr-terminal-closure.md) remains open.

the public `herdr terminal session control` stream has a separate phone input
gap: it accepts text/bytes, resize, scroll and release, but no mode-aware tap or
mouse command. its ansi frames are rendered cells, without the provider's
mouse-mode negotiation; forwarding xterm's tap bytes would guess the wrong
mode. the native herdr client uses an internal mouse message, whose private
protocol is outside this migration. [the terminal issue](issues/herdr-terminal-acceptance.md)
remains open. these two failures prevent a `proceed` decision even if the other
boundaries pass.

the inspected source is [herdr v0.9.1 at
`065ef9d6a531c49fb8bee7e818ef837065b21ee9`](https://github.com/herdrdev/herdr/tree/065ef9d6a531c49fb8bee7e818ef837065b21ee9).
the isolated macos arm64 [v0.9.1 release
binary](https://github.com/herdrdev/herdr/releases/tag/v0.9.1) reports v0.9.1
and socket protocol 22;
its downloaded sha-256 was
`5fc7a7e7adfaca56fa80aa89dcb025693357268dab8285b9ce2d08a2313c89de`.
the test used a temporary config/home and isolated sockets with automatic
agent resume disabled. this identifies the tested artifact, not an installed
fleet version. the exact release source for the closure path is
[`pane.close`](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/app/api/panes.rs#L1849),
[`close_selected_workspace`](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/app/actions.rs#L735),
and [`workspace.close`](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/app/api/workspaces.rs#L311).
the public stream is defined in
[`terminal_sessions.rs`](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/client/terminal_sessions.rs#L126)
and its [accepted input enum](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/client/terminal_sessions.rs#L160).

| boundary | platform and scenario | result | observation and limit |
| --- | --- | --- | --- |
| exact closure, original binary | darwin arm64; isolated parent worktree, linked terminal present, unrelated survivor; `confirm_close=false`; close parent's sole pane | `FAIL` | `pane.close` returned success and removed parent and linked workspaces. the linked terminal's original id was absent afterward; survivor remained. the unchanged binary failed the temporary assertion with exit 1. |
| synthetic wrapper launch | darwin arm64; isolated workspace with explicit cwd/env, then shell-quoted `exec` through `pane.send_input` | `PASS` | one root pane remained; the wrapper compared exact argv, environment marker and physical cwd internally. this does not prove the four real provider profiles, readiness or timeout behavior. |
| terminal metadata across socket reconnect and cold restart | darwin arm64; no-expiry synthetic tokens and pane label; stop/restart only the isolated server with resume disabled | `PASS` | tokens survived fresh socket connections. cold restore retained the label, cleared the tokens and allocated a new terminal id. this does not prove gateway-process restart or provider non-resume. |
| terminal identity through rename and move | darwin arm64; rename pane, move it to a new workspace, reuse its old display name | `PASS` | the terminal id stayed constant; public pane id changed and its old alias still resolved the original. this does not establish an exact worker lifetime or atomic mutation. |
| public terminal stream and controller loss | darwin arm64; isolated shell; control command and competing controller | `PASS` | first frame was full with sequence 1 and matching geometry; conflict produced `terminal.closed` with a free-text reason; losing the cli controller left the terminal present. none of this establishes gateway loss or phone rendering. |
| unsupported mouse command | darwin arm64; synthetic `terminal.mouse` sent to the public control command | `PASS` | the command was rejected without mouse delivery. this is a negative control confirming missing public input, not phone acceptance. |
| phone frame, key and touch-scroll transport | samsung sm-s906w, android 16/api 36; disposable `dev.niels.skidbladnir.herdrproof` 0.1.0/code 1 against isolated darwin relay and shell | `PASS` | a full frame was applied before input admission; one adb-injected keyevent on the physical phone increased the same shell's visible draft by one, compared only in memory. a phone swipe produced six public `terminal.scroll` commands. a second phone connection without takeover received `terminal.closed` while the cli controller remained. actual scrollback movement, conflict ui/takeover, ime, paste, selection/copy and mouse-aware taps remain unproved. the proof build came from base `c6661a587ae766b9c2f21d665546b21e3a19da0b` plus temp-only edits and bypassed production pairing. |
| closure guards, original binary | darwin arm64; isolated parent and linked terminal; `confirm_close=true` for `pane.close`, then `workspace.close(close_group=false)` | `PASS` | the commands returned `confirmation_required` and `workspace_group_close_required` respectively; both exact terminal ids survived. this does not repair the confirmation-disabled failure or guard a workspace containing other terminals. |
| phone mouse-aware tap through the public stream | pinned source; physical phone | `NOT_RUN` | source review establishes no public mode-aware input command; the physical mouse-aware interaction cannot be exercised through this bridge. |
| linux host, codex/claude profiles, headless orchestration, foreground replacement, write uncertainty, gateway process loss and provider non-resume | isolated live boundaries | `NOT_RUN` | source review and the darwin probes do not establish these behaviors. |

source review also found useful candidate primitives, not a closed contract.
`workspace.create` supplies the root pane, and a separately submitted `exec`
can launch an exact configured wrapper there without leaving an extra shell.
the create success precedes process readiness; a lost reply to the later input
leaves an uncertain, visible terminal and must never trigger a second create.
`agent.start` uses herdr's canonical executable rather than skid's configured
wrapper, so it is not an interchangeable launch route.
`PaneInfo.terminal_id` stays with a terminal on move while its pane id changes.
the pinned binary also preserved it on rename and name reuse; the old pane-id
alias resolved the original after movement. `pane.process_info` exposes
foreground pids but no process start identity, so agent references would need
skid's existing host process-start observer for pre-dispatch revalidation.
nonexpiring `pane.report_metadata` tokens remain while the server runs, so a
gateway-owned random token could bind an opaque reference across gateway
restart. cold restore allocates fresh terminals and empty tokens. assignment
and revalidation must be qualified live; a token does not make a later write
atomic. upstream metadata tokens are not snapshotted, so character, objective
and launch-profile retention across cold restart is a proposed loss needing
explicit acceptance. native claude history/status/stop loss likewise has no
recorded acceptance. [targeting](issues/herdr-agent-targeting.md) and
[profile/restore](issues/herdr-profile-restore.md) remain open.

the [independent mobile relay](https://github.com/benkraus/herdr-plugin-mobile-relay)
confirms that a separately supervised, bounded host bridge is a plausible
composition pattern. its browser, worktree, file, git, push and audit features
are outside skid's pr 1 contract and do not change the pinned herdr public
command limitations identified above.

the wire/operation table and affected product-contract amendments required for
`proceed` cannot be closed at this pin: `kill` and the closure step of `stop`
have no acceptable upstream effect, and phone tap/mouse has no acceptable
public input primitive. the smallest next decision is whether to scope and
qualify an upstream revision with an atomic expected-terminal close and a
public mode-aware tap/mouse path with typed close reasons. the candidate must
then repeat the important pr 1 proofs, including both confirmation settings,
before a new `proceed` decision. pr 2 remains unauthorized by this evidence.

the disposable phone probe's scope review added only
`LockedTerminalWebView.kt` inside its isolated source copy. that page-port
decoder is the narrow owner for a typed scroll event from the existing touch
route to herdr's public `terminal.scroll`; it did not alter repository source.
the proof still cannot provide a mouse command absent from upstream.
the proof package was uninstalled and absent after testing; the adb reverse
mapping was removed, and the installed skid package remained present. these
are cleanup observations, not phone behavior acceptance.
