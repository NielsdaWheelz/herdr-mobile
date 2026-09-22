# delivery and remaining work

[architecture](architecture.md) owns the system and shared invariants;
accepted feature specifications own detailed product contracts.
[the codebase map](codebase-map.md) locates their implementation.
this index records present scope and open work, not a release diary.

## planned herdr migration

[the migration plan](herdr-migration.md) owns the proposed runtime/desktop cutover,
retained android and cli surfaces, and each pr's scope, requirements and completion
criteria. delivery is feasibility/contract, one coherent skid cutover, jarvis
alignment, then deployment and fleet acceptance. all stages remain uncompleted.
isolated v0.9.1 darwin/linux and proof-phone results now supplement source
research; each boundary keeps its own result in the pr 1 record.

[the pr 1 spec](herdr-pr1.md) defines the investigation and contract closure.
it must resolve [profile/restart behavior](issues/herdr-profile-restore.md),
[exact worker targeting](issues/herdr-agent-targeting.md), and
[phone/desktop terminal interaction](issues/herdr-terminal-acceptance.md), plus
[native closure disclosure and qualification](herdr-pr1.md#2026-09-22-reopened-qualification-at-the-accepted-scope) before
implementation proceeds. current implemented scope below remains the tmux system.
the original `reconsider` finding remains historical evidence. the user-approved
[scope amendment](herdr-pr1.md#accepted-scope-amendment) accepts native group-close
effects with disclosure and excludes remote clicks/drags; keys, scrolling and
local selection/copy remain required. pr 1 qualification is reopened at v0.9.1,
without requiring upstream changes for those former gates. the remaining proofs
and wire contract are open. the user accepted terminal-only claude observation
and new unnamed shell lifetimes without metadata after cold restart; a
surviving native pane label is only a hint, as is an unmarked manual pane label;
actions need a fresh reference until named through skid. manual phone interaction was
explicitly skipped and remains `NOT_RUN`; provider startup/readiness and the
full phone key deck are unresolved: physical home/end failed in
application-cursor mode at the pinned public api. disposable heartbeat proof
restored geometry after abrupt loss, but pr 2's actual bridge is absent. there is no `proceed`
or pr 2 authorization.

## implemented scope

| capability | contract owner |
| --- | --- |
| host inventory, exact session lifetimes, launch, rename, kill and direct attachment | [architecture](architecture.md), [client and attachment](agent-control-ux.md), [rename](session-renaming.md) |
| sampled agent status, bounded reads, terminal input, interrupt and stop | [agent control](agent-control.md) |
| session labels, client grouping/filtering and dashboard restoration | [spaces](spaces.md), [dashboard continuity](dashboard-return-continuity.md) |
| standalone terminal and new terminal here | [terminal creation](shells.md) |
| organized desktop browser and fullscreen attachment return | [desktop browser](desktop-browser.md) |
| phone fleet connect/reconnect, encrypted pairings and quarantine | [fleet distribution](public-fleet-distribution.md), [architecture §6](architecture.md#6-android-surface) |
| phone dashboard, directory chooser and machine pressure | [refresh](dashboard-pull-to-refresh.md), [chooser](working-directory-chooser.md), [pressure](machine-pressure-rail.md) |
| terminal sizing, keys, touch, selection and input composition | [sizing](terminal-readable-sizing.md), [key deck](terminal-key-deck.md), [touch](terminal-touch-scroll.md), [selection](terminal-selection-copy.md) |
| visual language and generated assets | [design language](design-language.md) |

source implementation does not establish every runtime or human acceptance
criterion. feature specs retain their detailed acceptance requirements.

## release and operations

[release-pin.json](../release-pin.json) is the single committed owner of the
published version, source and artifact digests. it does not assert the currently
installed version of any host or phone. this cleanup changes source only;
it does not publish or deploy a release.

`dev-server` owns machine-local installation, services and configuration.
`scripts/fleet` owns `verify`, direct `invite`, and `provision-clients`.
`scripts/install-android` validates and installs an apk in place; installation
alone is not pairing or behavioral acceptance. [architecture §5](architecture.md#5-host-architecture)
and [§6](architecture.md#6-android-surface) own those boundaries.

new agent launches use deployment-owned permission bypass flags under
[architecture §2](architecture.md#2-fixed-contract). deployed configuration was
verified; new provider launches were not exercised in that change. existing
sessions retain their original launch policy.

## open work and acceptance

- [sequential cleanup](codebase-map.md): verified findings live in [issues](issues),
  one per issue. finish one reviewed pr before starting the next.
- [darwin desktop browser](issues/desktop-browser-runtime-acceptance.md): the
  real browser/pty/gateway/isolated-tmux journey remains skipped by user direction.
- [spaces and shells hands-on](issues/spaces-shells-hands-on.md): human workflow
  and usability acceptance remains unperformed; automated phone results do not
  supply it.
- [readable terminal sizing](terminal-readable-sizing.md#red--green--refactor-and-acceptance): the named
  fitting-grid/handback and human readability criteria remain unclaimed by the
  generic later release-suite result.
- [behavioral coverage](issues/test-system-reset.md): temporary change-specific
  tests are removed before commit. no retained suite protects the important
  behavior automatically. [testing policy](rules/testing.md) owns the workflow;
  `scripts/check verify` runs engineering checks and builds only.
- [terminal embedding](spaces-and-shells.md): custom desktop embedding work is
  paused while [herdr feasibility](herdr-migration.md) is qualified; there is no
  accepted production embedding contract.

other unperformed visual/device checks and explicitly waived shipment checks
remain with their feature owners. a waiver is not a pass. unavailable or
unexecuted boundaries remain `NOT_RUN` and require their applicable approval.

## historical evidence

[source-attributed release and acceptance records through this cleanup](https://github.com/NielsdaWheelz/skidbladnir/blob/5986a650d02106a2c10415a81a6ad956fe198665/docs/roadmap.md)
remain in git. they include the v0.5.0 failures, corrected v0.6.0 phone results,
and shipment waivers. removing their duplicate active-document tables neither
erases failures nor proves current acceptance. retired commands are historical
recipes, not executable gates or instructions to rebuild a harness.
