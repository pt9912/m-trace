# Build stage — installiert Workspace-Deps (Dev + Prod) für player-sdk
# und dashboard, baut beide Pakete und konsolidiert die produktiven
# Artefakte über `pnpm deploy --prod` in `/deploy`. Damit braucht die
# Runtime-Stage keinen weiteren pnpm-install-Schritt mehr — und vor
# allem keine dev-Dependencies wie vitest/tsup, deren transitive
# picomatch-Range-Deklarationen Trivy sonst als HIGH meldet
# (CVE-2026-33671; effektiv installiert ist 4.0.4, aber package.json-
# Range-Floors landen ohne `--prod` im Image). Analog Analyzer-Service.
# Digest ermitteln via `docker buildx imagetools inspect node:22-trixie-slim`.
FROM node:22-trixie-slim@sha256:7b8a0c89c54499bee567618f96578e1a12a800f062fbdbfd1fb6a443fa6f6284 AS build

WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/dashboard/package.json apps/dashboard/package.json
COPY packages/player-sdk/package.json packages/player-sdk/package.json

RUN pnpm install --frozen-lockfile --ignore-scripts --filter @pt9912/m-trace-dashboard...

COPY apps/dashboard apps/dashboard
COPY packages/player-sdk packages/player-sdk

RUN pnpm --filter @pt9912/player-sdk run build && pnpm --filter @pt9912/m-trace-dashboard run build

# `pnpm deploy --prod` konsolidiert die produktiven Artefakte in ein
# selbst-tragendes `/deploy`-Bundle ohne Dev-Deps.
#
# OHNE `--legacy`, bewusst — siehe Analyzer-Service-Dockerfile und
# `injectWorkspacePackages` in `pnpm-workspace.yaml`: die
# Alt-Implementierung legte hier einen Symlink nach
# `../../../app/packages/player-sdk` ab, der im Runtime-Image ins
# Leere zeigt. Das Dashboard ueberlebte das nur, weil adapter-node
# die player-sdk bereits zur Build-Zeit in `build/` bundelt und zur
# Laufzeit nichts nachlaedt — der analyzer-service, der sein
# Workspace-Paket zur Laufzeit `require`t, starb daran.
RUN pnpm --filter @pt9912/m-trace-dashboard deploy --prod /deploy

# Runtime stage — startet direkt aus dem deploy-Bundle, ohne pnpm-
# Binär oder weiteren install-Schritt. SvelteKit-adapter-node legt
# das produktive Bundle unter `build/` an; CMD startet von dort.
# Digest ermitteln via `docker buildx imagetools inspect node:22-trixie-slim`.
FROM node:22-trixie-slim@sha256:7b8a0c89c54499bee567618f96578e1a12a800f062fbdbfd1fb6a443fa6f6284 AS runtime

WORKDIR /app

# `node:*-slim`-Images bringen `npm` unter
# `/usr/local/lib/node_modules/npm/` mit. NPM trägt eine eigene
# `picomatch@4.0.3`-Kopie in seinen internen `node_modules`, die
# Trivy als HIGH meldet (CVE-2026-33671). Im Runtime-Image starten
# wir mit `node build` und brauchen npm nicht — entfernen schließt
# diese CVE-Quelle ohne Workarounds.
RUN rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx

# `node:22-trixie-slim` baeckt den Debian-Paketstand zum Zeitpunkt des
# Base-Image-Builds ein und bleibt bei einem gepinnten Tag hinter
# spaeteren Debian-Security-Updates zurueck (u.a. CVE-2026-11822/11824,
# FTS5-Memory-Corruption in libsqlite3-0; CVE-2026-86145/89161,
# PCRE2-Memory-Corruption in libpcre2-8-0 — beide im trixie/main-Repo
# bereits als *u2 verfuegbar, im Base-Image-Layer aber noch *u1). Volles
# `apt-get upgrade` statt Paket-fuer-Paket-Pins, damit kuenftige
# Debian-Point-Releases automatisch mitkommen statt bei jedem Nightly-
# Fund eine neue Zeile zu brauchen.
RUN apt-get update && apt-get upgrade -y \
 && rm -rf /var/lib/apt/lists/*

# Alle SUID-/SGID-Bits im Runtime abziehen. Die trixie-slim-Base
# bringt neun solcher Binaries mit, allesamt Passwort-/Account-
# Verwaltung plus die mount-Familie:
#   SUID root:   chsh, passwd, gpasswd, su, newgrp, chfn, mount, umount
#   SGID shadow: unix_chkpwd, chage, expiry
# Keines hat im Runtime etwas zu suchen (USER node, reiner
# Node-Startpfad). Jedes ist ein Privilege-Escalation- bzw.
# Shadow-Hash-Zugriffsvektor, falls jemand Code-Ausfuehrung als node
# erlangt — Defense-in-Depth, kein von aussen erreichbarer Pfad.
# Konkreter Anlass war CVE-2026-53614 (SUID-`mount(8)` erlaubt
# nosuid/noexec-Bypass via LIBMOUNT_FORCE_MOUNT2; dazu
# CVE-2026-53612/53613, TOCTOU im selben mount-Pfad). Die uebrigen
# sieben sind dieselbe Klasse und fallen mit derselben Zeile.
#
# Bewusst generisch per `find` statt einer Pfadliste: zieht die Base
# spaeter ein weiteres SUID-Binary ein, ist es automatisch erfasst —
# eine Liste muesste jemand nachfuehren, und niemand wuerde es.
#
# Wichtig: Trivy-Befunde bleiben davon UNBERUEHRT. Erkannt wird das
# dpkg-Paket, nicht das Dateibit — anders als beim npm-Rauswurf
# darueber, den Trivy ueber `package.json` sieht. Die Suppressions in
# `.security/vulnignore.yaml` decken deshalb den Scanner ab, dieses
# `chmod` den tatsaechlichen Vektor. Beides zusammen, nicht als
# Ersatz fuereinander.
RUN find / -xdev -type f \( -perm -4000 -o -perm -2000 \) -exec chmod ug-s {} +

ENV HOST=0.0.0.0
ENV PORT=5173
ENV NODE_ENV=production

COPY --from=build /deploy /app

USER node
EXPOSE 5173

CMD ["node", "build"]
