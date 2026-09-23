package dev.niels.skidbladnir

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.minimumInteractiveComponentSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.semantics.stateDescription
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp

private val terminalNamePattern = Regex("[A-Za-z0-9][A-Za-z0-9_-]{0,63}")

internal sealed interface RenamePhase {
    data class Editing(val stale: Boolean = false) : RenamePhase
    data object Sending : RenamePhase
    data class Reconciling(val sheetVisible: Boolean) : RenamePhase
}

internal data class RenameState(
    val target: TerminalTarget,
    val draft: String,
    val phase: RenamePhase,
    val error: String?,
)

internal data class RenameHttpTransition(
    val state: RenameState,
    val clearMutationFence: Boolean,
    val requireInventoryRead: Boolean,
)

internal data class TerminalRenameInventoryResult(
    val terminal: SkidbladnirUiState.Terminal,
    val detachTransport: Boolean,
)

internal const val RENAME_OUTCOME_UNKNOWN = "Rename outcome unknown. Checking the terminal."
internal const val RENAME_STALE_EDIT = "The terminal name changed. Review and try again."

internal fun isValidTerminalName(candidate: String): Boolean = terminalNamePattern.matches(candidate)

internal fun beginRename(target: TerminalTarget): RenameState = RenameState(
    target = target,
    draft = target.terminal.name.orEmpty(),
    phase = RenamePhase.Editing(),
    error = null,
)

internal fun updateRenameDraft(state: RenameState, draft: String): RenameState {
    if (state.phase !is RenamePhase.Editing) return state
    return state.copy(draft = draft, phase = RenamePhase.Editing(), error = null)
}

internal fun renameSubmissionAdmissible(
    state: RenameState,
    terminalTarget: TerminalTarget,
    terminalActionsAdmissible: Boolean,
): Boolean {
    val phase = state.phase as? RenamePhase.Editing ?: return false
    return terminalActionsAdmissible && !phase.stale &&
        sameSessionAuthority(state.target, terminalTarget) &&
        state.draft != state.target.terminal.name.orEmpty() && isValidTerminalName(state.draft)
}

internal fun beginRenameSending(
    state: RenameState,
    terminalTarget: TerminalTarget,
    terminalActionsAdmissible: Boolean,
): RenameState? = if (renameSubmissionAdmissible(state, terminalTarget, terminalActionsAdmissible)) {
    state.copy(phase = RenamePhase.Sending, error = null)
} else {
    null
}

internal fun dismissRename(state: RenameState): RenameState? = when (val phase = state.phase) {
    is RenamePhase.Editing -> null
    RenamePhase.Sending -> state
    is RenamePhase.Reconciling -> state.copy(phase = phase.copy(sheetVisible = false))
}

internal fun completeRenameHttp(
    state: RenameState,
    result: GatewayResult<ObservedTerminal>,
): RenameHttpTransition = when (result) {
    is GatewayResult.Success -> RenameHttpTransition(
        state = state.copy(phase = RenamePhase.Reconciling(sheetVisible = true), error = null),
        clearMutationFence = false,
        requireInventoryRead = true,
    )
    is GatewayResult.Failure -> when (val failure = result.failure) {
        GatewayFailure.Transport -> renameNeedsInventory(state, RENAME_OUTCOME_UNKNOWN)
        is GatewayFailure.Api -> when {
            failure.code in setOf(ApiErrorCode.Unauthenticated, ApiErrorCode.MachineIdentityMismatch) ->
                error("rename access failure escaped the controller access owner")
            failure.dispatch == MutationDispatch.NotSent && failure.code in setOf(
                ApiErrorCode.InvalidRequest, ApiErrorCode.RequestTooLarge, ApiErrorCode.NameInvalid,
            ) -> RenameHttpTransition(
                state.copy(phase = RenamePhase.Editing(), error = gatewayFailureMessage(failure)),
                clearMutationFence = true, requireInventoryRead = false,
            )
            else -> renameNeedsInventory(state, gatewayFailureMessage(failure))
        }
    }
}

