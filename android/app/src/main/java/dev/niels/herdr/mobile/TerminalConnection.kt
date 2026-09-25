package dev.niels.herdr.mobile

import android.os.Handler
import android.os.Looper
import java.util.Base64
import java.util.concurrent.atomic.AtomicBoolean
import kotlinx.serialization.Serializable
import kotlinx.serialization.SerializationException
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.booleanOrNull
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okio.ByteString
import org.json.JSONArray
import org.json.JSONObject

private const val MAXIMUM_TERMINAL_FRAME_BYTES = 3 * 1024 * 1024
private const val MAXIMUM_TERMINAL_ANSI_BYTES = 2 * 1024 * 1024
private const val MAXIMUM_TERMINAL_QUEUE_BYTES = 3 * 1024 * 1024L
private const val MAXIMUM_TERMINAL_INPUT_BYTES = 32 * 1024

internal enum class TerminalScrollSource { Wheel, PageKey }
internal enum class TerminalScrollDirection { Up, Down }

internal fun terminalLogicalKeyValid(key: String): Boolean =
    key in setOf(
        "enter", "escape", "tab", "backspace", "up", "down", "left", "right",
        "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
    ) || key.codePointCount(0, key.length) == 1 &&
        key.utf8ByteCountWithin(4) != null &&
        key.codePointAt(0).let {
            it >= 0x20 && !Character.isISOControl(it) &&
                (it == 0x20 || !Character.isWhitespace(it) && !Character.isSpaceChar(it))
        }

internal sealed interface TerminalInput {
    data class Text(val text: String) : TerminalInput
    data class Paste(val text: String) : TerminalInput
    data class Key(val key: String, val modifiers: List<String>) : TerminalInput
    data class Scroll(
        val source: TerminalScrollSource,
        val direction: TerminalScrollDirection,
        val lines: Int,
        val column: Int? = null,
        val row: Int? = null,
    ) : TerminalInput
}

internal data class TerminalFrame(
    val seq: String,
    val columns: Int,
    val rows: Int,
    val full: Boolean,
    val ansi: ByteArray,
)

internal enum class TerminalEndCode { ControlUnavailable, StreamLost, Detached, ProtocolError }

internal sealed interface TerminalServerEvent {
    data class Frame(val frame: TerminalFrame) : TerminalServerEvent
    data class End(val code: TerminalEndCode) : TerminalServerEvent
}
@Serializable private data class TerminalResize(val kind: String, val columns: Int, val rows: Int)
@Serializable private data class TerminalDetach(val kind: String)

internal fun decodeTerminalServerEvent(encoded: String): TerminalServerEvent = decodeProtocol {
    val objectValue = strictJsonObject(encoded)
    when (val kind = objectValue.requiredString("kind")) {
        "Frame" -> {
            objectValue.requireExactKeys(setOf("kind", "seq", "columns", "rows", "full", "ansiBase64"))
            val seq = objectValue.requiredString("seq")
            if (!seq.matches(Regex("[1-9][0-9]*")) || seq.toULongOrNull() == null) {
                throw SerializationException("invalid terminal frame sequence")
            }
            val columns = objectValue.requiredInteger("columns")
            val rows = objectValue.requiredInteger("rows")
            val fullValue = objectValue["full"] as? JsonPrimitive
            if (fullValue == null || fullValue.isString) {
                throw SerializationException("invalid terminal frame completeness")
            }
            val full = fullValue.booleanOrNull
                ?: throw SerializationException("invalid terminal frame completeness")
            if (columns !in TERMINAL_COLUMNS_RANGE || rows !in TERMINAL_ROWS_RANGE) {
                throw SerializationException("invalid terminal frame geometry")
            }
            val base64 = objectValue.requiredString("ansiBase64")
            if (base64.length > (MAXIMUM_TERMINAL_ANSI_BYTES + 2) / 3 * 4 ||
                base64.length % 4 != 0 || !base64.matches(Regex("[A-Za-z0-9+/]*={0,2}"))
            ) throw SerializationException("invalid terminal frame encoding")
            val ansi = try {
                Base64.getDecoder().decode(base64)
            } catch (_: IllegalArgumentException) {
                throw SerializationException("invalid terminal frame encoding")
            }
            if (ansi.size > MAXIMUM_TERMINAL_ANSI_BYTES ||
                Base64.getEncoder().encodeToString(ansi) != base64
            ) throw SerializationException("invalid terminal frame encoding")
            TerminalServerEvent.Frame(TerminalFrame(seq, columns, rows, full, ansi))
        }
        "End" -> {
            objectValue.requireExactKeys(setOf("kind", "code"))
            val code = when (objectValue.requiredString("code")) {
                "control_unavailable" -> TerminalEndCode.ControlUnavailable
                "stream_lost" -> TerminalEndCode.StreamLost
                "detached" -> TerminalEndCode.Detached
                "protocol_error" -> TerminalEndCode.ProtocolError
                else -> throw SerializationException("unknown terminal end code")
            }
            TerminalServerEvent.End(code)
        }
        else -> throw SerializationException("unknown terminal event kind")
    }
}

