# original skidbladnir: agent handoff

2026-09-25 update: the github name transfer succeeded. repository id
`1342599607` is `NielsdaWheelz/herdr-mobile`; id `1386409483` is
`NielsdaWheelz/skidbladnir`. known clone remotes are retargeted. the baseline
names below describe the pre-transfer investigation. immutable releases are
enabled for both repositories. herdr-mobile `v0.9.0` is published and pinned;
your independent release and live namespace handback remain pending.

coordination follow-up: dev-server's ten-file acknowledgment and generation
mode/digest admission correction are recorded in its local `94931a1` commit.
its provider-preservation correction is present but uncommitted; that older
commit alone must not be deployed. your last reported hosted-verified source
is `4eeba152bc4e37cc36eaa9a367c27efcb47aba20`, run `36204088437`; final
provider-contract documentation corrections remain uncommitted.

git write and network access are restored. continue from the existing checkout:
review/commit those corrections, preserve unrelated edits, push, and obtain
that exact final source's hosted check. no further message relay is needed:
the user has supplied both owners' acknowledgment and status.

local signing preflight passed per your handoff; fresh phone version/signer
observation still needs applicable current-turn authorization before selecting
`v0.9.0`/`9000`. publication is root-owned preparation and does not wait for
host namespace handback. live installation still waits for that handback.

implement the original tmux-backed product's side of the split below. inspect
your repository's architecture, roadmap, and agent rules first. this is a
bounded restoration and coexistence task; preserve the original architecture.
owner correction: your isolation must remain scoped to your product. herdr and
ordinary provider commands keep their existing binaries, homes and histories;
the earlier private herdr-home proposal is withdrawn.

## identity and scope

your repository is `NielsdaWheelz/skidbladnir`, github repository id
`1386409483`; the local checkout remains `skid-v1`. the source investigation
baseline was `927c55412eec7fa129a3325fb8f5ebf8051b6ea0`. the other repository,
id `1342599607`, is `NielsdaWheelz/herdr-mobile`. the root operator completed
and verified this same-owner github name transfer on 2026-09-25 without
replacing either history. immutable releases are enabled for both. verify
the intended repository id when preparing publication; old-name redirects
are no longer a safe source selector.

the owner requires both products to run concurrently and independently on
macbook, devbox, arch, and one android phone. installing, updating, stopping,
rolling back, or removing yours must leave herdr-mobile and upstream herdr
working. shared host-managed provider binaries and tailscale are allowed;
their upgrades are separate from either product's gateway maintenance.

retain your existing identities: binary `skidbladnir`, optional existing `skid`
cli, android `dev.niels.skidbladnir` with the skidbladnir label/ship icon,
`skidbladnir.service`, `dev.niels.skidbladnir`, `~/.config/skidbladnir`,
`~/.local/share/skidbladnir`, `~/.local/state/skidbladnir`, loopback `7341`,
tailscale https `8443`, and existing invitation/header formats.
herdr-mobile reserves its own names/paths, android `dev.niels.herdr.mobile`,
loopback `7342`, and https `8444`. upstream `herdr`, its services/socket,
and `~/.local/share/herdr` belong to the herdr side.

edit only your repo. one dev-server integrator owns deployment changes;
give that agent precise templates, pins, helper and profile requirements.
its separate assignment is [the dev-server handoff](dev-server-separation-handoff.md);
this prompt contains everything needed for your own source work.
do not rename github repositories, deploy, or reclaim live skid paths ahead
of the root operator's namespace handback. source preparation may proceed
independently. obey current-turn tmux/device authorization and logging rules.

## required work

1. audit release tooling before using it. its destination
   `NielsdaWheelz/skidbladnir` now resolves to your intended repository.
   your repo has no releases; its former `v0.6.0` pin referred to assets held
   by the other repo and is removed. obtain exact-source hosted checks on the
   committed restoration and coordinate fresh publication with the root
   operator; replace the pin only with those published artifact digests.
   do not depend on herdr-mobile releases or restore a whole old deployment.
