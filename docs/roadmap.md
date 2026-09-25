# delivery and remaining work

[architecture](architecture.md) owns the system and shared invariants;
accepted feature specifications own detailed product contracts.
[the codebase map](codebase-map.md) locates their implementation.
this index records present scope and open work, not a release diary.

## independent herdr-mobile and skidbladnir

2026-09-25: the [separation spec](herdr-mobile-separation.md) plans this repo's
rename to herdr-mobile and the original repo's reclamation of skidbladnir.
both products must coexist independently on all three hosts and android.
existing herdr panes are disposable; provider integration homes, github names,
deployment ownership, credentials, and phone packages separate explicitly.
the [original-repo handoff](skidbladnir-restoration-handoff.md) and
[dev-server handoff](dev-server-separation-handoff.md) scope those agents'
work; [the issue](issues/herdr-mobile-separation.md) tracks delivery.
this repo's source split is implemented: command/module, android package and
signer, wire, metadata, ports, generated mark, and release/fleet tooling.
engineering checks, signed candidate build, and temporary isolated gateway and
android parser probes passed. the [deployment contract](dev-server-handoff.md)
records exact interfaces and remaining inputs. github names are transferred
and repository ids preserved; known clone remotes are retargeted. publication,
fleet cutover, pane reset, and phone installation remain pending. all fleet and
phone coexistence acceptance remains pending.

## herdr pr 5

2026-09-24: this branch is the step 3 source candidate of [pr 5](herdr-pr5.md):
skid reduces to the android app and one phone gateway per host. the cli, peer
and attach clients, identity hook, process identity and agent read/send/keys
routes are gone; refs encode herdr ids, and only a named agent has one;
launch uses herdr's agent start and names the agent after a free dwarf. a
disposable herdr and gateway on the macbook exercised every phone route; no
release, host or phone change is claimed. step 4 must rewrite the host configs
to the reduced profile table in the same apply that pins this release, and the
phone updates after all three hosts.

## herdr migration

2026-09-23: pr 2 merged as pr 134 at `9999033`; no deployment is claimed.
[pr 3](herdr-pr3.md) is merged in jarvis at `fb5e4a9`; recorded boundary proofs
do not qualify the unmodified live provider journey. [the pr 4 spec](herdr-pr4.md)
now defines installation, that isolated proof, two bounded contract corrections,
coordinated activation and rollback. its source prerequisites are implemented:
truthful stop partials and `skidbladnir validate-host-config` in skid, the herdr
installer and herdr-era host configuration in dev-server, and stop-gated jarvis
activation with the cutover runbook in jarvis. v0.7.0 was qualified in isolation,
published, activated across the fleet and accepted on 2026-09-24; the tmux-era
v0.6.0 rollback and its legacy assets are retired. evidence and the owner's
waivers are in jarvis `docs/qualification/2026-09-23-herdr-pr4.md`. the earlier candidate/proof notes below
retain their historical scope.

[the migration plan](herdr-migration.md) owns the proposed runtime/desktop cutover,
retained android and cli surfaces, and each pr's scope, requirements and completion
criteria. delivery is feasibility/contract, one coherent skid cutover, jarvis
alignment, then deployment and fleet acceptance. this branch contains the pr 2
source candidate; no installation or fleet acceptance is claimed.
isolated v0.9.1 darwin/linux and proof-phone results now supplement source
research; each boundary keeps its own result in the pr 1 record.

