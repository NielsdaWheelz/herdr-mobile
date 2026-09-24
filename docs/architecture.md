# skíðblaðnir: product and architecture

this document describes the herdr pr 2 source candidate. it is not a claim that
any host or phone has been upgraded. [the roadmap](roadmap.md) records delivery
and open acceptance; [pr 1](herdr-pr1.md) owns the exact public wire and upstream
mappings, and [pr 2](herdr-pr2.md) owns this cutover's implementation and gates.
accepted feature specifications own their detailed interactions. [codebase
rules](rules/index.md) own implementation conventions.

## 1. Philosophy

herdr owns terminal and process lifetimes, terminal emulation, workspaces,
layout, and sampled observations. providers own execution and history. skid
owns an authenticated, machine-bound projection and direct controls over one
host's herdr server. each client composes independent gateways; no gateway
coordinates another host. there is no application database or replay engine.

one dwarf is one live herdr terminal. a provider may be observed within that
terminal; its agent identity and readiness are separate facts. a terminal can
exist without an agent, and an agent ref can become stale while its terminal
remains live. phone and desktop address the same worker without a focus broker.

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
| profiles | empty or the closed ordered `personal`, `work`, `work2`, `claude-work` table; each row fixes provider, command, home, arguments, and foreground signature |
| state | herdr panes, workspaces, process facts, and reserved `skid_*` metadata are runtime truth; android persists encrypted pairings and local presentation preferences |
| host app | go gateway, public herdr socket/client, platform process and pressure observation |
| clients | bare local `skid` executes configured `herdr client`; fleet cli and phone use exact gateway refs |
| delivery | one coordinated release in pr 4; source changes here do not install, publish, or update the phone |
| trust | agents run as the host user; same-uid adversarial containment is out of scope |

new agent launches use deployment-owned permission bypass flags. callers choose
only a declared profile and validated cwd; they cannot supply a command, account
home, permission option, or objective-as-prompt. a zero-profile host still has
terminal creation and process-based agent discovery. launch metadata names a
candidate profile, never proves the account of a running process.

### product language

skíðblaðnir is the app; the dashboard presents dwarves from dvergatal, the
append-only character catalogue. a dwarf's landmark is independent of its
operator-owned terminal name. `agent` means an observed foreground provider
program. errors and destructive confirmations name the actual terminal and
possible linked-workspace effects rather than relying on the metaphor.

### guarantees

- detach releases only the attachment; kill requires confirmation and an exact
  current terminal ref. native close can also end linked workers, and the
  confirmation says so.
- validate names, workspace labels, cwd, refs, input, and encoded sizes at their
  owners. reject stale and ambiguous targets; never substitute a refreshed ref
  for a mutation target.
- host credentials stay on their host. phone pairings remain encrypted and
  machine-bound. logs and evidence contain no terminal bytes, prompts,
  objectives, tokens, account data, or clipboard text.
- restart rediscovers runtime state. it never replays input or claims completion.
  a cold herdr restart gives new terminal lifetimes and loses old worker metadata.

### non-goals

provider transcript storage or search, semantic task state, orchestration,
automatic resume/retry/replay, a persistent worker registry, cross-host move or
broadcast, remote terminal clicks/drags, multi-user isolation, and a second
terminal runtime are excluded. retired hook status/history machinery, sqlite
lifecycle facts, provenance, contract codegen, and proof ledgers stay retired.

## 3. Platform facts and accepted limits

herdr's public api supplies pane/workspace discovery, metadata, launch,
mutation, bounded reads, visual scroll, and a separate terminal-control child.
its v0.9.1 key parser does not accept home, end, insert, delete, or modified
page keys, and trims modified non-ascii whitespace. android keeps home/end
visible and disabled. unsupported hardware, desktop, and cli keys fail locally
with an explanation. unmodified page up/down use public visual scroll. no client
guesses raw escape sequences to recover an unsupported key.

native pane close may close a linked group or be refused. herdr does not offer
compare-and-set metadata or an atomic check-and-write action. skid rereads
before a metadata claim and reads back a write, but a foreign writer may still
interleave. desktop and api writers can coexist with phone control. claude
status/history/halt confirmation is unavailable through the accepted public
contract; a bounded terminal read is evidence, not a provider transcript.

## 4. Product behavior

### dashboard and forge

