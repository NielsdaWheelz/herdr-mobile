package dev.niels.skidbladnir

import java.net.URI
import java.net.URISyntaxException
import java.nio.charset.StandardCharsets
import java.security.MessageDigest
import java.time.DateTimeException
import java.time.Instant
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter
import java.util.Locale
import java.util.Base64
import kotlinx.serialization.KSerializer
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.PrimitiveSerialDescriptor
import kotlinx.serialization.encodeToString
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.decodeFromJsonElement
import kotlinx.serialization.json.jsonArray

internal val productJson = Json { explicitNulls = false; ignoreUnknownKeys = false }

// justify-defect: the app and the gateway own one closed wire schema, so an undecodable protocol
// payload is a same-system contract violation. The reason is content-free by construction — a fixed
// literal or a failure class name, never the offending payload or a cause that embeds it — so the
// architecture §7 credential-free/content-free log guarantee holds even when the platform prints
// this fatal defect.
internal class ProtocolDecodeException(reason: String) :
    RuntimeException("Protocol payload could not be decoded: $reason.")

internal inline fun <Value> decodeProtocol(block: () -> Value): Value = try {
    block()
} catch (failure: ProtocolDecodeException) {
    throw failure
} catch (failure: SerializationException) {
    throw ProtocolDecodeException(failure.javaClass.simpleName)
} catch (failure: DateTimeException) {
    throw ProtocolDecodeException(failure.javaClass.simpleName)
} catch (failure: NoSuchElementException) {
    throw ProtocolDecodeException(failure.javaClass.simpleName)
} catch (failure: IllegalArgumentException) {
    throw ProtocolDecodeException(failure.javaClass.simpleName)
}

// The gateway encodes every protocol instant with Go's RFC3339Nano against a
// UTC clock, so exactly one shape is legal. ISO_INSTANT would also accept an
// offset form and a lower-case designator, which would let two encodings of the
// same moment cross a boundary the hard cut declares closed.
private val wireInstantPattern = Regex(
    """[0-9]{4}-[0-9]{2}-[0-9]{2}T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](?:\.[0-9]{0,8}[1-9])?Z""",
)
private val wireInstantBaseFormatter =
    DateTimeFormatter.ofPattern("uuuu-MM-dd'T'HH:mm:ss", Locale.ROOT).withZone(ZoneOffset.UTC)

internal fun acceptWireInstant(text: String): Instant {
    if (!wireInstantPattern.matches(text)) throw SerializationException("instant is not the canonical UTC wire form")
    return Instant.parse(text)
}

internal fun formatWireInstant(value: Instant): String {
    val base = try {
        wireInstantBaseFormatter.format(value)
    } catch (_: DateTimeException) {
        throw SerializationException("instant is outside the canonical UTC wire range")
    }
    val fraction = value.nano.takeIf { it != 0 }
        ?.toString()
        ?.padStart(9, '0')
        ?.trimEnd('0')
        ?.let { ".$it" }
        .orEmpty()
    val encoded = "$base${fraction}Z"
    if (!wireInstantPattern.matches(encoded)) {
        throw SerializationException("instant is outside the canonical UTC wire range")
    }
    return encoded
}

internal object IsoInstantSerializer : KSerializer<Instant> {
    override val descriptor =
        PrimitiveSerialDescriptor("dev.niels.skidbladnir.IsoInstant", PrimitiveKind.STRING)
    override fun deserialize(decoder: Decoder): Instant = acceptWireInstant(decoder.decodeString())
    override fun serialize(encoder: Encoder, value: Instant) = encoder.encodeString(formatWireInstant(value))
}

// Go's zero time is syntactically valid RFC 3339, but it cannot represent a captured projection.
private val unsetProjectionInstant = Instant.parse("0001-01-01T00:00:00Z")

internal class MachineHandle private constructor(val encoded: String) {
    companion object {
        private val pattern = Regex("mh-[0-9a-f]{32}")
        fun parse(candidate: String): MachineHandle? = candidate.takeIf(pattern::matches)?.let(::MachineHandle)
    }
    override fun equals(other: Any?): Boolean = other is MachineHandle && encoded == other.encoded
    override fun hashCode(): Int = encoded.hashCode()
    override fun toString(): String = encoded
}

internal class MachineLabel private constructor(val text: String) {
    companion object {
        fun parse(candidate: String): MachineLabel? {
            if (candidate.isEmpty() || candidate.length > 40 || candidate != candidate.trim()) return null
            if (candidate.hasDisplayUnsafeCodePoint()) return null
            return MachineLabel(candidate)
        }
    }
    override fun equals(other: Any?): Boolean = other is MachineLabel && text == other.text
    override fun hashCode(): Int = text.hashCode()
    override fun toString(): String = text
}

internal fun String.hasDisplayUnsafeCodePoint(): Boolean = codePoints().anyMatch { codePoint ->
    Character.isISOControl(codePoint) ||
        codePoint in 0xd800..0xdfff ||
        codePoint == 0x061c ||
        codePoint in 0x200e..0x200f ||
        codePoint in 0x2028..0x202e ||
        codePoint in 0x2066..0x2069
}