2026-09-23: pr 134 at `29f98e7` records passing host control and linux claude
submission/registration proofs; hosted engineering checks passed. the user
accepted the remaining phone/mac proof gaps, same-pid optional identity
carryover and temporary dns residue as
[non-blocking follow-ups](herdr-pr2.md#accepted-non-blocking-follow-ups).
do not require another test campaign or defer these as mandatory rollout gates.
unrun checks stay `NOT_RUN`, and known defects stay documented. this does not
claim completed code review or authorize merge/deployment; pr 3 alignment and
pr 4 release/installation work remain.

[the pr 1 spec](herdr-pr1.md) defines the investigation and contract closure.
remaining acceptance includes [profile/restart behavior](issues/herdr-profile-restore.md),
[exact worker targeting](issues/herdr-agent-targeting.md), and
[phone/desktop terminal interaction](issues/herdr-terminal-acceptance.md), plus
[native closure disclosure and qualification](herdr-pr1.md#2026-09-22-reopened-qualification-at-the-accepted-scope).
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
application-cursor mode at the pinned public api. the user later accepted
visible disabled home/end controls, unmodified page navigation through herdr's
public attached-scroll operation, and explicit rejection of unsupported
hardware/desktop/cli keys. this permits pr 2 candidate implementation; it is
a disclosed feature loss, not key acceptance. disposable heartbeat proof
restored geometry after abrupt loss; the pr 2 bridge still needs product live
proof. there is no deployment authorization.

[the pr 2 spec](herdr-pr2.md) defines this branch's candidate implementation,
exact ownership slices and acceptance. desktop input decoding remains explicit
product qualification work. accepted 2026-09-22: ordinary send requires recognized
idle; `--terminal` remains an explicit readiness override. readiness detection
still needs qualification; the policy decision is not proof.
jarvis alignment is required under the changed candidate wire format.

## implemented scope

| capability | contract owner |
| --- | --- |
| herdr-backed inventory, exact terminal lifetimes, launch, rename, move, kill and direct attachment | [architecture](architecture.md), [pr 2](herdr-pr2.md), [rename](session-renaming.md) |
| herdr agent status, interrupt and stop | [architecture §4](architecture.md#agent-status-and-controls) |
| real herdr workspaces, client grouping/filtering and dashboard restoration | [spaces](spaces.md), [dashboard continuity](dashboard-return-continuity.md) |
| standalone terminal and new terminal here | [terminal creation](shells.md) |
| phone fleet connect/reconnect, encrypted pairings and quarantine | [fleet distribution](public-fleet-distribution.md), [architecture §6](architecture.md#6-android-surface) |
| phone dashboard, directory chooser and machine pressure | [refresh](dashboard-pull-to-refresh.md), [chooser](working-directory-chooser.md), [pressure](machine-pressure-rail.md) |
| terminal sizing, keys, touch, selection and input composition | [sizing](terminal-readable-sizing.md), [key deck](terminal-key-deck.md), [touch](terminal-touch-scroll.md), [selection](terminal-selection-copy.md) |
| visual language and generated assets | [design language](design-language.md) |

source implementation does not establish every runtime or human acceptance
criterion. feature specs retain their detailed acceptance requirements.

## release and operations

[`release-pin.json`](../release-pin.json) owns the first published herdr-mobile
release, `v0.9.0` at `68a652d7ccbeaaf472ef1c5f3a4ea6949808bca4`, and all five
artifact digests. the old skid-named `v0.8.0` pin remains in git history at
`d8bb9c4`; it is not a separated rollback target. a release pin does not assert
any host or phone's installed version. the [deployment contract](dev-server-handoff.md)
records source/release qualification and outstanding cutover work.

`dev-server` owns machine-local installation, services and configuration.
`scripts/fleet` owns `verify` and direct `invite`; `invite` reads the peers'
private credentials from the macbook's `~/.config/herdr-mobile/client.json`.
`scripts/install-android` validates and installs an apk in place; installation
alone is not pairing or behavioral acceptance. [architecture §5](architecture.md#5-host-architecture)
and [§6](architecture.md#6-android-surface) own those boundaries.

new agent launches take their permission flags from the deployment's shell
aliases under [architecture §2](architecture.md#2-fixed-contract); herdr-mobile passes
none. existing sessions retain their original launch policy.

## open work and acceptance

- [sequential cleanup](codebase-map.md): verified findings live in [issues](issues),
  one per issue. finish one reviewed pr before starting the next.
- [phone terminal interaction](issues/herdr-terminal-acceptance.md): the
  gateway stream and linux/darwin phone behavior need product live proof.
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

other unperformed visual/device checks and explicitly waived shipment checks
remain with their feature owners. a waiver is not a pass. unavailable or
unexecuted boundaries remain `NOT_RUN` and require their applicable approval.

## historical evidence

[source-attributed release and acceptance records through this cleanup](https://github.com/NielsdaWheelz/herdr-mobile/blob/5986a650d02106a2c10415a81a6ad956fe198665/docs/roadmap.md)
remain in git. they include the v0.5.0 failures, corrected v0.6.0 phone results,
and shipment waivers. removing their duplicate active-document tables neither
erases failures nor proves current acceptance. retired commands are historical
recipes, not executable gates or instructions to rebuild a harness.
