// JSON wire types for the HTTP adapter. Kept separate from the
// driving-port types (BatchInput / EventInput) so JSON / Jackson
// concerns don't leak into the inner hexagon.
package dev.mtrace.api.adapters.driving.http

import com.fasterxml.jackson.annotation.JsonProperty
import dev.mtrace.api.hexagon.port.driving.BatchInput
import dev.mtrace.api.hexagon.port.driving.EventInput
import dev.mtrace.api.hexagon.port.driving.SDKInput

data class WireBatch(
    @JsonProperty("schema_version") val schemaVersion: String = "",
    @JsonProperty("events") val events: List<WireEvent> = emptyList(),
)

data class WireEvent(
    @JsonProperty("event_name") val eventName: String = "",
    @JsonProperty("project_id") val projectId: String = "",
    @JsonProperty("session_id") val sessionId: String = "",
    @JsonProperty("client_timestamp") val clientTimestamp: String = "",
    @JsonProperty("sequence_number") val sequenceNumber: Long? = null,
    @JsonProperty("sdk") val sdk: WireSDK = WireSDK(),
)

data class WireSDK(
    @JsonProperty("name") val name: String = "",
    @JsonProperty("version") val version: String = "",
)

internal fun WireBatch.toBatchInput(authToken: String): BatchInput = BatchInput(
    schemaVersion = schemaVersion,
    authToken = authToken,
    events = events.map { it.toEventInput() },
)

internal fun WireEvent.toEventInput(): EventInput = EventInput(
    eventName = eventName,
    projectId = projectId,
    sessionId = sessionId,
    clientTimestamp = clientTimestamp,
    sequenceNumber = sequenceNumber,
    sdk = SDKInput(name = sdk.name, version = sdk.version),
)
