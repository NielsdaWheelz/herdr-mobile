package dev.niels.skidbladnir

import java.io.IOException
import java.nio.ByteBuffer
import java.nio.charset.CharacterCodingException
import java.nio.charset.CodingErrorAction
import java.nio.charset.StandardCharsets
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicBoolean
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.decodeFromJsonElement
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response

private val jsonMediaType = "application/json; charset=utf-8".toMediaType()
internal const val MAXIMUM_HTTP_BODY_BYTES = 64 * 1024
private const val MAXIMUM_CONTROL_BODY_BYTES = 256 * 1024
private const val MAXIMUM_INVENTORY_BODY_BYTES = 1024 * 1024

internal class GatewayBearer private constructor(internal val encoded: String) {
    companion object {
        fun parse(candidate: String): GatewayBearer? =
            candidate.takeIf(::isCanonicalBase64Url256)?.let(::GatewayBearer)
    }

    override fun equals(other: Any?): Boolean = other is GatewayBearer && encoded == other.encoded
    override fun hashCode(): Int = encoded.hashCode()
    override fun toString(): String = "GatewayBearer(redacted)"
}

internal data class MachineCredential(
    val machine: PairedMachine,
    val bearer: GatewayBearer,
)

internal sealed interface GatewayResult<out Value> {
    data class Success<Value>(val value: Value) : GatewayResult<Value>
    data class Failure(val failure: GatewayFailure) : GatewayResult<Nothing>
}

internal enum class MutationDispatch { NotSent, Sent, Unknown }

internal data class MutationPartial(
    val stage: String? = null,
    val terminal: TerminalRecord? = null,
    val terminalStatus: String? = null,
    val agentStatus: String? = null,
)

internal sealed interface GatewayFailure {
    data class Api(
        val code: ApiErrorCode,
        val dispatch: MutationDispatch? = null,
        val partial: MutationPartial? = null,
    ) : GatewayFailure
    data object Transport : GatewayFailure
}

internal fun gatewayFailureMessage(failure: GatewayFailure): String = when (failure) {
    is GatewayFailure.Api -> buildString {
        append(apiErrorMessage(failure.code))
        failure.partial?.let { partial ->
            when (partial.stage) {
                "resource_created" -> append(" A terminal was created, but its identity was not confirmed.")
                "identified" -> append(" The terminal was identified; launch was not confirmed.")
                null -> Unit
                else -> error("unknown creation stage")
            }
            when (partial.agentStatus) {
                "interrupt_sent" -> append(" Interrupt sent; stopping is not confirmed.")
                "exited" -> append(" The original agent process exited.")
                "unconfirmed" -> append(" Agent state is unconfirmed.")
                null -> Unit
                else -> error("unknown partial agent state")
            }
            when (partial.terminalStatus) {
                "refused" -> append(" Terminal close was refused.")
                "unconfirmed" -> append(" Terminal close is unconfirmed.")
                "not_attempted" -> append(" Terminal close was not attempted.")
                null -> Unit
                else -> error("unknown partial terminal state")
            }
            partial.terminal?.let { append(" Terminal: ${terminalDisplayName(it)}.") }
        }
    }
    GatewayFailure.Transport -> "Could not reach this machine over your Tailnet."
}

internal fun createFailureIsDefinitive(failure: GatewayFailure): Boolean =
    failure is GatewayFailure.Api && failure.dispatch == MutationDispatch.NotSent

@Serializable private data class WirePairingMachine(val handle: String, val platform: MachinePlatform)
@Serializable private data class WirePairingResponse(val machine: WirePairingMachine, val bearer: String)

internal data class PairingResponse(
    val machine: MachineSummary,
    val bearer: GatewayBearer,
)

internal fun decodePairingResponse(encoded: String): PairingResponse = decodeProtocol {
    val wire = productJson.decodeFromJsonElement<WirePairingResponse>(strictJsonObject(encoded))
    PairingResponse(
        machine = MachineSummary(
            requireNotNull(MachineHandle.parse(wire.machine.handle)),
            wire.machine.platform,
        ),
        bearer = requireNotNull(GatewayBearer.parse(wire.bearer)),
    )
}

internal class GatewayClient {
    internal val http = OkHttpClient.Builder()
        .retryOnConnectionFailure(false)
        .followRedirects(false)
        .followSslRedirects(false)
        .pingInterval(15, TimeUnit.SECONDS)
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(15, TimeUnit.SECONDS)
        .writeTimeout(15, TimeUnit.SECONDS)
        .build()

