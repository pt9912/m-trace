package dev.mtrace.api.adapters.driving.http

import io.micronaut.http.MediaType
import io.micronaut.http.annotation.Controller
import io.micronaut.http.annotation.Get
import io.micronaut.http.annotation.Produces

@Controller("/api")
class HealthController {

    @Get("/health")
    @Produces(MediaType.APPLICATION_JSON)
    fun health(): Map<String, String> = mapOf("status" to "ok")
}
