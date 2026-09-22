# herdr native closure disclosure and qualification

problem: phone confirmations and cli/jarvis contracts have not yet been qualified
for herdr's native close effects and refusals. closing a pane can close its
workspace and linked git-worktree group, including other running terminals.

accepted 2026-09-22: the user approved these native effects with explicit disclosure
under [the scope amendment](../herdr-pr1.md#accepted-scope-amendment). exact-only
closure and a new atomic upstream close primitive are no longer requirements.
this issue tracks the remaining disclosure/mapping/proof work, not a demand to
remove native group semantics.

impact: preserving old single-terminal wording would misstate what kill/stop can
do. group members are independent terminals in related git checkouts; the link
does not mean they are views of the same worker. detach remains non-destructive.

evidence: release `065ef9d6a531c49fb8bee7e818ef837065b21ee9` routes
[last-pane closure](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/app/api/panes.rs#L1862)
through workspace closure. the upstream
[confirmation-disabled case](https://github.com/herdrdev/herdr/blob/065ef9d6a531c49fb8bee7e818ef837065b21ee9/src/app/actions.rs#L4246)
expects parent and linked workspaces to disappear. this is source evidence;
the pinned darwin arm64 binary reproduced this on an isolated server: closing
the parent's sole pane with `confirm_close=false` removed the linked terminal's
original identity while an unrelated workspace survived. the disposable
assertion failed with exit 1. on a separate isolated server,
`confirm_close=true` made `pane.close` return `confirmation_required`, and
`workspace.close(close_group=false)` returned `workspace_group_close_required`;
the parent and linked exact terminal ids survived. linux and layout-race
boundaries remain `NOT_RUN`.

pr 1 source review confirms that `pane.close` accepts a pane id without an
expected terminal id. on last-pane closure it calls `close_selected_workspace`,
whose linked-worktree guard depends on `confirm_close`. with confirmation off,
it can remove the parent and linked workspaces. `workspace.close` rejects a
linked group when `close_group=false`, but closes every terminal in its target
workspace and has no expected layout or terminal predicate. a gateway preflight
cannot make either command exact across a concurrent layout change. the original
`FAIL` remains evidence against the former exact-only requirement; accepting
native scope is not a rerun or a new behavioral pass.

resolution owner: [pr 1](../herdr-pr1.md). on isolated resources, close a sole
ordinary pane and a parent-worktree pane with a live linked workspace, under
both confirmation settings. test kill and stop's closure step through the
chosen adapter primitive, including layout change between lookup and dispatch;
confirm terminals outside the native group survive. revalidate the original
target without implicit substitution. no atomic target/effect-set claim.

resolved when: the public operation mapping, phone close/stop confirmation and
cli/jarvis tool contract disclose possible native group effects before dispatch;
the real-boundary proofs establish those effects or explicit refusal. preserve
upstream confirmation policy: `confirmation_required` must not trigger a broader
command, configuration change or retry. returned outcomes distinguish requested
target closure, provider halt and uncertainty; never claim an exact affected set
or confirmed halt of every linked worker. checkout directories/branches remain,
and detach leaves workers running.

next step: qualify the approved semantics and disclosure at the pinned revision.
no upstream exact-terminal close is required solely to resolve the former gate.
