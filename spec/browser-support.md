# Browser-Support-Matrix

> **Bezug**: F-58..F-67.

## Einstufung

| Umgebung | Status | Gate |
|---|---|---|
| Chrome Desktop, aktuelle stabile Version | supported | `make browser-e2e` Chromium |
| Firefox Desktop, aktuelle stabile Version | supported | `make browser-e2e` Firefox |
| Safari Desktop, aktuelle stabile Version | documented limitation | Basis-Playback über native HLS; keine vollständige hls.js-Eventparität |
| Chromium-basierte Browser | best effort | über Chromium-Gate mitabgedeckt, keine separate Matrix |
| iOS Safari | out of scope | kein MVP-Gate |
| Android Chrome | out of scope | kein MVP-Gate |
| Smart-TV Browser | out of scope | ausdrücklich nicht im MVP-Scope |
| Embedded WebViews | out of scope | ausdrücklich nicht im MVP-Scope |

## Regeln

- `supported` bedeutet: Die Demo-Integration muss im Browser-E2E-Gate
  laufen und hls.js als primären Integrationspfad nutzen.
- `documented limitation` bedeutet: Die Laufzeit darf nicht brechen,
  aber Event-Tiefe und Adapter-Parität sind eingeschränkt dokumentiert.
- `out of scope` bedeutet: Kein Release-Gate und keine implizite
  Support-Zusage.

Safari Desktop bleibt eingeschränkt (F-61), weil native HLS deutlich
weniger Playback-Introspektion liefert als hls.js. Die Demo-Route hat
einen nativen HLS-Fallback; der Player-SDK-hls.js-Adapter ist dort aber
nicht vollständig aktiv.
