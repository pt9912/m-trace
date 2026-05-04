import { writable, type Writable } from "svelte/store";
import { env } from "$env/dynamic/public";

/**
 * Mini-Statusquellen für den Dashboard-Status-Pfad
 * (plan-0.4.0 §5 H3, F-39).
 *
 * Drei dokumentierte Quellen:
 *  - API-Erreichbarkeit über `GET /api/health` (siehe `getHealth` in
 *    `lib/api.ts`)
 *  - SSE-Verbindungszustand des Dashboard-Clients (Placeholder bis
 *    Tranche-4 §5 H5; aktuell `not_yet_connected` bzw. `disabled`,
 *    sobald der SSE-Client eingewired ist)
 *  - Letzter Session-Read-Fehler aus realen `getSession`/
 *    `listSessions`-Aufrufen
 *
 * Aggregierte Invalid-/Dropped-/Rate-Limit-Zähler werden bewusst
 * NICHT konsumiert — Plan-DoD §5 verschiebt das auf Tranche 6.
 */

export type SseConnectionState =
  | "not_yet_connected"
  | "connecting"
  | "connected"
  | "polling_fallback"
  | "disabled";

export interface SseConnectionStatus {
  state: SseConnectionState;
  /** Letzter Statuswechsel; `null` wenn der Client noch nicht
   *  initialisiert ist. */
  changedAt: string | null;
  /** Optionale, kurze Diagnose (z. B. „network error"). */
  detail: string | null;
}

export interface ReadErrorRecord {
  /** Pfad oder Kontext, an dem der Fehler entstand. */
  source: string;
  /** Lesbare Fehlermeldung. */
  message: string;
  /** ISO-Timestamp. */
  occurredAt: string;
}

/**
 * SSE-Verbindungszustand des Dashboard-SSE-Clients. In Tranche 4
 * §5 H3 gibt es noch keinen SSE-Client; der Store steht auf
 * `not_yet_connected`. §5 H5 verdrahtet den fetch-basierten
 * SSE-Client und schreibt in diesen Store.
 */
export const sseConnection: Writable<SseConnectionStatus> = writable({
  state: "not_yet_connected",
  changedAt: null,
  detail: null
});

/**
 * Letzter Session-Read-Fehler. `null` wenn keiner aufgetreten ist
 * (oder wenn `clearLastReadError` aufgerufen wurde).
 */
export const lastReadError: Writable<ReadErrorRecord | null> = writable(null);

/**
 * Konsumenten der Read-API rufen das bei jedem Fehler auf, damit die
 * Status-Sektion den aktuellen Stand zeigt.
 */
export function recordReadError(source: string, err: unknown): void {
  const message = err instanceof Error ? err.message : String(err);
  lastReadError.set({
    source,
    message,
    occurredAt: new Date().toISOString()
  });
}

export function clearLastReadError(): void {
  lastReadError.set(null);
}

/**
 * Konfigurierbare Service-Links (plan-0.4.0 §5 F-40, Variante
 * "konfigurationsgetriebene Link-Section"). Caller liefern den
 * Schlüssel; Rückgabe ist die URL oder `undefined`, wenn die
 * Env-Variable nicht gesetzt ist. Konsumenten zeigen "configured but
 * unreachable" oder verstecken den Link, wenn URL fehlt.
 *
 * Dokumentierte Keys:
 *  - PUBLIC_GRAFANA_URL    → Grafana-Dashboard-Origin
 *  - PUBLIC_PROMETHEUS_URL → Prometheus-Web-UI-Origin
 *  - PUBLIC_MEDIAMTX_URL   → MediaMTX-Web-/HLS-Origin
 *  - PUBLIC_OTEL_HEALTH_URL → OTel-Collector-Health-Endpoint
 */
export type ServiceLinkKey =
  | "PUBLIC_GRAFANA_URL"
  | "PUBLIC_PROMETHEUS_URL"
  | "PUBLIC_MEDIAMTX_URL"
  | "PUBLIC_OTEL_HEALTH_URL";

