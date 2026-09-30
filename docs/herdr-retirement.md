# skid only cutover

2026-09-29. implemented and qualified under the amended scope;
[observations](herdr-retirement-qualification.md) record the exercised boundaries.
retire upstream herdr and herdr-mobile on macbook, devbox, arch and android;
move jarvis worker control to current skid. this is a scope amendment to
[architecture](architecture.md) and [coexistence](herdr-mobile-separation.md),
the specification defines requirements; observations establish deployment.
this change authorizes temporary behavioral tests
under [testing policy](rules/testing.md), not a permanent harness.

owner decisions: discard all remaining herdr workers at cutover; delete old
worker-receipt readers and accept lost interpretation. preserve raw action
records, canonical messages, provider accounts and history. the subsequent
implementation request authorizes this cutover.

execution amendment: jarvis keeps the shared codex app-server approach; a private
cognitive process/account is out of scope. the existing deployed cognition
connection is broken after dev-server retired its former broker. repair belongs
to a separate task. jarvis may remain disabled and stopped; source delivery,
worker-client qualification and herdr retirement proceed independently. activation
and normal cognition remain pending observations, not cutover gates.

## 1. scope and final state

| concern | final owner and behavior |
| --- | --- |
| terminals and processes | existing tmux and skid; gateways restart without killing terminals |
| execution and conversation history | existing codex/claude and account homes |
| desktop and phone | existing skid cli/browser and `dev.niels.skidbladnir`, existing pairings |
| jarvis workers | one current skid cli consumer; explicit terminal or conversation targets |
| host installation | dev-server; skid on loopback `7341`, tailscale https `8443 /v1` |
| retired | herdr servers/workers/hooks/gates, mobile gateways, `7342`, `8444 /v1`, `dev.niels.herdr.mobile` |

no new runtime, worker database, provider-home migration, transcript store,
credential-scope system, compatibility backend, automatic replay, transport
fallback, phone redesign or generic migration framework. jarvis gains no shell,
group, tracking, queue, wait or unread-management tool. cognition and unrelated
connectors stay under their existing contracts. historical repositories remain;
active installers, instructions and execution paths become skid-only.

source baseline: mobile `777d1dd`, skid `edf85f8`, jarvis `e6a6d20`,
dev-server `92d4e68`. refresh heads and worktree changes before implementation;
preserve unrelated edits. directory names are historical:

| repo | local directory | governing source |
| --- | --- | --- |
| herdr-mobile | `skidbladnir` | this specification |
| skidbladnir | `skid-v1` | `docs/native-agent-observation.md`, `internal/fleetclient/` |
| jarvis | `jarvis` | `SPEC.md`, `docs/operations.md`, `src/jarvis/agent_tools.py` |
| deployment | `dev-server` | `SPEC.md`, `workstation`, `ansible/playbooks/` |

## 2. composition and invariants

```text
jarvis tools -> existing policy/action recorder -> fixed skid cli subprocess
  -> private peer configuration -> each host's authenticated skid gateway
    -> tmux/session control OR existing short-lived native provider helper
phone and desktop -----------------------> same gateways and workers
```

skid owns refs, routing, live target validation, provider methods and outcomes.
jarvis owns intent, write authority, bounded decoding and durable settlement.
dev-server owns executable installation; jarvis deployment owns its private
client configuration. providers retain their existing shared homes.

- models receive no bearer, filesystem account home, endpoint or executable.
- echo opaque refs unchanged. names are creation/presentation fields, never
  replacement mutation targets. a new observation does not renew authority.
- a captured conversation survives terminal reassociation/deletion; a terminal
  operation still requires its exact tmux/process lifetime. native stop retains
  its captured turn and refuses a successor. inspection never substitutes refs.
- native unavailability never selects terminal input. terminal delivery proves
  bytes dispatched, native acceptance proves admission; neither proves completion.
- every write uses the existing one-attempt durable action path. after possible
  dispatch, lost/malformed replies remain unknown; no retry or replacement action.
- credential-free, content-free evidence only. result text may enter jarvis's
  authorized conversation path, never operational logs or test reports.

## 3. capability and api contract

### skid inspection addition

current `info --ref` refreshes a terminal and may report conversation b while
native writes using the original ref still target a. it cannot ground a native
write preview. expose the existing internal conversation inspection through:

```text
skid --config PATH inspect --ref REF --json
result = {label, machine, target:{ref, conversation, turn?},
          inspection:<existing inspect envelope>, observedRef?}
```

