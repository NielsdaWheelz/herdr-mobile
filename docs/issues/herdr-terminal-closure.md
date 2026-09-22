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
skid live reproduction remains `NOT_RUN`.

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
