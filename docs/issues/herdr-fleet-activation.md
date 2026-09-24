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
2026-09-23, round 2: jarvis activation runs the read-only `check-activation`
instead of the zero-pending rule; dev-server carries explicit deployment
identity, and the darwin lifecycle proofs ran on this mac through a disposable
deployment under native launchd (report section "2026-09-23 herdr pr 4 round 2").
2026-09-23, round 3: dev-server `8a794be` waits for supervisor teardown before
restarting a restored service and fixes three restore defects the proofs
surfaced (prior verified against the candidate's version, a failed first
activation disabling the label, `UP TO DATE` for a herdr launchd had parked);
the herdr restore, its genuine-prior-failure path and the skid restores ran
again on this mac under a disposable deployment; jarvis `23c830c` ran the real
`activate-release` one-shot, `serve` startup, recovery over compatible rows and
a clean stop in an isolated systemd container with no egress (report section
"round 3"). release-bytes qualification waits for the draft.
2026-09-24, round 4: dev-server `f12f1a7` makes the confirmed stop precede
every restore (an unstoppable candidate keeps its inputs; the prior's backups
are retained as `.apply.failed.*`), proven through both skid restore callers;
jarvis drops the redundant `reset-failed` from `activate-release`, proven on a
never-loaded unit, a clean reactivation and a refused unclean stop (report
section "round 4").

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
| qualified artifact digests | the draft's five `SHA256SUMS` lines, copied here when step 2 finishes on both platforms; publication (`gh release edit --draft=false`) changes no asset, and step 3 diffs `gh release download v0.7.0 --pattern SHA256SUMS` against this row before `check published-release`. empty until then |
| jarvis candidate | the merge of `herdr-pr4` into main, made before qualification: jarvis keys releases by commit, so the commit installed on the isolated host must be the one production activates |
| jarvis rollback | none through `activate-release`: production `f4e2ce6129c0add09ddb50355a8997a1c589d3d1` predates `check-activation` and is refused, not skipped. if the candidate fails, keep jarvis stopped and repair forward. from the candidate on, rollback targets carry the check. transitional `51f62c86322a66224d1576395b5795ae823c1f75` for the admission-journal cutover |

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
   - darwin: no nonproduction host. lifecycle evidence ran 2026-09-23 on this
     mac through the disposable explicit-identity deployment (root
     `/private/tmp/skq`, labels `dev.niels.skq.*`, port 7351; recipe in
     dev-server `SPEC.md`), same installer, units and unmodified binaries,
     stand-in workers, no ingress: supervision, gateway restart preserving
     worker, terminal ref and agent ref, cold herdr restart invalidating old
     refs with no resume, failed activation restoring the prior installation,
     truthful failure when the restore also fails, herdr restore (round 3: the
     stop waits for launchd's teardown, the prior herdr is restored and
     restarted, and a prior that cannot verify is reported as such) and the
     changed-inputs `ACTION` all PASS on a LOCAL build of `2a6bfaf`. that
     qualifies the installer and units, not the release. release-bytes
     qualification is pending and repeats the decisive boundaries with the
     draft's own artifact: `gh release download v0.7.0 --pattern SHA256SUMS
     --pattern skidbladnir-darwin-arm64.tar.gz`, `shasum -a 256 -c` on the
     tarball line, seed the disposable root's `artifacts/<darwin sha>/` as the
     linux bullet describes, apply (`STARTED  skid.runtime: v0.7.0` with
     `skidbladnir version` reading the merge commit), second apply
     `UP TO DATE`, gateway restart preserving the worker and refs, cold herdr
     restart invalidating old refs, failed activation restoring the prior,
     truthful failure when the restore fails; then copy the five `SHA256SUMS`
     lines into the release tuple. production was not touched (pids and
     installed plist unchanged) and must stay so.
   - the unmodified codex + claude journey through installed jarvis (spec §7
     row), provider spend approved; jarvis stopped recovery of a paid decision
     and dispatched read across activation; admission charges and history
     before and after activation.
   - the emergency-apk roundtrip on isolated android resources: signed
     candidate in place, then v0.7.1 in place, pairings intact both ways.
   record each row in the jarvis report. if candidate bytes change, delete the
   draft, rerun `scripts/release` and only the affected boundary.
3. publish: `gh release download v0.7.0 --pattern SHA256SUMS` must equal the
   tuple's qualified digests (the qualified bytes are the published bytes);
   `gh release edit v0.7.0 --draft=false --latest`;
   `scripts/check published-release v0.7.0 <sha>`; skid pin commit; dev-server
   pr out of draft and merged, with `git diff <candidate tip> main` empty.
4. before any host: on macbook move `~/.config/herdr/session.json` aside (pr 1
   proof residue); on devbox check `~/.config/herdr/sessions/`; the installer
   refuses both with an `ACTION`, never deletes them.
5. stage artifacts on every host under the installer lock with the standalone
   `herdr_prepare_artifact` and `skidbladnir_prepare_artifact` calls documented
   in dev-server (the recipe copies and renders the assets first); nothing
   switches.
6. jarvis: run the runbook's read-only inventory, resolve pending work, owner
   `pause`, stop, admission-journal cutover, `deploy/install-release`.
7. window: `./devbox apply` (jarvis stopped), `./workstation apply` on arch and
   macbook; expect `CHANGED  skid.unit` once on the macbook (the explicit `HOME`
   line) with bootout and bootstrap rather than kickstart; `scripts/fleet
   verify` from the macbook; do not run bare `herdr` between a stop and a
   reapply. smoke checks, not qualification: `launchctl print` shows `HOME` in
   the gateway's environment and a new pid; the installed plist equals the
   default rendering; `/v1/pressure` answers on 7341; `tailscale serve status`
   still reads the `/v1` mapping (ingress now runs after retention from
   `workstation`); `dev.niels.herdr` runs with ping `0.9.1`/`22` and
   `~/.config/herdr/session.json` under the real home; a second apply is
   `UP TO DATE`.
8. `scripts/install-android` in place; check pairing continuity only. the
   waived terminal, lifecycle and manual phone journeys stay `NOT_RUN` and
   gate nothing.
9. `deploy/activate-release <commit>` (its `check-activation` refuses in-flight
   or incompatible unfinished work and leaves compatible rows untouched; no
   override), `deploy/verify-containment`, owner `resume`.
10. after acceptance: remove the inventoried legacy assets per dev-server
    `docs/issues/skid-legacy-asset-retirement.md`; reconcile this record and the
    migration plan; delete this file.

resolved when: all three hosts advertise the pinned release and herdr; installed
jarvis reads each peer and controls approved test workers, and one unavailable
peer does not block the others; the signed apk is installed in place with
pairings intact; rollback, including its failure handling, was exercised on
isolated resources (darwin done 2026-09-23, linux pending); and every row of
the jarvis evidence table is `PASS` or a recorded waiver, with no `NOT_RUN` row
that the spec requires.