    private val closeScheduled = AtomicBoolean(false)

    fun closeAsync() {
        if (!closeScheduled.compareAndSet(false, true)) return
        val dispatcher = http.dispatcher
        val executor = dispatcher.executorService
        executor.execute {
            try {
                dispatcher.cancelAll()
                http.connectionPool.evictAll()
            } finally {
                executor.shutdown()
            }
        }
    }

    fun listTerminals(credential: MachineCredential): GatewayResult<TerminalsResponse> = executeJson(
        request = authorizedRequest(credential, listOf("v1", "terminals")).get().build(),
        expectedStatus = 200,
        decode = ::decodeTerminalsResponse,
        decodeFailure = ::decodeTerminalsHttpFailure,
        maximumBodyBytes = MAXIMUM_INVENTORY_BODY_BYTES,
    )

    fun readPressure(credential: MachineCredential): GatewayResult<PressureResponse> = executeJson(
        request = authorizedRequest(credential, listOf("v1", "pressure")).get().build(),
        expectedStatus = 200,
        decode = ::decodePressureResponse,
        decodeFailure = ::decodePressureHttpFailure,
    )

    fun listDirectory(
        credential: MachineCredential,
        directory: HomeDirectory,
        expectedMachine: MachineSummary,
    ): GatewayResult<DirectoryListing> {
        require(expectedMachine.handle == credential.machine.handle)
        return executeJson(
            request = directoryListingRequest(credential, directory),
            expectedStatus = 200,
            decode = { encoded -> decodeDirectoryListingResponse(encoded, expectedMachine) },
            decodeFailure = ::decodeDirectoryListingHttpFailure,
        )
    }

    internal fun directoryListingRequest(
        credential: MachineCredential,
        directory: HomeDirectory,
    ): Request = authorizedRequest(credential, listOf("v1", "directory-listings"))
        .post(encodeDirectoryListingRequest(directory).toRequestBody(jsonMediaType))
        .build()

    fun redeemPairing(machine: FleetInviteMachine): GatewayResult<PairingResponse> = executeJson(
        request = pairingRequest(machine),
        expectedStatus = 200,
        decode = ::decodePairingResponse,
        decodeFailure = ::decodePairingHttpFailure,
    )

    internal fun pairingRequest(machine: FleetInviteMachine): Request {
        val url = machine.machine.origin.encoded.toHttpUrl().newBuilder()
            .addPathSegment("v1")
            .addPathSegment("pairings")
            .build()
        return Request.Builder()
            .url(url)
            .header("Authorization", "Skidbladnir-Invite ${machine.pairingInviteToken.encoded}")
            .header("Accept", "application/json")
            .header("Skidbladnir-Machine", machine.machine.handle.encoded)
            .post(ByteArray(0).toRequestBody())
            .build()
    }

    fun createTerminal(credential: MachineCredential, draft: ForgeDraft): GatewayResult<CreatedTerminal> {
        require(draft.machineHandle == credential.machine.handle)
        return executeJson(
            authorizedRequest(credential, listOf("v1", "terminals"))
                .post(encodeCreateTerminalRequest(draft).toRequestBody(jsonMediaType)).build(),
            201, ::decodeCreatedTerminalResponse, ::decodeMutationHttpFailure,
        )
    }

    fun createShell(credential: MachineCredential, source: TerminalTarget): GatewayResult<CreatedTerminal> {
        require(source.machineHandle == credential.machine.handle)
        return executeJson(
            authorizedRequest(credential, listOf("v1", "terminals", source.terminal.ref, "shell"))
                .post("{}".toRequestBody(jsonMediaType)).build(),
            201, ::decodeCreatedTerminalResponse, ::decodeMutationHttpFailure,
        )
    }

    fun interruptAgent(credential: MachineCredential, target: TerminalTarget): GatewayResult<AgentWriteResult> = executeJson(
        agentRequest(credential, target, "interrupt"), 200, ::decodeAgentWriteResult, ::decodeMutationHttpFailure,
    )

    fun stopAgent(credential: MachineCredential, target: TerminalTarget): GatewayResult<AgentStopResult> = executeJson(
        agentRequest(credential, target, "stop"), 200, ::decodeAgentStopResult, ::decodeMutationHttpFailure,
    )

