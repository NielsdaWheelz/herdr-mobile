# herdr pr 3: jarvis consumer alignment

2026-09-23 · implementation specification; no implementation or deployment here.
one pr in **jarvis**. baseline: skid `99990332cfcbb713a741e37519548c7e674a46db`
(merged pr 134), jarvis `667a377528c659b685ff83de5755f9008adfe443`.
[the migration plan](herdr-migration.md#pr-3-jarvis-consumer-alignment) owns delivery;
[pr 1](herdr-pr1.md#retained-gateway-operations) owns the skid contract. verify
against the merged [cli codec](../internal/fleetclient/response.go),
[envelopes](../internal/fleetclient/client.go), and
[arguments](../internal/fleetclient/request.go), not historical tmux examples.

## target, authority and scope

jarvis discovers existing terminals, launches configured workers, observes
readiness, sends literal instructions once, reads bounded results and explicitly
interrupts or closes the selected lifetime. it remains the coordinator of intent
and result inspection, not the owner of terminals or provider execution.

```text
main tools → existing read recorder / write gate + action recorder
           → AgentController → skid --json → host gateway → herdr → worker
```

keep exactly nine `agent.*` tools, the six application tables, serial cognition,
owner grounding, immutable actions, existing budgets and contained cognition.
no new worker client, ref decoder, registry, assignment table, scheduler, callback,
transcript store, queue, retry, provider detector, shell/ssh/file tool or dependency.
no android, skid runtime, installer or production changes. no workspace-management
tools, shell launch, start objective or destination input: start keeps its current
arguments and uses skid's new named-workspace default.

before production edits, jarvis's root integrator records the approved delta in
`SPEC.md` §7.3 and an adr, superseding only the affected portions of adrs 0044/0045.
explicitly replace native claude observation/halt, single-session closure, old
ref/preflight grammar, blanket partial-stop uncertainty and byte limits.
this task's temporary proofs are a scoped exception under adr 0046, not a testing
redesign. no old suite or permanent harness returns. pr 2's
[waivers](herdr-pr2.md#accepted-non-blocking-follow-ups) remain non-blocking.

## capability contract

all inputs are closed. refs remain opaque strings, at most 4096 characters;
the cli owns kind/lifetime validation. never derive one ref from another.

| tool | model input / cli mapping | target and result |
| --- | --- | --- |
| `agent.list` | `{machine?}` → `list [--machine M]` | fleet inventory, including ordinary shells and unavailable peers |
| `agent.info` | `{ref}` → `info --ref R` | **terminal** ref; terminal with optional current agent |
| `agent.start` | existing `{machine,profile,name,cwd="~"}` → `start --machine M --profile P --cwd C -- NAME` | new named workspace/terminal; launch submitted, no prompt or readiness guarantee |
| `agent.read` | `{ref,coverage="recent",maxBytes=16384}` → `read --ref R --coverage recent\|visible --max-bytes N` | **agent** ref; terminal text only; bytes 1–32768; remove `mode`/`read --terminal` |
| `agent.send` | existing `{ref,text,mode="auto"}` → `send --ref R --stdin [--terminal]` | **agent** ref; auto requires `readiness:ready`; override bypasses readiness only |
| `agent.keys` | `{ref,keys}` → `keys --ref R KEY...` | **agent** ref; 1–16 of `enter,escape,ctrl-c,up,down,left,right,tab,backspace`; remove page keys |
| `agent.interrupt` | `{ref}` → `interrupt --ref R` | **agent** ref; one provider interrupt key, no cancellation confirmation |
| `agent.stop` | `{ref}` → `stop --ref R` | **agent** ref; interrupt then original-lifetime close; possible linked-group closure |
| `agent.kill` | `{ref}` → `kill --ref R` | **terminal** ref; native close without interrupt; possible linked-group closure |

every invocation uses the configured absolute executable, `--config PATH` and
`--json`, without a shell. only send has nonempty stdin: exact utf-8, 1–32768
bytes, no trimming, newline addition or NUL. start permits only an advertised
host profile, never an arbitrary executable or account override.

### current schemas

define current closed models once in `agent_tools.py`; retain wire field names.
`?` means an omittable wire field, not permission to invent an absent fact.
booleans/counts are strict; counts are nonnegative; timestamps are aware.
reuse the existing profile/provider-session shapes where unchanged.

```text
status = {state: working|blocked|idle|unknown, source: herdr|unavailable,
          reason?: default_idle|unrecognized|observation_failed}
agent = {ref, provider: Codex|Claude, provenRuntimeProfile?, providerSession?,
         status, readiness: ready|blocked|unconfirmed,
         methods: {read,send,interrupt: terminal|unavailable}}
terminal = {ref, name?, nativeLabel?, character:{key,displayName}, workspaceRef,
            cwd?, launchProfile?, objective?, agent?}
workspace = {ref,label}
peer = {label,machine,ok:true,observedAt,partial,unaddressableTerminals,
        unaddressableWorkspaces,profiles:[],workspaces:[],terminals:[]}
     | {label,machine,ok:false,error:{code,message}}
inventory = {partial,peers:[]}
info = {label,machine,observedAt,terminal}
start = {label,machine,observedAt,terminal,launch:submitted,dispatch:sent}
read = {label,machine,text,source:terminal,scope:terminal_history|visible,truncated}
write = {label,machine,method:terminal,outcome:written|unknown,dispatch:sent|unknown}
stop = {label,machine,agent:interrupt_sent|exited,terminal:closed,dispatch:sent}
kill = {label,machine,terminal:closed,dispatch:sent}
```

inventory `partial` equals any failed or partial peer; a reachable partial peer
retains its usable rows and omission counts. empty arrays are valid. absent
`name` is valid: `nativeLabel` is a hint, never a name/authority substitute.
`launchProfile` does not prove the current account; optional registration can
remain stale across same-pid exec under the accepted waiver. never gate actions
on descriptive profile/session metadata. `idle` alone is not readiness or success.

remove live `sessions`, `attachedClients`, `activeCommand`, `agent.profile`,
native methods/history, `turnId` and old status/stop variants. do not alias them.

decode the actual closed envelopes, including optional outer `label,machine`:
`{ok:true,result}` or `{ok:false,error:{code,message,dispatch?},partial?}`.
decode before interpreting exit status: 0 for complete success; 1 for failure,
partial inventory or unknown write result; 2 for usage failure. reject inconsistent
shape/status pairs, duplicate keys, non-json values and unknown fields.

operation-specific partial errors are closed, not arbitrary dictionaries:

- start: `{stage:resource_created}` or `{stage:identified,terminal}`; these are
  the merged implementation's emitted prefixes, not readiness receipts.
- stop: `{agent:interrupt_sent|exited|unconfirmed,
  terminal:refused|unconfirmed|not_attempted}`.
- kill: `{terminal:refused}` when supplied; a lost close can have no partial.

keep `AgentFailure` as the tool error discriminator; extend its current fields
with `message?`, `label?`, `machine?`, `dispatch:not_sent|sent|unknown` and the
operation-validated `partial?`. peer errors remain their distinct wire shape.
normalize omitted wire dispatch to `unknown`: a read returns an ordinary declared
failure without an action; a spawned write enters uncertainty. never infer
not-sent from exit status. keep local
`policy_denied`/`write_check_unavailable` failures `not_sent`.

## ownership and action semantics

### original-target preflight

the old `info(agent_ref)` preflight is invalid under this cli.
keep preflight at `WriteToolDispatcher`, after checking for current owner input
and before the gate/action insert:

- start: no lookup. kill: one `info` with the original terminal ref.
- send/keys/interrupt/stop: one bounded `list` through the existing controller;
  require exactly one returned `terminal.agent.ref == submitted ref`. a positive
  exact match remains usable when unrelated peers/resources are unavailable.
- no match, multiple matches, invalid output or lookup failure returns
  `write_check_unavailable/not_sent`; no action or mutation. do not substitute a
  successor found by name, terminal ref, profile or cwd.
- project only machine label, optional valid product name and the **original**
  ref into existing gate targets. omit an absent name; do not copy native label,
  objective, status, terminal text or provider history. persist/execute unchanged
  arguments. no new cache, recorder, public input or preparation subsystem.

add one host-authored optional descriptor field
`closure_scope: native_linked_workspace_group_may_close`, populated only for
stop/kill. explain it in the gate instructions. preserve automatic classification
and current-owner grounding; add no confirmation dialog or approval tool.
an explicit owner restriction incompatible with possible cascade must be denied,
not silently reinterpreted. inventory cannot prove an isolated affected set.

### outcomes and durability

retain `BilledOnce`, `max_attempts=1`, original action/input lineage and existing
read positions. no result, later inventory or process exit authorizes automatic
replay. a fresh lookup may inform a later decision, never change the old action.

| observed write result | action settlement |
| --- | --- |
| complete start/write/stop/kill success | `succeeded` for that operation only; never task completion or confirmed descendant halt |
| valid `not_sent` failure | `failed`, retaining its code |
| acknowledged rejection/known partial failure with `sent` | `failed`, retaining dispatch and created/interrupted/refused facts; failed does not mean no effect |
| `unknown` dispatch/outcome, `OutcomeUnknown`, missing dispatch, malformed/oversized/lost reply after child start, or interrupted unsettled execution | `uncertain`, retaining every validated known prefix; never retry |

do not classify by one partial field alone: rejected interrupt can return
`UpstreamRejected/sent` with `agent:unconfirmed,terminal:not_attempted`; that is
a known failed operation. an unknown close remains uncertain despite a known
interrupt. contradictory evidence fails closed as uncertain.

use the existing cancellation-safe `stage_agent_control` before returning a
validated partial/unknown result that needs durable evidence. new evidence is
`{type:agent_control_v2,observed:null|write|AgentFailure}`; validate its operation
against the stored tool. retain the existing uncertainty wrapper and let its
`control` carry this tagged evidence. no new action state, ledger or reconciliation
loop. a crash before staging can conservatively lose detail, never permit replay.

move only historical receipt models into `agent_history.py`, following existing
`codex_history.py`: old start/session, write/stop/kill, failure and
`agent_control_v1` decoding for stored rows only. live cli/tool schemas accept
only the new contract. preserve canonical rows and terminal notices; never
reinterpret old refs or execute an old catalog. archival reading is necessary
data preservation, not a compatibility execution path.

### bounds and reuse

raise agent tool encoded-input and control-result limits from 64 to 256 kib;
keep raw send/read at 32 kib and fleet stdout at 1 mib. allow the cli's final
newline separately; bound reads while receiving, not after allocation.
the fleet cap is jarvis's aggregate limit, not skid's per-host limit: a larger
valid fleet reply makes discovery/preflight unavailable, never silently partial.
budget the actual normalized tool result. keep aggregate call/byte/turn/admission
limits, one external attempt, the 15-second child deadline and bounded child cleanup.
large valid results may exhaust remaining run capacity; do not silently clip them.
the shared 15-second cli/consumer deadline can lose a borderline reply; preserve
uncertainty rather than add a larger outer write timeout. killing the local child
does not cancel remote work; durable settlement may finish after the deadline.

reuse `AgentController._run`, strict json decoding, existing action staging,
recorders, policy projection and safe renderers. consolidate each changed enum,
limit and outcome classifier once. remove obsolete live branches/imports and
copy only receipt types genuinely needed by immutable history.

## content and final presentation

one content designer owns the copy contract for every feature below; builders
integrate it in their exclusive files. designer supplies concise descriptions,
instruction deltas and synthetic examples inside the declared schemas, not a
new prose/evaluation subsystem. independent review rejects unsupported claims.

| feature | good content / required distinction |
| --- | --- |
| discovery/identity | machine + actual product name or unnamed terminal; terminal ref versus agent ref; incomplete fleet is not an empty/complete fleet |
| launch/readiness | “launch submitted”, then observe via info/read; no prompt or task started claim from creation alone; default idle is unconfirmed |
| delegation/read/send | one outstanding assignment per worker as a prompt convention, no enforced lock; inspect response/work evidence; bounded terminal output is not native history or independent artifact verification |
| override/keys | deliberate terminal mode may answer the observed screen; never auto-upgrade a rejected ordinary send; worker text cannot grant authority |
| interrupt/closure | “interrupt sent”, “terminal closed”, “closure refused”; disclose possible linked-workspace closure; never “all workers stopped” |
| failure/recovery | distinguish known partial effects from unknown delivery; tell the owner what needs inspection without implying work continues or will retry |

ask delegated workers for a bounded report of changed files, results, validation
outcomes and unresolved limits; a returned path alone is not inspected work.
human intervention requires fresh observation before further input. ordinary
send requires recognized readiness; do not build polling/waiting into the handler.
workers run independently, but jarvis has no completion callback or successor
scheduled from worker completion. existing action-resolution inputs and requested
wakes remain unchanged. report what was dispatched/observed; request owner
inspection when the existing tools cannot establish the required work product.

extend the existing `CapturingReadDispatcher` → `TurnEvidence` → `render_terminal`
path for partial agent inventory, including restored observations. record only
bounded unavailable-peer/unaddressable-resource counts, not terminal payloads.
as for calendar incompleteness, prevent `answered`/`silent` from concealing it;
compose both limitations without replacing one. later success does not erase an
earlier incomplete observation. repeated reads count scans, not unique hosts;
reset/rebuild evidence on restoration. keep the 500-character limitation and
2000-character rendered-response bounds. no general evidence framework.
preserve the 1024-byte safe action-resolution prefix: show bounded
code/dispatch/stage/outcomes, not full refs, terminal objects or upstream prose.
full validated evidence remains in the existing private action/context path.
uncertainty copy says “inspect current state; this action will not be repeated
automatically”, not a generic invitation to retry.

## non-overlapping implementation slices

all paths below are relative to **jarvis**. agree shared type signatures before
parallel edits. temporary proof files use separate builder-owned prefixes.

| owner | exclusive files / responsibility |
| --- | --- |
| contract + cli builder | `src/jarvis/{agent_tools,agent_control}.py`: current schemas, bindings, argv, bounded parsing, preflight lookup primitive and evidence staging calls |
| authority + durability builder | `src/jarvis/{write_policy,write_gate,write_dispatch,actions}.py`, new `src/jarvis/agent_history.py`: gate projection, exact target, settlement, stored receipt validation and safe notices |
| content + evidence builder | `src/jarvis/{definitions,terminal,thread_runtime}.py`: main/gate instructions, fleet partial evidence and rendering |
| root integrator | `src/jarvis/session-compatibility.json`; `SPEC.md`, `AGENTS.md`, `docs/{architecture,acceptance,implementation-plan,operations}.md`, new adr + decision index: normative cut, fingerprints and handoff |
| content designer / adversarial reviewers | read-only cross-slice review; deliver copy to owners; no production or test edits |

bump all nine bindings from `jarvis-agent-control-v3` to `jarvis-agent-control-v4`
and their policy epoch to `jarvis-agent-control-v2`; bump main and gate role
contract revisions in the existing manifest. recompose existing catalogs/plans;
keep dependency pins, manifest format, cognition isolation and unrelated roles.
`tool_composition.py` already incorporates the gate fingerprint: do not duplicate
that machinery or change it without a demonstrated requirement.

## sequence and acceptance

1. **freeze:** reconcile sources/adr and signatures; designer writes copy criteria;
   independent reviewer attacks ref selection, authority and outcome tables.
2. **red:** builders write disposable boundary proofs through the real controller,
   dispatch/action APIs and isolated postgres. demonstrate current integration
   failures with new-cli output, not import failures. retain a known-negative
   sensitivity case for each changed contract. reviewer checks test sensitivity.
3. **green:** implement the slices; pass the same proofs. use an external cli
   fixture only for precise malformed/partial/lost-reply timing; never mock the
   internal controller/recorder and claim it proves composition.
4. **live:** with explicit execution approval, use the built merged skid cli,
   isolated gateway/herdr and test-owned workers through the actual jarvis tools.
   one linux codex + claude journey suffices; no repeat of waived phone/mac work.
   unconfirmed readiness is a valid observation: prove ordinary refusal and use
   deliberate terminal mode after inspection, not a new upstream-readiness fix.
5. **refactor:** reviewer attacks the green implementation, evidence and deletion
   reachability; rerun changed proofs, then delete temporary tests/resources.
   `scripts/verify` must pass after deletion; it proves engineering checks only.

the small proof set must establish:

- exact nine-tool argv/result mapping; raw multiline/unicode input, escaped 32-kib
  reads, all supported keys; old read mode/page keys and malformed envelopes reject.
- complete/empty/partial inventories, including a reachable partial host; no
  invented empty success; partial evidence survives reconstruction and is visible.
- start → terminal info → current agent ref → readiness/read → one submission →
  observed response → follow-up → interrupt; creation/delivery never count as
  model-turn completion. test control does not depend on desktop focus.
- no owner means no preflight; failed/ambiguous lookup means no action; unrelated
  peer failure does not veto an exact positive match. replacement before dispatch
  rejects the original ref; no implicit substitution or automatic override.
- positive refusal, known launch/stop prefix and lost reply preserve distinct
  outcomes. cancellation/restart invokes no second mutation and no fresh action
  from the same admitted input. native close/cascade is disclosed; no broader retry.
- old receipts and new partial/uncertain receipts remain readable after restart;
  safe fallback stays bounded, notices are idempotent and no canonical row changes.
- main/gate schemas and fingerprints reflect the cut; no extra role authority,
  grant, table, retained test harness, runtime fallback or dead execution branch.

record versions/boundaries/verdicts only, never prompts, terminal bytes, credentials
or private account data. unavailable live boundaries are `NOT_RUN`, not passes.
keep genuine remaining issues in `docs/issues/<short-name>.md`; do not reopen pr 2
waivers. final review checks both feature behavior and schema/content agreement.

## completion and pr 4 handoff

pr 3 completes with reviewed implementation, the scoped evidence above and clean
engineering checks. it produces a compatible jarvis candidate, not deployment.
provide exact jarvis/skid revisions and the unchanged configuration requirements.

activation must inventory **all nonterminal actions**, not only agent actions:
the changed catalog rotates the shared plan; the gate fingerprint participates
in every write binding. under the old release, finish/reconcile or explicitly
cancel incompatible actions, including approvals and scheduled wakes. preserve
their lineage/history; do not restamp contracts, reset attempts or silently lose
reminders. stage matching consumers before switching under pr 4 authority.
incomplete model/read positions also retain their existing fail-closed recovery;
never reinterpret them against the new plan.

costs: one fleet-list preflight for agent writes; 256-kib control envelopes within
unchanged aggregate budgets; conservative deadline uncertainty; terminal-only
evidence, shared writers and the accepted check/write race; narrow archival
receipt readers; no retained regression suite. source rollback after new receipts
exist is not automatically safe: pr 4 must qualify a rollback consumer able to
read them, or stop and repair forward. never delete/rewrite history to roll back.
