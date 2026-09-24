# herdr pr 5: herdr-native fleet, phone-only skid

2026-09-24: draft specification from the owner's direction; no code, release or
host change authorized or claimed. the design questions are settled against
herdr's source at the pinned `065ef9d6` (v0.9.1) and an isolated server on the
macbook (§7). baseline: skid `f2a07b1` (v0.7.0 published),
jarvis `3dc3590` (`2a59355` active), dev-server `e58e1a1`. owner decisions
already taken: per-host phone gateways, not a devbox hub; jarvis's gate allows
agents plus the pane creation an agent needs, not `pane run`; devbox holds no
ssh access to a workstation except through jarvis's gate; no host declares
herdr saved machines (§6).

## 1. outcome

nothing in skid duplicates herdr. herdr 0.9.1 already provides remote attach
over ssh (`--remote`), a local cli that `ssh <host> herdr …` reaches on any host,
agent primitives (`agent start/get/read/explain/prompt --wait/wait/send-keys`),
stable pane ids, a blocked-agent refusal before any prompt, and built-in
codex/claude integrations. pr 1 through 4 kept a skid cli, peer client, attach
client, identity hooks, notifier and agent-control contracts for jarvis. none
of them is needed once humans and jarvis call herdr directly.

after this pr:

| owner | responsibility |
| --- | --- |
| herdr | terminals, panes, workspaces, agent detection and lifecycle, desktop and remote attach over ssh, agent integrations and notifications |
| skid gateway | one host's phone api over its local herdr: inventory projection, profile launch, stop, terminal stream, pressure, directory listing, pairing. nothing else |
| skid android | the phone app, unchanged in behaviour |
| jarvis | intent, authority and assignments; a thin herdr codec with its own write checks |
| dev-server | pinned herdr and gateway, herdr integrations, the jarvis gate and its keys |