    internal fun agentRequest(credential: MachineCredential, target: TerminalTarget, operation: String): Request {
        require(target.machineHandle == credential.machine.handle)
        require(operation == "interrupt" || operation == "stop")
        val ref = requireNotNull(target.terminal.agent?.ref)
        return authorizedRequest(credential, listOf("v1", "agents", ref, operation))
            .post("{}".toRequestBody(jsonMediaType)).build()
    }

    fun killTerminal(credential: MachineCredential, target: TerminalTarget): GatewayResult<TerminalCloseResult> =
        executeJson(killRequest(credential, target), 200, ::decodeTerminalCloseResult, ::decodeMutationHttpFailure)

    internal fun killRequest(credential: MachineCredential, target: TerminalTarget): Request {
        require(target.machineHandle == credential.machine.handle)
        return authorizedRequest(credential, listOf("v1", "terminals", target.terminal.ref))
            .delete("{}".toRequestBody(jsonMediaType)).build()
    }

    fun renameTerminal(credential: MachineCredential, target: TerminalTarget, name: String): GatewayResult<ObservedTerminal> =
        executeJson(renameRequest(credential, target, name), 200,
            ::decodeObservedTerminalResponse, ::decodeMutationHttpFailure)

    internal fun renameRequest(credential: MachineCredential, target: TerminalTarget, name: String): Request {
        require(target.machineHandle == credential.machine.handle)
        return authorizedRequest(credential, listOf("v1", "terminals", target.terminal.ref))
            .patch(encodeRenameTerminalRequest(name).toRequestBody(jsonMediaType)).build()
    }

    fun moveTerminal(
        credential: MachineCredential,
        target: TerminalTarget,
        destination: WorkspaceDestination,
    ): GatewayResult<ObservedTerminal> = executeJson(
        moveRequest(credential, target, destination), 200, ::decodeObservedTerminalResponse,
        ::decodeMutationHttpFailure,
    )

    internal fun moveRequest(
        credential: MachineCredential,
        target: TerminalTarget,
        destination: WorkspaceDestination,
    ): Request {
        require(target.machineHandle == credential.machine.handle)
        return authorizedRequest(credential, listOf("v1", "terminals", target.terminal.ref, "workspace"))
            .put(encodeMoveTerminalRequest(destination).toRequestBody(jsonMediaType)).build()
    }

    internal fun terminalRequest(credential: MachineCredential, target: TerminalTarget, takeover: Boolean = false): Request {
        require(target.machineHandle == credential.machine.handle)
        return authorizedRequest(credential, listOf("v1", "terminals", target.terminal.ref, "stream"))
            .header("Skidbladnir-Terminal-Takeover", takeover.toString()).build()
    }

    /**
     * Every ordinary bearer request is bound to a pinned machine; there is no headerless variant.
     * A gateway that answers this origin with another installation's identity fails with
     * `409 MachineIdentityMismatch` before it discloses anything.
     */
    private fun authorizedRequest(credential: MachineCredential, segments: List<String>): Request.Builder {
        val url = credential.machine.origin.encoded.toHttpUrl().newBuilder()
            .apply { segments.forEach(::addPathSegment) }
            .build()
        return Request.Builder()
            .url(url)
            .header("Authorization", "Bearer ${credential.bearer.encoded}")
            .header("Accept", "application/json")
            .header("Skidbladnir-Machine", credential.machine.handle.encoded)
    }

    private fun <Value> executeJson(
        request: Request,
        expectedStatus: Int,
        decode: (String) -> Value,
        decodeFailure: (Int, String) -> GatewayFailure = ::decodeGatewayHttpFailure,
        maximumBodyBytes: Int = MAXIMUM_CONTROL_BODY_BYTES,
    ): GatewayResult<Value> = try {
        http.newCall(request).execute().use { response ->
            decodeGatewayResponse(response, expectedStatus, decode, decodeFailure, maximumBodyBytes)
        }
    } catch (_: IOException) {
        GatewayResult.Failure(GatewayFailure.Transport)
    }

}

