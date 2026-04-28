import io.gitlab.arturbosch.detekt.Detekt
import io.gitlab.arturbosch.detekt.extensions.DetektExtension

plugins {
    val kotlinVersion = "2.1.20"
    kotlin("jvm") version kotlinVersion
    kotlin("plugin.allopen") version kotlinVersion
    id("com.google.devtools.ksp") version "2.1.20-1.0.31"
    id("io.micronaut.application") version "4.4.5"
    id("io.gitlab.arturbosch.detekt") version "1.23.8"
}

group = "dev.mtrace"
version = "0.1.0-spike"

application {
    applicationName = "api"
    mainClass.set("dev.mtrace.api.adapters.driving.http.ApplicationKt")
}

repositories {
    mavenCentral()
}

dependencies {
    // Micronaut HTTP server, Jackson and Kotlin support.
    ksp("io.micronaut:micronaut-http-validation")
    implementation("io.micronaut:micronaut-http-server-netty")
    implementation("io.micronaut:micronaut-jackson-databind")
    implementation("io.micronaut.kotlin:micronaut-kotlin-runtime")
    implementation("jakarta.annotation:jakarta.annotation-api")
    implementation("org.jetbrains.kotlin:kotlin-reflect:${rootProject.properties["kotlinVersion"]}")

    // Logging, YAML config and metrics.
    runtimeOnly("ch.qos.logback:logback-classic:${rootProject.properties["logbackVersion"]}")
    runtimeOnly("org.yaml:snakeyaml")
    implementation(
        "io.micrometer:micrometer-registry-prometheus:${rootProject.properties["micrometerVersion"]}",
    )

    // OpenTelemetry minimal setup (Spec §6.7: wired but silent).
    implementation("io.opentelemetry:opentelemetry-api:${rootProject.properties["opentelemetryVersion"]}")
    implementation("io.opentelemetry:opentelemetry-sdk:${rootProject.properties["opentelemetryVersion"]}")
    implementation(
        "io.opentelemetry:opentelemetry-sdk-metrics:${rootProject.properties["opentelemetryVersion"]}",
    )

    // Tests: Kotest + MockK + Micronaut-Test (plan §14.10).
    testImplementation("io.micronaut.test:micronaut-test-junit5")
    testImplementation("io.kotest:kotest-runner-junit5:${rootProject.properties["kotestVersion"]}")
    testImplementation(
        "io.kotest:kotest-assertions-core:${rootProject.properties["kotestVersion"]}",
    )
    testImplementation("io.mockk:mockk:${rootProject.properties["mockkVersion"]}")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

micronaut {
    version("4.7.0")
    runtime("netty")
    testRuntime("kotest5")
    processing {
        incremental(true)
        annotations("dev.mtrace.api.*")
    }
}

// High-level layout per docs/plan-spike.md §12.2: hexagon/ and
// adapters/ live directly under apps/api/, not under
// src/main/kotlin/dev/mtrace/api/. Custom srcDirs make this work
// with the Micronaut/Gradle defaults.
sourceSets {
    main {
        java.srcDirs(emptyList<String>())
        kotlin.srcDirs("hexagon", "adapters")
        resources.srcDirs("resources")
    }
    test {
        java.srcDirs(emptyList<String>())
        kotlin.srcDirs("test")
    }
}

kotlin {
    jvmToolchain(21)
}

tasks.withType<Test>().configureEach {
    useJUnitPlatform()
    dependsOn("detekt")
}

configure<DetektExtension> {
    buildUponDefaultConfig = true
    allRules = false
    parallel = true
    ignoreFailures = false
    config.setFrom(file("detekt.yml"))
}

tasks.withType<Detekt>().configureEach {
    jvmTarget = "21"
    // The default source discovery uses Java sourceSets and does not
    // see our custom srcDirs ("hexagon", "adapters", "test").
    // Setting source explicitly ensures detekt scans the actual code.
    setSource(files("hexagon", "adapters", "test"))
    reports {
        html.required.set(true)
        xml.required.set(true)
        sarif.required.set(true)
        txt.required.set(false)
        md.required.set(false)
    }
}

tasks.named("check") {
    dependsOn("detekt")
}

// Custom task used by the deps Dockerfile stage to warm the
// Gradle dependency cache without compiling sources. See
// docs/plan-spike.md §14.11.
//
// NB: Micronaut's Gradle plugin declares dependencies without
// versions and resolves them through a platform/BOM. Resolving
// every configuration (e.g. *DependenciesMetadata) before the BOM
// has been applied would try to download `io.micronaut:micronaut-
// inject:` (empty version) and fail. We therefore restrict warmup
// to the standard user-facing classpaths.
tasks.register("resolveAllDependencies") {
    group = "build setup"
    description = "Resolve compile/runtime/ksp classpaths to warm the Gradle cache."
    doLast {
        val targets = listOf(
            "compileClasspath",
            "runtimeClasspath",
            "testCompileClasspath",
            "testRuntimeClasspath",
            "kspClasspath",
            "kspTestClasspath",
            "detekt",
            "detektPlugins",
        )
        targets.forEach { name ->
            configurations.findByName(name)?.resolve()
        }
    }
}
