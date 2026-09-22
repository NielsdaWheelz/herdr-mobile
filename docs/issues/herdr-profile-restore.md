# herdr account profiles and cold restore

problem: the inspected herdr v0.9.1 cold-restore path does not retain per-pane
environment overrides. skid's account profiles rely on distinct provider homes.

impact: successful initial launch does not establish account-correct automatic
resume. the migration must not silently resume a worker under another profile
or claim that its restored account selection is proven.

evidence: at release commit `065ef9d6a531c49fb8bee7e818ef837065b21ee9`,
[saved pane fields](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/persist/snapshot.rs#L98)
omit environment, and
[restore](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/persist/restore.rs#L520)
constructs empty extra environment. this is source evidence; live profile/restore
acceptance remains `NOT_RUN`.

pr 1's isolated darwin binary proof passed one synthetic exact-wrapper launch:
`workspace.create` supplied one root pane with chosen cwd/environment, and
`pane.send_input` submitted a shell-quoted `exec` of an absolute wrapper. the
wrapper compared argv, environment and physical cwd internally. this proves a
candidate public launch route without an extra shell, not the four configured
provider rows or cold-restart behavior. `agent.start` chooses herdr's canonical
executable and is not equivalent to skid's configured wrapper.

an isolated cold restart with resume disabled retained a pane label but cleared
synthetic metadata tokens and allocated a new terminal id. no provider was
running in that probe, so it does not establish absence of unintended provider
relaunch or correct behavior for any of the four real profiles.

resolution owner: [migration pr 1](../herdr-migration.md#pr-1-feasibility-and-implementation-contract).
use isolated test workers to verify all four configured profiles, cwd and launch
flags. restart only the test-owned server and inspect the resulting behavior
without logging account data. record exactly what was preserved or replaced.

resolved when: the accepted initial contract disables automatic provider resume,
and a live proof shows correct explicit launches plus restart without unintended
provider relaunch. alternatively, qualify an upstream path that preserves the
correct profile. adopting the explicit no-resume contract resolves this migration
issue without requiring an upstream restore feature.

2026-09-22 isolated v0.9.1 darwin arm64 and linux x86_64 proof: all four
actual configured command rows were launched through test-owned servers. a
content-free verifier compared absolute executable, argv, provider-home
environment and physical cwd before executing each; shell-only and zero-profile
cases passed with one root pane. disabling automatic resume then cold-restarting
only those servers removed the old terminal ids and relaunched no provider. the
user accepted the resulting new shell lifetimes without prior launch profile,
objective or dwarf metadata. interactive readiness remains `NOT_RUN`: configured
providers showed startup/update/form blockers or default-idle fallback within
ten seconds; no account/configuration was changed. the sole content-free claude
`SessionStart` runtime-profile identity registration is unproved on herdr.
v0.9.1's `pane.process_info.tty` is always absent; inherited `HERDR_PANE_ID`
and foreground-process identity offer a source-supported route. keep this issue
open until that retained capability and interactive readiness are qualified, or
their loss receives separate explicit acceptance.