humans attach to any host's herdr from a workstation with `herdr --remote
<ssh-target>` and run herdr commands on it with `ssh <host> herdr …`. jarvis controls all three hosts through an ssh
forced-command gate on each. the phone keeps pairing with the three gateways.

## 2. invariants

- one herdr server per host is the only terminal owner; every client reaches it
  through its socket, locally or over ssh.
- all three hosts run the same herdr pin. remote attach compares protocol
  numbers and herdr bumps the protocol even in patch releases, so a herdr pin
  change is one fleet-wide step, and jarvis's codec is qualified against it.
- a gateway talks only to its own host's herdr. no host holds ssh authority for
  the phone.
- jarvis's process never reads the owner's files or credentials; everything it
  does happens in a herdr pane the owner can see. the gate adds no authority
  beyond what skid's gateway gave it: a shell pane driven with keys was already
  command execution.
- a stale reference never reaches a different pane or agent: after replacement,
  close or herdr restart, a write fails before dispatch.
- an uncertain write is never replayed (herdr's own rule: a connection failure
  does not prove a mutation was not applied).
- readiness has one owner: herdr's detection. neither skid nor jarvis parses
  screen text.

## 3. deleted

skid:

- `internal/agentcli`, `internal/fleetclient`, `internal/terminalclient`,
  `internal/agenthook`; bare `skid`, `skid enter`, `agent-hook` and the agentcli
  fallthrough in `cmd/skidbladnir`; the `skid` executable name.
- `agentcontrol` read, send and keys; `sessions.Info`; the hook registration
  code in `agentruntime`; the per-pane lifetime tokens and the process
  identity (pid, start identity, command fingerprint) in references, and
  `internal/process` if nothing else uses it.
- routes `GET /v1/terminals/{ref}`, `POST /v1/agents/{ref}/read|send|keys`,
  `GET /healthz`.
- `scripts/fleet provision-clients`; per-host `client.json` except the
  macbook's credential file that `scripts/fleet verify` and `invite` read.
- superseded docs: `agent-control*.md`, `agent-identity-projection.md`,
  `jarvis-codex-control.md`; pr 1 to 4 specs stay as history.

dev-server:

- the `~/.local/bin/skid` link and its protected-link handling; `skid-notify`,
  `agent-hooks-*.json`, the `claude-agent-identity` plugin and
  `skidbladnir_install_integrations`; the `--plugin-dir` argument in the host
  configs; the devbox copy `/usr/local/libexec/skidbladnir` for jarvis.
- by hand once: the `notify = [".../skid-notify"]` line in the macbook's three
  codex `config.toml` files.

jarvis:

- `agent_control.py` (the skid cli codec), `JARVIS_AGENT_CLI_PATH`,
  `JARVIS_AGENT_CLIENT_CONFIG_PATH` and `/etc/jarvis/agent-client.json`, with
  their install and containment checks. adr 0044, 0045 and 0048 are superseded
  by a new adr.

## 4. kept, and why

the gateway keeps exactly the phone's routes: `GET/POST /v1/terminals`, the
terminal stream, shell, rename, workspace move, delete, agent interrupt and
stop, pressure, directory listings, pairings and pairing invites. herdr has no
network api, and phone pairing, bearer auth, pressure, directory browsing and
dwarves have no herdr equivalent.

inside those routes the gateway stops re-implementing herdr:

- launch creates the pane with `--env` for the profile's account home
  (`CODEX_HOME` or `CLAUDE_CONFIG_DIR`), then `agent start <name> --kind`,
  instead of typing `exec <wrapper>`. `agent start` types the bare `codex` or
  `claude` at the pane shell's prompt, so the account wrappers must respect a
  preset account home (§6). the agent's herdr name is its dwarf's name in
  herdr's grammar (`[a-z][a-z0-9_-]{0,31}`, unique among a server's live
  agents, e.g. `haugspori` for `norse.haugspori`), so the phone's name and
  herdr's are one.
- status shows herdr's `agent get`/`agent explain` result as it is. the phone's
  readiness label stays a projection, with no second rule.
- stop stays: herdr has no single stop, so it remains interrupt, re-check and
  `pane close`, with its truthful partials.
- references become thin encodings of herdr ids: machine, pane id and
  `terminal_id`, plus the agent name for agent references. herdr reissues pane
  ids after a restart but never repeats a `terminal_id`, and it clears an
  agent's name when the agent exits or is replaced. before a write the gateway
  re-reads the pane and requires the same `terminal_id` (and, for an agent, the
  same name), which is herdr's own check in `agent start`. the check is not
  atomic; the race stays accepted. references stay opaque to the phone.
- the phone's session card loses `provenRuntimeProfile` (it came from the
  deleted hook) and shows `launchProfile`. android changes only if its decoder
  requires the field.

## 5. jarvis over herdr

transport. dev-server owns jarvis's side of the gate on devbox:
`/etc/jarvis-herdr/` holds jarvis's key (`id_ed25519`, jarvis 0600, generated on
devbox; its public half is committed to dev-server as the trust root), an
`ssh_config` naming the three hosts (`devbox` as `niels@localhost`, `macbook`
and `arch` as `nnandal` over the tailnet; `BatchMode`, `IdentitiesOnly`,
`StrictHostKeyChecking yes`) and a `known_hosts` built from committed host
keys. jarvis runs `ssh -F /etc/jarvis-herdr/ssh_config <label> <herdr args>`.
each host's owner account (`niels` on devbox, `nnandal` on macbook and arch)
authorizes the key once as `restrict,command="$HOME/.local/libexec/herdr-gate"
<key>`; the gate lives in the owner's account because it guards jarvis, not the
owner. devbox reaches
itself as `niels@localhost` through the same gate, so all three hosts look the
same. the gate splits `SSH_ORIGINAL_COMMAND` into argv without a shell, accepts
only the allowlist below, and execs the host's own pinned herdr, which talks to
its local server. the gate is policy hygiene, not containment: `agent start
--pane`, `pane split` and `pane close` accept any pane id, and a prompt to an
agent running with `--yolo` is already arbitrary execution as the owner. what
containment keeps is that jarvis's own process never reads the owner's files
and acts only through visible panes. jarvis's codec only starts agents in panes
it has just created. anything else exits nonzero without running herdr. jarvis
never uses `--machine`: that path runs `sh` probes and a bridge over several ssh
sessions, which a forced command would break, and it adds nothing when the gate
already runs herdr on the target. cli output is herdr's api envelope
`{id, result:{type, …}}`, defined by `herdr api schema`; `agent read` prints
plain text. the codec ignores unknown fields. dev-server installs the gate and the
authorized key; jarvis owns the allowlist's content through its adr, and
`verify-containment` checks the key's mode and the three `restrict,command=`
lines.

allowlist, owner decision "agents plus pane creation":

| purpose | herdr commands |
| --- | --- |
| inventory | `agent list`, `agent get`, `pane list`, `workspace list` |
| observe | `agent read`, `agent explain`, `agent wait` |
| create | `workspace create --cwd --env`, `pane split --env` (env limited to `CODEX_HOME` for the three codex homes and `CLAUDE_CONFIG_DIR` for `.claude-work`; the personal claude runs with it unset), `agent start --kind codex|claude` |
| drive | `agent prompt [--wait --timeout]`, `agent send-keys` |
| end | `pane close` |

excluded: `pane run`, `pane send-text`, `pane wait-output`, and every server,
config, machine, integration, update, worktree, plugin and notification
command.

tools. the nine `agent.*` tool names stay, mapped as follows:

- list: per-host `agent list` and `pane list`; an unreachable host is partial.
- info: `agent get`.
- start: create a pane with env, then `agent start`.
- read: `agent read`.
- send: `agent prompt`, which refuses a blocked agent. the terminal-mode text
  bypass goes; menus are answered with keys after inspection.
- keys: `agent send-keys`.
- interrupt: `send-keys ctrl-c`.
- stop: interrupt, re-check, `pane close`.
- kill: `pane close`.

jarvis starts agents with unique names and keeps (machine, name, pane id,
`terminal_id`) as the reference. before every write it re-reads `agent get
<name>` and requires the same pane and `terminal_id`, then writes to the name.
a restarted server, a replaced agent or a reused pane id fails that check. the check/write race
stays accepted. action recording, write grounding and non-replay are unchanged.

activation. the catalog changes, so this is a jarvis cutover like pr 4: the new
release's `check-activation` passes only with no pending agent action. forward
only; the previous release needs the skid cli, which is removed last.

## 6. humans across machines

no saved machines. herdr's saved ssh machines (`endpoints.json`, the only
selector `--machine` accepts) make every open herdr window hold ssh sessions to
each machine, and on the far end herdr's `remote-client-bridge` starts `herdr
server` itself when none is listening. during a herdr restart or pin change, or
on arch before login, that would start an unmanaged server with the wrong
environment in the supervised server's place; 0.9.1 has no attach-only switch.
so a human uses `herdr --remote <ssh-target>` to attach the desktop to another
host's herdr, and `ssh <host> herdr agent|pane|workspace …` for commands, which
run the target's own pinned herdr. `issues/herdr-saved-machines.md` in
dev-server records what would bring `--machine` back.

account wrappers. `agent start` types the bare `codex`/`claude`, and a pane's
`--env` sets the new shell's environment at spawn (it is not persisted across a
herdr restore). the `codex` wrapper today forces `CODEX_HOME` to the personal
home, so a pane created for work or work2 would still run personal. the bare
`codex` wrapper changes to respect a preset home and default to personal
(`: "${CODEX_HOME:=…}"`); `codex-work` and `codex-work2` keep forcing theirs,
because their name is the choice. bare `claude` is the real binary and already
honors `CLAUDE_CONFIG_DIR`; `claude-work` keeps forcing its dir. proven in
isolation: with the respecting wrapper the pane's home reaches the codex
process; with today's it does not.

integrations: dev-server runs `herdr integration install codex` once per codex
account home and `claude` once per claude config dir, with `CODEX_HOME` or
`CLAUDE_CONFIG_DIR` set. codex gets `herdr-agent-state.sh`, a SessionStart hook
in `hooks.json` and `[features] hooks = true`; claude gets
`hooks/herdr-agent-state.sh` and a SessionStart hook in `settings.json`. the
hook needs python3 and runs only inside a herdr pane. herdr's installer and
skid's both write `hooks.json` and `settings.json`, so the switch happens at
once in delivery step 4, when skid's hook goes; nothing in steps 2 and 3 needs
herdr's hook. codex asks once per
profile and host to trust the new hook; answer it deliberately as in pr 4.

## 7. settled questions

1. pane ids: reused after a restart. workspace numbers come from an unsaved
   counter and restart at `w1` without a session file; restored panes keep their
   ids but run new shells. no api field identifies a server instance.
   `terminal_id` is reissued on every restore and never repeated, and herdr's
   own `agent start` uses it to notice a replaced pane (§4, §5).
2. `agent start` types the bare agent name plus `--` args at the shell prompt;
   the shell's aliases, functions and PATH decide what runs, and it sets no
   environment. `--env` exists on `workspace create`, `tab create` and
   `pane split` (§6, account wrappers).
3. saved machines live in `~/.local/state/herdr/client/endpoints.json`, which
   could be written directly; `machine add` is interactive and may install or
   stop a remote server. `--machine` needs a running remote server and equal
   protocol numbers. saved machines are not used: every open herdr window
   connects to them and the remote bridge starts servers (§6).
4. json: the cli prints the api envelope defined by `herdr api schema`, except
   plain-text reads. the protocol changed within minor lines (0.7.4 → 0.7.5,
   0.8.0 → 0.8.2), `schema_version` stays 1, and herdr's only stated policy is
   to ignore unknown fields. the herdr pin is the compatibility boundary (§2).

known upstream detection gaps, unchanged by this pr and owned by herdr: codex
0.156's trust menus and codex's sign-in screen both read as idle and ready
(`issues/codex-menu-readiness.md`), so `agent prompt`'s blocked refusal does not
catch them.

## 8. delivery

1. dev-server (additive): the gate on every host, jarvis's key, ssh config
   and known hosts on devbox, the authorized key in each owner account, and
   the respecting `codex` wrapper. skid and jarvis are unchanged. the wrapper
   lives in `codex-shared.py`, whose hash is devbox's codex identity, so the
   devbox apply needs `--restart-codex` and runs in step 2's window while
   jarvis is stopped. arch's part waits until arch is reachable.
2. jarvis: the herdr codec release, qualified in isolation against the gate,
   then activated under the pr 4 runbook (pause, stop, check, activate,
   containment, resume).
3. skid v0.8.0: the reductions in §3 and §4, qualified on the phone's journeys
   (list, launch per profile, stream, interrupt, stop, pairing continuity).
4. dev-server: pin v0.8.0; remove the skid link, hooks, notifier, plugin and
   the jarvis cli copy; install herdr's codex and claude integrations in their
   place.

each step keeps the previous one working: jarvis switches before the cli
disappears, and the phone's routes never change shape.

## 9. done when

- from each workstation, `herdr --remote` attaches to the other hosts and
  `ssh <host> herdr agent list` answers.
- jarvis starts, reads, prompts, interrupts and stops a codex and a claude
  agent on each host through the gate. a command outside the allowlist is
  refused. `verify-containment` passes.
- the phone lists, launches per profile, streams, interrupts and stops on all
  three hosts with its existing pairings.
- `git grep` finds no skid cli, peer client, attach client, hook or notifier
  code, and no host has `~/.local/bin/skid`, `skid-notify`, `agent-hooks` or
  `/usr/local/libexec/skidbladnir`.