`conversation` and optional `turn` are captured from the supplied ref;
`inspection` contains either current `ConversationRuntime` or the existing typed
failure. after local ref/config admission, return captured identity even when
native inspection fails; malformed refs and unknown machines fail the outer
envelope. `label` and `machine` come from admitted configuration. require a
conversation-bearing ref. preserve the original ref byte-for-byte; successful
inspection must match its conversation. only on inspection success, `observedRef`
is a separate conversation-only ref encoding that same conversation and its
newly observed runtime/turn. a later explicitly authorized action may choose it;
preflight never substitutes it into the current action. missing terminal is
irrelevant. reuse
`/v1/conversations/inspect`, reference parsing and response validation; no new
gateway route, provider method, name resolution or wire version negotiation.
retain the internal inspect result used by existing clients; compose this cli
projection at its owner instead of changing unrelated consumers.
exit zero requires successful inner inspection; inner failure exits one while
retaining the outer target. the adapter reads that structured result either way.

### jarvis tools

hard-cut the catalog to these nine `agent.*` tools. remove `interrupt` and
`kill`; no aliases. all inputs are closed objects. refs are opaque strings of
1–4096 characters; machine is `macbook|devbox|arch`; profile is
`personal|work|work2|claude-work`. start keeps the existing jarvis name grammar
and cwd bounds. explicit mode/scope fields below are required, not inferred.

| tool | input beyond `ref` | skid command and meaning |
| --- | --- | --- |
| list | `machine?`; no ref | `list [--machine …] --json`; unavailable peers remain present; partial is not empty |
| info | `target:terminal|conversation` | terminal: `info --ref`; conversation: `inspect --ref` |
| start | `machine, profile, name, cwd?`; no ref | `start NAME --machine … --profile …`; no initial prompt/readiness promise |
| read | `source:latest|history|terminal`, `maxBytes` default 16384, max 32768 | `read --ref`, with `--history` or `--terminal`; reads never acknowledge unread |
| send | `text` | `send --ref --input peer --stdin`; native peer admission only |
| text | `text` | `text --ref --stdin`; explicit terminal paste and submit |
| keys | `keys` | `keys --ref`; 1–16 skid keys, including `ctrl-c`, not herdr's `ctrl+c` |
| stop | `mode:native|terminal` | `stop --ref [--terminal]`; stop captured work, retain terminal |
| close | `scope:conversation_and_terminal|terminal_only` | `close --ref [--terminal-only]`; separate halt and closure outcomes |

all commands include fixed `--config` and `--json`; use argv without a shell,
text through stdin, never caller-supplied flags. text is nonempty utf-8, at most
32768 bytes, no nul. key vocabulary is exactly skid's current eleven-key set
in `internal/fleetclient/request.go`. no user-attributed native input or queue.
claude native send reports unavailable before dispatch; explicit `text` remains
available when its terminal/process target is valid.

use skid's existing closed envelope:

```text
success = {ok:true, result:<operation result>}
failure = {ok:false, error:{code, dispatch:not_sent|unknown, conversation?}}
```

`internal/fleetclient/response.go` and `internal/agentruntime/control.go` own result
schemas. jarvis models the consumed current shapes at one adapter boundary;
preserve optional absence, method, output scope, truncation, partial inventory,
captured conversation on failed creation, and separate close outcomes.
do not collapse results to `written` or `closed`. parse the envelope before
process exit status: valid partial inventory and unconfirmed outcomes can exit
nonzero. settle unconfirmed writes as uncertain, not success. a failure with a
known created conversation remains partial even when terminal creation was not sent.
native receipts and process presence are separate facts; unsupported state is
not idle. no exhaustive duplicate provider schema or code generation.

reuse skid's 15-second command budget. cap subprocess stdout at 1 mib for
inventory and 64 kib otherwise; discard stderr directly to the null device
(zero captured bytes). retain a valid owned stdout receipt even if stdin closes
early or the subprocess has not exited. never log discarded diagnostics.
allow 20 seconds per subprocess: 20 seconds for reads/start, 45 seconds for one
preflight plus one write, 65 seconds for two preflight reads plus one write
(compound close or native authority grounded by terminal name). drain stdout
while supplying stdin. preflight failure and failure to
spawn the mutating command are not-sent; after spawn, timeout/output overflow/bad envelope
on a write is unknown unless a valid owned receipt proves otherwise. cleanup
terminates only that subprocess, never workers or provider daemons.