internal fun String.utf8ByteCountWithin(maximum: Int): Int? {
    if (length > maximum) return null
    var byteCount = 0
    var index = 0
    while (index < length) {
        val character = this[index]
        val increment = when {
            character.code <= 0x7f -> 1
            character.code <= 0x7ff -> 2
            character.isHighSurrogate() &&
                index + 1 < length &&
                this[index + 1].isLowSurrogate() -> {
                index += 1
                4
            }
            character.isSurrogate() -> return null
            else -> 3
        }
        byteCount += increment
        if (byteCount > maximum) return null
        index += 1
    }
    return byteCount
}

internal class MachineOrigin private constructor(val encoded: String) {
    companion object {
        fun parse(candidate: String): MachineOrigin? {
            val uri = try {
                URI(candidate)
            } catch (_: URISyntaxException) {
                // justify-ignore-error: an origin that is not even a URI is simply not an origin; the
                // only classification this parser owns is accepted or rejected.
                return null
            }
            if (uri.scheme != "https" || uri.host.isNullOrEmpty() || uri.port != 8443) return null
            if (uri.rawUserInfo != null || uri.rawQuery != null || uri.rawFragment != null) return null
            if (uri.rawPath !in setOf("", "/") || candidate.any(Char::isWhitespace)) return null
            // URI.getHost() already returns the RFC 2732 bracketed form for an IPv6 literal, so the
            // canonical authority is the lowercased host verbatim. Re-bracketing it would produce a
            // value that neither this parser nor OkHttp can read back.
            return MachineOrigin("https://${uri.host.lowercase(Locale.ROOT)}:8443/")
        }
    }
    override fun equals(other: Any?): Boolean = other is MachineOrigin && encoded == other.encoded
    override fun hashCode(): Int = encoded.hashCode()
    override fun toString(): String = encoded
}

internal class ProfileKey private constructor(val encoded: String) {
    companion object {
        private val pattern = Regex("[a-z][a-z0-9_-]{0,31}")
        fun parse(candidate: String): ProfileKey? =
            candidate.takeIf(pattern::matches)?.let(::ProfileKey)
    }
    override fun equals(other: Any?): Boolean = other is ProfileKey && encoded == other.encoded
    override fun hashCode(): Int = encoded.hashCode()
    override fun toString(): String = encoded
}

internal data class PairedMachine(val handle: MachineHandle, val label: MachineLabel, val origin: MachineOrigin)
internal data class TerminalTarget(val machineHandle: MachineHandle, val terminal: TerminalRecord)
internal enum class MachinePlatform { Linux, Darwin }
internal data class MachineSummary(val handle: MachineHandle, val platform: MachinePlatform)
@Serializable internal enum class AgentProvider { Codex, Claude }
internal data class ProfileChoice(val key: ProfileKey, val label: String, val provider: AgentProvider)

@Serializable private data class WireMachineSummary(val handle: String, val platform: WireMachinePlatform)
@Serializable private enum class WireMachinePlatform { Linux, Darwin }
@Serializable private data class WireProfileChoice(
    val key: String,
    val label: String,
    val provider: AgentProvider,
)
@Serializable internal data class CharacterSummary(val key: String, val displayName: String)

internal data class AgentRuntime(
    val ref: String,
    val provider: AgentProvider,
    val status: AgentStatus,
    val readiness: AgentReadiness,
    val methods: AgentMethods,
) {
    init {
        require(isOpaqueRef(ref))
        require(readiness != AgentReadiness.Ready || status.state == AgentState.Idle)
        require(readiness != AgentReadiness.Blocked || status.state == AgentState.Blocked)
    }
}

internal data class TerminalRecord(
    val ref: String,
    val name: String? = null,
    val nativeLabel: String? = null,
    val character: CharacterSummary,
    val workspaceRef: String,
    val launchProfile: ProfileKey? = null,
    val objective: String? = null,
    val cwd: String? = null,
    val agent: AgentRuntime? = null,
)

internal data class WorkspaceRecord(val ref: String, val label: WorkspaceLabel)

private val opaqueRefAlphabet = Regex("[A-Za-z0-9_-]+")
private fun isOpaqueRef(value: String): Boolean {
    if (value.length !in 1..4096 || value.length % 4 == 1 || !opaqueRefAlphabet.matches(value)) return false
    return try {
        Base64.getUrlEncoder().withoutPadding().encodeToString(Base64.getUrlDecoder().decode(value)) == value
    } catch (_: IllegalArgumentException) {
        false
    }
}

internal data class TerminalsResponse(
    val machine: MachineSummary,
    val observedAt: Instant,
    val partial: Boolean,
    val unaddressableTerminals: Int,
    val unaddressableWorkspaces: Int,
    val profiles: List<ProfileChoice>,
    val workspaces: List<WorkspaceRecord>,
    val terminals: List<TerminalRecord>,
)

@Serializable
private data class WireTerminalsResponse(
    val machine: WireMachineSummary,
    @Serializable(with = IsoInstantSerializer::class) val observedAt: Instant,
    val partial: Boolean,
    val unaddressableTerminals: Int,
    val unaddressableWorkspaces: Int,
    val profiles: List<WireProfileChoice>,
    val workspaces: List<WireWorkspaceRecord>,
    val terminals: List<WireTerminalRecord>,
)

@Serializable
private data class WireCreatedTerminalResponse(
    @Serializable(with = IsoInstantSerializer::class) val observedAt: Instant,
    val terminal: WireTerminalRecord,
    val launch: String,
    val dispatch: String,
)

