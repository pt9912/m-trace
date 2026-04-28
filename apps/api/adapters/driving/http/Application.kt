// Micronaut entry point. Per docs/plan-spike.md §12.2 the
// application bootstrap lives in adapters/driving/http/ because the
// HTTP server is the inbound (driving) adapter — analogous to
// d-migrate's adapters/driving/cli/.../Main.kt.
package dev.mtrace.api.adapters.driving.http

import io.micronaut.runtime.Micronaut

fun main(args: Array<String>) {
    Micronaut.run(*args)
}
