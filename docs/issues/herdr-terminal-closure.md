# herdr terminal closure scope

problem: closing a pane can close its workspace when that was its last pane.
for a parent worktree, the inspected upstream path can also close linked
workspaces when close confirmation is disabled.

impact: translating skid kill/stop directly to pane close may destroy unrelated
terminals, including manually created work that skid did not organize.
excluding skid-managed worktrees does not remove this discovery/control case.

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
cannot make either command exact across a concurrent layout change. this is a
blocking capability gap at the pinned source, not a skid adapter defect.

resolution owner: [pr 1](../herdr-pr1.md). on isolated resources, close a sole
ordinary pane and a parent-worktree pane with a live linked workspace, under
both confirmation settings. test kill and stop's closure step through the
chosen adapter primitive, including layout change between lookup and dispatch.

resolved when: the accepted primitive closes only the confirmed terminal or
rejects before wider destruction. removing its resulting empty tab/workspace is
permitted; destroying another terminal is not. a preflight followed by an
unguarded cascading command is insufficient. qualify upstream support if needed;
do not reparent user panes, override user confirmation policy or silently widen
the confirmation to make the test pass.

smallest next decision: scope an upstream operation that atomically checks the
expected terminal lifetime and refuses closure if it would remove any other
terminal, independent of the user's confirmation setting. qualify a revision
containing that operation before reconsidering pr 2.