keep `BilledOnce`, one executor entry and `max_attempts=1`. require current
owner input under existing policy before preparatory lookups. before the
write gate, native operations inspect their original captured
target; terminal operations inspect their original terminal. compound close
describes both original targets. a native-only write requires successful native
inspection; never borrow a terminal's name to ground it unless that terminal's
observed conversation equals the captured conversation. explicit conversation
authority does not require a surviving terminal. feed normalized target facts
to existing policy; exclude worker text. an unavailable native halt does not forbid an explicitly
authorized exact terminal closure: retain skid's partial-close behavior. execute
the original arguments once after authority is granted; gateway validation owns
races after preflight. bump tool, policy, implementation and session revisions
where their existing owners require it; no new revision registry.

## 4. jarvis deployment and historical data

| item | contract |
| --- | --- |
| cli | `/usr/local/libexec/skidbladnir`, regular file, root:root `0755`; same admitted artifact/pin as devbox gateway |
| client file | `/etc/jarvis/agent-client.json`, regular file, jarvis:jarvis `0600` |
| settings | `JARVIS_AGENT_CLI_PATH`, `JARVIS_AGENT_CLIENT_CONFIG_PATH`; remove `JARVIS_HERDR_*` |
| containment | retain existing service restrictions, including `ProtectHome`; no link into the development user's home |

reuse the existing skid configuration, with all three peers:

```text
{peers:[{label, origin:"https://HOST:8443", machine:"mh-…", bearer}],
 defaultMachine:"mh-…"}
```

`fleetclient/config.go` remains its validator: unique identities/origins/labels,
canonical bearer, fixed tls port and private file. generate through existing
human fleet provisioning; supply that private file explicitly to jarvis's
`deploy/install-private-state`. no second authored peer table or automatic
credential minting. jarvis never relies on `defaultMachine` for writes.

stage/validate before any activation. changing cli bytes or client configuration
requires jarvis paused and cleanly stopped; check this before switching either
the devbox gateway or its root cli. identical apply is inert. dev-server
installs the cli from its admitted devbox artifact; jarvis deploy installs its
private file. check under actual service restrictions before resuming. bearer
rotation requires explicit client redistribution. preserve the existing cognition
contract and account/history. do not restore removed broker files or provision
private cognition as a retirement side effect. qualify the worker client separately
under the actual service uid and relevant restrictions while jarvis is down.

delete `agent_history.py`, `codex_history.py` and all old worker receipt
decoders/dispatch branches. before removal, old code must settle every
nonterminal action, finish or explicitly park/reconcile unfinished turns under
existing rules, materialize action-resolution messages and drain their delivery.
the catalog change affects all pending actions, not just workers.

terminal old worker rows become opaque archives BEFORE current tool lookup,
input/evidence decoding or uncertainty rendering. validate common immutable
digest/state/timestamp invariants; retain raw arguments/results without parsing
their retired payloads. render only action id, tool, recorded status and
`receipt details unavailable after cutover`. never reinterpret them as current
receipts. current-revision malformed data remains a defect; old nonterminal
rows block activation. this archive path is not a legacy codec or replay path.

## 5. host and phone retirement

one bounded operator procedure; no permanent herdr lifecycle manager. preserve
existing mobile removal until used, then remove it and its product selector.
ordinary apply must never silently discard workers.

| remove after verification | preserve |
| --- | --- |
| linux `herdr.service`, `herdr-mobile.service`; mac `dev.niels.herdr`, `dev.niels.herdr-mobile` | skid units, cli/browser, native helper/plugin and existing workers |
| `~/.local/bin/herdr`, `herdr-mobile`, `herdr-mobile-launch`; owned config/share/state deployment files and dev-server receipts | all skid roots, credentials, signing files and human client configs; permanent mobile signing backups |
| exact herdr hooks/scripts in `.codex`, `.codex-work`, `.codex-work2`, `.claude`, `.claude-work` | unrelated settings/hooks, authentication, histories and memories |
| `~/.local/libexec/herdr-gate`, exact forced-command ssh authorizations, `/etc/jarvis-herdr` | unrelated ssh access, shared codex services and provider binaries |
| exact tailscale https `8444 /v1` handler | `8443 /v1`, other handlers and tailscale itself; never reset serve |
| android `dev.niels.herdr.mobile` | `dev.niels.skidbladnir`, data, signer and pairings |

identify default, named and unmanaged herdr servers by executable/socket/process
ownership; disable their restart paths before stopping them. discard only proven
herdr-owned workers. never blanket-kill codex/claude or mutate ordinary tmux.
remove exact integration entries while the pinned uninstall capability is still
available, or edit the identified json entries atomically after parse validation;
preserve other values. inventory project-local registrations too, without reading
or reporting prompt content. unknown ownership stops deletion of that item.

