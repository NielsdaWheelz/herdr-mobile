# android tolerates an unemitted launch_submitted start stage

problem: `GatewayClient.kt` accepts and renders a `launch_submitted` create-stage
partial, but the merged host emits only `resource_created` and `identified`, and
[pr 4](../herdr-pr4.md#exact-start-contract) removes the stage from the pr 1 wire
contract.

impact: dead tolerance only. no host output reaches that branch, so no behavior
changes; the phone would render a stage the contract no longer names if a host
ever emitted it.

evidence: `android/app/src/main/java/dev/niels/skidbladnir/GatewayClient.kt`
(message branch and accepted stage set) at skid `86d7dcf`; no `launch_submitted`
under `internal/` or `cmd/`. pr 4 §6 plans no android production change, so the
cut was deliberately left out of that pr.

resolved when: a later android change removes the branch and the accepted value,
and the phone build passes engineering checks. no device gate is needed for a
dead-branch removal; do not bundle it into a coordinated fleet activation.