// One geometry bound: the page publishes only fitted whole-cell grids inside it
// (and ViewportTooSmall below it), and the resize transport carries nothing else.
internal val TERMINAL_COLUMNS_RANGE = 20..1024
internal val TERMINAL_ROWS_RANGE = 5..512

internal fun encodeTerminalResize(columns: Int, rows: Int): String {
    if (columns !in TERMINAL_COLUMNS_RANGE || rows !in TERMINAL_ROWS_RANGE) {
        throw IllegalArgumentException("terminal geometry out of bounds")
    }
    return productJson.encodeToString(TerminalResize("Resize", columns, rows))
}
internal fun encodeTerminalDetach(): String = productJson.encodeToString(TerminalDetach("Detach"))

internal fun encodeTerminalInput(input: TerminalInput): String {
    val message = JSONObject()
    when (input) {
        is TerminalInput.Text -> {
            require(input.text.isNotEmpty() && input.text.utf8ByteCountWithin(MAXIMUM_TERMINAL_INPUT_BYTES) != null)
            require(input.text.none { it.code < 0x20 || it.code == 0x7f })
            message.put("kind", "Text").put("text", input.text)
        }
        is TerminalInput.Paste -> {
            require(input.text.isNotEmpty() && input.text.utf8ByteCountWithin(MAXIMUM_TERMINAL_INPUT_BYTES) != null)
            message.put("kind", "Paste").put("text", input.text)
        }
        is TerminalInput.Key -> {
            require(terminalLogicalKeyValid(input.key))
            require(input.modifiers == listOf("ctrl", "alt", "shift").filter { it in input.modifiers })
            message.put("kind", "Key").put("key", input.key)
                .put("modifiers", JSONArray(input.modifiers))
        }
        is TerminalInput.Scroll -> {
            require(input.lines in 1..512)
            require(input.column == null || input.column in 0..65535)
            require(input.row == null || input.row in 0..65535)
            require(input.source != TerminalScrollSource.PageKey ||
                input.column == null && input.row == null && input.lines in TERMINAL_ROWS_RANGE)
            message.put("kind", "Scroll")
                .put("source", if (input.source == TerminalScrollSource.Wheel) "wheel" else "page_key")
                .put("direction", if (input.direction == TerminalScrollDirection.Up) "up" else "down")
                .put("lines", input.lines)
            input.column?.let { message.put("column", it) }
            input.row?.let { message.put("row", it) }
        }
    }
    return message.toString()
}

private fun JsonObject.requiredInteger(key: String): Int {
    val member = this[key]
    if (member !is JsonPrimitive || member.isString) throw SerializationException("missing or non-number $key")
    return member.content.toIntOrNull() ?: throw SerializationException("$key is not an integer")
}

internal interface TerminalConnectionObserver {
    fun onFrame(frame: TerminalFrame)
    fun onEnd(code: TerminalEndCode)
    fun onFailure(code: ApiErrorCode)
}

