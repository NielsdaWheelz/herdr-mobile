# codex menus read as ready

problem: herdr v0.9.1's bundled codex detection reports `idle ready` while
codex shows a blocking menu: the directory-trust prompt on first start in a
directory, then the hook-trust prompt for skid's installed identity hooks.

impact: skid and jarvis treat the worker as ready, so an ordinary send is
accepted and its text lands in the menu, where a digit or enter can answer the
prompt. claude's equivalent prompt is detected as `blocked` and refused
correctly. after the window, the first codex start per profile and host shows
the hook-trust prompt until someone answers it.

evidence (2026-09-24, isolated linux qualification, codex-cli 0.156.1 profile
`personal`, herdr 0.9.1): after `start`, `info` read `idle ready` while the
visible read showed a two-option directory-trust menu; after it was answered
by keys, a second hook-trust menu also read `idle ready`. jarvis row "live
codex + claude journey" in its `docs/qualification/2026-09-23-herdr-pr4.md`.

mitigation: after each host's window apply and before jarvis `resume`, start
one codex agent per profile through skid, inspect it, answer the trust prompts
deliberately, and stop it.

resolved when: codex's trust menus read as blocked or unknown in herdr (an
upstream detection-manifest change or a skid-side rule), shown by an isolated
start in an untrusted directory refusing an ordinary send.
