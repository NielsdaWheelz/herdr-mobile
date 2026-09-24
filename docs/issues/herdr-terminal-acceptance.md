# herdr phone and desktop terminal acceptance

2026-09-24 herdr pr 5: the skid desktop client is retired; the desktop is
herdr's own client, and the desktop and cli checks below are history. the phone
journey remains.

status 2026-09-23: the remaining phone journey and mac-local cli/private-ca
checks are [accepted non-blocking follow-ups](../herdr-pr2.md#accepted-non-blocking-follow-ups).
no new test campaign is required; unrun checks remain `NOT_RUN`. the resolution
criteria below describe future verification, not a merge or cutover gate.

problem: herdr's rendered-frame bridge has not been exercised through skid's
complete retained android input/scroll/lifecycle journey. its direct controller
owns sizing while desktop input can still reach the same terminal.

impact: raw-byte assumptions, scroll routing, automatic terminal replies,
phone keyboard resizing and desktop handback may need different ownership.
upstream support alone does not prove the existing phone experience works.

accepted 2026-09-22: [the scope amendment](../herdr-pr1.md#accepted-scope-amendment)
attaches one exact provider terminal, not herdr's workspace ui. remote application
clicks/drags are excluded. keys, ime/dictation, paste, public swipe scrolling,
phone-local input focus and selection/copy remain required. missing general
mouse input is no longer an adoption blocker.

evidence: the documented
[terminal bridge](https://herdr.dev/docs/persistence-remote/#direct-terminal-attach)
offers observe/control streams and one writable direct controller. the inspected
v0.9.1 source separately routes ordinary desktop input. full skid/herdr phone
and host interaction acceptance remains open; bounded results follow.

pr 1 source review found a public input gap at this pin. `herdr terminal session
control` accepts `terminal.input`, `terminal.resize`, `terminal.scroll` and
`terminal.release`, but no tap or mouse command. its stdout publishes rendered
ansi frames and discards internal mode events. those frames do not carry the
provider's mouse-mode negotiation, so the existing xterm tap router cannot
reconstruct when or how to send mouse input. the native client uses an internal
`AttachMouse` message; the migration contract forbids implementing that private
protocol. this is a limitation of the now-excluded remote pointer interaction,
not public scrolling: upstream
[mode-aware wheel routing](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/server/pane_input.rs#L128)
can produce application mouse-wheel reports, alternate-scroll input or host
scrollback movement. actual provider scrolling still needs proof.
the bounded physical-phone proof on samsung sm-s906w/android 16
applied the first full frame before allowing input, delivered one key to the
same isolated shell as checked by an in-memory visible read, and routed a swipe
as six public `terminal.scroll` commands. it did not establish scrollback
movement, ime/dictation, paste, selection/copy, or mouse-aware tap behavior.
a second no-takeover phone connection received `terminal.closed` while a cli
controller remained; the phone's conflict message and explicit takeover were
not verified on device.
force-stopping only the proof app during an adb-reversed connection left the
test relay's websocket established until an explicit takeover disconnected
it. remote terminal survival was not in doubt; timely controller release and
geometry handback after abrupt loss remain unproved. this observation is tied
to the test relay and adb reverse, not a claim about a production bridge.

the public control stream also reports attach conflict inside a free-form
`terminal.closed.reason`, not a typed error. an exact v0.9.1 text match can
drive a disposable probe; the accepted product contract instead reports coarse
`control_unavailable` before the first frame and never treats human prose as
a stable code.
the stream reader accepts up to 32 mib, though its terminal attach producer
emits under a 2-mib cap. the retired tmux bridge's frame bound was 64 kib;
full-frame and queue limits need measurement and an explicit bounded bridge
decision before cutover.

an isolated v0.9.1 darwin control stream at the intended 1024×512 geometry
produced a full frame of 1,272,539 decoded ansi bytes for 16-cell truecolor
stripes and 2,016,731 bytes for 8-cell stripes. the latter public json line
was 2,689,077 bytes. both are valid emitted frames but exceed pr 1's former
1-mib decoded cap; the latter also exceeds its former 2-mib line and queue caps.
upstream skipped a denser per-cell-color render at its own 2,097,152-byte
normal-frame limit. pr 1 now proposes source-derived 2-mib decoded and
3-mib encoded bounds. the terminal attach producer does not use herdr's
separate 32-mib graphics allowance, despite the public control reader's
larger input cap. this is a measured failure of the old contract, not yet a
product pass:
largest-frame phone rendering, memory and attachment-only overflow closure
are `NOT_RUN`.

resolution owner: [migration pr 1](../herdr-migration.md#pr-1-feasibility-and-implementation-contract).
use test-owned codex, claude and shell terminals on isolated linux/darwin servers.
exercise the existing phone renderer with typing, synthetic multiline paste,
keys, alternate-screen history, touch scroll and keyboard resize. the user
waived dictation, gboard paste, local selection/copy and rotation; record them
as `NOT_RUN`, without inferring success. prove actual codex/claude history
movement, not just receipt of scroll commands. taps focus input without remote
clicks; selection stays local.
attach desktop concurrently, then detach/background/disconnect
the phone and verify sizing handback and continued worker execution.

resolved when: a bounded approved live/device journey proves the bridge is usable
and fixes input, reply, scroll, geometry, takeover and cleanup ownership in the
migration contract. pr 2 must repeat the important cases through the implemented
product. if basic operation requires rebuilding substantial terminal machinery,
record that negative result and reconsider adoption before pr 2.

next step: complete the retained phone journey at the pinned revision. do not
require a new mouse command. qualify typed stream outcomes and decide frame
admission bounds from measured full frames rather than silently inheriting the
upstream maximum. those open contract questions and release/geometry handback
are not resolved by accepting the narrower interaction scope.

2026-09-22 physical samsung sm-s906w/android 16 proof, separate
`dev.niels.skidbladnir.herdrproof` package and pinned darwin herdr v0.9.1:
one shell swipe moved host history; no-takeover acquisition gave a coarse
pre-frame error; explicit takeover displaced the test cli controller. public
`pane.send_input` and direct bracketed-paste recognition carried synthetic
multiline input exactly. clean detach returned geometry to a test desktop in
under 0.8 seconds, preserving the terminal/shell. without a heartbeat, abrupt
app loss left the adb-reversed tcp connection and phone geometry held; a
disposable two-second heartbeat with a six-second unanswered deadline restored
desktop geometry after about five seconds, preserving the worker. two later
home/background runs returned desktop geometry within one second, also without
ending the worker; one earlier home/background attempt kept phone geometry
after five seconds, so release is inconsistent and still open. adb typing
emitted nine input events but no exact in-memory host match, so typing is
unproved. user-skipped dictation, gboard paste, local
selection/copy and rotation remain `NOT_RUN`. real codex/claude phone use is
`NOT_RUN` because the isolated provider sessions were startup-blocked.
v0.9.1's public logical-key parser rejects home/end/page keys, including
modified forms. a physical key-deck probe against a test-owned
application-cursor-mode worker failed home/end while page-up/down passed;
raw fixed csi is therefore insufficient. this is a real red at the retained
input boundary. keep this issue open for the
key-deck, provider and linux stream boundaries; pr 2 must repeat release
through its actual websocket bridge.

a later official v0.9.1 physical-phone proof matched nine adb-injected
synthetic input bytes exactly once in a test-owned raw host reader, then failed
the same comparison on a fresh worker despite nine foreground input messages.
automated typing therefore remains open. background release remained red: its
`onStop` callback and ordered writer ran, but the
relay received no release and herdr kept phone geometry beyond seven seconds.
moving release to `onPause` in that temporary adapter delivered
`terminal.release` and returned geometry from phone 42×40 to desktop 28×53
within two seconds, with the worker alive. one green run identifies a plausible
lifecycle boundary but does not establish reliable product behavior; `onPause`
can also run for transient overlays. retain the background issue until the
actual phone controller passes repeated acquisition and steady-state cases.

## pr 2 input contract review

problem: the earlier stream key set described only the phone deck; it omitted
enter/backspace and desktop control/function/editing keys. `skid enter` currently
forwards raw tty chunks, which cannot become committed text without decoding.
the candidate now defines one shared semantic key vocabulary. upstream
[`parse_key_combo`](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/config/keybinds.rs#L1231)
also lacks insert/delete/page names at v0.9.1. extend the public mode-aware path
for a future full desktop key contract. the current candidate visibly disables
home/end on the phone, supports unmodified page visual navigation through
`terminal.scroll`, and rejects remaining unsupported hardware/desktop keys.

reuse candidate: the existing ultraviolet dependency offers a terminal input
reader with key/paste/reply events. its
[pinned source](https://github.com/charmbracelet/ultraviolet/blob/f5a850f9c2b7/terminal_reader.go#L334)
discards valid replacement characters in paste and accumulates paste without a
size limit. a temporary hermetic probe reproduced both defects and blocked
cancellation at that pin; current upstream source
`4e49372c11f9827d149a13ffaca652cec39466d9` still has them. a local
upstream patch passed focused, package and race checks on darwin, but no
published revision can be pinned. accepting the current reader would
contradict literal input and bounded buffering.

a separate temporary skid-side decoder probe composes the already pinned
`x/ansi.DecodeSequence` for escape framing, ultraviolet's public `EventDecoder`
for named keys/replies, and go's utf-8 decoder for literal text and paste. it
passed fragmented replacement-rune, multiline and detach-looking literal
paste, 32-kib overflow, one modified arrow, escape-timeout and unknown-sequence
checks. this avoids the defective `TerminalReader` and copies no parser.
the resulting scanner still needs a real tty and linux/darwin product proof;
unrecognized terminfo-only input must end the attachment explicitly. the probe
currently drops ultraviolet's ambiguous modified-f3/cursor-position
`MultiEvent` and appends a whole caller chunk before cap checks. product code
must explicitly handle or reject that ambiguity and feed bounded chunks, with
tests, before this decoder path qualifies. it also treats a key with both
`Text` and ctrl/alt modifiers as plain text; preserve modifiers when mapping
kitty keyboard events and test them. no valid key may disappear silently.

the [latest published herdr preview](https://github.com/herdrdev/herdr/releases/tag/preview-2026-09-21-0ff0f27e2226)
still rejects home, end, insert, delete and page keys through its public pane
input parser. a temporary upstream patch on that source passed parser and
mode-aware api tests, including application-cursor home/end, but has no
published release or linux/darwin/phone proof. herdr's
[contribution policy](https://github.com/herdrdev/herdr/blob/0ff0f27e222633c97ba4291f6b9be4137002ca84/CONTRIBUTING.md)
does not accept unsolicited implementation pull requests. local patches do not
qualify the key path. the user later accepted visible disabled controls and
explicit rejection of unsupported keys so the rest of pr 2 can be implemented
at v0.9.1. no herdr discussion or pull request was posted. a future
independently released public key operation needs mode-aware phone/desktop proof
before those controls are re-enabled.

resolved when: the skid-side decoder preserves fragmented utf-8 (including
`U+FFFD`), multiline/bracketed paste and detach-looking paste literally, rejects
oversize input without unbounded accumulation, and cancels cleanly. prove
the supported named key/modifier mapping through actual `skid enter` and
gateway/herdr on
linux and darwin. the [pr 2 spec](../herdr-pr2.md#attached-terminal) owns integration;
these checks do not waive the physical-phone gaps above.

## pr 2 source candidate, 2026-09-22

the production gateway, cli decoder, and android page are now wired to the
pinned public protocol. temporary socket probes passed inventory, metadata
claim, exact typed text, full-frame admission, stale-ref rejection after
websocket admission, detach and child release; they were synthetic, not herdr
live acceptance. the cli decoder's temporary probes reproduced and fixed
per-operation response decoding and json expansion of maximum text/read bytes.
android compilation and javascript syntax pass. independent source review found
and repaired unbounded pending main-thread frame callbacks, unsupported named
hardware-key fallthrough, and a scroll-line cap mismatch.

the approved pr 2 live run used isolated official v0.9.1 servers and the
source candidate on darwin and arch linux. both gateway streams delivered a
valid full frame, exact synthetic text plus enter, and detach within two
seconds. with an independent native desktop attached on each host, the worker
held phone-style geometry at 20×40 and returned to desktop geometry 29×73
after detach; the shell survived. darwin cli `enter` also passed typing,
paste, resize and tty restoration on a test-owned terminal. temporary
gateway websocket probes reproduced and repaired pre-frame input admission,
stalled-input release and stale-ref handling; those probes were deleted.

a separate proof apk built from the candidate with only its application id
changed, installed and launched on samsung sm-s906w/android 16, then was
removed without replacing the installed app. full phone terminal acceptance
is `NOT_RUN`: production pairing requires three named machines over canonical
https on port 8443, while the isolated test boundary exposed one darwin http
gateway; the device keyguard also obscured the proof activity. no product
transport or pairing bypass was used. exact phone typing, synthetic paste,
scroll, takeover, background/acquisition release, loss, bearer revocation,
largest-frame rendering and phone geometry return remain unproved. dictation,
gboard paste, local copy and rotation remain user-waived `NOT_RUN`.

## pr 134 phone attempt, 2026-09-23

three isolated candidate gateways were reachable from the physical phone at
distinct canonical https origins through a temporary adb network tunnel. phone
curl validated the public certificate chain and received the expected
unauthenticated response from each gateway. the proof app opened the real qr
scanner, but no invite was created or scanned. the user stopped the qr work
before pairing, so the phone terminal journey and all phone lifetime checks
above remain `NOT_RUN`. no scanner result, pairing, tls or product-code bypass
was used. the proof app, display, proxy, tunnel and local certificate/key were
removed; the production app remains. the public dns registration could not be
removed through the service api and is tracked in
[its own issue](temporary-dns-registration.md).