private data class PendingResize(
    val encoded: String,
    val byteCount: Int,
)

internal class TerminalConnection(
    private val client: GatewayClient,
    private val credential: MachineCredential,
    private val target: TerminalTarget,
    initialViewport: TerminalViewport.Fitted,
    private val takeover: Boolean,
    private val observer: TerminalConnectionObserver,
) : WebSocketListener() {
    private val stopped = AtomicBoolean(false)
    private val monitor = Any()
    private val main = Handler(Looper.getMainLooper())
    private var socket: WebSocket? = null
    private var started = false
    private var opened = false
    private var receivedFrame = false
    private var pendingFrameBytes = 0L
    private var pendingResize: PendingResize? = null
    private var resizeDrainScheduled = false
    private val resizeDrain = Runnable(::drainResize)

    // The gateway creates no PTY until a valid Resize arrives, so a
    // connection cannot exist without its measured geometry already queued as the
    // first client frame the open-time flush will send.
    init {
        resize(initialViewport.columns, initialViewport.rows)
    }

    fun start() {
        synchronized(monitor) {
            check(!started) // justify-service-invariant-check: each connection object owns exactly one WebSocket lifetime.
            started = true
            if (stopped.get()) return
            socket = client.http.newWebSocket(client.terminalRequest(credential, target, takeover), this)
        }
    }

    override fun onOpen(webSocket: WebSocket, response: Response) {
        synchronized(monitor) {
            if (stopped.get()) {
                webSocket.cancel()
                return
            }
            opened = true
            flushResizeLocked()
        }
    }

    override fun onMessage(webSocket: WebSocket, text: String) {
        if (stopped.get()) return
        val frameBytes = text.utf8ByteCountWithin(MAXIMUM_TERMINAL_FRAME_BYTES)
        if (frameBytes == null) {
            end(TerminalEndCode.ProtocolError)
            return
        }
        val event = try {
            decodeTerminalServerEvent(text)
        } catch (_: ProtocolDecodeException) {
            end(TerminalEndCode.ProtocolError)
            return
        }
        when (event) {
            is TerminalServerEvent.Frame -> {
                synchronized(monitor) {
                    if (stopped.get()) return
                    if (!receivedFrame && !event.frame.full ||
                        pendingFrameBytes + frameBytes > MAXIMUM_TERMINAL_QUEUE_BYTES) {
                        end(TerminalEndCode.ProtocolError)
                        return
                    }
                    receivedFrame = true
                    pendingFrameBytes += frameBytes
                }
                if (!main.post {
                        try {
                            if (!stopped.get()) observer.onFrame(event.frame)
                        } finally {
                            synchronized(monitor) { pendingFrameBytes -= frameBytes }
                        }
                    }) {
                    synchronized(monitor) { pendingFrameBytes -= frameBytes }
                    end(TerminalEndCode.StreamLost)
                }
            }
            is TerminalServerEvent.End -> end(event.code)
        }
    }

    override fun onMessage(webSocket: WebSocket, bytes: ByteString) {
        end(TerminalEndCode.ProtocolError)
    }

    override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
        end(TerminalEndCode.StreamLost)
    }

    override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
        end(TerminalEndCode.StreamLost)
    }

    override fun onFailure(webSocket: WebSocket, throwable: Throwable, response: Response?) {
        if (stopped.get()) return
        val code = if (response == null) {
            terminalUpgradeFailureCode(null, null)
        } else {
            response.use {
                val bytes = it.body?.byteStream()?.readNBytes(MAXIMUM_HTTP_BODY_BYTES + 1) ?: ByteArray(0)
                if (bytes.size > MAXIMUM_HTTP_BODY_BYTES) ApiErrorCode.ReconnectRequired
                else terminalUpgradeFailureCode(it.code, decodeStrictUtf8(bytes))
            }
        }
        fail(code)
    }

    fun resize(columns: Int, rows: Int) {
        val encoded = try {
            encodeTerminalResize(columns, rows)
        } catch (_: IllegalArgumentException) {
            // justify-defect: the owned page emits only geometry inside the closed protocol bounds,
            // and the reason stays a fixed literal so no terminal content can reach a log.
            throw ProtocolDecodeException("terminal geometry left the closed protocol bounds")
        }
        synchronized(monitor) {
            if (stopped.get()) return
            pendingResize = PendingResize(encoded, encoded.toByteArray(Charsets.UTF_8).size)
            flushResizeLocked()
        }
    }

    fun send(input: TerminalInput) {
        val encoded = encodeTerminalInput(input)
        synchronized(monitor) {
            val activeSocket = socket
            if (!receivedFrame || stopped.get() || activeSocket == null) return
            val pendingResizeBytes = pendingResize?.byteCount ?: 0
            if (encoded.toByteArray(Charsets.UTF_8).size.toLong() + pendingResizeBytes + activeSocket.queueSize() >
                MAXIMUM_TERMINAL_QUEUE_BYTES
            ) {
                end(TerminalEndCode.ProtocolError)
                return
            }
            flushResizeBeforeInputLocked(activeSocket)
            if (stopped.get()) return
            if (!activeSocket.send(encoded)) end(TerminalEndCode.StreamLost)
        }
    }

    fun detach() {
        if (!stopped.compareAndSet(false, true)) return
        synchronized(monitor) {
            receivedFrame = false
            pendingResize = null
            resizeDrainScheduled = false
            main.removeCallbacks(resizeDrain)
            socket?.send(encodeTerminalDetach())
            socket?.close(1000, "Detach")
            socket = null
        }
    }

    fun terminalUnavailable() {
        end(TerminalEndCode.StreamLost)
    }

    private fun fail(code: ApiErrorCode) {
        if (!stopped.compareAndSet(false, true)) return
        synchronized(monitor) {
            receivedFrame = false
            pendingResize = null
            resizeDrainScheduled = false
            main.removeCallbacks(resizeDrain)
            socket?.cancel()
            socket = null
        }
        observer.onFailure(code)
    }

    private fun end(code: TerminalEndCode) {
        if (!stopped.compareAndSet(false, true)) return
        synchronized(monitor) {
            receivedFrame = false
            pendingResize = null
            resizeDrainScheduled = false
            main.removeCallbacks(resizeDrain)
            socket?.cancel()
            socket = null
        }
        observer.onEnd(code)
    }

    private fun drainResize() {
        synchronized(monitor) {
            resizeDrainScheduled = false
            if (!stopped.get()) flushResizeLocked()
        }
    }

    private fun flushResizeLocked() {
        val activeSocket = socket ?: return
        val resize = pendingResize ?: return
        if (!opened) return
        val queueSize = activeSocket.queueSize()
        if (queueSize + resize.byteCount > MAXIMUM_TERMINAL_QUEUE_BYTES) {
            fail(ApiErrorCode.ReconnectRequired)
            return
        }
        if (queueSize > 0) {
            scheduleResizeDrainLocked()
            return
        }
        pendingResize = null
        if (!activeSocket.send(resize.encoded)) {
            fail(ApiErrorCode.ReconnectRequired)
        }
    }

    private fun scheduleResizeDrainLocked() {
        if (resizeDrainScheduled) return
        resizeDrainScheduled = true
        main.postDelayed(resizeDrain, 25)
    }

    private fun flushResizeBeforeInputLocked(activeSocket: WebSocket) {
        val resize = pendingResize ?: return
        pendingResize = null
        resizeDrainScheduled = false
        main.removeCallbacks(resizeDrain)
        if (!activeSocket.send(resize.encoded)) fail(ApiErrorCode.ReconnectRequired)
    }
}

internal fun terminalUpgradeFailureCode(status: Int?, encoded: String?): ApiErrorCode {
    if (status == null) {
        require(encoded == null)
        return ApiErrorCode.ReconnectRequired
    }
    val failure = decodeGatewayHttpFailure(status, requireNotNull(encoded))
    return when (failure) {
        is GatewayFailure.Api -> failure.code
        GatewayFailure.Transport -> ApiErrorCode.ReconnectRequired
    }
}
