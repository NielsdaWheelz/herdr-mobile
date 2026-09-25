package dev.niels.herdr.mobile

import android.os.Bundle
import androidx.compose.foundation.lazy.grid.LazyGridState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.savedstate.SavedStateRegistry

internal sealed interface DashboardScope {
    data object All : DashboardScope
    data class Machine(val handle: MachineHandle) : DashboardScope
}

internal sealed interface DashboardItemKey {
    val encoded: String
    data class Workspace(val machine: MachineHandle, val ref: String) : DashboardItemKey {
        override val encoded: String get() = "workspace:${machine.encoded}:$ref"
    }
    companion object {
        fun parse(encoded: String): DashboardItemKey? = when {
            DashboardCardKey.isFingerprint(encoded) -> DashboardCardKey(encoded)
            encoded.startsWith("workspace:") -> {
                val parts = encoded.split(':', limit = 3)
                val machine = parts.getOrNull(1)?.let(MachineHandle::parse)
                val ref = parts.getOrNull(2)
                if (machine == null || ref.isNullOrEmpty()) null else Workspace(machine, ref)
            }
            else -> null
        }
    }
}

@JvmInline
internal value class DashboardCardKey(val lifetimeFingerprint: String) : DashboardItemKey {
    override val encoded: String get() = lifetimeFingerprint
    init { require(isFingerprint(lifetimeFingerprint)) }
    companion object {
        private val FINGERPRINT = Regex("[0-9a-f]{64}")
        fun isFingerprint(value: String): Boolean = FINGERPRINT.matches(value)
    }
}

internal data class DashboardViewport(val anchor: DashboardItemKey?, val fallbackIndex: Int, val offsetPx: Int) {
    init {
        require(fallbackIndex >= 0 && offsetPx >= 0)
        require(anchor != null || fallbackIndex == 0 && offsetPx == 0)
    }
}

internal data class DashboardEntrySnapshot(
    val schemaVersion: Int,
    val scope: DashboardScope,
    val viewport: DashboardViewport,
) {
    init { require(schemaVersion == 3) }
}

internal class DashboardEntryState(restoredSnapshot: DashboardEntrySnapshot? = null) {
    private var currentScope by mutableStateOf(restoredSnapshot?.scope ?: DashboardScope.All)
    private var currentWorkspace by mutableStateOf<DashboardWorkspaceSelection>(DashboardWorkspaceSelection.All)
    private var pendingSnapshot by mutableStateOf(restoredSnapshot)
    private var ownedGridState by mutableStateOf(LazyGridState())
    private var acceptedHandles: Set<MachineHandle>? = null
    private var installed = false

    val scope: DashboardScope get() = currentScope
    val workspace: DashboardWorkspaceSelection get() = currentWorkspace
    val gridState: LazyGridState get() = ownedGridState
    val restorationPending: Boolean get() = pendingSnapshot != null

    fun acceptFleet(handles: Set<MachineHandle>) {
        acceptedHandles = handles.toSet()
        val scopeAccepted = when (val scope = currentScope) {
            DashboardScope.All -> true
            is DashboardScope.Machine -> scope.handle in handles
        }
        if (handles.isEmpty() || !scopeAccepted) resetAll()
    }

    fun selectScope(scope: DashboardScope) {
        if (scope == currentScope) return
        when (scope) {
            DashboardScope.All -> Unit
            is DashboardScope.Machine -> require(scope.handle in checkNotNull(acceptedHandles))
        }
        currentScope = scope
        currentWorkspace = DashboardWorkspaceSelection.All
        pendingSnapshot = null
    }

    fun selectWorkspace(selection: DashboardWorkspaceSelection) {
        if (selection == currentWorkspace) return
        currentWorkspace = selection
        pendingSnapshot = null
        ownedGridState = LazyGridState()
    }

    fun followCreatedMembership(target: TerminalTarget) {
        val selection = DashboardWorkspaceSelection.Exact(target.machineHandle, target.terminal.workspaceRef)
        if (currentWorkspace != selection) selectWorkspace(selection)
    }