android composes paired host inventories. each card carries machine identity,
terminal ref, optional current agent ref, runtime facts, name, workspace, and
deterministic landmark. one unavailable host does not disable actions on another.
the forge takes an explicit machine, validated host directory, and terminal or
that host's declared profile. creation returns launch/partial facts, not
readiness. a shell is a terminal choice and needs no profile. [the directory
chooser](working-directory-chooser.md), [spaces](spaces.md), [shells](shells.md),
[rename](session-renaming.md), and [pressure](machine-pressure-rail.md) own their
surface contracts.

### attachment and agent controls

phone and fleet cli attach one exact terminal through the gateway websocket.
initial fitted geometry precedes public control acquisition; a valid full frame
precedes input. the phone applies it before admitting input to that attempt.
subsequent frames are sequenced and bounded. a disconnect ends only that
attachment. conflict/takeover is explicit and never replays input. terminal
text, paste, keys, scroll, resize, and detach have distinct meanings; terminal
output is never a source of outbound emulator replies.

[agent controls](agent-control.md) own sampled readiness, bounded reads,
ordinary send, explicit terminal send, key sequences, interrupt, and stop.
ordinary send requires independently recognized idle. `send --terminal` is an
explicit readiness override, still bound to the original current agent. send
returns dispatch evidence, not provider effect or task completion. stop
interrupts once, then revalidates before close; an unknown interrupt does not
proceed to close. native close refusal remains refusal.

### identity registration

[identity projection](agent-identity-projection.md) owns the content-free,
process-lifetime-bound `SessionStart` registration. deployment-owned codex and
claude hooks may register a documented session id after matching inherited pane
identity, foreground process group, process ancestry/start and profile facts.
hooks do not report status, activity, history, prompt payloads,
or completion. absent or stale registration does not block honest process and
terminal observation. a codex completion notifier may emit BEL as terminal-local
presentation; it stores no state and has no product authority.

## 5. Host architecture

`internal/herdr` alone speaks the pinned public socket and owns one
terminal-control child per attachment. `sessions` discovers panes/workspaces,
claims reserved lifetime metadata, resolves original refs, and performs
validated mutations. `agentcontrol` enriches observed terminals and enforces
agent action policy; it depends on sessions, never the reverse. `agentruntime`,
`process`, and `agenthook` own launch/foreground/registration facts. `gateway`
composes these concrete modules with auth, pairing, pressure, directory, http,
and websocket lifetimes. `fleetclient` and `agentcli` own peer routing, strict
result decoding, and exact selection. `terminal` and `terminalclient` own typed
stream frames and cancellable local tty input.
`skidbladnir validate-host-config` admits a deployment-owned configuration
through the same `hostconfig` loader without reading the runtime; dev-server
calls it before staging a generation. validity is not runtime readiness.

terminal, agent, and workspace refs encode different lifetimes. rename, move,
and gateway restart preserve a terminal ref; runtime restart makes it stale.
agent replacement or bearer rotation makes an agent ref stale. metadata
conflict/exhaustion makes a resource unaddressable and inventory partial. labels
are for observation and selection; they never authenticate a mutation. name
selection rejects duplicates. no empty-host or guessed ref projection occurs.

`/v1/terminals` owns inventory, create, info, rename, move, shell, kill, and
stream. `/v1/agents` owns read, send, keys, interrupt, and stop. machine/auth,
pairing, directory, and pressure retain their separate boundaries. successful
host operations expose observed/partial facts and `not_sent | sent | unknown`
dispatch where applicable. the ten-second host budget and fifteen-second client
deadline bound ordinary operations. [the pr 1 operation table](herdr-pr1.md#retained-gateway-operations)
is the exact wire contract.

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

pairings survive this source cutover. the existing directory, pressure, forge,
filter, dashboard, and terminal presentation surfaces remain. after a native
close, the affected host inventory is refreshed because linked closure may
remove several cards. user-waived dictation, gboard paste, local copy, and
rotation checks remain `NOT_RUN` until observed.

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

this branch implements the pr 2 source candidate. pr 3 must align jarvis's
strict consumer to the changed contract. pr 4 owns installation, coordinated
activation, and rollback. neither source implementation nor a synthetic probe
is fleet acceptance. [the roadmap](roadmap.md) records remaining live and phone
proofs.

push, unread-result attention, provenance, copied provider history, durable
receipts and replay remain excluded. a new capability requires an explicit
scope and acceptance-criterion change; removal of legacy code does not
authorize its return.

## 9. Verification

[testing policy](rules/testing.md) owns temporary integration/live probes and
cleanup. `scripts/check verify` performs engineering checks and builds only;
it is not behavioral acceptance. unavailable devices, unexecuted gates, and
waived manual checks are `NOT_RUN`, never passes. pr 1 and pr 2 record each
boundary's actual evidence and remaining blockers.
