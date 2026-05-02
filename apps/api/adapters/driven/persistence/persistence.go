// Package persistence ist Sammelort für Driven-Adapter, die Sessions
// und Playback-Events durable halten. Die konkreten Implementierungen
// leben in Sub-Paketen:
//
//   - persistence/inmemory: prozesslokal, nicht restart-stabil, nur
//     für Tests und expliziten Dev-Fallback (ADR-0002 §8.2).
//   - persistence/sqlite:   SQLite-Datei, restart-stabil, Default ab
//     `0.4.0` im Compose-Lab (ADR-0002 §8.1).
//
// Dieses Paket selbst hält nur shared Konstanten und Helfer, die von
// beiden Adapter-Klassen genutzt werden — keine Anwendungs- oder
// Domain-Logik.
package persistence

// SessionEndedEventName ist der Event-Name, der eine Session direkt
// auf State=Ended schaltet (plan-0.1.0.md §5.1 Sub-Item 8;
// telemetry-model.md §1.3). Wird vom UpsertFromEvents-Pfad jedes
// Adapters konsumiert.
const SessionEndedEventName = "session_ended"