@Serializable private data class WireWorkspaceRecord(val ref: String, val label: String)

@Serializable
private data class WireAgentRuntime(
    val ref: String,
    val provider: AgentProvider,
    val status: AgentStatus,
    val readiness: AgentReadiness,
    val methods: AgentMethods,
)

@Serializable
private data class WireTerminalRecord(
    val ref: String,
    val name: String? = null,
    val nativeLabel: String? = null,
    val character: CharacterSummary,
    val workspaceRef: String,
    val launchProfile: String? = null,
    val objective: String? = null,
    val cwd: String? = null,
    val agent: WireAgentRuntime? = null,
)

internal sealed interface LaunchChoice {
    data class Agent(val profile: ProfileKey) : LaunchChoice
    data object Terminal : LaunchChoice
}

internal sealed interface WorkspaceDestination {
    data class Existing(val workspaceRef: String) : WorkspaceDestination
    data class New(val label: WorkspaceLabel?) : WorkspaceDestination
}

internal data class ForgeDraft(
    val machineHandle: MachineHandle,
    val cwd: String,
    val launch: LaunchChoice,
    val name: String,
    val objective: String,
    val destination: WorkspaceDestination?,
)

internal const val MAXIMUM_WORKING_DIRECTORY_BYTES = 4_096
internal const val MAXIMUM_DIRECTORY_LISTING_CHILDREN = 256
internal const val MAXIMUM_DIRECTORY_LISTING_PATH_TEXT_BYTES = 32 * 1_024

internal class HomeDirectory private constructor(val encoded: String) {
    companion object {
        val Home = HomeDirectory("~")

        fun parse(candidate: String): HomeDirectory? {
            if (candidate == "~") return Home
            if (!candidate.startsWith("~/") || candidate.endsWith('/')) return null
            if (candidate.utf8ByteCountWithin(MAXIMUM_WORKING_DIRECTORY_BYTES) == null) return null
            if (candidate.hasDisplayUnsafeCodePoint()) return null
            val components = candidate.substring(2).split('/')
            if (components.any { component -> component.isEmpty() || component == "." || component == ".." }) {
                return null
            }
            return HomeDirectory(candidate)
        }
    }

    val basename: String get() = if (this == Home) "~" else encoded.substringAfterLast('/')
    val hidden: Boolean get() = this != Home && basename.startsWith('.')

    internal fun parent(): ParentDirectory = if (this == Home) {
        ParentDirectory.Absent
    } else {
        ParentDirectory.Available(
            checkNotNull(parse(encoded.substringBeforeLast('/').ifEmpty { "~" })),
        )
    }

    internal fun isDirectChildOf(parent: HomeDirectory): Boolean =
        parent() == ParentDirectory.Available(parent)

    override fun equals(other: Any?): Boolean = other is HomeDirectory && encoded == other.encoded
    override fun hashCode(): Int = encoded.hashCode()
    override fun toString(): String = encoded
}

internal enum class DirectoryEntryKind { Directory, SymbolicLink }
internal sealed interface ParentDirectory {
    data object Absent : ParentDirectory
    data class Available(val directory: HomeDirectory) : ParentDirectory
}
internal enum class DirectoryOmissions { None, Present }
internal data class DirectoryEntry(val directory: HomeDirectory, val kind: DirectoryEntryKind)
internal data class DirectoryListing(
    val machine: MachineSummary,
    val directory: HomeDirectory,
    val parent: ParentDirectory,
    val children: List<DirectoryEntry>,
    val omissions: DirectoryOmissions,
)

@Serializable
private data class WireDirectoryListingResponse(
    val machine: WireMachineSummary,
    val directory: String,
    val parentDirectory: String? = null,
    val children: List<WireDirectoryEntry>,
    val omitted: Boolean,
)

@Serializable
private data class WireDirectoryEntry(
    val directory: String,
    val kind: WireDirectoryEntryKind,
)

@Serializable
private enum class WireDirectoryEntryKind { Directory, SymbolicLink }

internal fun decodeDirectoryListingResponse(
    encoded: String,
    expectedMachine: MachineSummary,
): DirectoryListing = decodeProtocol {
    val element = strictJsonObject(encoded)
    element.requireAbsentOrNonNull(setOf("parentDirectory"))
    val wire = productJson.decodeFromJsonElement<WireDirectoryListingResponse>(element)
    val machineHandle = requireNotNull(MachineHandle.parse(wire.machine.handle))
    val machine = MachineSummary(machineHandle, acceptMachinePlatform(wire.machine.platform))
    require(machine == expectedMachine)
    val directory = requireNotNull(HomeDirectory.parse(wire.directory))
    val parent = wire.parentDirectory?.let { encodedParent ->
        ParentDirectory.Available(requireNotNull(HomeDirectory.parse(encodedParent)))
    } ?: ParentDirectory.Absent
    require(parent == directory.parent())
    require(wire.children.size <= MAXIMUM_DIRECTORY_LISTING_CHILDREN)
    val children = wire.children.map { child ->
        val childDirectory = requireNotNull(HomeDirectory.parse(child.directory))
        require(childDirectory.isDirectChildOf(directory))
        DirectoryEntry(
            directory = childDirectory,
            kind = when (child.kind) {
                WireDirectoryEntryKind.Directory -> DirectoryEntryKind.Directory
                WireDirectoryEntryKind.SymbolicLink -> DirectoryEntryKind.SymbolicLink
            },
        )
    }
    require(children.map(DirectoryEntry::directory).allUnique())
    require(
        children == children.sortedWith { first, second ->
            compareCaseInsensitiveUtf8(first.directory.basename, second.directory.basename)
        },
    )
    val pathTextBytes = sequenceOf(directory) +
        when (parent) {
            ParentDirectory.Absent -> emptySequence()
            is ParentDirectory.Available -> sequenceOf(parent.directory)
        } + children.asSequence().map(DirectoryEntry::directory)
    require(
        pathTextBytes.sumOf { path -> path.encoded.toByteArray(StandardCharsets.UTF_8).size } <=
            MAXIMUM_DIRECTORY_LISTING_PATH_TEXT_BYTES,
    )
    DirectoryListing(
        machine = machine,
        directory = directory,
        parent = parent,
        children = children,
        omissions = if (wire.omitted) DirectoryOmissions.Present else DirectoryOmissions.None,
    )
}

