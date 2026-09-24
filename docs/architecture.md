# skíðblaðnir: product and architecture

this document describes the herdr pr 5 source candidate: the android app and one
phone gateway per host. it is not a claim that any host or phone has been
upgraded. [the roadmap](roadmap.md) records delivery and open acceptance;
[pr 5](herdr-pr5.md) owns this reduction, [pr 1](herdr-pr1.md) the retained
wire and [pr 2](herdr-pr2.md) the stream. accepted feature specifications own
their detailed interactions. [codebase rules](rules/index.md) own
implementation conventions.

## 1. Philosophy

herdr owns terminals, panes, workspaces, agent detection and lifecycle,
desktop and remote attach over ssh, and agent integrations.
providers own execution and history. skid owns nothing herdr already does: one
authenticated gateway per host projects that host's herdr for the phone and
adds what herdr has no equivalent for, namely phone pairing, bearer auth,
pressure, directory browsing, profile launch and dwarves. the phone composes
independent gateways; no gateway coordinates another host or holds ssh
authority. there is no application database or replay engine.

one dwarf is one live herdr terminal. herdr may detect an agent in it; the
agent's identity and readiness are herdr's facts. a terminal can exist without
an agent, and an agent ref can become stale while its terminal remains live.
humans attach with `herdr --remote` and run `ssh <host> herdr …`; jarvis goes
through its ssh gate. phone, desktop and jarvis address the same worker without
a focus broker.

## 2. Fixed contract

| concern | decision |
| --- | --- |
| product | skíðblaðnir; one user, one tailnet, three hosts |
| hosts | linux/systemd user service on devbox and arch; darwin/launchagent on macbook |
| runtime | one independently supervised herdr v0.9.1 server per host; no request starts or replaces it |
| phone | android 16/api 36 compose dashboard and source-pinned xterm.js renderer |
| ingress | one tailscale serve tls `:8443` origin per machine to a loopback gateway; no funnel or public ingress |
| machine identity | immutable random `mh-` installation handle; a label, origin, bearer, or platform is not identity |
| authentication | independent bearer per gateway and one-use five-minute pairing invitation; `/v1` requests bind bearer and pinned machine handle |
| profiles | empty or the closed ordered `personal`, `work`, `work2`, `claude-work` table; each row fixes label, provider and the pane environment holding its account home |
| state | herdr panes, workspaces, agents and reserved `skid_*` pane metadata (name flag, launch profile, objective) are runtime truth; android persists encrypted pairings and local presentation preferences |
| host app | go gateway over the local public herdr socket, platform pressure and directory observation |
| clients | the phone only; humans and jarvis call herdr directly |
| delivery | skid v0.8.0 is herdr pr 5 step 3; source changes install, publish, or update nothing |
| trust | agents run as the host user; same-uid adversarial containment is out of scope |

callers choose only a declared profile and a validated cwd; they cannot supply
a command, account home, argument or objective-as-prompt. a launch creates the
pane with the profile's environment and asks herdr to start the provider's
bare `codex` or `claude` there. the pane shell resolves that name: the
deployment's `codex` wrapper respects a preset `CODEX_HOME`, and the
deployment's shell aliases add the permission flags. skid passes no arguments,
because codex refuses a repeated `--yolo`. a zero-profile host still has
terminal creation and herdr's agent detection. launch metadata names a
candidate profile, never proves the account of a running process.

### product language

skíðblaðnir is the app; the dashboard presents dwarves from dvergatal, the
append-only character catalogue. a terminal's dwarf derives from its herdr
terminal id; an agent the gateway launches takes its dwarf's name as its herdr
name (`haugspori` for `norse.haugspori`, `haugspori-2` while another live agent
holds it), so the phone's name and herdr's are one. a dwarf's landmark is
independent of its operator-owned terminal name. `agent` means a codex or
claude process herdr detects in a terminal. errors and destructive
confirmations name the actual terminal and possible linked-workspace effects
rather than relying on the metaphor.

### guarantees

- detach releases only the attachment; kill requires confirmation and an exact
  current terminal ref. native close can also end linked workers, and the
  confirmation says so.
