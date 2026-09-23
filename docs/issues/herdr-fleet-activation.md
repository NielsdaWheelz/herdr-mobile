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
| skid candidate | `v0.7.0` from the merge of `herdr-pr4` into main; android version code 7000; `scripts/release` builds the five assets and a github draft whose digests are final; `gh release edit --draft=false --latest` publishes only after isolated qualification |
| emergency apk | `v0.7.1` (code 7001) built locally with `scripts/build release-assets` from a clean `2d6184c63d62396f69342200e4229cc902ca140c` checkout; same package and signer; never published; `scripts/check-release --source` verifies it |
| next forward | `v0.7.2` (7002) if the emergency apk is ever installed |
| skid rollback | `v0.6.0`, source `2d6184c63d62396f69342200e4229cc902ca140c`, the pin at skid and dev-server main today and the release every host advertises through `readlink ~/.local/share/skidbladnir/current` |
| dev-server candidate | the `herdr-pr4` tip after its v0.7.0 pin commit. the installer's behavior is its tree, not its commit, so the later merge into main must leave the main tree identical to that tip (`git diff <tip> main` empty) |
| dev-server rollback | commit `bb8e218bbfb10a9077aadce2983d510c8366a2f9`: origin/main, the `herdr-pr4` merge-base, and the library the deleted rollback proof re-applied over herdr-era state (darwin only; the linux rollback runs in isolated qualification). rollback is that checkout plus `./workstation apply` or `./devbox apply`; it reinstalls v0.6.0 with the retained native-control copies and leaves herdr supervised. the macbook checkout is at it; confirm devbox's and arch's last applied checkout read the same commit (read-only) before the window |
| jarvis candidate | the merge of `herdr-pr4` into main, made before qualification: jarvis keys releases by commit, so the commit installed on the isolated host must be the one production activates |
| jarvis rollback | production `f4e2ce6129c0add09ddb50355a8997a1c589d3d1` with its own cli and unit, allowed only before new receipts or positions exist (spec §5) and only once qualified against the normalized admission journal (adr 0047, runbook rollback section); that qualification is not recorded, so until it is, keep jarvis stopped and repair forward. transitional `51f62c86322a66224d1576395b5795ae823c1f75` for the admission-journal cutover |

## remaining steps, each under its own approval

merges freeze the candidate commits and come first. qualification precedes
publication (spec §4.2 then §4.3); publication precedes the skid pin, the
dev-server merge, staging and the window. waived phone journeys never return as
gates.

1. merge skid `herdr-pr4`; wait for main `verify`; merge jarvis `herdr-pr4`
   (its candidate commit is now frozen); `scripts/release v0.7.0` creates the
   draft. do not publish. commit the dev-server pin on its `herdr-pr4` branch
   from `SHA256SUMS`; that tip is the dev-server candidate.
2. isolated qualification, all on nonproduction resources:
   - linux host with a trusted `https://HOST:8443`, isolated gateway, herdr,
     jarvis postgres and runtime state, test-owned workers, dev-server at the
     candidate tip. the draft is not publicly fetchable, so
     `gh release download v0.7.0` on the macbook, check the linux tarball
     against `SHA256SUMS` by hand, and seed the host's
     `~/.local/share/skidbladnir/artifacts/<linux sha>/` with the three
     extracted members at modes 755/644/644 plus `identity.sha256` (0600) from
     `skidbladnir_payload_hashes | dev_server_sha256_stream`. the installer
     trusts a seeded directory and never downloads there; the fetch path is
     proven only by `check published-release` and the fleet staging step.
   - real systemd supervision of herdr and the gateway; gateway restart
     preserving an exact worker and ref; cold herdr restart invalidating old
     refs without provider resume; failed gateway activation restoring the
     prior bytes and unit with the worker alive, and the restore path where the
     prior also fails to start; the codex completion bell inside a herdr pane.
   - darwin has no isolated host: the macbook runs the production gateway. its
     launchd supervision is first observed in the window, macbook before the
     linux hosts, with the dev-server rollback checkout ready.
   - the unmodified codex + claude journey through installed jarvis (spec §7
     row), provider spend approved; jarvis stopped recovery of a paid decision
     and dispatched read across activation; admission charges and history
     before and after activation.
   - the emergency-apk roundtrip on isolated android resources: signed
     candidate in place, then v0.7.1 in place, pairings intact both ways.
   record each row in the jarvis report. if candidate bytes change, delete the
   draft, rerun `scripts/release` and only the affected boundary.
3. publish: `gh release edit v0.7.0 --draft=false --latest`;
   `scripts/check published-release v0.7.0 <sha>`; skid pin commit; dev-server
   pr out of draft and merged, with `git diff <candidate tip> main` empty.
4. before any host: on macbook move `~/.config/herdr/session.json` aside (pr 1
   proof residue); on devbox check `~/.config/herdr/sessions/`; the installer
   refuses both with an `ACTION`, never deletes them.
5. stage artifacts on every host under the installer lock with the standalone
   `herdr_prepare_artifact` and `skidbladnir_prepare_artifact` calls documented
   in dev-server; nothing switches.
6. jarvis: run the runbook's read-only inventory, resolve pending work, owner
   `pause`, stop, admission-journal cutover, `deploy/install-release`.
7. window: `./workstation apply` on the macbook first (darwin supervision is
   unobserved until then), then on arch, then `./devbox apply` (jarvis
   stopped); `scripts/fleet verify` from the macbook; do not run bare `herdr`
   between a stop and a reapply.
8. `scripts/install-android` in place; check pairing continuity only. the
   waived terminal, lifecycle and manual phone journeys stay `NOT_RUN` and
   gate nothing.
9. `deploy/activate-release <commit>` (it refuses any pending action; there is
   no override), `deploy/verify-containment`, owner `resume`.
10. after acceptance: remove the inventoried legacy assets per dev-server
    `docs/issues/skid-legacy-asset-retirement.md`; reconcile this record and the
    migration plan; delete this file.

resolved when: all three hosts advertise the pinned release and herdr; installed
jarvis reads each peer and controls approved test workers, and one unavailable
peer does not block the others; the signed apk is installed in place with
pairings intact; rollback, including its failure handling, was exercised on
isolated linux resources; and every row of the jarvis evidence table is `PASS`
or a recorded waiver, with no `NOT_RUN` row that the spec requires.
