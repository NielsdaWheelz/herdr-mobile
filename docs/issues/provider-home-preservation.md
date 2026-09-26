# ordinary/herdr provider preservation awaits live qualification

problem: dev-server `1296309` restores ordinary/herdr providers, but actual
provider continuity remains unqualified. the older `94931a1` contains the
withdrawn global router/private herdr-home plan and must not be deployed alone.

impact: deploying the older commit would hide existing provider configuration,
authentication, histories and memories behind fresh homes. keeping the files
on disk would not preserve normal command access to them.

source correction (2026-09-25): read-only inspection confirms the restored
`ai-profile`, baseline account launcher installation, normal herdr integration
targets/service defaults, gate and mobile profiles. herdr no longer calls
private-home provisioning. original skid's later account correction is tracked
[separately](skid-provider-continuity.md). the
owner's `docs/gateway-separation-validation.md` records eight disposable shell
paths, state-preservation fixtures and static checks; those were not rerun by
the root operator. no live apply or provider-state change is reported.

herdr-mobile's published `v0.9.0` already accepts the existing-home profile
config. jarvis's unnecessary remap is reverted at `e6a6d20`, pr 42 is closed,
and its source/spec match the baseline; local `scripts/verify` passed.

remaining: qualify actual provider/history continuity and cross-runtime hooks at home and a shared
project under the normal cutover rules; source/fixture success is not that
live evidence. do not relocate homes or replace user configuration to make
qualification pass.

resolved when the corrected deployment revision is committed and selected,
ordinary/herdr commands retain their existing homes and account/override
behavior, and actual login/config/history continuity is verified without
recording user content. gateway maintenance must not own provider state.
