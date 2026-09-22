# herdr exact worker targeting

problem: a herdr name or pane selector addresses its current occupant. the proposed
adapter has not established how an older skid reference prevents mutation of a
replacement worker, particularly across separate lookup and dispatch operations.

impact: wrapping a fresh lookup in an opaque reference cannot by itself preserve
the original worker's observed identity. rename, pane movement and server restart
also need explicit terminal-lifetime semantics.

accepted 2026-09-22: retain pre-dispatch revalidation and no implicit retargeting;
replacement between check and write remains possible. current skid also calls
`resolveAgent` before a separate `Paste`/`Keys` in
[`internal/sessions/control.go`](../../internal/sessions/control.go); its mutation
lock does not prevent independent provider exit. this is not an existing atomic
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
