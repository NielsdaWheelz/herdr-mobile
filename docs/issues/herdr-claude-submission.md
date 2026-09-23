# herdr claude submission is not qualified

problem: a wrapper-launched claude worker reached a consistent public idle
observation, but accepted public input did not produce observed processing.

impact: `readiness:ready` could admit a send whose only proven effect is
upstream input-queue acceptance. the contract already distinguishes `written`
from provider processing;
the retained claude send journey still lacks acceptance.

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

resolved when: on isolated official release builds, a recognized ready claude
worker processes one literal submission through the chosen public operation,
with observed working or a distinguishable provider response. verify the
original worker remains the target and that the gateway truthfully reports
delivery uncertainty. repeat through implemented skid before cutover; do not
infer task completion from a write acknowledgement.