internal fun compareCaseInsensitiveUtf8(first: String, second: String): Int {
    val folded = compareAsciiFoldedUtf8(first, second)
    return if (folded != 0) folded else compareUtf8(first, second)
}

private fun compareAsciiFoldedUtf8(first: String, second: String): Int {
    val firstBytes = first.toByteArray(StandardCharsets.UTF_8)
    val secondBytes = second.toByteArray(StandardCharsets.UTF_8)
    for (index in 0 until minOf(firstBytes.size, secondBytes.size)) {
        val firstByte = asciiLowercase(firstBytes[index].toInt() and 0xff)
        val secondByte = asciiLowercase(secondBytes[index].toInt() and 0xff)
        if (firstByte != secondByte) return firstByte - secondByte
    }
    return firstBytes.size - secondBytes.size
}

private fun asciiLowercase(byte: Int): Int = if (byte in 0x41..0x5a) byte + 0x20 else byte

private fun compareUtf8(first: String, second: String): Int {
    val firstBytes = first.toByteArray(StandardCharsets.UTF_8)
    val secondBytes = second.toByteArray(StandardCharsets.UTF_8)
    for (index in 0 until minOf(firstBytes.size, secondBytes.size)) {
        val difference = (firstBytes[index].toInt() and 0xff) - (secondBytes[index].toInt() and 0xff)
        if (difference != 0) return difference
    }
    return firstBytes.size - secondBytes.size
}

internal data class ForgeForm(
    val machineHandle: MachineHandle?,
    val cwd: String,
    val launch: LaunchChoice?,
    val name: String,
    val objective: String,
    val destination: WorkspaceDraft = WorkspaceDraft.New(""),
) {
    constructor(draft: ForgeDraft) : this(
        draft.machineHandle, draft.cwd, draft.launch, draft.name, draft.objective,
        when (val destination = draft.destination) {
            is WorkspaceDestination.Existing -> WorkspaceDraft.Existing(destination.workspaceRef)
            is WorkspaceDestination.New -> WorkspaceDraft.New(destination.label?.text.orEmpty())
            null -> WorkspaceDraft.New("")
        },
    )

    fun submission(): ForgeDraft? {
        if (machineHandle == null || launch == null || cwd.isBlank() || !validTerminalName(name)) return null
        val chosen = when (val choice = destination) {
            is WorkspaceDraft.Existing -> WorkspaceDestination.Existing(choice.workspaceRef)
            is WorkspaceDraft.New -> WorkspaceDestination.New(
                if (choice.label.isEmpty()) null else WorkspaceLabel.fromDraft(choice.label) ?: return null,
            )
        }
        return ForgeDraft(machineHandle, cwd, launch, name, objective, chosen)
    }
}

internal fun changeForgeDraft(current: ForgeForm, proposed: ForgeForm): ForgeForm =
    if (proposed.machineHandle == current.machineHandle) proposed else proposed.copy(
        cwd = "",
        launch = proposed.launch.takeIf { it == LaunchChoice.Terminal },
        destination = WorkspaceDraft.New(""),
    )

internal fun forgeActionLabel(label: MachineLabel): String = "Create on ${label.text}"
internal fun terminalDisplayName(terminal: TerminalRecord): String =
    terminal.name ?: "unnamed terminal ${terminal.ref.takeLast(8)}"
internal fun killActionLabel(label: MachineLabel, target: TerminalTarget, terminalOnly: Boolean = false): String =
    "${if (target.terminal.agent == null || terminalOnly) "Close" else "Stop"} ${terminalDisplayName(target.terminal)} on ${label.text}"
internal fun killConfirmationTitle(label: MachineLabel, target: TerminalTarget, terminalOnly: Boolean = false): String =
    killActionLabel(label, target, terminalOnly) + "? linked workspaces and their running terminals may also close."

private fun validTerminalName(value: String): Boolean =
    value.length in 1..64 && value.matches(Regex("[A-Za-z0-9][A-Za-z0-9_-]*"))

@Serializable private data class WireDestination(
    val kind: String,
    val workspaceRef: String? = null,
    val label: String? = null,
)
@Serializable private data class CreateTerminalRequest(
    val kind: String,
    val cwd: String,
    val name: String,
    val profile: String? = null,
    val objective: String? = null,
    val destination: WireDestination? = null,
)
@Serializable private data class DirectoryListingRequest(val directory: String)
@Serializable private data class RenameTerminalRequest(val name: String)