private fun renameNeedsInventory(state: RenameState, error: String?): RenameHttpTransition =
    RenameHttpTransition(
        state = state.copy(
            phase = RenamePhase.Reconciling(sheetVisible = true),
            error = error,
        ),
        clearMutationFence = false,
        requireInventoryRead = true,
    )

internal fun clearMetadataMutationFence(inventory: InventoryState, fence: Long): InventoryState =
    if (inventory is InventoryState.Superseded && inventory.requiredMutationFence == fence) {
        InventoryState.Fresh(inventory.snapshot)
    } else {
        inventory
    }

internal fun reconcileTerminalRename(
    terminal: SkidbladnirUiState.Terminal,
): TerminalRenameInventoryResult {
    val snapshot = terminal.machine.inventory.lastSnapshot()?.inventory
        ?: return TerminalRenameInventoryResult(terminal, detachTransport = false)
    val authoritative = snapshot.terminals.singleOrNull {
        it.ref == terminal.target.terminal.ref
    }
    val rename = terminal.rename
    if (authoritative == null) {
        if (rename == null) return TerminalRenameInventoryResult(terminal, detachTransport = false)
        return TerminalRenameInventoryResult(
            terminal = terminal.copy(
                rename = null,
                connection = TerminalUiStatus.ReconnectRequired(
                    "${terminal.machine.machine.label.text}: that terminal lifetime is no longer available.",
                ),
            ),
            detachTransport = true,
        )
    }

    val authoritativeTarget = TerminalTarget(terminal.target.machineHandle,
        authoritative.copy(agent = terminal.target.terminal.agent))
    if (rename == null) {
        return TerminalRenameInventoryResult(
            terminal.copy(target = authoritativeTarget),
            detachTransport = false,
        )
    }
    val resolvedRename = when (val phase = rename.phase) {
        RenamePhase.Sending -> rename
        is RenamePhase.Editing -> if (sameSessionAuthority(rename.target, authoritativeTarget)) {
            rename
        } else {
            rename.copy(
                target = authoritativeTarget,
                phase = RenamePhase.Editing(stale = true),
                error = RENAME_STALE_EDIT,
            )
        }
        is RenamePhase.Reconciling -> when {
            authoritative.name == rename.draft -> null
            phase.sheetVisible -> rename.copy(
                target = authoritativeTarget,
                phase = RenamePhase.Editing(stale = true),
                error = RENAME_STALE_EDIT,
            )
            else -> null
        }
    }
    return TerminalRenameInventoryResult(
        terminal = terminal.copy(target = authoritativeTarget, rename = resolvedRename),
        detachTransport = false,
    )
}

private fun sameSessionAuthority(first: TerminalTarget, second: TerminalTarget): Boolean =
    first.machineHandle == second.machineHandle && first.terminal.ref == second.terminal.ref &&
        first.terminal.name == second.terminal.name