internal fun <Value> decodeGatewayResponse(
    response: Response,
    expectedStatus: Int,
    decode: (String) -> Value,
    decodeFailure: (Int, String) -> GatewayFailure = ::decodeGatewayHttpFailure,
    maximumBodyBytes: Int = MAXIMUM_CONTROL_BODY_BYTES,
): GatewayResult<Value> {
    val body = response.body
    val mediaType = body?.contentType()
    if (mediaType?.type != "application" || mediaType.subtype != "json") {
        return GatewayResult.Failure(GatewayFailure.Transport)
    }
    // OkHttp presents this ResponseBody after transparent content decompression. Reading one byte
    // beyond the selected protocol bound distinguishes an exact-limit body without buffering
    // an attacker-controlled response through ResponseBody.string().
    val bound = if (response.code == expectedStatus) maximumBodyBytes else MAXIMUM_HTTP_BODY_BYTES
    val bytes = body?.byteStream()?.readNBytes(bound + 1) ?: ByteArray(0)
    if (bytes.size > bound) return GatewayResult.Failure(GatewayFailure.Transport)
    val encoded = decodeStrictUtf8(bytes)
    return if (response.code != expectedStatus) {
        GatewayResult.Failure(decodeFailure(response.code, encoded))
    } else {
        GatewayResult.Success(decodeProtocol { decode(encoded) })
    }
}

internal fun decodeStrictUtf8(bytes: ByteArray): String = try {
    StandardCharsets.UTF_8.newDecoder()
        .onMalformedInput(CodingErrorAction.REPORT)
        .onUnmappableCharacter(CodingErrorAction.REPORT)
        .decode(ByteBuffer.wrap(bytes))
        .toString()
} catch (_: CharacterCodingException) {
    throw ProtocolDecodeException("HTTP response body is not UTF-8")
}

@Serializable private data class WireTerminalCloseResult(val terminal: String, val dispatch: String)
internal data class TerminalCloseResult(val terminal: String, val dispatch: MutationDispatch)
internal fun decodeTerminalCloseResult(encoded: String): TerminalCloseResult = decodeProtocol {
    val value = strictJsonObject(encoded)
    value.requireExactKeys(setOf("terminal", "dispatch"))
    val wire = productJson.decodeFromJsonElement<WireTerminalCloseResult>(value)
    require(wire.terminal == "closed" && wire.dispatch == "sent")
    TerminalCloseResult(wire.terminal, MutationDispatch.Sent)
}

@Serializable private data class WireMoveRequest(val destination: WireMoveDestination)
@Serializable private data class WireMoveDestination(
    val kind: String,
    val workspaceRef: String? = null,
    val label: String? = null,
)
internal fun encodeMoveTerminalRequest(destination: WorkspaceDestination): String =
    productJson.encodeToString(WireMoveRequest(when (destination) {
        is WorkspaceDestination.Existing -> WireMoveDestination("existing", workspaceRef = destination.workspaceRef)
        is WorkspaceDestination.New -> WireMoveDestination("new", label = requireNotNull(destination.label).text)
    }))

private fun decodeLegacyHttpFailure(status: Int, encoded: String, allowed: Set<ApiErrorCode>): GatewayFailure = decodeProtocol {
    val value = strictJsonObject(encoded)
    value.requireExactKeys(setOf("code", "message"))
    val code = parseApiErrorCode(value.requiredString("code"))
    require(code in allowed && value.requiredString("message").isNotEmpty())
    require(status == apiErrorHttpStatus(code))
    GatewayFailure.Api(code)
}

internal fun decodeGatewayHttpFailure(status: Int, encoded: String): GatewayFailure =
    decodeApiHttpFailure(status, encoded, requireDispatch = false)

internal fun decodeTerminalsHttpFailure(status: Int, encoded: String): GatewayFailure =
    decodeApiHttpFailure(status, encoded, requireDispatch = false)

internal fun decodePressureHttpFailure(status: Int, encoded: String): GatewayFailure =
    decodeLegacyHttpFailure(status, encoded, setOf(
        ApiErrorCode.Unauthenticated, ApiErrorCode.MachineIdentityMismatch, ApiErrorCode.InternalError,
    ))

internal fun decodePairingHttpFailure(status: Int, encoded: String): GatewayFailure =
    decodeLegacyHttpFailure(status, encoded, setOf(
        ApiErrorCode.PairingInviteRejected, ApiErrorCode.InvalidRequest, ApiErrorCode.InternalError,
    ))