the existing mobile remover deletes bearers, identities and generations. its
devbox ingress preflight can block removal; workstation removal can finish the
runtime while ingress fails. inspect each postcondition, repair only remaining
owned residue, never claim transactionality. missing owned files mean completed;
foreign content means operator repair. remove verified retired-product
dev-server staging assets, instructions and temporary rollback copies after
their rollback window. preserve permanent signing keys/backups privately;
directories containing only these are allowed residue, not active installations.

## 6. ownership and content design

each row has one implementer and one content designer; a different reviewer
challenges its contract, tests, implementation and cleanup at every handoff.
designers own copy inside their row, not another row's files. root owns contract
changes and assigns any newly discovered shared file before edits.

| slice | exclusive files | designer's definition of good |
| --- | --- | --- |
| skid inspection and obsolete guards | skid `internal/agentcli/run.go`, `internal/fleetclient/{client,response}.go`, `cmd/skidbladnir/main.go`, `internal/{runtimeenv/environment,agentruntime/profile,agentcontrol/native,tmux/client}.go`, `deployment/providers/{provider-command,shell-init}`, relevant cli/native docs; root handles release/pins | captured target and current observation visibly separate; no refreshed authority |
| jarvis adapter and archive cut | jarvis `src/jarvis/{agent_control,agent_tools,agent_history,codex_history,actions,write_dispatch,write_policy,settings,service,cli,definitions,tool_composition,session}.py`, `session-compatibility.json`; `deploy/{install-private-state,verify-containment,activate-release}`; affected spec/adr/operations docs | each tool names target, effect, uncertainty and unavailable cases; archive copy claims only recorded status |
| deployment retirement | dev-server `workstation`, `devbox`, `lib/{herdr,herdr-mobile,common,gateway-ingress,gateway-runtime,skidbladnir}.sh`, `assets/{herdr,herdr-mobile}/`, `assets/agent-instructions.md`, obsolete herdr references only in `assets/codex/codex-shared.py`, `ansible/playbooks/{apply,gateway}.yml`, `ansible/roles/{herdr,jarvis_herdr,skidbladnir,workspace_assets}/`, affected docs | host/component, observed condition, completed effect, next step; no blanket success or secret output |
| cutover and phone | root: release pins/publication, installed hosts/phone, mobile retirement docs, qualification and issue records | package/host-specific progress, truthful partials; retained skid ui needs no redesign |
| acceptance review | verifier reads all slices; writes only its review findings | every pass names an exercised boundary; no device/no boundary is `NOT_RUN` |

reuse existing skid codecs, cli parser, auth, partial outcomes, jarvis action/gate/
activation machinery and deployment staging/ingress helpers. remove the herdr
profile-home table, ssh codec, refs, selectors and orphaned helpers at their
owners. consolidate product branches made redundant by retirement where callers
prove it; do not rewrite the shared installer or unrelated cognition machinery.
remove skid's now-dead herdr coexistence guards only where their whole call path
is obsolete; preserve independent launch-environment isolation.

content examples: `native input accepted; completion unconfirmed`,
`terminal closed; conversation stop unconfirmed`,
`write outcome unknown; not replayed`,
`archived action; recorded status: succeeded; receipt details unavailable after cutover`.
forbidden claims: ready from process presence, completed from accepted, fleet
removed from one host, or repaired history from a successful restart.

## 7. ordered delivery and acceptance

each step: designer freezes its bounded contract/copy; reviewer attacks the
assumptions; implementer demonstrates a sensitive red; implement and obtain
green; adversarially review uncertainty/ownership; refactor and rerun affected
checks; delete temporary tests after recording content-free results. pure
deletions use caller/build/link evidence where no changed behavior exists.

1. freeze the contract and reconcile it into each repo's governing docs. stage
   the inspection addition, jarvis adapter/archive cut and deployment changes.
   remove unconditional herdr preflight/install roles and change the shared
   instruction template before any ordinary apply; retain only removal helpers.
   no broad revert of jarvis or resurrection of old adapters.
2. qualify the candidate end-to-end in isolation. use one disposable test driver
   per boundary, not a framework. mutation fixtures create their own isolated
   tmux socket, gateway/client configuration and exact workers. no production
   test seam, permanent harness or retired gate returns.
3. stage a normal signed skid release containing the cli addition, its unchanged
   helper pin, and matching host/apk artifacts. apply the existing release
   contract. pause/settle/clean-stop jarvis before switching its cli/config.
   install across all three hosts and update the skid apk in place; verify
   version/signature and preserve pairings. no fleet pass inferred from a pin.
   this is a coordinated maintenance interval: old clients reject the new wire
   schema. update the phone immediately after the hosts, then have the owner
   quit/reopen already-running desktop browsers. prove session visibility before
   uninstalling herdr; do not kill or recreate workers to repair client access.