export function configuredServiceUrl(
  key: ServiceLinkKey,
  source: Readonly<Record<string, string | undefined>> = env
): string | undefined {
  const raw = source[key];
  if (typeof raw !== "string" || raw.trim() === "") {
    return undefined;
  }
  return raw.trim();
}

/**
 * Service-Link-Eintrag für die Status-Page. Vereint Konfig-Lookup,
 * Probe-URL-Ableitung und initialen Status — die Page rendert nur
 * noch und ruft `probeServices` für die Reachability-Updates.
 */
export interface ServiceLink {
  name: string;
  envKey: ServiceLinkKey;
  /** ID der konfigurierten Origin (für UI-Hinweise wie
   *  "PUBLIC_GRAFANA_URL not set"). */
  configHint: string;
  openUrl?: string;
  probeUrl?: string;
  status: "connected" | "inactive" | "not_configured";
}

const serviceDefinitions: ReadonlyArray<{
  name: string;
  envKey: ServiceLinkKey;
  /** Berechnet die optionale Probe-URL aus der konfigurierten
   *  Origin. */
  probePath?: (origin: string) => string;
  /** Ob die konfigurierte Origin auch ein Open-Link wird. Default
   *  true. */
  hasOpenUrl?: boolean;
}> = [
  {
    name: "Prometheus",
    envKey: "PUBLIC_PROMETHEUS_URL",
    probePath: (origin) => `${origin.replace(/\/$/, "")}/-/ready`
  },
  {
    name: "Grafana",
    envKey: "PUBLIC_GRAFANA_URL",
    probePath: (origin) => `${origin.replace(/\/$/, "")}/api/health`
  },
  {
    name: "OTel Collector",
    envKey: "PUBLIC_OTEL_HEALTH_URL",
    hasOpenUrl: false
  },
  {
    name: "MediaMTX",
    envKey: "PUBLIC_MEDIAMTX_URL"
  }
];

/**
 * Baut die Service-Link-Liste für die Status-Page aus
 * `$env/dynamic/public`. Ohne gesetzte Env-Variable bekommt ein
 * Service `status="not_configured"`. Reine Funktion ohne UI-
 * Abhängigkeit, damit die Logik isoliert testbar ist.
 */
export function buildServiceLinks(
  source: Readonly<Record<string, string | undefined>> = env
): ServiceLink[] {
  return serviceDefinitions.map((def) => {
    const origin = configuredServiceUrl(def.envKey, source);
    if (!origin) {
      return {
        name: def.name,
        envKey: def.envKey,
        configHint: def.envKey,
        status: "not_configured"
      };
    }
    return {
      name: def.name,
      envKey: def.envKey,
      configHint: def.envKey,
      openUrl: def.hasOpenUrl === false ? undefined : origin,
      probeUrl: def.probePath ? def.probePath(origin) : origin,
      status: "inactive"
    };
  });
}

/**
 * Probet alle in `links` aufgeführten Services parallel und mappt
 * `connected`/`inactive`. Services ohne `probeUrl` (z. B. weil nicht
 * konfiguriert) bleiben auf ihrem aktuellen Status.
 */
export async function probeServices(links: ServiceLink[]): Promise<ServiceLink[]> {
  const probes = links.map(async (link): Promise<ServiceLink> => {
    if (!link.probeUrl) {
      return link;
    }
    try {
      await fetch(link.probeUrl, { mode: "no-cors", cache: "no-store" });
      return { ...link, status: "connected" };
    } catch {
      return { ...link, status: "inactive" };
    }
  });
  return Promise.all(probes);
}

/**
 * Reactive Service-Link-Liste für die Status-Page. Defaults auf
 * `buildServiceLinks()` (Env-getrieben). Tests können den Store
 * direkt überschreiben (z. B. mit `set([{...openUrl: "..."}])`),
 * um UI-Branches zu pinnen, die ohne Env-Mocking nicht erreichbar
 * sind.
 */
export const observabilityServices: Writable<ServiceLink[]> = writable(buildServiceLinks());