internal fun decodeTerminalsResponse(encoded: String): TerminalsResponse = decodeProtocol {
    val element = strictJsonObject(encoded)
    element.getValue("terminals").jsonArray.forEach { encodedTerminal ->
        (encodedTerminal as? JsonObject ?: throw SerializationException("terminal is not an object"))
            .requireTerminalOptionalFields()
    }
    val wire = productJson.decodeFromJsonElement<WireTerminalsResponse>(element)
    val profiles = wire.profiles.map { profile ->
        require(profile.label.isNotEmpty())
        ProfileChoice(requireNotNull(ProfileKey.parse(profile.key)), profile.label, profile.provider)
    }
    require(profiles.map(ProfileChoice::key).allUnique())
    val workspaces = wire.workspaces.map {
        require(isOpaqueRef(it.ref))
        WorkspaceRecord(it.ref, requireNotNull(WorkspaceLabel.parse(it.label)))
    }
    require(workspaces.map(WorkspaceRecord::ref).allUnique())
    val terminals = wire.terminals.map(::acceptTerminal)
    require(terminals.map(TerminalRecord::ref).allUnique())
    require(terminals.mapNotNull { it.agent?.ref }.allUnique())
    require(terminals.all { terminal -> workspaces.any { it.ref == terminal.workspaceRef } })
    require(wire.unaddressableTerminals >= 0 && wire.unaddressableWorkspaces >= 0)
    require(wire.partial || wire.unaddressableTerminals == 0 && wire.unaddressableWorkspaces == 0)
    terminals.forEach { terminal ->
        terminal.launchProfile?.let { profile -> require(profiles.any { it.key == profile }) }
    }
    TerminalsResponse(
        MachineSummary(requireNotNull(MachineHandle.parse(wire.machine.handle)),
            acceptMachinePlatform(wire.machine.platform)),
        acceptProjectionInstant(wire.observedAt), wire.partial, wire.unaddressableTerminals,
        wire.unaddressableWorkspaces, profiles, workspaces, terminals,
    )
}

internal data class CreatedTerminal(val observedAt: Instant, val terminal: TerminalRecord, val launch: String)
internal fun decodeCreatedTerminalResponse(encoded: String): CreatedTerminal = decodeProtocol {
    val element = strictJsonObject(encoded)
    element.requireExactKeys(setOf("observedAt", "terminal", "launch", "dispatch"))
    element.requiredObject("terminal").requireTerminalOptionalFields()
    val wire = productJson.decodeFromJsonElement<WireCreatedTerminalResponse>(element)
    require(wire.launch in setOf("submitted", "not_requested") && wire.dispatch == "sent")
    CreatedTerminal(acceptProjectionInstant(wire.observedAt), acceptTerminal(wire.terminal), wire.launch)
}

internal data class ObservedTerminal(val observedAt: Instant, val terminal: TerminalRecord)
@Serializable private data class WireObservedTerminal(
    @Serializable(with = IsoInstantSerializer::class) val observedAt: Instant,
    val terminal: WireTerminalRecord,
    val dispatch: String,
)
internal fun decodeObservedTerminalResponse(encoded: String): ObservedTerminal = decodeProtocol {
    val element = strictJsonObject(encoded)
    element.requireExactKeys(setOf("observedAt", "terminal", "dispatch"))
    element.requiredObject("terminal").requireTerminalOptionalFields()
    val wire = productJson.decodeFromJsonElement<WireObservedTerminal>(element)
    require(wire.dispatch == "sent")
    ObservedTerminal(acceptProjectionInstant(wire.observedAt), acceptTerminal(wire.terminal))
}

internal fun encodeCreateTerminalRequest(draft: ForgeDraft): String = productJson.encodeToString(
    CreateTerminalRequest(
        kind = if (draft.launch is LaunchChoice.Agent) "agent" else "terminal",
        cwd = draft.cwd,
        name = draft.name,
        profile = (draft.launch as? LaunchChoice.Agent)?.profile?.encoded,
        objective = draft.objective.ifEmpty { null },
        destination = when (val destination = draft.destination) {
            is WorkspaceDestination.Existing -> WireDestination("existing", workspaceRef = destination.workspaceRef)
            is WorkspaceDestination.New -> WireDestination("new", label = destination.label?.text)
            null -> null
        },
    ),
)
internal fun encodeDirectoryListingRequest(directory: HomeDirectory): String =
    productJson.encodeToString(DirectoryListingRequest(directory.encoded))
internal fun encodeRenameTerminalRequest(name: String): String =
    productJson.encodeToString(RenameTerminalRequest(name))

internal data class InventorySnapshot(val inventory: TerminalsResponse, val receivedAtElapsedMillis: Long)

internal sealed interface InventoryState {
    data object Reading : InventoryState
    data class Fresh(val snapshot: InventorySnapshot) : InventoryState
    /** A mutation landed on this machine and the snapshot has not caught up with it yet. */
    data class Superseded(val snapshot: InventorySnapshot, val requiredMutationFence: Long) : InventoryState
    data class Stale(val snapshot: InventorySnapshot, val cause: GatewayFailure) : InventoryState
    data class Unreachable(val cause: GatewayFailure) : InventoryState
}