- validate names, workspace labels, cwd, refs and encoded sizes at their
  owners. reject stale targets before dispatch; never substitute a refreshed
  ref for a mutation target.
- host credentials stay on their host. phone pairings remain encrypted and
  machine-bound. logs and evidence contain no terminal bytes, prompts,
  objectives, tokens, account data, or clipboard text.
- restart rediscovers runtime state. it never replays input or claims completion.
  a herdr restart gives new terminal ids, so every earlier terminal and agent
  ref is stale.

### non-goals

a skid cli, peer client, attach client, identity hook or notifier; provider
transcript storage or search, semantic task state, orchestration, automatic
resume/retry/replay, a persistent worker registry, cross-host move or
broadcast, remote terminal clicks/drags, multi-user isolation, and a second
terminal runtime are excluded. retired hook status/history machinery, sqlite
lifecycle facts, provenance, contract codegen, and proof ledgers stay retired.

## 3. Platform facts and accepted limits

herdr's public api supplies pane/workspace/agent discovery, metadata, launch,
agent start and send-keys, mutation, visual scroll, and a separate
terminal-control child. its v0.9.1 key parser does not accept home, end,
insert, delete, or modified page keys, and trims modified non-ascii
whitespace. android keeps home/end visible and disabled. unmodified page
up/down use public visual scroll. no client guesses raw escape sequences to
recover an unsupported key.

herdr never repeats a `terminal_id` and reissues every one on restore. a pane
id is workspace-scoped: it changes when the pane moves and is reused after a
restart. herdr clears an agent's name when that agent exits or another replaces
it. workspace numbers restart from the restored maximum, so a closed
workspace's id can return after a restart. no api field identifies a server
instance.

native pane close may close a linked group or be refused. herdr offers no
compare-and-set write: skid re-reads its target before every write, but another
herdr client may still interleave between check and write. an agent herdr
detected without a name (one a human typed) is identified only by its
terminal, so its replacement by another unnamed agent there is not detectable.
herdr's readiness is a projection of its screen rules; codex's trust menus and
sign-in screen read as idle ([issue](issues/codex-menu-readiness.md)).

## 4. Product behavior

### dashboard and forge

android composes paired host inventories. each card carries machine identity,
terminal ref, optional current agent ref, herdr's agent status, name,
workspace, launch profile and deterministic landmark. one unavailable host does
not disable actions on another. the forge takes an explicit machine, validated
host directory, and terminal or that host's declared profile. creation returns
launch/partial facts, not readiness. a shell is a terminal choice and needs no
profile. [the directory chooser](working-directory-chooser.md),
[spaces](spaces.md), [shells](shells.md), [rename](session-renaming.md), and
[pressure](machine-pressure-rail.md) own their surface contracts.

### attachment

the phone attaches one exact terminal through the gateway websocket. initial
fitted geometry precedes public control acquisition; a valid full frame
precedes input. the phone applies it before admitting input to that attempt.
subsequent frames are sequenced and bounded. a disconnect ends only that
attachment. conflict/takeover is explicit and never replays input. terminal
text, paste, keys, scroll, resize, and detach have distinct meanings; terminal
output is never a source of outbound emulator replies.

### agent status and controls

status is herdr's `agent get` state. readiness is `ready` or `blocked` only
when `agent explain` matched that same visible screen rule for the same
unchanged state, and `unconfirmed` otherwise; skid adds no rule and reads no
terminal text. interrupt sends the provider's interrupt key (escape for codex,
ctrl-c for claude) through herdr's agent send-keys and returns dispatch
evidence, not provider effect. stop interrupts once, re-reads the original
target, then closes the terminal: the same agent, or none because it exited,
permits the close; another agent refuses it. an unknown interrupt does not
proceed to close. native close refusal remains refusal, and each partial says
how far the stop got.

## 5. Host architecture

