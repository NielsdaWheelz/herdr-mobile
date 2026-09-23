# herdr fleet activation

problem: [pr 4](../herdr-pr4.md)'s source prerequisites exist on the `herdr-pr4`
branches of skid, jarvis and dev-server, but nothing is published, installed,
qualified on a nonproduction host or activated. installed hosts still run the
tmux-era v0.6.0 gateway; jarvis production runs the pre-pr 3 release.

impact: the merged herdr candidate is not in use anywhere. the phone, cli and
jarvis keep addressing tmux sessions; the pr 3 consumer stays undeployed.

evidence (2026-09-23): skid `herdr-pr4` adds truthful stop partials,
`skidbladnir validate-host-config` and a herdr probe in `scripts/fleet verify`;
dev-server `herdr-pr4` adds `lib/herdr.sh`, `assets/herdr/*` and the herdr-era
host configs; jarvis `herdr-pr4` adds stop-gated activation and the
"herdr cutover (pr 4)" section of its `docs/operations.md` (on that branch until
it merges after skid).
boundary proofs ran on isolated resources and were deleted; every live row is
`NOT_RUN` in jarvis `docs/qualification/2026-09-23-herdr-pr4.md`.

## release tuple

| item | value |
| --- | --- |
| herdr | v0.9.1, source `065ef9d6a531c49fb8bee7e818ef837065b21ee9`; linux `2a02fed16beb651ef006e1d43f048f652ca4dc58ad053cd2d44450563d5c54b7`, darwin `5fc7a7e7adfaca56fa80aa89dcb025693357268dab8285b9ce2d08a2313c89de` |
| skid candidate | `v0.7.0` from the merge of `herdr-pr4` into main; android version code 7000; publish through `scripts/release` then `gh release edit --draft=false` |
| emergency apk | `v0.7.1` (code 7001) built locally with `scripts/build release-assets` from a clean `2d6184c63d62396f69342200e4229cc902ca140c` checkout; same package and signer; never published; `scripts/check-release --source` verifies it |
| next forward | `v0.7.2` (7002) if the emergency apk is ever installed |
| dev-server | pins `v0.7.0` in `assets/skidbladnir/release-pin.json` after publication; rollback = the last pre-pin commit on main plus apply |
| jarvis | the merge of `herdr-pr4` into main; production `f4e2ce6…`; transitional `51f62c86…` for the admission-journal cutover |

## remaining steps, each under its own approval

1. merge skid `herdr-pr4`; wait for main `verify`; `scripts/release v0.7.0`;
   publish; `scripts/check published-release v0.7.0 <sha>`; pin in both repos.
2. merge dev-server and jarvis `herdr-pr4` (jarvis after skid).
3. before any host: on macbook move `~/.config/herdr/session.json` aside (pr 1
   proof residue); on devbox check `~/.config/herdr/sessions/`; the installer
   refuses both with an `ACTION`, never deletes them.
4. stage artifacts on every host under the installer lock with the standalone
   `herdr_prepare_artifact` and `skidbladnir_prepare_artifact` calls documented
   in dev-server; nothing switches.
5. qualify the unmodified candidate on a nonproduction linux host with a trusted
   `https://HOST:8443`, isolated gateway/herdr and test-owned workers; run the
   jarvis codex + claude journey there. provider spend needs approval.
6. jarvis: run the runbook's read-only inventory, resolve pending work, owner
   `pause`, stop, admission-journal cutover, `deploy/install-release`.
7. window: `./devbox apply` (jarvis stopped), `./workstation apply` on arch and
   macbook; `scripts/fleet verify` from the macbook; do not run bare `herdr`
   between a stop and a reapply.
8. `scripts/install-android` in place; check pairing continuity, not the waived
   terminal journeys.
9. `deploy/activate-release <commit>` (never `--accept-pending-actions` here),
   `deploy/verify-containment`, owner `resume`.
10. after acceptance: remove the inventoried legacy assets per dev-server
    `docs/issues/skid-legacy-asset-retirement.md`; reconcile this record and the
    migration plan; delete this file.

resolved when: all three hosts advertise the pinned release and herdr, the
signed phone and native desktop share test workers, jarvis completes the
journey through the installed cli, one unavailable peer does not block the
others, rollback was exercised on isolated resources, and the evidence table in
the jarvis report holds no pending row.