@Composable
internal fun TerminalRenameControl(
    machine: PairedMachine,
    target: TerminalTarget,
    presence: String,
    presenceColor: Color,
    enabled: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val interactionSource = remember { MutableInteractionSource() }
    Surface(
        color = DeepSurface,
        border = BorderStroke(1.dp, (if (enabled) Gold else Muted).copy(alpha = 0.40f)),
        shape = NidavellirShapes.Chip,
        modifier = modifier
            .clickable(
                interactionSource = interactionSource,
                indication = AngularIndication(NidavellirShapes.Chip),
                enabled = enabled,
                role = Role.Button,
                onClick = onClick,
            )
            .semantics(mergeDescendants = true) {
                contentDescription = "Rename ${terminalDisplayName(target.terminal)} on ${machine.label.text}"
                stateDescription = presence
            },
    ) {
        Column(
            modifier = Modifier
                .minimumInteractiveComponentSize()
                .padding(horizontal = 12.dp, vertical = 4.dp),
            verticalArrangement = Arrangement.Center,
        ) {
            Text(
                text = "${machine.label.text} · ${terminalDisplayName(target.terminal)}",
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Row {
                Text(
                    text = "Rename",
                    color = presenceColor,
                    style = MaterialTheme.typography.labelSmall,
                    fontFamily = NidavellirType.Data,
                    maxLines = 1,
                )
                Text(
                    text = " · ",
                    color = presenceColor,
                    style = MaterialTheme.typography.labelSmall,
                    fontFamily = NidavellirType.Data,
                    maxLines = 1,
                )
                Text(
                    text = "${machine.label.text} · $presence",
                    color = presenceColor,
                    style = MaterialTheme.typography.labelSmall,
                    fontFamily = NidavellirType.Data,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
internal fun SessionRenameSheet(
    machine: PairedMachine,
    terminalTarget: TerminalTarget,
    state: RenameState,
    terminalActionsAdmissible: Boolean,
    onDraftChange: (String) -> Unit,
    onDismiss: () -> Unit,
    onSubmit: () -> Unit,
) {
    val phase = state.phase
    if (phase is RenamePhase.Reconciling && !phase.sheetVisible) return
    val fieldsEnabled = phase is RenamePhase.Editing
    val canSubmit = renameSubmissionAdmissible(state, terminalTarget, terminalActionsAdmissible)
    val focusRequester = remember { FocusRequester() }
    LaunchedEffect(state.target, fieldsEnabled) {
        if (fieldsEnabled) focusRequester.requestFocus()
    }
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        shape = NidavellirShapes.Sheet,
        containerColor = DeepSurface,
        sheetGesturesEnabled = phase != RenamePhase.Sending,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .verticalScroll(rememberScrollState())
                .imePadding()
                .padding(horizontal = 20.dp)
                .padding(bottom = 28.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = "Rename terminal",
                style = MaterialTheme.typography.headlineSmall,
                fontFamily = NidavellirType.Display,
                fontWeight = FontWeight.SemiBold,
            )
            Text(
                text = "${terminalDisplayName(state.target.terminal)} on ${machine.label.text}",
                color = Muted,
                fontFamily = NidavellirType.Data,
            )
            OutlinedTextField(
                value = state.draft,
                onValueChange = onDraftChange,
                enabled = fieldsEnabled,
                singleLine = true,
                isError = state.error != null,
                label = { Text("Tmux name") },
                keyboardOptions = KeyboardOptions(
                    capitalization = KeyboardCapitalization.None,
                    autoCorrectEnabled = false,
                    keyboardType = KeyboardType.Ascii,
                    imeAction = ImeAction.Done,
                ),
                keyboardActions = KeyboardActions(onDone = { if (canSubmit) onSubmit() }),
                modifier = Modifier
                    .fillMaxWidth()
                    .focusRequester(focusRequester),
            )
            Text(
                text = "1–64 letters, numbers, underscores, or hyphens",
                color = Muted,
                style = MaterialTheme.typography.labelSmall,
            )
            state.error?.let { error ->
                Text(
                    text = error,
                    color = noticeToneColor(NoticeTone.Failure),
                )
            }
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                OutlinedButton(
                    onClick = onDismiss,
                    enabled = phase != RenamePhase.Sending,
                ) {
                    Text("Cancel")
                }
                Spacer(Modifier.width(8.dp))
                Button(
                    onClick = onSubmit,
                    enabled = canSubmit,
                    shape = NidavellirShapes.Chip,
                ) {
                    if (phase == RenamePhase.Sending) {
                        CircularProgressIndicator(Modifier.size(18.dp), strokeWidth = 2.dp)
                        Spacer(Modifier.width(8.dp))
                    }
                    Text("Rename")
                }
            }
        }
    }
}
