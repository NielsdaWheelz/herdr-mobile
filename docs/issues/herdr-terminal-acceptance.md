# herdr phone and desktop terminal acceptance

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
drive a disposable probe, but a production phone contract should request an
upstream typed close reason rather than make human prose into a stable code.
the stream reader admits upstream frames up to 32 mib while skid's current
terminal frame bound is 64 kib; full-frame and queue limits need measurement
and an explicit bounded bridge decision before cutover.

resolution owner: [migration pr 1](../herdr-migration.md#pr-1-feasibility-and-implementation-contract).
use test-owned codex, claude and shell terminals on isolated linux/darwin servers.
exercise the existing phone renderer with typing, composition/dictation, multiline
paste, keys, alternate-screen history, touch scroll, selection/copy, rotation and
keyboard resize. prove actual codex/claude history movement, not just receipt of
scroll commands. taps focus input without remote clicks; selection stays local.
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
