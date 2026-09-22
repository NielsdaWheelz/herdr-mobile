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

resolution owner: [migration pr 1](../herdr-migration.md#pr-1-feasibility-and-implementation-contract).
on isolated resources, capture a target, replace its worker, then exercise each
retained mutation with the original reference. also move/rename the original
terminal and recreate the server. confirm no unrelated terminal is affected.

resolved when: the [pr 1 contract](../herdr-pr1.md) and real-boundary proof establish
original-lifetime revalidation, rejection of already-replaced workers, no implicit
target substitution and exact terminal closure, with upstream support added if
necessary. characterize the remaining check/write race explicitly; no atomic
claim. do not add a process supervisor or identity database merely to conceal an
unsupported upstream boundary.
