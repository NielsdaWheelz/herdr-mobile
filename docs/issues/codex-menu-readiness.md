# codex screens read as ready (herdr detection)

problem: herdr v0.9.1's bundled codex detection reports `idle` and ready for
codex screens that are not at a prompt: the directory-trust and hook-trust
menus of codex 0.156.1, and codex's "Sign in with ChatGPT" screen (`agent
start` returned `interactive_ready=true` there). skid and jarvis only project
herdr's `agent.explain` (skid never reads screen text), so the misread is
herdr's, in its codex detection rules.

impact: an ordinary send is accepted and its text lands in the menu or sign-in
screen, where a digit or enter can answer it; herdr's own `agent prompt` blocked
refusal does not catch these screens. the same codex menus under 0.155.1 read
correctly (`unconfirmed` for hook review, `blocked` for the update prompt), so
the rules lag codex's ui changes.

evidence (2026-09-24): the isolated linux qualification (codex 0.156.1,
profile `personal`): `info` read `idle ready` over a two-option directory-trust
menu and then a hook-trust menu (jarvis `docs/qualification/2026-09-23-herdr-pr4.md`,
"live codex + claude journey"). an isolated herdr on the macbook: `agent start
--kind codex` with an unauthenticated home returned `agent_started`,
`agent_status=idle`, `interactive_ready=true` on the sign-in screen
(`docs/herdr-pr5.md` §7).

resolved when: a herdr release reads these screens as blocked or unknown,
shown by an isolated `agent start` in an untrusted directory and one with an
unauthenticated home refusing `agent prompt`. filing upstream at herdrdev/herdr
is the owner's call.