internal sealed interface MachineAccess {
    data object Ready : MachineAccess
    data object AuthRequired : MachineAccess
    data object IdentityChanged : MachineAccess
}

internal fun InventoryState.lastSnapshot(): InventorySnapshot? = when (this) {
    is InventoryState.Fresh -> snapshot
    is InventoryState.Superseded -> snapshot
    is InventoryState.Stale -> snapshot
    InventoryState.Reading, is InventoryState.Unreachable -> null
}

/** Single owner of the read-failure downgrade: a failed read never discards another machine's facts. */
internal fun InventoryState.downgraded(cause: GatewayFailure): InventoryState = when (this) {
    is InventoryState.Fresh -> InventoryState.Stale(snapshot, cause)
    is InventoryState.Superseded -> InventoryState.Stale(snapshot, cause)
    is InventoryState.Stale -> InventoryState.Stale(snapshot, cause)
    InventoryState.Reading, is InventoryState.Unreachable -> InventoryState.Unreachable(cause)
}

internal data class MachineState(
    val machine: PairedMachine,
    val access: MachineAccess,
    val inventory: InventoryState,
    val pressure: PressureState,
) {
    val canMutate: Boolean get() = when (access) {
        MachineAccess.Ready -> inventory is InventoryState.Fresh
        MachineAccess.AuthRequired, MachineAccess.IdentityChanged -> false
    }

    val canForge: Boolean get() = canMutate

    fun inventoryFailed(cause: GatewayFailure): MachineState = copy(inventory = inventory.downgraded(cause))
}

/**
 * Single classifier for what a machine can currently do. Every machine-state message, colour,
 * and Forge affordance reads this one derivation, so a new access or inventory variant breaks the
 * build in exactly one place.
 */
internal sealed interface MachineAvailability {
    data object Ready : MachineAvailability
    data object Refreshing : MachineAvailability
    data object AuthRequired : MachineAvailability
    data object IdentityChanged : MachineAvailability
    data object Reading : MachineAvailability
    data class Stale(val cause: GatewayFailure) : MachineAvailability
    data class Unavailable(val cause: GatewayFailure) : MachineAvailability
}

internal fun machineAvailability(machine: MachineState): MachineAvailability = when (machine.access) {
    MachineAccess.AuthRequired -> MachineAvailability.AuthRequired
    MachineAccess.IdentityChanged -> MachineAvailability.IdentityChanged
    MachineAccess.Ready -> when (val inventory = machine.inventory) {
        InventoryState.Reading -> MachineAvailability.Reading
        is InventoryState.Fresh -> MachineAvailability.Ready
        is InventoryState.Superseded -> MachineAvailability.Refreshing
        is InventoryState.Stale -> MachineAvailability.Stale(inventory.cause)
        is InventoryState.Unreachable -> MachineAvailability.Unavailable(inventory.cause)
    }
}

internal data class MachineNotice(val message: String, val tone: NoticeTone)

internal data class SessionAvailabilityContent(val label: String, val tone: NoticeTone)

/**
 * Single owner of how loud a machine state is. Trust events are the only failures: a broken bearer
 * or a changed identity means we no longer know who we are talking to. Everything else is absent or
 * ageing knowledge, and architecture.md treats one host being out as normal federated operation, so
 * an outage withdraws the alarm colour while its message still names it literally.
 */
internal fun availabilityTone(availability: MachineAvailability): NoticeTone = when (availability) {
    MachineAvailability.AuthRequired, MachineAvailability.IdentityChanged -> NoticeTone.Failure
    MachineAvailability.Ready,
    MachineAvailability.Refreshing,
    MachineAvailability.Reading,
    is MachineAvailability.Stale,
    is MachineAvailability.Unavailable,
    -> NoticeTone.Degraded
}

/**
 * Single owner of the retained-card availability marker. A card can remain visible while actions
 * are fenced, but the marker must name the actual reason instead of collapsing every fence into
 * staleness. Reading and unavailable inventories have no retained card to annotate.
 */
internal fun sessionAvailabilityContent(machine: MachineState): SessionAvailabilityContent? {
    val availability = machineAvailability(machine)
    val label = when (availability) {
        MachineAvailability.Ready -> return null
        MachineAvailability.Refreshing -> "REFRESHING · actions disabled"
        MachineAvailability.AuthRequired -> "AUTH REQUIRED · actions disabled"
        MachineAvailability.IdentityChanged -> "IDENTITY CHANGED · actions disabled"
        MachineAvailability.Reading, is MachineAvailability.Unavailable -> return null
        is MachineAvailability.Stale -> "STALE · actions disabled"
    }
    return SessionAvailabilityContent(label, availabilityTone(availability))
}

/**
 * Single owner of machine-state prose and its severity: one `when` over [MachineAvailability] yields
 * both, so a message and a tone read at different granularities stop being representable.
 */
