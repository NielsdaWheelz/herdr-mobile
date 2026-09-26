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

coordination evidence: original skid's handoff and fleet verifier now agree on
the ten-file receipt. dev-server's adoption notes are stale, and its generation
admission still omits the specified directory mode and computed digest-suffix
checks; that owner must resolve the original repo's
`docs/issues/restoration-receipt-validation.md`. the corrected deployment
handoff derives the host-only pin from the upstream five-asset pin instead of
copying its incompatible schema. publication must not wait on live namespace
handback; doing so would invert the cutover dependencies.

known blockers to claiming completion: original publication, live deployment,
correction of dev-server's provider routing, original-skid setup and fleet/phone
acceptance remain pending. device/tmux checks
need the applicable current-turn authorization. acceptance is `NOT_RUN`;
isolated probes and published artifacts do not prove deployed behavior.

resolved when: every acceptance row in the spec passes for both products,
on all applicable hosts and the phone, and neither ordinary installation nor
maintenance depends on the other's owned state. delete this issue then.
