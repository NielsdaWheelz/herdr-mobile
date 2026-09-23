# herdr claude submission qualification

problem: earlier wrapper-launched claude workers reached public idle, but
accepted public input did not produce observed processing. the current devbox
product proof closes that gap on linux; the macbook work profile is signed out.

impact: `readiness:ready` means the observed terminal accepts input; `written`
means delivery, not a completed provider turn. macbook work-profile model
processing remains unqualified until that account is authenticated.

evidence: an isolated official herdr v0.9.1 darwin arm64 server reported
`agent.get:idle` and a matched `agent.explain.visible_idle` for `claude-work`.
one `pane.send_input` text-plus-enter returned success but had no observed
working state, command execution or reply within 25 seconds. a separate fresh
worker received text and enter in two successful writes 300 ms apart with the
same absence. another fresh worker received one successful public
`agent.prompt`, whose source queues text and enter with the same delay; 35
seconds again showed only idle and no observed processing. no accepted send
was retried. source and public responses establish upstream input-queue
acceptance, not provider processing or why claude did not act. all test-owned
resources were removed.

resolved when: on an isolated official release build, a recognized ready
macbook `claude-work` worker processes one literal submission through skid and
produces a distinguishable provider response on the original process. devbox
now meets this condition; do not infer task completion from a write
acknowledgement.

2026-09-23: on an isolated official v0.9.1 linux server, the exact deployed
devbox `claude-work` wrapper was signed in and reached matched visible idle.
one literal public text-plus-enter submission changed the same original
foreground process from working back to idle and produced a distinguishable
provider answer in that terminal. the answer matched a random synthetic
challenge checked only in memory; no content was recorded or retried. this
qualifies the upstream operation on an authenticated work profile. the exact
macbook work wrapper is currently signed out: its original process handled one
submission and returned a distinct authentication refusal, but no model turn. the
macbook personal profile remained in account selection, with readiness
unconfirmed. the required repeat through the product route passed on the
authenticated devbox profile: the candidate gateway created a ready worker,
one product send returned http 200 `written/sent`, and a bounded product read
found a distinct provider reply. the original pid, agent ref and terminal ref
stayed unchanged. no send was replayed. the working transition was not
captured by sampling, so the reply is the processing evidence. macbook work
authentication is still a host-specific blocker for its own model-turn proof.

the product's process-bound `SessionStart` proof also passed on a fresh original
devbox worker: recognized ready/idle, `provenRuntimeProfile:claude-work`, and a
provider session id appeared together in inventory. the first single-token
attempt was truncated at herdr's 80-character metadata value limit and was
correctly omitted. a bounded multi-token registration repaired that cause.
after this real worker exited, a different unregistered foreground worker in
the same pane did not inherit the optional profile/session facts; the original
agent ref returned http 409 `AgentStale` on read. temporary plugin, workers,
server, gateway and tests were removed after the proof.