4. retain private code/config/artifact rollback inputs. stage jarvis while
   disabled, paused and cleanly stopped. prove zero incompatible/in-flight actions,
   stale turns and pending old worker resolutions/delivery. discard permission
   does not cancel unrelated durable work. qualify worker-client reads against all
   three production gateways under the actual service uid/restrictions. activation
   and owner resume follow the separate shared-cognition repair, using unchanged
   activation checks; they do not gate retirement.
5. disable herdr/mobile restart and ingress, discard herdr workers as authorized,
   verify skid and the jarvis worker client, then purge owned components and
   uninstall the mobile apk.
   remove remaining cleanup-only source assets/selectors after use; run ordinary
   host apply once to prove no runtime, hook or instruction resurrection.
6. root records source/artifact identity, boundary, verdict and remaining issue
   per host/phone. delete resolved issues; retain unresolved ones under
   `docs/issues/` with evidence and resolution criteria. mark mobile retired;
   keep checkout paths/history rather than disrupting provider project identity.

| acceptance | red and required final observation |
| --- | --- |
| exact native target | new inspect + policy identify a after terminal tracks b and after terminal deletion; failed inspection retains captured identity; explicit observedRef permits a later stop of newly observed work, never substitution into a stale write |
| worker lifecycle | actual jarvis tool/action path → cli → gateway → helper/provider; codex launch/read/native peer send/stop/close, claude native-send refusal plus explicit terminal text; honest creation partials and terminal closure despite unavailable native halt |
| authority and uncertainty | stale terminal/process/turn refused; changed machine pin refused; unsupported native input sends nothing; lost write reply settles unknown once, including restart/recovery, with no second effect |
| archive hard cut | old succeeded/failed/uncertain/cancelled worker rows, including removed `agent.kill`, `agent.interrupt` and `codex.*`, remain stored and render opaque; current malformed receipts fail; old unfinished actions/resolutions block activation; cognition code is unchanged, runtime activation is separate |
| deployment convergence | all three hosts pass an isolated owned-worker smoke; actual jarvis uid/relevant service restrictions read each production gateway while the main service is stopped; invalid config refuses; changed cli while active refuses, identical apply is inert |
| precise retirement | seeded unrelated hooks/settings/ssh/serve entries survive; interrupted removal completes on rerun; no herdr server/unit/hook/gate/owned ingress remains after normal apply |
| retained phone | herdr package absent; skid package/signature/data/pairings retained; owner confirms normal attach on each host; automated phone checks do not mutate baseline tmux |

cover linux and darwin mechanisms once each; repeat host identity/routing/removal
postconditions on all three. do not multiply provider/profile permutations
without a concrete differing path. seeded disposable provider content only;
report no prompts, outputs, tokens or account data. ordinary engineering checks
remain required but cannot establish behavioral acceptance.

live/integration, tmux and adb operations still need explicit approval in the
execution turn under `AGENTS.md`. this spec supplies no standing opt-in.
scripted tmux mutations target only exact test-created lifetimes on isolated
`-L` sockets. owner-performed phone attachment supplies the retained production
interaction check. unperformed boundaries remain `NOT_RUN` and block the
corresponding completion claim.

## 8. decisions and accepted costs

- full existing gateway bearers replace ssh command allowlists. jarvis policy
  constrains its tools; it is not a second credential-scope security boundary.
- subprocess transport reuses skid's authoritative client at the cost of a
  process per call and coordinated cli maintenance. no parallel http client.
- small cli inspection addition closes the authority gap; require a normal
  release rather than duplicating opaque-reference parsing inside jarvis.
- historical typed receipt interpretation is lost by owner choice. raw records
  and already-materialized canonical messages survive; no conversion migration.
- discarding herdr workers is irreversible process loss. provider history is
  retained; rollback cannot restore a running terminal or unsaved turn.
- before destructive retirement, restore staged code/config while stopped if
  necessary. after it, recover a compatible skid/jarvis deployment or leave
  jarvis paused and repair. no automatic herdr resurrection, mixed-contract
  actions or replay. no prior skid-backed jarvis release is assumed available.
- temporary tests leave no retained behavioral regression protection. preserve
  engineering checks and concise evidence; no proof ledger or permanent matrix.
- one-time retirement avoids perpetual legacy lifecycle code. an overlooked
  host requires this procedure and explicit repair, not silent fallback.

- jarvis remains down until its shared cognition connection is repaired separately.
  worker-client probes prove gateway access and restrictions, not a functioning
  cognitive loop. this accepted omission does not delay herdr retirement.
