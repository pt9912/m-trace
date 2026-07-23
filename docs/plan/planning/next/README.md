# next/

Zwischenstufe im Planning-Lifecycle des v3.5.0-Kanons:

```
open/  →  next/  →  in-progress/  →  done/
```

- **`open/`** — Backlog: geschnitten, aber noch nicht eingeplant.
- **`next/`** — als Nächstes vorgesehen: priorisiert und startbereit, aber noch
  nicht in Arbeit. Diese Stufe.
- **`in-progress/`** — aktiv in Umsetzung.
- **`done/`** — abgeschlossen; trägt eine Closure-Note (ADR-0010, für neue
  Pläne).

Ein Plan wandert per `git mv` zwischen den Stufen (reiner Move, dann
Inhalts-/Link-Anpassung als zweiter Commit — `AGENTS.md` §3.3). Solange keine
Arbeit ansteht, bleibt dieses Verzeichnis leer (nur diese README).