internal fun decodeDirectoryListingHttpFailure(status: Int, encoded: String): GatewayFailure =
    decodeLegacyHttpFailure(status, encoded, setOf(
        ApiErrorCode.Unauthenticated, ApiErrorCode.InvalidRequest, ApiErrorCode.RequestTooLarge,
        ApiErrorCode.DirectoryListingUnavailable, ApiErrorCode.DirectoryListingTooLarge,
        ApiErrorCode.MachineIdentityMismatch, ApiErrorCode.InternalError,
    ))

internal fun decodeMutationHttpFailure(status: Int, encoded: String): GatewayFailure =
    decodeApiHttpFailure(status, encoded, requireDispatch = true)

private fun decodeApiHttpFailure(status: Int, encoded: String, requireDispatch: Boolean): GatewayFailure = decodeProtocol {
    val value = strictJsonObject(encoded)
    require(value.keys.all { it in setOf("code", "message", "dispatch", "partial") })
    value.requireAbsentOrNonNull(setOf("dispatch", "partial"))
    val code = parseApiErrorCode(value.requiredString("code"))
    require(value.requiredString("message").isNotEmpty())
    require(status == apiErrorHttpStatus(code))
    val dispatch = when (value["dispatch"]?.let { value.requiredString("dispatch") }) {
        null -> null
        "not_sent" -> MutationDispatch.NotSent
        "sent" -> MutationDispatch.Sent
        "unknown" -> MutationDispatch.Unknown
        else -> throw SerializationException("invalid mutation dispatch")
    }
    require(!requireDispatch || dispatch != null)
    val partialValue = value["partial"]
    if (partialValue != null && partialValue !is kotlinx.serialization.json.JsonObject) {
        throw SerializationException("invalid mutation partial")
    }
    val partial = (partialValue as? kotlinx.serialization.json.JsonObject)?.let { objectValue ->
        require(objectValue.keys.all { it in setOf("stage", "terminal", "agent") })
        objectValue.requireAbsentOrNonNull(setOf("stage", "terminal", "agent"))
        val terminal = objectValue["terminal"]
        val terminalRecord = (terminal as? kotlinx.serialization.json.JsonObject)?.let { objectRecord ->
            decodePartialTerminal(objectRecord)
        }
        val terminalStatus = if (terminal != null && terminalRecord == null) objectValue.requiredString("terminal") else null
        val stage = objectValue["stage"]?.let { objectValue.requiredString("stage") }
        val agentStatus = objectValue["agent"]?.let { objectValue.requiredString("agent") }
        require(stage == null || stage in setOf("resource_created", "identified"))
        require(agentStatus == null || agentStatus in setOf("interrupt_sent", "exited", "unconfirmed"))
        require(terminalStatus == null || terminalStatus in setOf("refused", "unconfirmed", "not_attempted"))
        MutationPartial(
            stage = stage,
            terminal = terminalRecord,
            terminalStatus = terminalStatus,
            agentStatus = agentStatus,
        )
    }
    GatewayFailure.Api(code, dispatch, partial)
}

private fun decodePartialTerminal(value: kotlinx.serialization.json.JsonObject): TerminalRecord =
    decodeTerminalRecord(value)

private fun apiErrorHttpStatus(code: ApiErrorCode): Int = when (code) {
    ApiErrorCode.Unauthenticated, ApiErrorCode.PairingInviteRejected -> 401
    ApiErrorCode.InvalidRequest -> 400
    ApiErrorCode.RequestTooLarge -> 413
    ApiErrorCode.TerminalNotFound -> 404
    ApiErrorCode.TerminalStale, ApiErrorCode.AgentStale, ApiErrorCode.WorkspaceStale,
    ApiErrorCode.ClosureConfirmationRequired, ApiErrorCode.MachineIdentityMismatch -> 409
    ApiErrorCode.ProfileUnknown, ApiErrorCode.WorkingDirectoryInvalid, ApiErrorCode.NameInvalid,
    ApiErrorCode.ObjectiveInvalid, ApiErrorCode.DirectoryListingUnavailable,
    ApiErrorCode.DirectoryListingTooLarge -> 422
    ApiErrorCode.MetadataUnavailable, ApiErrorCode.UpstreamRejected -> 502
    ApiErrorCode.HerdrUnavailable -> 503
    ApiErrorCode.OutcomeUnknown -> 504
    ApiErrorCode.InternalError -> 500
    ApiErrorCode.ReconnectRequired -> throw SerializationException("websocket only code")
}
