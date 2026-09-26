# separated source awaits fleet and phone cutover

problem: the source and github identities are separated, but the installed
fleet and phone still require the coordinated transition from the herdr-backed
skid installation to two independent products.

impact: premature original-skid activation can replace the old herdr gateway
or reuse its credentials/rollback generations. provider homes, project hook
discovery, native helper behavior and phone coexistence are not yet qualified.

baseline evidence (2026-09-25): source inspection at herdr-backed `d8bb9c4`,
original `927c554`, dev-server `8498933`, jarvis `c739f2d`; read-only host/phone
observations are in the [spec](../herdr-mobile-separation.md#2-investigated-baseline).
the observed phone held herdr-backed `0.8.0` under `dev.niels.skidbladnir`.

completed source/release work:

- this repo's separate command/module, android package/signer, paths, ports,
  headers, invitation kind, metadata and tooling are merged at `68a652d`.
- immutable `v0.9.0` is published from that source; the exact five public asset
  digests are in [`release-pin.json`](../../release-pin.json).
- both github names now resolve to the intended repository ids; known clone
  remotes on all three hosts were retargeted and immutable releases enabled.
- engineering checks, signed artifact checks, isolated native gateway/metadata
  probes and compiled android origin/invitation probes passed. the
  [deployment contract](../dev-server-handoff.md) records their limits.
- the unnecessary jarvis worker-home remap was reverted at `e6a6d20` and
  [pr 42](https://github.com/NielsdaWheelz/jarvis/pull/42) closed. existing
  worker homes/spec are restored; no live provider state changed.

remaining work: original skid's release/pin; reviewed dev-server deployment;
the [provider-home correction](provider-home-preservation.md), existing
provider continuity and hook-discovery qualification; original-skid login/trust
and native-helper live compatibility; jarvis action settlement; `8444` reachability;
old host generation retirement and namespace handback; phone reset/re-pairing;
concurrent operation and independent restart/reinstall/rollback/removal.
the new android signer has a same-host backup; the owner still needs an
off-machine copy for disk-loss recovery.

coordination evidence: original skid's handoff and fleet verifier agree on
the ten-file receipt. dev-server committed admission mode/digest-suffix checks,
contract acknowledgment and the real mobile pin in `94931a1`. its later
provider-routing correction is still uncommitted; that older commit alone must
not be deployed. read-only source review confirms the correction, and the
owner records disposable/static results in its validation document.

original skid is at `4eeba152bc4e37cc36eaa9a367c27efcb47aba20`, reported as
hosted-verified in run `36204088437`. six corrected documentation files remain
uncommitted; the final release source needs its own exact hosted check. current
local signing preflight passed according to its handoff; a fresh phone
version/signer observation remains required before finalizing `v0.9.0`/`9000`.

git write and network access are restored. each owner is finalizing its
reviewed separation changes in the existing checkout. record the final commits
and obtain the exact original-source hosted check before publication. the
user has relayed both handoffs; connector messaging is not a prerequisite.
an old checked commit does not cover later corrections.

original publication, original-skid setup, host namespace handback and live
fleet/phone acceptance remain pending. device/tmux checks need their applicable
current-turn authorization. acceptance is `NOT_RUN`; artifacts and disposable
fixtures do not prove deployed behavior.

resolved when: every acceptance row in the spec passes for both products,
on all applicable hosts and the phone, and neither ordinary installation nor
maintenance depends on the other's owned state. delete this issue then.