internal fun machineNotice(machine: MachineState): MachineNotice? {
    val label = machine.machine.label.text
    val availability = machineAvailability(machine)
    val tone = availabilityTone(availability)
    return when (availability) {
        MachineAvailability.AuthRequired ->
            MachineNotice("$label: authentication required. Actions disabled.", tone)
        MachineAvailability.IdentityChanged ->
            MachineNotice("$label: identity changed. Fleet reset is required.", tone)
        MachineAvailability.Refreshing ->
            MachineNotice("$label: confirming the latest terminal inventory. Actions disabled.", tone)
        MachineAvailability.Reading -> MachineNotice("$label: reading terminals.", tone)
        is MachineAvailability.Stale -> MachineNotice(
                "$label: ${gatewayFailureMessage(availability.cause)} Prior terminals are STALE; actions disabled. " +
                "Pull down to check again.",
            tone,
        )
        is MachineAvailability.Unavailable ->
            MachineNotice("$label: ${gatewayFailureMessage(availability.cause)} Pull down to check again.", tone)
        MachineAvailability.Ready -> {
            val projection = (machine.inventory as InventoryState.Fresh).snapshot.inventory
            if (projection.partial) {
                MachineNotice("$label: inventory partial; ${projection.unaddressableTerminals} terminals and " +
                    "${projection.unaddressableWorkspaces} workspaces could not be addressed.", tone)
            } else when (machine.pressure) {
                is PressureState.Stale ->
                    MachineNotice("$label: pressure is STALE. Terminals remain current.", tone)
                is PressureState.Unavailable ->
                    MachineNotice("$label: pressure unavailable. Terminals remain current.", tone)
                PressureState.Reading, is PressureState.Fresh -> null
            }
        }
    }
}

internal data class VisibleSession(
    val machine: PairedMachine,
    val target: TerminalTarget,
) {
    val cardKey: DashboardCardKey = dashboardCardKey(target)
}

internal fun dashboardCardKey(target: TerminalTarget): DashboardCardKey {
    val digest = MessageDigest.getInstance("SHA-256")
    digest.update("skidbladnir.dashboard-card.v1".encodeToByteArray())
    listOf(
        target.machineHandle.encoded,
        target.terminal.ref,
    ).forEach { value ->
        val bytes = value.encodeToByteArray()
        digest.update(
            byteArrayOf(
                (bytes.size ushr 24).toByte(),
                (bytes.size ushr 16).toByte(),
                (bytes.size ushr 8).toByte(),
                bytes.size.toByte(),
            ),
        )
        digest.update(bytes)
    }
    val hexadecimal = "0123456789abcdef"
    return DashboardCardKey(buildString(64) {
        digest.digest().forEach { byte ->
            val value = byte.toInt() and 0xff
            append(hexadecimal[value ushr 4])
            append(hexadecimal[value and 0x0f])
        }
    })
}


internal fun visibleInventoryTargets(
    liveMachineHandles: Collection<MachineHandle>,
    scope: DashboardScope,
): Set<MachineHandle> = when (scope) {
    DashboardScope.All -> liveMachineHandles.toSet()
    is DashboardScope.Machine -> if (scope.handle in liveMachineHandles) setOf(scope.handle) else emptySet()
}

internal fun visibleSessions(machines: List<MachineState>, scope: DashboardScope): List<VisibleSession> = machines
    .filter { scope == DashboardScope.All || (scope as? DashboardScope.Machine)?.handle == it.machine.handle }
    .flatMap { state -> state.inventory.lastSnapshot()?.inventory?.terminals.orEmpty().map {
        VisibleSession(state.machine, TerminalTarget(state.machine.handle, it))
    } }
    .sortedWith(compareBy<VisibleSession> { it.machine.label.text.lowercase(Locale.ROOT) }
        .thenBy { it.machine.label.text }
        .thenBy { it.machine.handle.encoded }
        .thenBy { terminalDisplayName(it.target.terminal).lowercase(Locale.ROOT) }
        .thenBy { terminalDisplayName(it.target.terminal) }
        .thenBy { it.target.terminal.ref })

internal enum class ApiErrorCode(val wireName: String) {
    Unauthenticated("Unauthenticated"), MachineIdentityMismatch("MachineIdentityMismatch"),
    InvalidRequest("InvalidRequest"), RequestTooLarge("RequestTooLarge"),
    TerminalNotFound("TerminalNotFound"), TerminalStale("TerminalStale"),
    AgentStale("AgentStale"), WorkspaceStale("WorkspaceStale"),
    MetadataUnavailable("MetadataUnavailable"), ProfileUnknown("ProfileUnknown"),
    WorkingDirectoryInvalid("WorkingDirectoryInvalid"), NameInvalid("NameInvalid"),
    NameAmbiguous("NameAmbiguous"), ObjectiveInvalid("ObjectiveInvalid"),
    ReadinessUnconfirmed("ReadinessUnconfirmed"), MethodUnavailable("MethodUnavailable"),
    ClosureConfirmationRequired("ClosureConfirmationRequired"), HerdrUnavailable("HerdrUnavailable"),
    UpstreamRejected("UpstreamRejected"), OutcomeUnknown("OutcomeUnknown"),
    PairingInviteRejected("PairingInviteRejected"),
    DirectoryListingUnavailable("DirectoryListingUnavailable"),
    DirectoryListingTooLarge("DirectoryListingTooLarge"),
    InternalError("InternalError"),
    ReconnectRequired("ReconnectRequired"),
}

