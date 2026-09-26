# deployment runbook must drop the unnecessary upstream reset

problem: the root spec and handoffs now preserve the running herdr runtime,
but dev-server `1296309` still directs stopping it and discarding its session
snapshot. the user permits losing old panes; that is not a requirement to
kill workers. provider relocation is withdrawn for both products.

impact: the plan introduces avoidable worker interruption and jarvis action
settlement into a gateway/package rename.

evidence: `internal/sessions/manager.go`, `project`, still projects native
labels, cwd and live agents without the renamed mobile metadata.
`internal/sessions/metadata.go`, `decodeObjective`, returns an empty objective
when the new key is absent. the runtime does not require a reset to read these
panes. no compatibility reader is needed to tolerate missing mobile decoration.

accepted correction: remove the unconditional reset from the deployment
runbook; replace only the old gateway. accept missing old mobile decoration. perform
an upstream restart only for a separately established runtime/service change,
with worker coordination if it is actually needed.

resolved when the owners agree on that cutover contract and the new gateway
is qualified against the unchanged herdr runtime. this is source evidence;
live coexistence remains `NOT_RUN`. no runtime or snapshot was changed by the
audit.
