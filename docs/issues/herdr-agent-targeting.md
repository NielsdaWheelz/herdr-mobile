# herdr exact worker targeting

problem: a herdr name or pane selector addresses its current occupant. skid's
original references now reject observed replacement, but revalidation and
upstream dispatch remain separate operations.

impact: another direct herdr writer can replace a worker between skid's final
check and dispatch. rename, pane movement and server restart require distinct
terminal and worker lifetime semantics.

accepted 2026-09-22: retain pre-dispatch revalidation and no implicit retargeting;
replacement between check and write remains possible. the retired tmux skid
implementation also resolved an agent before a separate input dispatch; its mutation
lock did not prevent independent provider exit. this is not an existing atomic
delivery guarantee and pr 1 must not claim one.

evidence: the v0.9.1
[prompt schema](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/api/schema/agents.rs#L179)
carries a target string; the
[prompt handler](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/app/api/agents.rs#L109)
resolves it and checks the current agent. this does not establish a caller-supplied
expected prior occupant. no failing live reproduction has been run.

pr 1 source review found a possible reference owner within upstream metadata.
`PaneInfo.terminal_id` stays with a terminal across pane moves, while its
workspace-qualified pane id changes. a gateway-owned random, nonexpiring
`pane.report_metadata` token could distinguish the same terminal after gateway
restart from a cold-restored terminal, which receives empty tokens. the gateway
must verify the token after assignment, never restamp an old reference, and
revalidate the original terminal and observed foreground process before
dispatch. this does not make the separate terminal write atomic. gateway-process
restart and worker-replacement proofs remain open; no reference encoding is
accepted from source inspection alone.

an isolated darwin v0.9.1 binary probe preserved the same terminal id through
rename and workspace movement. the public pane id changed, and its old alias
still resolved the original after another pane reused the old display name.
`pane.process_info` exposes foreground pids but no process start identity;
reuse skid's host process observer to distinguish replacement and pid reuse.
agent-replacement and check/write scheduling remain unproved.

resolution owner: [migration pr 1](../herdr-migration.md#pr-1-feasibility-and-implementation-contract).
on isolated resources, capture a target, replace its worker, then exercise each
retained mutation with the original reference. also move/rename the original
terminal and recreate the server. non-closure controls must not reach another
worker; close/stop effects follow the approved native group-closure contract.

resolved when: the [pr 1 contract](../herdr-pr1.md) and real-boundary proof establish
original-lifetime revalidation, rejection of already-replaced workers and no
implicit target substitution. closure scope is qualified separately in
[the native closure qualification](../herdr-pr1.md#2026-09-22-reopened-qualification-at-the-accepted-scope), not an exact-only requirement.
characterize the remaining check/write race explicitly; no atomic
claim. do not add a process supervisor or identity database merely to conceal an
unsupported upstream boundary.

2026-09-22 isolated v0.9.1 darwin/linux proof: terminal id and metadata token
survived rename, move, display-name reuse and new gateway client processes;
cold restart changed the terminal id and cleared pane/workspace tokens. a
synthetic same-pid/start `exec` changed the observed command signature, so the
proposed agent ref also carries a protected executable/argv signature. a raw
stale-worker write reached a successor as a negative control; the proposed
gateway must reject it before dispatch. headless control of two workers kept
the unfocused worker unchanged; a real ephemeral codex worker invoked the same
test-owned controller on darwin. the accepted check/write gap remains. pr 2
must repeat the original-ref matrix and one-dispatch behavior in product code.

2026-09-22 pr 2 candidate: isolated darwin and linux gateways each launched
all four configured profiles through the product route. three codex workers
on each host were recognized as blocked; a bounded terminal read succeeded, ordinary send to a
blocked worker rejected before dispatch, and explicit terminal-mode send
reported one upstream write. claude launched and remained a foreground process
on linux, but herdr did not classify it as an agent in either isolated fleet, so process-
bound claude targeting is still `NOT_RUN`. temporary gateway probes verified
stale terminal rejection after websocket admission and one partial dispatch
path; they were removed. a full live original-agent-ref replacement matrix,
lost-reply uncertainty, native closure cascade and real cross-worker product
orchestration remain open. the check/write gap remains accepted and explicit.

2026-09-23 pr 134 host proofs: a test-owned unix relay forwarded one real
`workspace.create` to herdr, read its success, then dropped the reply. the
darwin product gateway returned http 504 `OutcomeUnknown`, dispatch `unknown`;
one new labelled workspace existed afterward and the upstream create count was
one. the same drop-after-success product result, single creation and cleanup
passed on arch linux. no request was retried. both workspaces and relays were
removed. ordinary pane closure through the product passed on darwin and arch
linux with official v0.9.1. a public-herdr-created linked worktree cascaded
when its parent was closed on both hosts. an unrelated terminal and both
checkout directories survived. on the isolated linux `confirm_close=true` fixture,
product close returned http 409 `ClosureConfirmationRequired`, dispatch `sent`
and partial terminal `refused`; parent, linked pane and unrelated terminal
survived. all closure resources were removed.

the darwin and arch linux product gateways then preserved one original codex
agent ref through rename, move and gateway restart. after that worker exited
and a successor occupied the same pane, stale read, send, keys, interrupt and
stop each returned http 409 `AgentStale`, dispatch `not_sent`. the successor,
an unfocused second worker and a separate name-reuse shell survived. a real
separate official v0.9.1 runtime-restart proof on each host rejected old
terminal info/kill and old agent read/send/stop with `TerminalNotFound`,
dispatch `not_sent`, after a new real worker started; its new ref survived.
the stop repair translated public `ctrl-c` to herdr's `ctrl+c` spelling and
held skid's mutation lock across final original-ref revalidation and native
close. live negatives rejected an unclassified foreground successor and a
same-pid `exec` successor; a returned shell and a still-original worker took
their intended stop paths. direct external herdr mutation can still race the
upstream close.
a real linux codex source used the candidate cli through private https to
control a separate linux codex target, which performed one synthetic filesystem
effect.
a real darwin codex source invoked the same cli over ssh on linux to control a
separate darwin codex target, which performed a distinct synthetic effect.
both source workers processed one instruction and both target effects were
observed. the mac-local cli was not qualified with the private test ca because
darwin's verifier did not use the process-local trust file; no system or
keychain trust was changed. test-owned panes, tunnels and certificate material
were removed. the accepted external check/write race remains; this proof
does not claim atomic protection against another direct herdr writer.

the claude product proof separately showed process-bound optional identity on
one original ready worker. after it exited, a different recognized foreground
worker in the same pane did not inherit its profile/session metadata, and the
old agent ref's read returned http 409 `AgentStale`. the retained registration
binds pid, start identity and provider; a same-pid `exec` may retain optional
fields until a new hook claim, while action refs also bind the command
fingerprint. no stronger optional identity claim is made.
