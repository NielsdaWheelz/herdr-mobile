# provider relocation must be withdrawn before apply

problem: the separation's dev-server working tree makes global `~/bin`
provider commands select new `.local/share/herdr/providers/` homes. herdr
services, integrations, gate and mobile profiles select the same new homes.
`provider_homes_prepare` creates instructions/settings, not the existing user
configuration, authentication, histories, memories or project trust.

impact: an apply would strand normal command access to existing provider state
despite leaving its files on disk. the owner explicitly rejected this design:
herdr must use the normal existing providers and homes. disposable herdr panes
do not authorize provider relocation. do not apply the conflicting source.

evidence (2026-09-25): dev-server working-tree `assets/routers/ai-profile`,
`ai_install_profiles` in `lib/ai-tools.sh`, `lib/provider-homes.sh`,
`assets/herdr-mobile/host-config.json`, `lib/herdr.sh`, herdr service assets
and `assets/herdr/herdr-gate` implement the rejected mapping. its committed
`8498933` baseline uses the existing account homes. pinned native herdr
`065ef9d` resolves the normal home or provider environment override; it has
no requirement for new homes. its hooks guard on herdr runtime context.

correction owned here: the spec and handoffs now retain `.codex`, `.codex-work`,
`.codex-work2`, `.claude`, `.claude-work`, existing binaries and ordinary
command semantics. jarvis's unshipped remap was reverted at `e6a6d20`; pr 42
is closed. no account-home migration or provider-state repair was performed.
the restored jarvis source/spec match the baseline and `scripts/verify` passed.
the digest-verified published herdr-mobile `v0.9.0` binary accepted the four
existing-home profile paths through `validate-host-config`; no provider ran.

dev-server owns the remaining source correction: restore its existing global
commands/account declaration, integration targets, service environment and
gate; correct the mobile template; remove herdr private-home provisioning and
the proposed removal of its normal-home integrations. preserve unrelated
gateway separation and context scrubbing. do not delete any newly created
directory, copy account trees, rewrite user configuration or log user content.
original-skid isolation remains scoped to its own launches and terminals.

resolved when disposable routing checks prove ordinary/herdr commands retain
their previous homes and explicit account/environment behavior, the published
gateway accepts the corrected profile config, and no apply path relocates or
reinitializes normal provider state. real cross-runtime hook behavior at home
and a shared project remains a separate live acceptance boundary; possible
project-hook discovery does not prove that relocation is necessary. verify
existing login/config/history continuity without recording their contents.
