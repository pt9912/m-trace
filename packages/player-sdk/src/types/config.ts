import type { PlaybackEventBatch } from "./events";

export interface Transport {
  send(batch: PlaybackEventBatch): Promise<void>;
}

/**
 * Liefert den Wert für den W3C-`traceparent`-Header pro Batch-Send.
 *
 * Format laut https://www.w3.org/TR/trace-context/#traceparent-header-field-values:
 * `00-<trace_id 32 lower-hex>-<parent_id 16 lower-hex>-<flags 2 hex>`.
 *
 * **Synchron**: Der Provider muss synchron antworten — er wird im
 * Hot Path direkt vor `fetch()` aufgerufen. Eine `Promise`-Rückgabe
 * wird nicht awaited und landet als ungültiger Header beim Server.
 *
 * **Geworfene Fehler werden geschluckt**: Provider-Throws fängt das
 * SDK still und sendet den Batch ohne Header weiter — Tracing darf
 * den Event-Pfad nicht sabotieren.
 *
 * Der SDK-HTTP-Transport sendet den Header nur, wenn die Funktion
 * einen nicht-leeren String zurückgibt — `undefined` oder `""`
 * → kein Header (Server-Fallback erzeugt eigene Trace, siehe
 * spec/telemetry-model.md §2.5). Das SDK validiert das Format
 * **nicht**: ein vom Provider gelieferter Müllstring landet beim
 * Server, der ihn als Parse-Error markiert (`mtrace.trace.parse_error=true`)
 * und zur eigenen Trace-ID zurückfällt.
 */
export type TraceParentProvider = () => string | undefined;

export interface PlayerSDKConfig {
  endpoint: string;
  token: string;
  projectId: string;
  sessionId?: string;
  batchSize?: number;
  flushIntervalMs?: number;
  sampleRate?: number;
  maxQueueEvents?: number;
  transport?: Transport;
  /**
   * Optionaler Provider für den `traceparent`-Header (W3C Trace
   * Context). Wenn gesetzt und nicht-leer, propagiert das SDK den
   * Wert pro Batch-Send. Konsumenten, die OpenTelemetry-JS oder
   * eine eigene Tracing-Schicht nutzen, können hier den aktuell
   * aktiven Span-Context bereitstellen — z. B.
   * `() => activeSpan?.spanContext() && formatTraceparent(activeSpan)`.
   *
   * Ohne Provider sendet das SDK keinen Header, der Server
   * generiert einen Root-Span. Abwärtskompatibel mit Backends < 0.4.0
   * (HTTP-Standard ignoriert unbekannte Header).
   */
  traceparent?: TraceParentProvider;
}
