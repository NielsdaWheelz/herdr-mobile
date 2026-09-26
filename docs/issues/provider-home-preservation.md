# provider preservation correction awaits commit and live qualification

problem: dev-server's working tree now restores ordinary/herdr providers, but
that correction is uncommitted. its local commit `94931a1` still contains the
withdrawn global router and private herdr-home plan. it is not an acceptable
deployment candidate by itself. git write access is now restored; the owner
is finalizing the corrected revision.

impact: deploying the older commit would hide existing provider configuration,
authentication, histories and memories behind fresh homes. keeping the files
on disk would not preserve normal command access to them.

source correction (2026-09-25): read-only inspection confirms the restored
`ai-profile`, baseline account launcher installation, normal herdr integration
targets/service defaults, gate and mobile profiles. herdr no longer calls
private-home provisioning. original-skid isolation remains scoped. the
owner's `docs/gateway-separation-validation.md` records eight disposable shell
paths, state-preservation fixtures and static checks; those were not rerun by
the root operator. no live apply or provider-state change is reported.

herdr-mobile's published `v0.9.0` already accepts the existing-home profile
config. jarvis's unnecessary remap is reverted at `e6a6d20`, pr 42 is closed,
and its source/spec match the baseline; local `scripts/verify` passed.

remaining: finalize and push only the corrected deployment slice, preserving
unrelated edits. qualify actual
provider/history continuity and cross-runtime hooks at home and a shared
project under the normal cutover rules; source/fixture success is not that
live evidence. do not relocate homes or replace user configuration to make
qualification pass.

resolved when the corrected deployment revision is committed and selected,
ordinary/herdr commands retain their existing homes and account/override
behavior, and actual login/config/history continuity is verified without
recording user content. gateway maintenance must not own provider state.