2. retain the original android signing key/certificate. the usb phone was
   observed running herdr-backed `0.8.0`, code `8000`, under YOUR package id;
   its signer matches the existing committed skid certificate. publish an
   increasing code, planned `v0.9.0` / `9000` if unused. recheck installed
   version first. the operator installs/pairs the new herdr app first, then
   force-stops and clears only your obsolete package data, installs yours,
   and pairs fresh. do not depend on a downgrade or stored-data compatibility.
3. use fresh gateway bearers, machine handles, and private fleet config after
   handback. no copied/symlinked credentials or old herdr-backed generations
   in your rollback chain. your installer and verifier must work with the
   other product absent and must not inspect its signing files or state.
4. keep four profiles, with private homes under
   `~/.local/share/skidbladnir/providers/`: `codex-personal`, `codex-work`,
   `codex-work2`, `claude-work`, plus `claude-personal` for manual shell use.
   invoke absolute native providers with explicit
   home environment, original permission arguments, and your claude identity
   plugin. shared work wrappers override homes and are unsuitable. install
   your hooks only in these homes. authenticate/trust normally; do not copy
   existing account trees, discovery sockets, caches, or credentials.
   cover shell terminals too: typing bare/account provider commands after
   login-shell startup must select your homes. supply product-scoped defaults
   and minimal wrappers where required, active only in your own marked shells;
   ordinary and herdr commands must retain their existing account homes.
   explicit profile selection wins. do not mutate existing user tmux sessions
   or invent a general runtime-selection framework.
5. restore and pin the required `nativeControlPath` helper. historical inputs:
   `llm-calling@ec97adeb9ddd0f91b141f89cc42cff7cc7efdb8f`, uv `0.11.28`,
   python `3.12.13`, claude sdk `0.2.130`. inspect and qualify rather than
   assuming compatibility with current providers. use a skid-owned helper
   installation/command, e.g. `skidbladnir-provider-runtime-control`; verify
   its claude subprocess resolves the intended executable and private home.
6. prevent inherited `HERDR_*` context entering your gateway, tmux startup,
   or provider children. never modify unrelated live tmux sessions/server
   environment. the owner rejected herdr provider-home relocation: herdr and
   ordinary shells keep `.codex`, `.codex-work`, `.codex-work2`, `.claude`,
   `.claude-work` and their existing provider state. preserve native herdr
   integrations there. the integrator may remove only proven obsolete skid
   registrations, preserving user configuration, hooks, trust and cognition.
   prove hook isolation at `cwd=$HOME` and a shared project, including inline
   and plugin sources. do not add a shared hook dispatcher or modify upstream
   herdr integrations.

## deliverables and acceptance

deliver reviewed source changes, engineering/release checks, a precise
dev-server handoff, and the exact release/source/package/signer identities.
write that deployment handoff to `docs/dev-server-handoff.md` in your own
checkout and return its absolute path for relay to the dev-server agent.
include the exact host-config schema and validator invocation, profile rows
and shell defaults, helper command/version/dependencies and launch interface,
and qualification results or explicit pending prerequisites. provide this
source contract before release publication; add the published
`release-pin.json` reference when available. do not invent future digests.
record unresolved problems in your own `docs/issues/`; never report unexecuted
behavior as passing. follow your repo's retired-test policy; temporary probes
must not become a new retained harness.

after namespace handback, prove on all three hosts: both products active;
your forge profiles and manually typed provider commands in your marked
terminals use your private homes, while ordinary/herdr commands retain existing
homes and histories; invites and
credentials cannot cross products; independent restart, reinstall and rollback
preserve the other's workers/attachments/files. prove scoped removal on a
disposable installation. tmux probes may mutate only their own isolated `-L`
socket resources. prove both phone apps pair, launch, attach, and control their
own workers, with distinct icons and independent data. return content-free
results and remaining blockers to the integrator.

the owner permits existing HERDR panes to die. this does not permit killing
unrelated tmux sessions or changing jarvis's cognition services. the root
operator owns the herdr reset and the live phone/package transition.