`internal/herdr` alone speaks the pinned public socket and owns one
terminal-control child per attachment. `sessions` projects panes, workspaces
and agents, launches, performs validated mutations and agent controls, and
re-reads every target before its write. `profile` owns launch profiles;
`hostconfig` admits the deployment's host configuration, including through
`skidbladnir validate-host-config`, which dev-server calls before staging a
generation; validity is not runtime readiness. `reference` owns the opaque ref
encoding. `gateway` composes these with auth, pairing, pressure, directory,
http, and websocket lifetimes; `terminal` owns the typed stream frames.

refs are thin encodings of herdr ids. a terminal ref is its `terminal_id`, so
rename, move and gateway restart preserve it and a herdr restart makes it
stale. an agent ref adds herdr's agent name; before a write the gateway finds
the pane hosting that terminal and requires herdr to report the same agent
name there, which is herdr's own check in `agent start`. a workspace ref is
herdr's workspace id and only places a new tab or a moved pane. labels are for
observation and selection; they never authenticate a mutation. a workspace
with an invalid label is unaddressable and makes inventory partial.

`/v1/terminals` owns inventory, create, rename, move, shell, kill, and stream.
`/v1/agents/{ref}` owns interrupt and stop. machine/auth, pairing, directory,
and pressure retain their separate boundaries. successful host operations
expose observed/partial facts and `not_sent | sent | unknown` dispatch where
applicable. the ten-second host budget and fifteen-second client deadline bound
ordinary operations. herdr refuses `agent start` while a new pane's shell is
still starting; the gateway retries only that refusal, for two seconds, after
re-reading the pane.

one websocket attempt owns one control child. acquisition waits at most ten
seconds for geometry and first full frame; bearer revalidation and ping/pong
watch the admitted stream. backpressure, frame limits, and sequence errors end
the attachment, never the worker or server. background release invalidates
phone input admission before releasing the child. a new attempt starts with
fresh discovery. [the pr 2 stream contract](herdr-pr2.md#attached-terminal)
owns timing and frame details.

## 6. Android surface

`SkidbladnirController` owns selected targets, foreground state, attempt
generation, navigation, and inventory refresh. `TerminalConnection` owns one
socket; the terminal page owns geometry, renderer application, viewport-local
selection/copy, key deck state, and gestures. native code owns credentials and
transport; the WebView receives only bounded frame data and semantic input
callbacks. foreground loss invalidates input before callbacks can reopen it.

a card names its launch profile when herdr's agent is that profile's provider,
and the provider with an unknown profile otherwise. pairings survive this
source cutover. after a native close, the affected host inventory is refreshed
because linked closure may remove several cards. user-waived dictation, gboard
paste, local copy, and rotation checks remain `NOT_RUN` until observed.

## 7. Security

all gateway `/v1` operations require the host bearer and pinned machine handle;
websocket attachment carries the same authority. the selected exact ref is
validated at admission and again before control acquisition, and each input
resolves the original target. bearer revalidation stops a stream whose token
was replaced. phone WebView code never receives the bearer, raw herdr socket,
provider homes, or command paths.

pairing invitations are short-lived, single-use, and machine-bound. tailscale
serve provides tls to each loopback gateway. authentication is not inferred
from a caller-supplied label or from the terminal's native label. logs contain
only bounded typed operational facts, never content or credentials. explicit
clipboard copy uses the android system clipboard; it is a user action, not a
secrecy guarantee.

## 8. Upgrade ladder

this branch is [pr 5](herdr-pr5.md#8-delivery) step 3. step 4 pins it in
dev-server, rewrites the host configs to the reduced profile table, and removes
the skid link, hooks, notifier, plugin and jarvis cli copy. hosts update before
the phone. neither source implementation nor a synthetic probe is fleet
acceptance. [the roadmap](roadmap.md) records remaining live and phone proofs.

push, unread-result attention, provenance, copied provider history, durable
receipts and replay remain excluded. a new capability requires an explicit
scope and acceptance-criterion change; removal of legacy code does not
authorize its return.

## 9. Verification

[testing policy](rules/testing.md) owns temporary integration/live probes and
cleanup. `scripts/check verify` performs engineering checks and builds only;
it is not behavioral acceptance. unavailable devices, unexecuted gates, and
waived manual checks are `NOT_RUN`, never passes. each pr records its
boundaries' actual evidence and remaining blockers.
