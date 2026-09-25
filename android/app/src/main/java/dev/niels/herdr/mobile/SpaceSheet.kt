package dev.niels.herdr.mobile

import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material3.Button
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

@Composable
internal fun WorkspaceTextAction(label: String, enabled: Boolean, onClick: () -> Unit, description: String) {
    TextButton(onClick = onClick, enabled = enabled,
        modifier = Modifier.semantics { contentDescription = description }) { Text(label) }
}

@Composable
internal fun WorkspaceField(
    draft: WorkspaceDraft,
    workspaces: List<WorkspaceTarget>,
    machineHandle: MachineHandle?,
    enabled: Boolean,
    onChange: (WorkspaceDraft) -> Unit,
) {
    val available = workspaces.filter { it.machineHandle == machineHandle }
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text("Workspace", style = MaterialTheme.typography.labelLarge, color = Muted)
        Row(Modifier.horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            FilterChip(selected = draft is WorkspaceDraft.New, enabled = enabled,
                onClick = { onChange(WorkspaceDraft.New("")) },
                label = { Text("new workspace") })
            available.forEach { target ->
                FilterChip(selected = draft is WorkspaceDraft.Existing && draft.workspaceRef == target.workspace.ref,
                    enabled = enabled, onClick = { onChange(WorkspaceDraft.Existing(target.workspace.ref)) },
                    label = { Text("${target.workspace.label.text} · ${target.workspace.ref.takeLast(6)}") })
            }
        }
        if (draft is WorkspaceDraft.New) {
            OutlinedTextField(
                value = draft.label,
                onValueChange = { onChange(WorkspaceDraft.New(it)) },
                enabled = enabled,
                modifier = Modifier.fillMaxWidth(),
                label = { Text("new workspace label${if (draft.label.isEmpty()) " (terminal name)" else ""}") },
                singleLine = true,
                isError = draft.label.isNotEmpty() && WorkspaceLabel.fromDraft(draft.label) == null,
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun WorkspaceSheet(
    editor: WorkspaceEditor,
    machine: MachineState,
    workspaces: List<WorkspaceTarget>,
    onChange: (WorkspaceDraft) -> Unit,
    onDismiss: () -> Unit,
    onSubmit: () -> Unit,
) {
    val sending = editor.phase == WorkspacePhase.Sending
    ModalBottomSheet(onDismissRequest = onDismiss, containerColor = DeepSurface) {
        Column(Modifier.padding(horizontal = 20.dp, vertical = 20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("Move terminal", style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.SemiBold)
            Text("${terminalDisplayName(editor.target.terminal)} on ${machine.machine.label.text}")
            Text("Moving changes the native layout. An emptied workspace may close.", color = Muted)
            WorkspaceField(editor.destination, workspaces, machine.machine.handle, !sending, onChange)
            editor.error?.let { Text(it, color = noticeToneColor(NoticeTone.Failure)) }
            Button(onClick = onSubmit, enabled = workspaceSubmissionAdmissible(editor, machine),
                modifier = Modifier.fillMaxWidth()) { Text(if (sending) "moving…" else "Move terminal") }
            TextButton(onClick = onDismiss, enabled = !sending) { Text("Back") }
        }
    }
}

@Composable
internal fun WorkspaceSelector(
    selection: DashboardWorkspaceSelection,
    workspaces: List<WorkspaceTarget>,
    onSelect: (DashboardWorkspaceSelection) -> Unit,
) {
    Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()).padding(horizontal = 12.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        FilterChip(selected = selection == DashboardWorkspaceSelection.All,
            onClick = { onSelect(DashboardWorkspaceSelection.All) }, label = { Text("all workspaces") })
        if (selection is DashboardWorkspaceSelection.Exact && workspaces.none {
                it.machineHandle == selection.machineHandle && it.workspace.ref == selection.workspaceRef
            }) {
            FilterChip(selected = true, enabled = false, onClick = {},
                label = { Text("workspace no longer available") })
        }
        workspaces.forEach { target ->
            val choice = DashboardWorkspaceSelection.Exact(target.machineHandle, target.workspace.ref)
            FilterChip(selected = selection == choice, onClick = { onSelect(choice) },
                label = { Text("${target.workspace.label.text} · ${target.workspace.ref.takeLast(6)}") })
        }
    }
}
