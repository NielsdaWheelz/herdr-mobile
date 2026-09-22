# herdr phone and desktop terminal acceptance

problem: herdr's rendered-frame bridge has not been exercised through skid's
android renderer/input path. its direct controller owns sizing while desktop
input can still reach the same terminal.

impact: raw-byte assumptions, scroll routing, automatic terminal replies,
phone keyboard resizing and desktop handback may need different ownership.
upstream support alone does not prove the existing phone experience works.

evidence: the documented
[terminal bridge](https://herdr.dev/docs/persistence-remote/#direct-terminal-attach)
offers observe/control streams and one writable direct controller. the inspected
v0.9.1 source separately routes ordinary desktop input. all skid/herdr phone and
host interaction acceptance remains `NOT_RUN`.

resolution owner: [migration pr 1](../herdr-migration.md#pr-1-feasibility-and-implementation-contract).
use test-owned codex, claude and shell terminals on isolated linux/darwin servers.
exercise the existing phone renderer with typing, composition/dictation, multiline
paste, keys, alternate-screen history, touch scroll, selection/copy, rotation and
keyboard resize. attach desktop concurrently, then detach/background/disconnect
the phone and verify sizing handback and continued worker execution.

resolved when: a bounded approved live/device journey proves the bridge is usable
and fixes input, reply, scroll, geometry, takeover and cleanup ownership in the
migration contract. pr 2 must repeat the important cases through the implemented
product. if basic operation requires rebuilding substantial terminal machinery,
record that negative result and reconsider adoption before pr 2.
