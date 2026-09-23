# herdr pr 1: feasibility and contract closure

2026-09-22 · qualification reopened after an approved scope amendment.
[the migration plan](herdr-migration.md) owns delivery order. this pr owns the
evidence and precise contract needed to authorize pr 2. current production
contracts remain unchanged. the historical terminal-only closure proof failed;
its requirement is superseded below, not its result. unperformed runtime/device
acceptance remains `NOT_RUN`; this amendment is not a `proceed` decision.

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
| runtime restart — metadata loss accepted by user 2026-09-22 | one shared herdr server per host; disable automatic provider resume. client/gateway loss preserves workers; cold runtime restart ends their old lifetimes. restored shells are new unnamed terminal lifetimes without the prior launch profile, objective or dwarf identity. upstream's surviving pane label is a secondary hint, never the old name/ref/worker; actions require a fresh ref. restoration is not live process migration. |
| manual native labels — accepted by user 2026-09-22 | a newly discovered manual pane's native label is also hint-only until named through skid. it remains discoverable and actionable by fresh ref; name selection deliberately does not use an unmarked native label. |
| grouping — proposed replacement | use actual host-bound herdr workspaces. equal labels group visually across hosts, never identify mutations. moving a terminal can change desktop layout; retire cosmetic space assignment and clearing, not conceal them behind the old command. |
| provider capabilities — accepted by user 2026-09-22 | herdr observations and bounded terminal reads for both providers; no second skid detector. native claude history, status and halt confirmation are lost. jarvis must inspect bounded terminal output and treat stop as unconfirmed without an observed exit. |
| phone control | opening acquires direct control without takeover; conflict offers explicit takeover or back. no observe-first mode or fallback. phone sets geometry while connected; desktop/api input can still occur. |
| stop after uncertain interrupt — proposed | do not issue close after the interrupt reply is lost; report `terminal:not_attempted` and let the caller inspect before a deliberate kill. current codex stop may attempt close despite interrupt error, so this is a conservative behavior change, not transparent parity. |
| desktop entry — proposed replacement | no-argument `skid` opens the configured local herdr desktop; `skid enter` retains exact fleet-target attachment through the gateway stream. no inference from gateway origins to ssh hosts; native herdr remote navigation remains upstream-owned. qualify both paths. |

the existing [targeting issue](issues/herdr-agent-targeting.md) tracks whether the
accepted revalidation boundary is implementable. current skid also checks then
writes; its mutex does not prevent independent process exit. terminal destruction
must start from the revalidated original terminal, never a name-reused replacement;
its permitted effect scope follows the amendment below.

## accepted scope amendment

2026-09-22, approved by the user after reviewing the pinned-baseline findings:

- phone attachment displays and controls one exact provider terminal, not herdr's
  workspace ui. retain keys, ime/dictation, paste, swipe scrolling, phone-local
  selection/copy, sizing and detach. taps focus input/show the keyboard;
  clicks and drags into the remote application are out of scope. no public mouse
  command or private mouse protocol is required.
