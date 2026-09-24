# herdr fleet activation

problem: the [pr 4](../herdr-pr4.md) cutover ran on 2026-09-24, but acceptance
is incomplete: the legacy assets and the workstation codex servers are still
installed, and three required evidence rows never ran.

impact: rollback inputs stay on disk until retirement; the unrun rows mean the
android data roundtrip, jarvis's paid-decision recovery across activation and
the live journey on the production fleet are observed in use, not qualified.

evidence (2026-09-24): skid #135 and jarvis #40 merged (`cf25c7d`, `2a59355`);
the v0.7.0 draft qualified in isolation (jarvis
`docs/qualification/2026-09-23-herdr-pr4.md`, "release qualification") and was
published with unchanged digests; `scripts/check published-release` passed and
skid pinned it (`8f7f795`); dev-server #119 merged with a tree equal to the
qualified tip (`b453ccb`). the window: pr 1 residue moved aside on macbook and
devbox (`~/.local/state/herdr-pr1-residue-2026-09-24/`, including the macbook's
remote agent-detection manifests the preflight named); artifacts staged on all
three hosts; jarvis inventory zero, owner `pause`, clean stop (130), admission
journal current through `51f62c86`, `2a59355` installed. `./devbox apply`
switched devbox. the macbook's `./workstation apply` failed at its codex step
(codex 0.156.1 publishes the app-server socket as a symlink), so dev-server #120
(`737c199`) retired the workstation codex servers and pinned codex 0.155.1 on
devbox, changing no herdr or skid installer file; the devbox reapplied with
`--restart-codex` while jarvis was stopped. macbook and arch then switched:
herdr `0.9.1/22` supervised, `RESTARTED  skid.runtime: v0.7.0`, one
`CHANGED  skid.unit` on the macbook with `HOME` in the gateway environment, the
tailscale `/v1` mapping intact, a second macbook apply `UP TO DATE`; arch
defers a reboot it already owed. `scripts/fleet verify` passed for all three.
the phone updated in place (code 6000 → 7000, first install time unchanged) and
shows all three machines without reenrollment. codex trust: one agent per
profile and host through skid, no prompt sent; devbox work and work2 trusted
skid's changed identity hook, devbox personal skipped codex's update prompt
(`Update now` preselected; herdr read it `blocked`). jarvis activation:
`compatible=0 incompatible=0 in_flight=0`, started paused, `NRestarts=0`,
discord connected, `verify-containment` passed, activation clock
`2026-09-24T19:46:33.581387+00:00`, owner `resume`.
the phone and jarvis now see herdr terminals only; tmux sessions drain through
tmux (no pty transfer).

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
| qualified artifact digests | draft `v0.7.0` at `cf25c7df2442d18cc6f6d030a2402b14bc9378ef`, qualified 2026-09-24: `e25f20e491da5d659333a524bd8f33f3572d9851b2b4fab0cd251bcbd27fe0b5  android-signing-cert.sha256`, `00559b775ce4989a04130de77cbffd4a1347564653c1d1245e186f8fbb9cc61f  skidbladnir-android.apk`, `20b5aaa055730169fb381fc53fce513efb91810148f47b3ffca75663cf99534f  skidbladnir-darwin-arm64.tar.gz`, `0dcbe67be21f690a021823411b9e07f2c4d0cd399419606cddf1290cdb6666ca  skidbladnir-linux-amd64.tar.gz`. dev-server candidate tip `4827bba958196d64f28c840553941f102f8c08ef`; jarvis candidate `2a59355fcff5bba1bee556fd877b8dc02dd37673`; emergency apk `0436a0469fd9827dff2e5730faf9b1508a2231efe4c49d0a6d5dc655568839fa` (local) |
| jarvis candidate | the merge of `herdr-pr4` into main, made before qualification: jarvis keys releases by commit, so the commit installed on the isolated host must be the one production activates |
| jarvis rollback | none through `activate-release`: production `f4e2ce6129c0add09ddb50355a8997a1c589d3d1` predates `check-activation` and is refused, not skipped. if the candidate fails, keep jarvis stopped and repair forward. from the candidate on, rollback targets carry the check. transitional `51f62c86322a66224d1576395b5795ae823c1f75` for the admission-journal cutover |

## remaining

1. retire the legacy assets per dev-server
   `docs/issues/skid-legacy-asset-retirement.md` once the owner accepts the
   cutover.
2. drain the workstation codex servers per dev-server
   `docs/issues/workstation-codex-shared-retirement.md` once no interactive codex
   is attached to them.
3. rollback note: `bb8e218`'s `./workstation apply` now fails at its codex step
   on macbook (codex 0.156.1), so a workstation skid rollback needs a dev-server
   commit that pins v0.6.0 without the workstation codex services, or the
   library entry the qualification used. devbox rollback is unaffected.
4. unrun rows, owner decision each: the android emergency-apk roundtrip (needs
   an sdk 36 image and an invite naming the three real machines, so not
   isolatable as specified); jarvis paid-decision and dispatched-read recovery
   across activation (needed discord ingress); findings
   [codex menus read as ready](codex-menu-readiness.md), jarvis
   `docs/issues/discord-login-failure-exit.md`, dev-server
   `docs/issues/codex-daemon-socket.md`.

resolved when: the legacy assets and workstation codex servers are removed,
the record and migration plan are reconciled, and each unrun row is run or
waived by the owner; then delete this file.