internal fun apiErrorMessage(code: ApiErrorCode): String = when (code) {
    ApiErrorCode.Unauthenticated -> "Authentication required."
    ApiErrorCode.MachineIdentityMismatch -> "The machine identity changed. Fleet reset is required."
    ApiErrorCode.InvalidRequest -> "The request is not valid."
    ApiErrorCode.RequestTooLarge -> "The request is too large."
    ApiErrorCode.TerminalNotFound -> "That terminal no longer exists."
    ApiErrorCode.TerminalStale -> "The terminal changed. Refresh and try again."
    ApiErrorCode.AgentStale -> "The agent changed. Refresh and try again."
    ApiErrorCode.WorkspaceStale -> "The workspace changed. Refresh and try again."
    ApiErrorCode.MetadataUnavailable -> "The terminal could not be identified. Inspect the inventory."
    ApiErrorCode.ProfileUnknown -> "Choose an available profile."
    ApiErrorCode.WorkingDirectoryInvalid -> "Choose a valid working directory."
    ApiErrorCode.NameInvalid -> "Use 1–64 letters, numbers, underscores, or hyphens, beginning with a letter or number."
    ApiErrorCode.NameAmbiguous -> "That name matches several terminals. Choose an exact terminal."
    ApiErrorCode.ObjectiveInvalid -> "Use 1–240 characters without terminal controls."
    ApiErrorCode.ReadinessUnconfirmed -> "The agent is not confirmed ready. Inspect it or choose a deliberate override."
    ApiErrorCode.MethodUnavailable -> "This control is unavailable for the agent."
    ApiErrorCode.ClosureConfirmationRequired -> "Herdr refused to close this terminal without confirmation."
    ApiErrorCode.HerdrUnavailable -> "Herdr is unavailable on this machine."
    ApiErrorCode.UpstreamRejected -> "Herdr refused the request."
    ApiErrorCode.OutcomeUnknown -> "Outcome unknown. Inspect the current terminal before trying again."
    ApiErrorCode.PairingInviteRejected -> "This fleet invite is invalid, expired, or already used."
    ApiErrorCode.DirectoryListingUnavailable -> "This directory cannot be browsed. Enter the path instead."
    ApiErrorCode.DirectoryListingTooLarge -> "This directory has too many folders to show. Enter the path instead."
    ApiErrorCode.InternalError -> "Skíðblaðnir could not complete the request."
    ApiErrorCode.ReconnectRequired -> "Reconnect required."
}

internal fun parseApiErrorCode(value: String): ApiErrorCode =
    ApiErrorCode.entries.singleOrNull { it.wireName == value } ?: throw SerializationException("unknown API error code")

internal data class SessionStatusContent(val label: String, val accessibilityLabel: String)

internal fun sessionStatusContent(agent: AgentRuntime?, fresh: Boolean): SessionStatusContent {
    val state = agent?.status?.state?.name?.uppercase() ?: "TERMINAL"
    val readiness = when (agent?.readiness) {
        AgentReadiness.Ready -> " · ready"
        AgentReadiness.Blocked -> " · blocked"
        AgentReadiness.Unconfirmed -> " · unconfirmed"
        null -> ""
    }
    val label = state + readiness
    val spoken = (if (fresh) "" else "Last observed: ") + label.lowercase()
    return SessionStatusContent(label, spoken)
}

private fun JsonObject.requireTerminalOptionalFields() {
    requireAbsentOrNonNull(setOf("name", "nativeLabel", "launchProfile", "objective", "cwd", "agent"))
    (this["agent"] as? JsonObject)?.let { agent ->
        (agent["status"] as? JsonObject)?.requireAbsentOrNonNull(setOf("reason"))
    }
}
private fun <Value> List<Value>.allUnique(): Boolean = distinct().size == size

private fun acceptProjectionInstant(value: Instant): Instant {
    require(value != unsetProjectionInstant)
    return value
}

private fun acceptMachinePlatform(platform: WireMachinePlatform): MachinePlatform = when (platform) {
    WireMachinePlatform.Linux -> MachinePlatform.Linux
    WireMachinePlatform.Darwin -> MachinePlatform.Darwin
}

private fun acceptTerminal(wire: WireTerminalRecord): TerminalRecord = TerminalRecord(
    ref = wire.ref,
    name = wire.name,
    nativeLabel = wire.nativeLabel,
    character = wire.character,
    workspaceRef = wire.workspaceRef,
    launchProfile = wire.launchProfile?.let { requireNotNull(ProfileKey.parse(it)) },
    objective = wire.objective,
    cwd = wire.cwd,
    agent = wire.agent?.let(::acceptAgentRuntime),
).also { terminal ->
    require(isOpaqueRef(terminal.ref) && isOpaqueRef(terminal.workspaceRef))
    require(terminal.name == null || validTerminalName(terminal.name))
    require(terminal.nativeLabel?.let { it.isNotEmpty() && !it.hasDisplayUnsafeCodePoint() } != false)
    require(terminal.cwd?.let(WorkingDirectoryPath::parse) != null || terminal.cwd == null)
    require(terminal.character.key.isNotEmpty() && terminal.character.displayName.isNotEmpty())
}

internal fun decodeTerminalRecord(value: JsonObject): TerminalRecord = decodeProtocol {
    value.requireTerminalOptionalFields()
    acceptTerminal(productJson.decodeFromJsonElement<WireTerminalRecord>(value))
}

private fun acceptAgentRuntime(wire: WireAgentRuntime): AgentRuntime = AgentRuntime(
    ref = wire.ref,
    provider = wire.provider,
    status = wire.status,
    readiness = wire.readiness,
    methods = wire.methods,
)

