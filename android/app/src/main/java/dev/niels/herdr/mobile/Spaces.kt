package dev.niels.herdr.mobile

import android.icu.text.Normalizer2
import android.icu.text.UnicodeSet

internal const val WORKSPACE_INVALID = "use 1–64 nfc characters; only interior ordinary spaces, without display controls."
internal const val WORKSPACE_OUTCOME_UNKNOWN = "move outcome unknown. inspect the current workspace before another attempt."

internal class WorkspaceLabel private constructor(val text: String) {
    companion object {
        private val normalizer = Normalizer2.getNFCInstance()
        private val repertoire = UnicodeSet("[:age=15.0:]").freeze()
        private fun normalize(candidate: String): String = buildString {
            var start = 0
            while (start < candidate.length) {
                val knownEnd = repertoire.span(candidate, start, UnicodeSet.SpanCondition.CONTAINED)
                append(normalizer.normalize(candidate.subSequence(start, knownEnd)))
                val end = repertoire.span(candidate, knownEnd, UnicodeSet.SpanCondition.NOT_CONTAINED)
                append(candidate, knownEnd, end)
                start = end
            }
        }
        fun parse(candidate: String): WorkspaceLabel? {
            if (candidate.isEmpty() || candidate.codePointCount(0, candidate.length) > 64 ||
                candidate.utf8ByteCountWithin(256) == null || normalize(candidate) != candidate ||
                candidate.hasDisplayUnsafeCodePoint() || candidate.first() == ' ' || candidate.last() == ' ' ||
                candidate.codePoints().anyMatch {
                    it == 0x00a0 || it == 0x1680 || it in 0x2000..0x200a ||
                        it == 0x202f || it == 0x205f || it == 0x3000
                }
            ) return null
            return WorkspaceLabel(candidate)
        }
        fun fromDraft(candidate: String): WorkspaceLabel? = parse(normalize(candidate))
    }
    override fun equals(other: Any?): Boolean = other is WorkspaceLabel && text == other.text
    override fun hashCode(): Int = text.hashCode()
}

internal data class WorkspaceTarget(val machineHandle: MachineHandle, val workspace: WorkspaceRecord)

internal sealed interface DashboardWorkspaceSelection {
    data object All : DashboardWorkspaceSelection
    data class Exact(val machineHandle: MachineHandle, val workspaceRef: String) : DashboardWorkspaceSelection
}

internal fun DashboardWorkspaceSelection.matches(target: TerminalTarget): Boolean = when (this) {
    DashboardWorkspaceSelection.All -> true
    is DashboardWorkspaceSelection.Exact ->
        target.machineHandle == machineHandle && target.terminal.workspaceRef == workspaceRef
}

internal fun DashboardWorkspaceSelection.displayLabel(workspaces: List<WorkspaceTarget>): String = when (this) {
    DashboardWorkspaceSelection.All -> "all workspaces"
    is DashboardWorkspaceSelection.Exact -> workspaces.singleOrNull {
        it.machineHandle == machineHandle && it.workspace.ref == workspaceRef
    }?.let { "workspace: ${it.workspace.label.text}" } ?: "workspace no longer available"
}

internal sealed interface WorkspaceDraft {
    data class Existing(val workspaceRef: String) : WorkspaceDraft
    data class New(val label: String) : WorkspaceDraft
}

internal fun observedWorkspaces(machines: List<MachineState>): List<WorkspaceTarget> = machines.flatMap { machine ->
    machine.inventory.lastSnapshot()?.inventory?.workspaces.orEmpty().map { workspace ->
        WorkspaceTarget(machine.machine.handle, workspace)
    }
}.sortedWith(compareBy<WorkspaceTarget> { it.workspace.label.text.lowercase() }
    .thenBy { it.machineHandle.encoded }.thenBy { it.workspace.ref })

