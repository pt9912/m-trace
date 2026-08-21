# Build stage — installiert Workspace-Deps (Dev + Prod) für player-sdk
# und dashboard, baut beide Pakete und konsolidiert die produktiven
# Artefakte über `pnpm deploy --prod` in `/deploy`. Damit braucht die
# Runtime-Stage keinen weiteren pnpm-install-Schritt mehr — und vor
# allem keine dev-Dependencies wie vitest/tsup, deren transitive
# picomatch-Range-Deklarationen Trivy sonst als HIGH meldet
# (CVE-2026-33671; effektiv installiert ist 4.0.4, aber package.json-
# Range-Floors landen ohne `--prod` im Image). Analog Analyzer-Service.
FROM node:22-trixie-slim AS build

WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/dashboard/package.json apps/dashboard/package.json
COPY packages/player-sdk/package.json packages/player-sdk/package.json

RUN pnpm install --frozen-lockfile --ignore-scripts --filter @pt9912/m-trace-dashboard...

COPY apps/dashboard apps/dashboard
COPY packages/player-sdk packages/player-sdk

RUN pnpm --filter @pt9912/player-sdk run build && pnpm --filter @pt9912/m-trace-dashboard run build

# `pnpm deploy --prod --legacy` flacht die Workspace-Symlinks in eine
# selbst-tragende node_modules-Hierarchie ab und lässt dev-Deps weg.
# `--legacy` ist nötig, weil pnpm v10+ sonst inject-workspace-packages
# erwartet (siehe Analyzer-Service-Dockerfile).
RUN pnpm --filter @pt9912/m-trace-dashboard deploy --prod --legacy /deploy

# Runtime stage — startet direkt aus dem deploy-Bundle, ohne pnpm-
# Binär oder weiteren install-Schritt. SvelteKit-adapter-node legt
# das produktive Bundle unter `build/` an; CMD startet von dort.
FROM node:22-trixie-slim AS runtime

WORKDIR /app

# `node:*-slim`-Images bringen `npm` unter
# `/usr/local/lib/node_modules/npm/` mit. NPM trägt eine eigene
# `picomatch@4.0.3`-Kopie in seinen internen `node_modules`, die
# Trivy als HIGH meldet (CVE-2026-33671). Im Runtime-Image starten
# wir mit `node build` und brauchen npm nicht — entfernen schließt
# diese CVE-Quelle ohne Workarounds.
RUN rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx

# SUID-Bits von `mount`/`umount` abziehen. Beide liegen in der
# trixie-slim-Base als `-rwsr-xr-x root`; das Runtime ruft sie nie auf
# (USER node, reiner Node-Startpfad). CVE-2026-53614 nutzt genau dieses
# SUID-`mount(8)` fuer einen nosuid/noexec-Bypass via
# LIBMOUNT_FORCE_MOUNT2, CVE-2026-53612/53613 sind TOCTOU-Races im
# selben mount-Pfad. Ohne SUID ist der Eskalationspfad zu — auch
# solange die Base `util-linux 2.41-5` statt des bereits paketierten
# Trixie-Fixes `2.41.5-0+deb13u1` traegt.
#
# Wichtig: der Trivy-Befund BLEIBT dadurch bestehen. Erkannt wird das
# dpkg-Paket, nicht das Dateibit — anders als beim npm-Rauswurf
# darueber, den Trivy ueber `package.json` sieht. Die Suppression in
# `.security/vulnignore.yaml` deckt deshalb den Scanner ab, dieses
# `chmod` den tatsaechlichen Vektor. Beides zusammen, nicht als Ersatz
# fuereinander.
RUN chmod u-s /usr/bin/mount /usr/bin/umount

ENV HOST=0.0.0.0
ENV PORT=5173
ENV NODE_ENV=production

COPY --from=build /deploy /app

USER node
EXPOSE 5173

CMD ["node", "build"]