    fun selectTerminalAccessLoss(handle: MachineHandle) {
        require(handle in checkNotNull(acceptedHandles))
        currentScope = DashboardScope.Machine(handle)
        currentWorkspace = DashboardWorkspaceSelection.All
        pendingSnapshot = null
        ownedGridState = LazyGridState()
    }

    fun resetAll() {
        currentScope = DashboardScope.All
        currentWorkspace = DashboardWorkspaceSelection.All
        pendingSnapshot = null
        ownedGridState = LazyGridState()
    }

    fun restoreOnce(keys: List<DashboardItemKey>) {
        val restored = pendingSnapshot ?: return
        if (keys.isNotEmpty()) {
            val index = restored.viewport.anchor?.let(keys::indexOf)?.takeIf { it >= 0 }
                ?: restored.viewport.fallbackIndex.coerceAtMost(keys.lastIndex)
            gridState.requestScrollToItem(index, restored.viewport.offsetPx)
        }
        pendingSnapshot = null
    }

    fun snapshot(): DashboardEntrySnapshot {
        pendingSnapshot?.let { return it }
        val index = gridState.firstVisibleItemIndex
        val key = gridState.layoutInfo.visibleItemsInfo.singleOrNull { it.index == index }?.key as? String
        val anchor = key?.let(DashboardItemKey::parse)
        return DashboardEntrySnapshot(3, currentScope, if (anchor == null) TOP_VIEWPORT else
            DashboardViewport(anchor, index, gridState.firstVisibleItemScrollOffset))
    }

    fun install(savedStateRegistry: SavedStateRegistry) {
        check(!installed)
        installed = true
        savedStateRegistry.consumeRestoredStateForKey(REGISTRY_KEY)?.let { encoded ->
            val restored = decodeSnapshot(encoded)
            if (restored == null) resetAll() else {
                currentScope = restored.scope
                pendingSnapshot = restored
            }
        }
        savedStateRegistry.registerSavedStateProvider(REGISTRY_KEY) { encodeSnapshot(snapshot()) }
    }

    private companion object {
        const val REGISTRY_KEY = "dev.niels.herdr.mobile.dashboard-entry"
        val TOP_VIEWPORT = DashboardViewport(null, 0, 0)

        fun encodeSnapshot(snapshot: DashboardEntrySnapshot): Bundle = Bundle().apply {
            putInt("version", 3)
            when (val scope = snapshot.scope) {
                DashboardScope.All -> putString("scopeKind", "all")
                is DashboardScope.Machine -> {
                    putString("scopeKind", "machine")
                    putString("scopeMachine", scope.handle.encoded)
                }
            }
            putString("anchor", snapshot.viewport.anchor?.encoded.orEmpty())
            putInt("fallbackIndex", snapshot.viewport.fallbackIndex)
            putInt("offsetPx", snapshot.viewport.offsetPx)
        }

        fun decodeSnapshot(encoded: Bundle): DashboardEntrySnapshot? {
            if (encoded.requiredInt("version") != 3) return null
            val keys = mutableSetOf("version", "scopeKind", "anchor", "fallbackIndex", "offsetPx")
            val scope = when (encoded.requiredString("scopeKind")) {
                "all" -> DashboardScope.All
                "machine" -> {
                    keys += "scopeMachine"
                    DashboardScope.Machine(checkNotNull(MachineHandle.parse(encoded.requiredString("scopeMachine"))))
                }
                else -> error("invalid dashboard scope")
            }
            check(encoded.keySet() == keys)
            val anchorText = encoded.requiredString("anchor")
            val anchor = if (anchorText.isEmpty()) null else checkNotNull(DashboardItemKey.parse(anchorText))
            return DashboardEntrySnapshot(3, scope,
                DashboardViewport(anchor, encoded.requiredInt("fallbackIndex"), encoded.requiredInt("offsetPx")))
        }

        fun Bundle.requiredInt(key: String): Int {
            check(containsKey(key))
            val lower = getInt(key, Int.MIN_VALUE)
            val upper = getInt(key, Int.MAX_VALUE)
            check(lower == upper)
            return lower
        }
        fun Bundle.requiredString(key: String): String {
            check(containsKey(key))
            return checkNotNull(getString(key))
        }
    }
}