internal sealed interface DashboardItem {
    val key: DashboardItemKey
    data class Heading(val workspace: WorkspaceTarget) : DashboardItem {
        override val key = DashboardItemKey.Workspace(workspace.machineHandle, workspace.workspace.ref)
    }
    data class Session(val visible: VisibleSession) : DashboardItem { override val key = visible.cardKey }
}

internal fun dashboardItems(
    machines: List<MachineState>,
    scope: DashboardScope,
    selection: DashboardWorkspaceSelection,
): List<DashboardItem> {
    val targets = visibleSessions(machines, scope).filter { selection.matches(it.target) }
    val workspaces = observedWorkspaces(machines).associateBy { it.machineHandle to it.workspace.ref }
    return targets.groupBy { it.target.machineHandle to it.target.terminal.workspaceRef }
        .entries.sortedWith(compareBy({ workspaces[it.key]?.workspace?.label?.text?.lowercase().orEmpty() },
            { it.key.first.encoded }, { it.key.second }))
        .flatMap { (key, terminals) ->
            val workspace = workspaces[key] ?: return@flatMap emptyList()
            listOf(DashboardItem.Heading(workspace)) + terminals.map(DashboardItem::Session)
        }
}

internal sealed interface WorkspacePhase {
    data object Editing : WorkspacePhase
    data object Sending : WorkspacePhase
    data class Checking(val acknowledged: Boolean) : WorkspacePhase
}

internal data class WorkspaceEditor(
    val target: TerminalTarget,
    val destination: WorkspaceDraft,
    val phase: WorkspacePhase = WorkspacePhase.Editing,
    val error: String? = null,
)

internal fun sameTerminalLifetime(first: TerminalTarget, second: TerminalTarget): Boolean =
    first.machineHandle == second.machineHandle && first.terminal.ref == second.terminal.ref

internal fun workspaceSubmissionAdmissible(editor: WorkspaceEditor, machine: MachineState): Boolean {
    if (editor.phase != WorkspacePhase.Editing || !machine.canMutate ||
        machine.machine.handle != editor.target.machineHandle) return false
    val inventory = machine.inventory.lastSnapshot()?.inventory ?: return false
    val current = inventory.terminals.singleOrNull { it.ref == editor.target.terminal.ref } ?: return false
    return when (val destination = editor.destination) {
        is WorkspaceDraft.Existing -> destination.workspaceRef != current.workspaceRef &&
            inventory.workspaces.any { it.ref == destination.workspaceRef }
        is WorkspaceDraft.New -> WorkspaceLabel.fromDraft(destination.label) != null
    }
}

internal fun completeWorkspaceHttp(editor: WorkspaceEditor, result: GatewayResult<ObservedTerminal>): WorkspaceEditor =
    when (result) {
        is GatewayResult.Success -> editor.copy(phase = WorkspacePhase.Checking(true), error = null)
        is GatewayResult.Failure -> {
            val failure = result.failure
            val notSent = failure is GatewayFailure.Api && failure.dispatch == MutationDispatch.NotSent
            editor.copy(
                phase = if (notSent) WorkspacePhase.Editing else WorkspacePhase.Checking(false),
                error = if (notSent) gatewayFailureMessage(failure) else WORKSPACE_OUTCOME_UNKNOWN,
            )
        }
    }

internal fun reconcileWorkspaceEditor(editor: WorkspaceEditor, machine: MachineState): WorkspaceEditor? {
    if (editor.phase == WorkspacePhase.Sending) return editor
    val current = machine.inventory.lastSnapshot()?.inventory?.terminals?.singleOrNull {
        it.ref == editor.target.terminal.ref
    } ?: return null
    return when (val phase = editor.phase) {
        WorkspacePhase.Sending -> error("sending is handled before inventory reconciliation")
        WorkspacePhase.Editing -> editor.copy(target = editor.target.copy(terminal = current))
        is WorkspacePhase.Checking -> if (phase.acknowledged) null else editor.copy(
            target = editor.target.copy(terminal = current), phase = WorkspacePhase.Editing,
        )
    }
}