- route swipes through public `terminal.scroll`. upstream already owns
  [mode-aware wheel routing](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/server/pane_input.rs#L128),
  including application mouse-wheel reports. excluding remote clicks does not
  exclude scrolling or prove that codex/claude history scrolling works.
- `kill` and the closure step of `stop` adopt herdr's native pane-close semantics.
  closing the last pane may close its workspace and linked git-worktree group,
  including their independent terminals. the group relates checkouts, not shared
  agent processes. closing it does not delete checkout directories or branches.
  exact-terminal-only closure and an atomic expected-terminal close are no longer
  migration requirements. original-reference revalidation/no retargeting remain.
- every phone close/stop confirmation names the selected terminal and machine
  and warns that herdr may also close the linked workspace group and its running
  terminals. cli help and jarvis's tool contract disclose the same effect before
  use; their wire contract must not imply single-terminal scope. this warning
  describes execution-time group scope, not a frozen preview of group members.
  results report only observed effects; `stop` does not claim native halt of
  every provider affected by a cascade. lost replies remain unknown, never retried.
- respect upstream confirmation policy. a returned `confirmation_required` is a
  refusal, not success or permission to switch to a broader command, override
  configuration or retry. phone confirmation does not bypass upstream refusal.
  qualify both confirmation settings and the exact public operation mapping.
  detach still releases only the attachment and leaves workers running.

costs: no remote terminal clicks/drags; closing one selected dwarf can end other
agents in its herdr group; some closes can be refused by upstream policy. these
are accepted product semantics, not unresolved upstream defects. qualification
resumes at v0.9.1 without requiring those two upstream additions. the wire table,
remaining live proofs and other capability-loss decisions still gate pr 2.

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
  [the operation table](#retained-gateway-operations). `written` means
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
tab/workspace and native linked-group closure under the
[accepted scope](#accepted-scope-amendment). see
[closure disclosure and qualification](#2026-09-22-reopened-qualification-at-the-accepted-scope).

<a id="candidate-pr-2-contract-proposed-not-implemented"></a>
## candidate pr 2 contract

this section is the sole candidate wire contract. the preceding requirements
state its acceptance tests. this branch implements the candidate source; live
acceptance and deployment remain open. facts below refer only to pinned v0.9.1;
they are not claims about the installed product. [the evidence](#investigation-result-at-the-pinned-baseline)
and subsequent results determine whether this candidate can be activated.

### identity, inventory and metadata

all `/v1` calls retain the current bearer and pinned `Skidbladnir-Machine`
header, strict json, peer-oriented partial results, observation timestamps,
256-kib http request/control-success caps, 64-kib error-response cap, 1-mib
inventory cap and 15-second client deadline. a 32768-byte read can exceed
64 kib on the wire when json escapes control bytes; 256 kib admits its worst
case plus the response envelope. host operations spend at most ten seconds. no upstream error
text, terminal content, cwd, argv, provider home or token value enters logs.
the inventory cap applies to each gateway response. the phone admits that cap
only on `/v1/terminals`; its other successful http response cap is 256 kib
and its error-response cap remains 64 kib. the fleet
cli emits every accepted peer projection, bounded by the existing 64-kib peer
configuration file and each peer's inventory cap, without a separate 1-mib
aggregate limit.

three opaque refs are unpadded canonical base64url of strict json, at most 4096
characters, with no null or extra fields:

```json
{"kind":"terminal","machine":"mh-…","terminalId":"term_…","identityToken":"…"}
{"kind":"agent","machine":"mh-…","terminalId":"term_…","identityToken":"…","agent":{"pid":1,"startIdentity":"…","commandFingerprint":"…","provider":"Codex"}}
{"kind":"workspace","machine":"mh-…","workspaceId":"w1","identityToken":"…"}
```

the gateway generates each 128-bit random base64url `identityToken` once per
observed terminal/workspace lifetime using upstream `pane.report_metadata` or
`workspace.report_metadata` (`source:"user:skidbladnir"`, token key
`skid_lifetime`, no `ttl_ms`). it rereads the token after assignment. the manager's
existing mutation lock serializes reread-if-absent/write/readback across discovery
and creation. `pane.list` supplies `terminal_id`, current `pane_id`,
`workspace_id`, label, cwd and metadata;
`workspace.list` supplies actual host-bound workspace ids and labels. the gateway
finds exactly one current pane by the original terminal id and checks its token
before each operation. if a manual pane has exhausted upstream's 32 metadata-token
keys, inventory reports it as unaddressable and the host result as partial; it
never overwrites conflicting existing keys or invents a ref. upstream's token
map is shared: `source` is not a namespace or ownership lock. reserve `skid_*`
by convention, without claiming protection against other same-user writers.
name and pane id never resolve an old ref. the agent ref additionally matches
the current foreground pid and start identity using
`pane.process_info` plus the existing linux/darwin `internal/process` observer.
`commandFingerprint` is lowercase hex hmac-sha256, keyed by the existing decoded
32-byte host bearer. input is the fixed ascii domain `skid-command-v1` followed
by a nul byte, then the executable and argv: unsigned 32-bit big-endian executable
byte length, executable bytes, unsigned 32-bit big-endian argument count, then
each argument's unsigned 32-bit big-endian byte length and bytes in order.
auth computes it without exposing its key; raw account arguments never enter a
ref or log. this is private command equality, not a signed reference or authority;
bearer replacement invalidates agent refs. [pr 2](herdr-pr2.md#contracts-and-composition)
requires that narrow distinction in identity rules at cutover.
provider classification must still agree. this defines an observed
kernel-process/command lifetime; a same-pid `exec` with identical executable
and argv is indistinguishable, while mutable process titles can conservatively
stale a ref. neither case warrants substituting a fresh worker.
an unobservable worker is unknown/stale, never a guessed profile. a separate
process exit or layout change between this check and herdr's command remains
possible. gateway restart rereads tokens; cold herdr restart clears pane and
workspace tokens and allocates new terminal ids. a restored workspace may keep
its upstream id but is a new skid reference. no registry, expiry, restamping of old refs, atomic
compare-and-write or implicit replacement target exists.

`GET /v1/terminals` returns exactly
`{machine:{handle,platform},observedAt,partial,unaddressableTerminals,
unaddressableWorkspaces,profiles:[{key,label,provider}],
workspaces:[{ref,label}],terminals:[terminal]}`. the counts include metadata
claim failures, without exposing their contents. a pane whose workspace token
cannot be claimed also counts as an unaddressable terminal: it has no truthful
`workspaceRef`. fleet `skid list --json` wraps
this as `{ok:true,result:{partial,peers:[peer]}}`: a successful peer is
`{label,machine,ok:true,observedAt,partial,unaddressableTerminals,
unaddressableWorkspaces,profiles,workspaces,terminals}`; an unavailable peer is
`{label,machine,ok:false,error:{code,message}}`. global `partial` is true if
any peer fails or reports partial. each `terminal` is
`{ref,name?,nativeLabel?,character:{key,displayName},workspaceRef,cwd?,
launchProfile?,objective?,agent?}`. `nativeLabel` is the observed upstream pane
label, when present; it is a secondary hint and never a name selector. without
an eligible `name`, clients display `unnamed terminal` plus a short id and,
if present, `herdr label: <nativeLabel>`; actions require a fresh ref.
the gateway omits an observed native label containing controls or bidi direction
formatting before projection; it remains a hint, never identity. `character` is a deterministic catalogue selection from the machine and terminal
lifetime, so manual panes need no stored persona. a cold-restored new lifetime
gets a new assignment, as accepted. `agent` is
`{ref,provider,provenRuntimeProfile?,providerSession?,status,readiness,methods}`;
`providerSession` is `{id?,name?}` only when the process-bound claude
registration proves it; `methods` is
`{read,send,interrupt}` with each value `terminal | unavailable`. `status` is
`{state,source,reason?}`; `state` is
`working | blocked | idle | unknown`, `source` is `herdr | unavailable`, and
`reason` is `default_idle | unrecognized | observation_failed`
when known. `readiness` is `ready | blocked | unconfirmed` and is independent of
state. a terminal without a proven provider foreground has no `agent`.

map `agent.get.agent_status` `working/blocked/idle` directly, `done` to `idle`
(unseen idle, never successful task completion), and every other or failed
observation to `unknown`. sample `agent.explain` separately: a matched visible
idle rule permits `ready` only if the foreground worker still matches and no
newer contradictory sample is known; upstream `visible_blocker` plus a
consistent blocked sample gives `blocked`; a
default-known-agent idle fallback, skipped screen detection, or disagreement
between samples gives `unconfirmed`. `reason:default_idle` applies only when
the explain fallback actually names that condition. unknown rule ids omit
reason. wrapper launches cannot use upstream `interactive_ready`, which is
specific to managed `agent.start`. no local screen detector
upgrades uncertainty into readiness. `agent.explain` may contain screen previews;
the gateway consumes only rule, visible-idle, fallback and skip fields. it never
forwards or logs previews.

accepted by the user 2026-09-22: ordinary send requires recognized idle
(`readiness:ready`); unlike current skid, it does not admit working agents.
otherwise reject before dispatch with `ReadinessUnconfirmed`, without waiting
or queuing. `skid send --terminal` (`mode:terminal`) bypasses readiness only;
original-worker revalidation and all input/dispatch rules still apply. the cost
is deliberate override or later resubmission when a worker is busy; there is no
automatic retry. this decision does not prove provider readiness detection.

upstream pane label remains the sole stored name text. a nonexpiring
`pane.report_metadata` token (`source:"user:skidbladnir"`, key `skid_named`,
value `1`) says skid deliberately named this terminal. `name` projects the
current pane label only when the marker is present and that label satisfies
skid's name grammar; an invalid native rename leaves `name` absent and the
observed label as a hint. valid native renames update the product name without
a duplicate name store. `skid start`, `skid shell` and `skid rename` set/read
back that marker. cold restore clears it but may
retain the pane label, so the new shell is unnamed with `nativeLabel` as a hint.
a newly discovered manual pane also has no marker; its native label is hint-only
until named through skid. the user accepted this manual-name-selection cost on
2026-09-22. both kinds remain selectable by fresh exact ref. upstream
workspace identity/label owns grouping. `skid_launch_profile` owns only the
selected launch row; it does not prove the current worker's account.
`provenRuntimeProfile` is omitted unless
a process-bound registration is independently proven; `agent_session` is not
account identity. the sole allowed `SessionStart` registration must be retargeted
to inherited `HERDR_PANE_ID` and `HERDR_SOCKET_PATH`, with public pane identity
and existing host process observation checking the provider ancestor and current
foreground pid/start. v0.9.1's `pane.process_info.tty` is always absent, so it
cannot replace the old tty proof. without a live registration proof the field
stays omitted;
dropping the capability altogether has not been accepted. optional objective
metadata uses canonical padding-free
base64url of utf-8, split into consecutive 80-character `skid_objective_00`
through `_15` values. write `skid_objective_count` first; decode only the exact
declared count and a valid original objective. missing or extra chunks make it
absent. one report can carry at most 16 keys, so a maximal 240-scalar objective
needs a second report and complete readback before publication. partial metadata is omitted, never
presented as complete. metadata tokens, objective and launch profile do not
survive cold herdr restart; a restored shell is a new terminal lifetime.

product terminal names preserve today's 1–64 ascii letter/digit/underscore/hyphen grammar,
starting with a letter or digit. generated names follow the same grammar;
duplicate upstream labels are permitted. skid may precheck a generated name,
but cannot promise uniqueness against native desktop edits. exact refs remain
decisive; name selection fails as ambiguous if multiple terminals match.
`nativeLabel` never participates in name resolution, even if it matches one
terminal uniquely.
workspace labels retain today's 1–64 nfc-scalar, 256-byte, ordinary-space and
display-control rules. equal labels on different hosts group only visually.
within a host, `workspaceRef` identifies one real workspace; name-based
destination selection rejects duplicate labels. no `unassigned` resource or cosmetic membership
exists. `workspace.create`/`tab.create` already supply one root pane; gateway
uses that pane for a new dwarf, never creates a second shell. moving a pane via
`pane.move` may change layout and remove an emptied source tab/workspace.

### retained gateway operations

`R` below means an original opaque ref in the path; the server decodes and
revalidates it. every mutating result includes `dispatch: not_sent | sent |
unknown`. `not_sent` requires proof that no upstream mutation was issued;
`sent` means an upstream response was received, not that the worker acted;
`unknown` covers a deadline or lost reply after possible dispatch and never
authorizes automatic retry. typed errors retain `{code,message,dispatch?}`:
`Unauthenticated`, `MachineIdentityMismatch`, `InvalidRequest`,
`RequestTooLarge`, `TerminalNotFound`, `TerminalStale`, `AgentStale`,
`WorkspaceStale`, `MetadataUnavailable`, `ProfileUnknown`, `WorkingDirectoryInvalid`,
`NameInvalid`, `NameAmbiguous`, `ObjectiveInvalid`, `ReadinessUnconfirmed`,
`MethodUnavailable`, `ClosureConfirmationRequired`, `HerdrUnavailable`,
`UpstreamRejected`, `OutcomeUnknown`. upstream prose is sanitized. a positive
upstream refusal is `dispatch:sent` with no claimed effect; transport loss after
dispatch is `OutcomeUnknown` or a partial result. downstream clients do not
infer safety to retry from http status or process exit.

create and shell are ordered multi-step mutations. a non-201 response is the
same error shape with optional `partial:{stage,terminal?}`. `stage` is
`resource_created | identified | launch_submitted`; `terminal` appears only
after token readback produced a valid ref. an unknown create reply has
`OutcomeUnknown,dispatch:unknown` and no partial ref. a known create followed
by failed token claim has `stage:resource_created` and no ref; inventory counts
the unaddressable pane. failed launch after identification returns
`stage:identified,terminal` even if its dispatch is unknown. no stage claims
readiness or permits a second launch. cli wraps host errors as
`{ok:false,error:{code,message,dispatch},partial?}` and exits 1. pre-dispatch
validation is `dispatch:not_sent`; upstream refusal is `dispatch:sent` with
no claimed effect; deadline or lost reply after dispatch is `unknown`.
for kill, an upstream `confirmation_required` is http 409
`{code:"ClosureConfirmationRequired",message,dispatch:"sent",
partial:{terminal:"refused"}}`; stop uses the same error with
`partial:{agent:"interrupt_sent"|"exited"|"unconfirmed",
terminal:"refused"}`. a lost close reply is http 504 `OutcomeUnknown`
with the known agent step, if any, and `terminal:"unconfirmed"` in partial.
unknown interrupt reply returns `terminal:"not_attempted"`. these are
errors, never successful close results; cli wraps them as `ok:false` and exits
1. the phone and jarvis show the partial outcome, not a success toast.

for every selected-host cli command below except `list`, the successful
`--json` result is the named host success shape plus `label` (fleet peer label)
and `machine` (pinned handle), with no renamed or omitted host fields. thus
`info` returns `{label,machine,observedAt,terminal}`, start/shell add
`launch,dispatch`, read adds `text,source,scope,truncated`, and mutation results
add their shown outcome and `dispatch`. `list` alone returns the fleet shape
above. human output renders the same facts; read text goes to stdout and its
coverage/truncation to stderr. a selected-host failure retains the same
`{ok:false,error,partial?}` cli envelope with the peer label/machine only when
known. every noninteractive command has `--json`, exact `--ref` or
case-sensitive name with optional `--machine` selectors, and `--stdin` only for
literal send text. unqualified name selection refuses a partial fleet result.

| route / cli operation | exact request and success | target, effect boundary, failure and upstream primitive |
| --- | --- | --- |
| `GET /v1/terminals`; `skid list` | empty request; inventory above; `partial` fleet projection includes unavailable peers or unaddressable resources | `workspace.list`, `pane.list`, `agent.list`, `pane.process_info`, plus one-time metadata token claim/readback on newly seen manual resources. that claim mutates upstream metadata and may fail or have an unknown reply; report the resource unaddressable, never replace an unknown token. any failed host is explicit, never an empty inventory |
| `GET /v1/terminals/R`; `skid info` | no body; `{observedAt,terminal}` with a fresh current agent ref | terminal lifetime only; missing/stale ref fails rather than refreshing a mutation target |
| `POST /v1/terminals`; `skid start` | `{kind:"agent",profile,cwd,name,objective?,destination?}` or `{kind:"terminal",cwd,name,destination?}`; `destination` is `{kind:"existing",workspaceRef}` or `{kind:"new",label?}`; omitted means a new workspace labelled with the terminal name; `201 {observedAt,terminal,launch:"submitted"|"not_requested",dispatch:"sent"}` | validate cwd/profile/destination before create. `workspace.create` or `tab.create` with `focus:false`, chosen `cwd` and profile `env` allocates one root pane; `pane.rename` sets its validated label before `pane.report_metadata` sets/readbacks `skid_named` and other metadata, then one shell-quoted `exec` of the configured absolute command plus `agentruntime.LaunchArguments` goes through `pane.send_input`: claude-work prepends `--name` and the terminal name; codex uses configured arguments. objective is never a prompt. creation success precedes readiness and does not assert startup. if create succeeds but a later step fails, return the new terminal and partial stage when known; unknown create reply leaves the effect unknown, never a second launch or compensating close |
| `POST /v1/terminals/R/shell`; `skid shell` | `{}`; same `201` terminal envelope with `launch:"not_requested",dispatch:"sent"` | revalidate original source terminal, sample its `foreground_cwd` when available or `cwd` and exact workspace, then `tab.create` with `focus:false` for one independent shell there. `pane.rename` sets its generated label before `pane.report_metadata` sets/readbacks `skid_named`. unreadable cwd fails before create; no split, source replacement or profile inheritance |
| `PATCH /v1/terminals/R`; `skid rename` | `{name}`; `{observedAt,terminal,dispatch:"sent"}` | original terminal; one `pane.rename` on the current pane id after revalidation. if `skid_named` is absent, set/read back that marker; a known marker refusal returns http 502 `{code:"MetadataUnavailable",message,dispatch:"sent",partial:{terminal}}`, whose current projection has the new native hint but no product name. a lost reply remains `OutcomeUnknown`, never retried. a later competing native rename can win; client confirms by inventory |
| `PUT /v1/terminals/R/workspace`; `skid move` | `{destination}` in the create union, with `label` required for `kind:"new"`; `{observedAt,terminal,dispatch:"sent"}` | original terminal and exact destination workspace ref, or explicitly labelled new workspace; `pane.move` to `new_tab` or `new_workspace`. no label-based upstream mutation, no exact-layout lock; report current workspace after response |
| `POST /v1/agents/R/read`; `skid read` | `{coverage?:"recent"|"visible",maxBytes?}`; coverage defaults `recent`, bytes default 16384, max 32768; `{text,source:"terminal",scope:"visible"|"terminal_history",truncated}` | original agent; `recent` maps to upstream `pane.read` `recent_unwrapped`, max 1000 lines, and scope `terminal_history`; `visible` maps to `pane.read` `visible` and scope `visible`. keep the newest complete-utf-8-codepoint suffix under the byte limit for either scope and mark truncation; upstream line truncation also marks it. `truncated:false` never means complete provider history; failed read returns typed failure |
| `POST /v1/agents/R/send`; `skid send` | `{text,mode:"auto"|"terminal"}` (mode defaults auto); `{method:"terminal",outcome:"written"|"unknown",dispatch}` | original agent; 1–32768 utf-8 bytes. auto requires independently recognized readiness; terminal mode deliberately addresses a dialog/unclassified screen. one `pane.send_input{text,keys:["enter"]}` queues paste plus submit. `written` means accepted into the upstream pty queue, not provider processing or task success |
| `POST /v1/agents/R/keys`; `skid keys` | `{keys:[logicalKey]}` with 1–16 keys from `enter,escape,ctrl-c,up,down,left,right,tab,backspace`; same write result | original agent; validate the entire list before one mode-aware `pane.send_input{keys}`. unsupported keys, including page keys, fail before dispatch |
| `POST /v1/agents/R/interrupt`; `skid interrupt` | `{}`; same write result | original agent; one `pane.send_input` key: codex `escape`, claude `ctrl-c`; `written` does not prove cancellation |
| `POST /v1/agents/R/stop`; `skid stop` | `{}`; success `{agent:"interrupt_sent"|"exited",terminal:"closed",dispatch:"sent"}`; other outcomes use the partial error above | original agent; send one interrupt, then revalidate original terminal. `exited` requires observing the original pid/start absent through `internal/process`, not merely no longer foreground; a live replacement rejects closure. unknown interrupt reply stops before closure. one `pane.close` follows; `confirmation_required` is a partial refusal, never a broader command. closure does not confirm descendant or linked-provider halt |
| `DELETE /v1/terminals/R`; `skid kill` | `{}`; success `{terminal:"closed",dispatch:"sent"}`; other outcomes use the partial error above | original terminal; one `pane.close` after revalidation. final-pane closure may remove its workspace and linked worktree group; no exact affected-set or halt claim. `confirmation_required` remains a refusal; no `workspace.close` fallback |
| `GET /v1/terminals/R/stream`; `skid enter` / phone attach | authenticated websocket; header `Skidbladnir-Terminal-Takeover: false|true`; `true` only after explicit user action | original terminal, control via `herdr terminal session control <terminal_id>`, never workspace ui. controller acquisition and first full frame gate input. disconnect/release owns only child/socket, not worker/server; reattach starts from fresh discovery |

the following host operations do not dispatch to herdr and retain their current
strict codecs, authentication, limits, deadline and typed errors. their schemas
are imported unchanged into this candidate; pr 2 must leave those DTOs and
clients untouched except route-composition changes:

| route | request and success | boundary |
| --- | --- | --- |
| `POST /v1/pairing-invites` | authenticated empty body; `201 {pairingInviteToken,expiresAt,machine}` | current in-memory five-minute slot, replaced on new invite; [pairing wire](public-fleet-distribution.md#4-capability-contract) |
| `POST /v1/pairings` | invite authorization and expected machine, empty body; `200 {machine,bearer}` | atomic one-use slot redemption; invalid/expired token is `PairingInviteRejected`; [pairing wire](public-fleet-distribution.md#4-capability-contract) |
| `POST /v1/directory-listings` | `{directory}`; `200 {machine,directory,parentDirectory?,children:[{directory,kind}],omitted}` | immediate canonical-home directory listing, no herdr cwd inference; [directory wire](working-directory-chooser.md#http-api) |
| `GET /v1/pressure` | empty request; `200 {unsupported,current,history}` with the unchanged closed metric/signal variants | host pressure sampler, independent of terminal runtime; [pressure contract](architecture.md#4-product-behavior) |

these routes retain `Unauthenticated`, `MachineIdentityMismatch`,
`InvalidRequest`, `RequestTooLarge` and their route-specific errors; each
failure is before any herdr mutation. pairing slot effects are governed by the
existing pairing contract, not by `dispatch` for terminal writes.

`skid start` and `skid move` accept mutually exclusive `--workspace-ref REF`
or `--new-workspace LABEL`; move requires one and its new-workspace label is
required. start defaults to a new workspace labelled with its terminal name.
destinations are host-bound; name-based workspace lookup is not a cli mutation
interface. retain the existing exact target,
profile/terminal, cwd, name, stdin and json options where applicable.

`skid` without operands and with a tty execs the configured local pinned
`herdr client` against its existing server socket; absent server fails without
autostart. [pr 2's local configuration](herdr-pr2.md#local-configuration) owns the
single host-config lookup and runtime address. `skid enter` keeps exact cross-host
attachment through the selected gateway stream, irrespective of desktop focus.
no-tty bare `skid` exits with
usage status 2. explicit name selection remains exact, case-sensitive and
fleet-complete; machine qualification or a ref resolves ambiguity. `--json`
keeps one `{ok,result|error}` envelope and exits 0 for confirmed complete
results, 1 for partial/unknown/refusal, 2 for usage. `skid space`, `--unassigned`,
`--space` and `--clear` are retired; `skid move` names the real destination and
reports layout effects. `skid read --terminal` becomes `skid read --coverage
visible`; omitted coverage reads recent terminal history. `skid send --terminal`
remains the deliberate input override. jarvis keeps the original returned ref
through its metadata read and revises strict peer, terminal, workspace, status,
read, write, error and partial-closure models. pr 3 changes
`src/jarvis/agent_tools.py`, `agent_control.py`, `write_policy.py`,
`write_dispatch.py`, `definitions.py` and `session-compatibility.json`
together: list/info projections and start destination,
read coverage, `dispatch:sent`, refused/partial outcome staging, effect-target
projection, fallback start receipts, owner/tool wording and existing binding
revision strings. a jarvis start with no selected workspace
uses the new named-workspace default above. tool descriptions disclose possible
linked-group closure before invoking stop/kill. no jarvis adapter gains herdr
access or retries an unknown write.

the phone maps inventory `terminal.ref` and `workspaceRef` to exact targets;
forge chooses one existing workspace or explicitly creates one. move and close
copy names the selected terminal and machine, with `linked workspaces and their
running terminals may also close`. the close/stop button cannot override an
upstream `confirmation_required`; detach says `leave terminal running`. terminal
attachment reports `control_unavailable` before the first complete frame,
`stream_lost` afterward, `detached` after local release, and `protocol_error`
for invalid/oversized frames. its conflict copy says `another connection may be
active; take over or go back` because upstream v0.9.1 exposes only a free-text
close reason. takeover always revalidates the same original ref; it is never
an automatic reconnect or input retry. desktop/api input can still reach the
same terminal while the phone owns the direct controller.

the proposed websocket uses strict json text messages, one envelope per message.
server `{"kind":"Frame","seq":string,"columns":uint16,"rows":uint16,
"full":bool,"ansiBase64":string}` projects upstream `terminal.frame`
`seq,width,height,full,encoding:"ansi",bytes`; `{"kind":"End","code":
"control_unavailable"|"stream_lost"|"detached"|"protocol_error"}` is final.
`terminal.closed.reason` remains sanitized diagnostic prose, never a typed
conflict code. `seq` is positive canonical unsigned decimal (a nonzero leading
digit followed by digits), within uint64. native clients validate it numerically;
javascript receives/acknowledges the unchanged string, never a rounded number.
before a first full frame, any upstream refusal is
`control_unavailable`; after one, unexpected closure is `stream_lost`.
the first full frame resets xterm and sets geometry; the phone page acknowledges
application to its native connection owner before that owner admits input. this
acknowledgement is local to the phone, not a websocket message. later frames must
have increasing contiguous sequence numbers. a gap or malformed frame
closes this attachment, not its worker.

initial geometry and acquisition follow [pr 2's bounded admission](herdr-pr2.md#attached-terminal).
proposed handback targets are two seconds after explicit detach/background and
eight seconds after abrupt loss; qualify those through pr 1, then repeat through
the product bridge. heartbeat and credential revalidation have separate purposes.

client messages are `{"kind":"Text","text":string}`,
`{"kind":"Paste","text":string}`, `{"kind":"Key","key":logicalKey,
"modifiers":["ctrl"|"alt"|"shift"]}`. `logicalKey` is
`enter|escape|tab|backspace|up|down|left|right|f1…f12`
or one printable non-whitespace unicode scalar, plus ascii space (excluding
control characters). herdr trims other whitespace before parsing named keys;
those scalars remain available as unmodified committed `Text`, while modified
whitespace keys are locally rejected with an explanation.
modifiers are unique, in `ctrl,alt,shift` order, and may be empty. the phone's
slash/hyphen buttons send `/` and `-`; upstream aliases belong only in the gateway.
this vocabulary covers committed input, hardware keyboards and desktop `enter`,
not just the visible deck. ctrl/alt toggles are phone-local
state and never standalone remote keys. other client messages are
`{"kind":"Scroll","source":"wheel"|"page_key","direction":"up"|"down","lines":uint16,
"column"?:uint16,"row"?:uint16}`, `{"kind":"Resize","columns":uint16,
"rows":uint16,"cellWidthPx"?:uint32,"cellHeightPx"?:uint32}` and
`{"kind":"Detach"}`. the gateway serializes text, paste and supported keys
through one ordered public json-socket queue: `pane.send_text` for typed utf-8,
`pane.send_input{text}` for paste, `pane.send_input{keys:[...]}` for mode-aware
logical keys. use the qualified upstream key names and canonical modifier order;
map space, plus and punctuation to their public parser aliases. `Text` carries
committed utf-8, without c0/del controls; enter/tab/backspace/control chords use
`Key`, and pasted newlines remain `Paste`, never implicit submit.
`Scroll` maps to upstream `terminal.scroll` with zero modifiers. wheel scroll
retains original coordinates when present; unmodified page up/down uses
`source:"page_key"`, the current row count as `lines`, and no coordinates.
herdr moves its scrollback in a normal shell and sends the page key in an
alternate-screen app. `Resize` maps to `terminal.resize`; `Detach` to
`terminal.release` and child cleanup. keyboard ordering holds within that one
queue; concurrent desktop input or direct scroll has no total order with it.
the phone keeps home and end visible but disabled. page up/down remain active
for unmodified visual navigation and are disabled while ctrl/alt is armed,
without consuming that modifier. hardware/desktop home/end/insert/delete and
modified page keys are locally rejected with an explanation; unmodified page
keys use the same public scroll operation as the deck. direct agent-control
page keys remain unavailable. v0.9.1 rejects these named keys through
`pane.send_input`, and raw csi lacks public application-mode encoding. this
accepted feature loss replaces the former full-key cutover gate. ordinary
supported control chords and shifted keys still require proof. no upstream
change has been requested or submitted. [the desktop decoder
gate](issues/herdr-terminal-acceptance.md#pr-2-input-contract-review)
records the pinned reader's defects and a skid-side composition of its public
decoder primitives. no raw-byte fallback or silently dropped keyboard input;
known unsupported keys are consumed locally with an explanation and leave the
attachment alive. undecodable sequences end the attachment with an explanation,
leaving the worker alive. terminal replies stay local.
do not forward xterm's emulator-generated replies as user text: herdr already
emulates the provider terminal. selection/copy stay in the phone's rendered
viewport and invalidate on full redraw/resize; neither becomes remote selection.
phone text/paste is at most 32768 utf-8 bytes, one key per message, scroll
lines 1–512, and geometry uses the current 20–1024 columns/5–512 rows. proposed
admission caps are 2 mib for client websocket input, 3 mib for upstream json
lines and outbound websocket frames, 2 mib for decoded ansi, and 3 mib for
the encoded outbound queue. the public terminal attach producer caps its
serialized terminal frame at 2 mib; its control reader's separate 32-mib cap
does not enlarge that output. base64 of the emitted frame fits below 3 mib.
the earlier 1/2-mib proposal rejected valid measured full frames. overflow
ends this attachment with `protocol_error` and leaves the worker alive. the
larger bounds require pr 2 stress checks for actual phone rendering, memory and
attachment-only overflow closure; source limits alone are not live proof.
server websocket ping/pong every two seconds with a six-second unanswered
deadline must close its controller child so upstream returns geometry to the
desktop even when adb reverse retains a dead tcp connection. backgrounding
releases deliberately; reconnect discovers a fresh terminal ref and never
replays input. the exact handback deadline remains under live proof.

### subsystem boundaries for pr 2

`internal/herdr` alone owns the json socket codec, upstream error mapping,
bounded terminal-session child and framing. `internal/sessions` owns inventory,
metadata, workspace/terminal refs and creation; `internal/agentcontrol` owns
foreground revalidation and readiness/control policy, depending on sessions,
never the reverse. `internal/gateway` owns auth/routes and connection-scoped
stream processes; `internal/fleetclient` owns peer routing and opaque refs;
`internal/agentcli` owns command parsing/help. android's `ProductModel.kt` and
`GatewayClient.kt` own strict wire types, `SkidbladnirController.kt` owns
attachment/action lifetime, and `TerminalConnection.kt`,
`LockedTerminalWebView.kt` plus `terminal.js` own input/frame presentation.
`dev-server` adds the pinned herdr service/socket/binary and profile launch
configuration in pr 4; no gateway request owns that service. jarvis changes
only its cli consumer/tool descriptions in pr 3; this candidate requires that pr.

pr 2 removes `internal/tmux`, `internal/sessionui`, obsolete tmux
`internal/terminalclient`/`internal/terminal` attachment code,
`internal/space`, tmux-specific parts of `internal/agenthook`, duplicate status
detectors and native claude helper only after their callers are changed. retain
the sole process-bound claude `SessionStart` identity registration by adapting
its target proof to `HERDR_PANE_ID`/socket and foreground process identity; do not infer a runtime
profile from `skid_launch_profile`. this adaptation needs a live claude proof
before cutover, and no status/history hook is added. inspect imports,
`cmd/skidbladnir/{main,terminal_exec}.go`, `internal/logging`, gateway and
fleet codecs, `go.mod`, android `Spaces.kt`, `DashboardEntryState.kt`,
`SessionRename.kt`, `TerminalKeyDeck.kt`, `TerminalSelection.kt` and the
dashboard/forge/persistence callers before deletion. preserve shared
`internal/process`, `internal/agentruntime` profile validation,
`internal/strictjson`, auth/pairing, pressure, directory, catalogue and
provider account wrappers where their contracts still apply. no legacy reader,
dual backend or generic bridge framework remains.

## phone boundary and content design

retain native fleet, forge, directory chooser, pressure, dwarf landmarks,
key deck, composition, selection and font sizing. no visual redesign or new art.
the designer for each slice defines content alongside its schema, before coding:

| feature / designer owner | what good content must communicate |
| --- | --- |
| runtime/control — runtime slice | separate creation, readiness, delivery and completion. e.g. `terminal created; startup not confirmed`, `delivery unknown; inspect before sending again`, `interrupt sent; cancellation unconfirmed`. never label inferred status native. |
| grouping/creation — phone slice | name the machine, exact terminal and actual workspace; say when a new workspace will be created. use `move to workspace` and disclose layout effects. unnamed workspace labels are presentation, not fabricated resources. no dwarf lore in operational errors. |
| attachment/input — phone slice | `another connection may be active; take over or go back`; the pre-frame error is coarse, so never assert a conflict was identified. loss pauses input on this phone, not remote work. disclose that desktop input remains active. close/stop confirms the terminal and machine and warns that linked workspaces and their running terminals may also close; detach never implies stop. |
| acceptance — root | concise observed facts, not assurances. state platform/version, boundary, result and limitation without terminal content or account data. |

qualify and write down one owner for each stream behavior:

- herdr emulates the provider terminal; xterm renders frames. distinguish user
  text/keys/paste from emulator-generated replies. do not blindly forward today's
  merged `onData` stream or strip ime input to make the experiment pass.
- define full-frame admission, subsequent-frame ordering, reconnect reset,
  decoded-byte/frame/queue limits and overflow closure. input stays disabled until
  control and initial geometry/frame are established. measure against existing
  bounds; any necessary bound change is an explicit contract delta.
- herdr owns attached scroll position and modes; prove public wheel/page scrolling,
  logical keys and alternate-screen history. no remote click/drag input; upstream
  may use mouse-wheel reports internally. reuse gestures, not assumptions that
  local xterm scrollback is authoritative.
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
| lifetime | exercise retained mutations after rename, move, foreground replacement, terminal/name reuse and server restart. already-replaced targets reject; record the accepted check/write race. qualify disclosed native close effects, including parent-worktree last-pane cascade or refusal under both confirmation settings; outside-group terminals survive. phone/cli/jarvis contracts must not promise single-terminal closure or native halt of every affected agent. |
| launch/readiness | all four profiles preserve configured executable/environment/cwd/flags; verify internally and report equality only. shell and zero-profile host work; startup dialog/timeout differs from failed creation. restart the isolated server with resume disabled and prove no unintended provider relaunch. |
| observation/read | codex and claude working, input wait, idle, unfamiliar screen and unavailable observation have honest source/readiness; reads declare bounded visible/history coverage. no inferred task success or hidden native-capability loss. |
| uncertainty/lifetime owners | suppress a reply after one real synthetic mutation: caller reports unknown within budget, dispatch count remains one after reconnect. kill the probe client/bridge and gateway separately; workers survive and rediscover. cold server restart is tested separately. |
| phone/desktop | attach one shell, codex or claude terminal, without herdr chrome: typing, synthetic multiline paste, supported keys, actual swipe/history movement, keyboard resize, conflict/takeover, detach, background and abrupt loss. disabled deck keys stay visible and inert; unsupported hardware/desktop keys explain their rejection. taps focus input without remote clicks. desktop shares the worker and regains sizing. qualify host stream/lifetime behavior on linux and darwin, without multiplying every input case across every profile. user-waived dictation, gboard paste, local copy and rotation stay `NOT_RUN` |

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
closure contract. accepted limitations belong
in the contract; delete resolved issue records. negative findings mean
`reconsider`, with the responsible blocker and smallest next decision stated.

hard cutover occurs in pr 2, not here: no legacy runtime/protocol paths, fallback
detectors, compatibility negotiation or dual backend. no persistent registry,
task scheduler, history copy, new provider protocol, generalized test machinery,
live pty migration or unrelated cleanup. deleting temporary tests leaves no
retained behavioral regression suite; that is the explicit verification cost.

## investigation result at the pinned baseline

historical result: **reconsider under the original requirements.** the table
below retains its original results. the [approved amendment](#accepted-scope-amendment)
removes exact-only closure and remote-click parity as gates; it does not turn
the old failure into a pass or supply missing live evidence.

the former terminal-only closure contract is impossible through
v0.9.1's public commands. its `pane.close` can remove other terminals when the
target is the final pane of a parent worktree workspace and `confirm_close=false`.
`workspace.close(close_group=false)` prevents that group cascade but closes every
terminal in its workspace. neither command accepts an expected terminal lifetime
or a predicate that the layout still contains only the confirmed terminal. a
gateway check before either command cannot constrain a later desktop layout
change. this is a gap against that former contract, not a request for a skid-side
preflight. the [new qualification](#2026-09-22-reopened-qualification-at-the-accepted-scope) tracks truthful
disclosure and qualification of the accepted native effect.

the public `herdr terminal session control` stream has a separate phone input
gap: it accepts text/bytes, resize, scroll and release, but no mode-aware tap or
mouse command. its ansi frames are rendered cells, without the provider's
mouse-mode negotiation; forwarding xterm's tap bytes would guess the wrong
mode. the native herdr client uses an internal mouse message, whose private
protocol is outside this migration. [the terminal issue](issues/herdr-terminal-acceptance.md)
remains open for the retained input/scroll/lifecycle requirements. these findings
prevented the original `proceed` decision; remote pointer input is now excluded.

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

the next step is resumed qualification at this pin under the approved scope,
not mandatory upstream closure/mouse changes. close the wire/operation table
with disclosed native closure/refusal semantics and single-terminal phone input;
complete the outstanding profile, targeting, uncertainty, lifecycle and actual
scrolling/input proofs. typed stream outcomes and bounded frame admission still
need a qualified contract. native claude and metadata-loss decisions remain open.
pr 2 requires a new evidence-backed `proceed` decision; approval of this scope
amendment is neither that decision nor permission for live/device operations.

the disposable phone probe's scope review added only
`LockedTerminalWebView.kt` inside its isolated source copy. that page-port
decoder is the narrow owner for a typed scroll event from the existing touch
route to herdr's public `terminal.scroll`; it did not alter repository source.
the proof still cannot provide a mouse command absent from upstream.
the proof package was uninstalled and absent after testing; the adb reverse
mapping was removed, and the installed skid package remained present. these
are cleanup observations, not phone behavior acceptance.

## 2026-09-22 reopened qualification at the accepted scope

this appended record does not relabel any historical `FAIL` or `NOT_RUN` above.
all host probes used source `065ef9d6a531c49fb8bee7e818ef837065b21ee9`
and herdr v0.9.1: darwin 26.4.1 arm64 binary sha256
`5fc7a7e7adfaca56fa80aa89dcb025693357268dab8285b9ce2d08a2313c89de`,
arch linux 7.2.3 x86_64 binary sha256
`2a02fed16beb651ef006e1d43f048f652ca4dc58ad053cd2d44450563d5c54b7`.
only isolated servers, terminals, workers and synthetic input were used. the
phone was samsung sm-s906w/android 16 with a separate temporary proof package
0.1.0/code 1, built from git archive `ff114c8` plus disposable edits in the
approved android slice.

| boundary | result | content-free observation and limit |
| --- | --- | --- |
| configured launch, darwin and linux | `PASS` | all four actual rows: verifier compared absolute command, argv including claude name, provider-home environment and physical cwd before exec; foreground provider appeared. shell and zero-profile cases used one root pane. this proves launch, not readiness. |
| interactive readiness, darwin and linux | `NOT_RUN` | startup/update/form blockers or default-idle fallback remained at the ten-second budget. no recognized interactive idle/working/input-wait journey was available without changing account/configuration. no prompts were submitted. |
| disabled resume and references, darwin and linux | `PASS` | cold restart ended old ids/agents, made fresh shell lifetimes and cleared pane/workspace metadata. rename, move, name reuse and fresh client processes preserved live terminal identity. synthetic same-pid exec changed command signature. remaining check/write race is accepted. |
| headless control, darwin and linux | `PASS` | two synthetic shell workers were controlled independently of desktop focus; the other worker stayed unchanged. darwin additionally passed an ephemeral real codex worker invoking the same test-owned controller against another shell. interactive provider readiness was separate and unproved. |
| bounded reads and uncertainty, darwin and linux | `PASS` | recent/visible reads were bounded; a one-line recent read returned only an empty trailing line with `truncated:true`, so absence of content is not absence of work. a temporary bridge withheld one mutation reply past ten seconds: caller saw unknown, effect count was one after rediscovery, no replay. a lost create reply left one terminal; a deliberate second create showed why blind retry duplicates. this proves the probe design, not an unimplemented product gateway. |
| closure and layout change, darwin and linux | `PASS` | ordinary sole-pane close succeeded under both confirmation settings. linked parent close cascaded only with confirmation off; confirmation on refused and kept the group. an unrelated workspace and checkout dirs survived. removing a second pane between preflight and close changed the effect as disclosed; no atomic affected-set guarantee. |
| local desktop client | darwin `PASS`; linux existing-server attach `NOT_RUN` | darwin client attached to isolated existing server and an absent server failed without autostart. linux absent-server refusal passed; the existing-server test pty exited early without a trustworthy product diagnosis. |
| phone shell stream, scroll and takeover | darwin host/physical phone `PASS` | a swipe moved actual host scrollback; no-takeover acquisition gave a coarse pre-frame outcome; explicit takeover displaced only the test controller. synthetic two-line paste reached a test worker through public input. no provider phone journey is inferred. |
| phone key deck | `FAIL` | in a test-owned application-cursor-mode worker, physical home/end were wrong and page-up/down worked. public `pane.send_input` rejects those named keys; fixed raw csi cannot meet the retained mode-aware requirement. |
| phone clean detach and abrupt loss | darwin host/physical phone `PASS` with disposable heartbeat for abrupt loss | desktop 29 rows became phone 33; clean detach returned 29 in under 0.8 seconds, and force-stop plus a two-second heartbeat/six-second unanswered deadline in about five seconds. the same terminal/shell survived. without heartbeat, adb reverse kept a dead tcp relay and phone geometry held: a genuine initial failure repaired only in the temporary adapter, not in product code. |
| phone background release | `FAIL` on one run; later `PASS` twice | one home/background attempt kept phone geometry after five seconds. two later runs returned desktop geometry in under one second and observed local stop/detach. the intermittent failure is not explained, so background acceptance remains open. |
| manual phone interaction and provider terminals | `NOT_RUN` | user skipped dictation, gboard paste, local selection/copy and rotation. adb typing produced nine input events but no exact in-memory host match, so typing remains unproved. configured codex/claude sessions were startup-blocked, so their phone typing/history journey was unavailable. linux phone stream was not exercised. |
| process-bound claude identity | `NOT_RUN` | source permits a content-free `SessionStart` route using inherited herdr pane identity and host process observation, but no hook was installed or live registration proved. runtime profile/provider session must not be inferred from launch metadata. |

the temporary adapters supplied real red/green evidence for input mode and
abrupt-loss release; no assertion was deliberately broken to manufacture red.
the reviewer challenged assertion sensitivity, partial creates, source versus
product boundaries, conflict wording, frame caps and the retained key deck.
the decoded full/incremental phone frames measured at most 1,617/1,019 bytes
in the sampled shell session, and the largest observed json line was 2,382
bytes. these are lower-bound samples, not proof that the then-proposed 1-mib
decoded cap admits every upstream-valid frame (the public reader permits
32 mib, while terminal attach production is capped at 2 mib). exceeding the
cap deliberately disconnects the phone and leaves the worker alive; pr 2 needs
a largest-intended-geometry/output stress check before fixing the cap.

cleanup was verified: both hosts' isolated servers, workers, sockets and
temporary probe files were removed; the proof package and test-owned adb
forwarding were removed; the installed skid package remained present. during
phone diagnostics, one `adb logcat -c` cleared the device-wide diagnostic
buffer despite a tag filter on the command. no skid data or pairings were
changed, and no further log clears were used. this was an investigation mistake,
not part of the accepted proof procedure.

the user accepted terminal-only claude observation and cold-restart metadata
loss. the user further chose to show any surviving native pane label only as a
hint on a new unnamed terminal, and likewise for newly discovered manual panes,
with a fresh ref required for actions, in
addition to the earlier pointer/closure amendment. the
full key deck, consistent background release, phone typing, interactive
provider readiness, process-bound claude identity, provider phone journey and
skipped human phone actions have no acceptance or
proof. at pinned v0.9.1 the home/end failure alone prevents `proceed`; a
qualified public mode-aware key addition is the smallest responsible remedy.
pr 1 status was **incomplete** at that record. the user later authorized work
on pr 2, conditional on its entry gate, and chose to defer cutover while the
retained key vocabulary lacks a qualified public herdr operation.

after that record, an isolated official v0.9.1 darwin probe reached recognized
blocked, ready, working and done observations for codex and delivered one
submission. claude reached `agent.get:idle` with matched visible idle, but
successful public one-call input, separately delayed text/enter input, and
`agent.prompt` on fresh workers produced no observed processing or reply
within 25–35 seconds. this is an open [claude submission
issue](issues/herdr-claude-submission.md), not proof that the provider received
or rejected text. process-bound claude identity was `NOT_RUN`. the isolated
resources were removed; no account or provider configuration changed.

a separate later physical-phone proof against official v0.9.1 matched nine
adb-injected synthetic typing bytes once, then failed a fresh repeat despite
nine foreground input messages. exact typing remains open.
background release failed through the temporary `onStop` path and passed one
temporary `onPause` repair: herdr returned geometry to desktop within two
seconds and kept the worker alive. repeatability and product-controller proof
remain open. dictation, gboard paste, local copy and rotation remain `NOT_RUN`
under the user's waiver.
