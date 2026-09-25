package dev.niels.skidbladnir

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.decodeFromJsonElement

@Serializable
internal enum class AgentState {
    @SerialName("working") Working,
    @SerialName("blocked") Blocked,
    @SerialName("idle") Idle,
    @SerialName("unknown") Unknown,
}

@Serializable
internal enum class AgentStatusSource {
    @SerialName("herdr") Herdr,
    @SerialName("unavailable") Unavailable,
}

@Serializable
internal enum class AgentReadiness {
    @SerialName("ready") Ready,
    @SerialName("blocked") Blocked,
    @SerialName("unconfirmed") Unconfirmed,
}

@Serializable
internal enum class AgentReason {
    @SerialName("default_idle") DefaultIdle,
    @SerialName("unrecognized") Unrecognized,
    @SerialName("observation_failed") ObservationFailed,
}

@Serializable
internal data class AgentStatus(val state: AgentState, val source: AgentStatusSource, val reason: AgentReason? = null)

@Serializable
internal data class AgentWriteResult(val outcome: String, val dispatch: String)

@Serializable
internal data class AgentStopResult(val agent: String, val terminal: String, val dispatch: String)

internal fun decodeAgentWriteResult(encoded: String): AgentWriteResult = decodeProtocol {
    productJson.decodeFromJsonElement<AgentWriteResult>(strictJsonObject(encoded)).also {
        require((it.outcome == "written" && it.dispatch == "sent") ||
            (it.outcome == "unknown" && it.dispatch == "unknown"))
    }
}

internal fun decodeAgentStopResult(encoded: String): AgentStopResult = decodeProtocol {
    productJson.decodeFromJsonElement<AgentStopResult>(strictJsonObject(encoded)).also {
        require(it.agent in setOf("interrupt_sent", "exited"))
        require(it.terminal == "closed" && it.dispatch == "sent")
    }
}

internal fun agentStopMessage(machine: MachineLabel, result: AgentStopResult): String =
    "${machine.text}: interrupt ${result.agent}; terminal closed. linked workspaces may also have closed."

internal fun agentInterruptMessage(machine: MachineLabel, result: AgentWriteResult): String = when (result.outcome) {
    "written" -> "${machine.text}: interrupt key sent; stopping is not yet confirmed."
    "unknown" -> "${machine.text}: interrupt outcome unknown."
    else -> error("unrecognized interrupt outcome")
}
