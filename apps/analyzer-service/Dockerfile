# Build stage — installiert die Workspace-Deps (Dev + Prod) für
# stream-analyzer und analyzer-service, baut beide Pakete und
# konsolidiert die produktiven Artefakte über `pnpm deploy --prod`
# in `/deploy`. Damit braucht die Runtime-Stage keinen weiteren
# pnpm-install-Schritt mehr (plan-0.3.0 §9 Tranche 7.5/3).
FROM node:22-trixie-slim AS build
WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@11.20.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/analyzer-service/package.json apps/analyzer-service/
COPY packages/stream-analyzer/package.json packages/stream-analyzer/

RUN pnpm install --frozen-lockfile --ignore-scripts \
    --filter @pt9912/stream-analyzer... \
    --filter @pt9912/analyzer-service...

COPY packages/stream-analyzer packages/stream-analyzer
COPY apps/analyzer-service apps/analyzer-service

RUN pnpm --filter @pt9912/stream-analyzer run build \
 && pnpm --filter @pt9912/analyzer-service run build

# `pnpm deploy --prod` konsolidiert die produktiven Artefakte in ein
# selbst-tragendes `/deploy`-Bundle ohne Dev-Deps, das die
# Runtime-Stage als read-only uebernehmen kann.
#
# OHNE `--legacy`, bewusst: das Flag erzwingt die Alt-Implementierung,
# und die legte in `/deploy/node_modules/@pt9912/stream-analyzer` nur
# einen Symlink nach `../../../workspace/packages/stream-analyzer` ab.
# Im Build-Stage ist der aufloesbar — in die Runtime-Stage wandert
# aber nur `/deploy`, dort zeigte er ins Leere und das Image starb
# beim Start mit `Cannot find module '@pt9912/stream-analyzer'`. Das
# Bundle war also entgegen dem Kommentar NICHT selbst-tragend.
# Voraussetzung fuer den Verzicht auf `--legacy` ist
# `injectWorkspacePackages: true` in `pnpm-workspace.yaml` (dort
# ausfuehrlich begruendet); pnpm verlangt das seit v10 explizit.
RUN pnpm --filter @pt9912/analyzer-service deploy --prod /deploy

# Runtime stage — startet direkt aus dem deploy-Bundle, ohne pnpm
# binär oder einen weiteren install-Schritt.
FROM node:22-trixie-slim AS runtime
WORKDIR /app

# `node:*-trixie-slim`-Images bringen `npm` unter
# `/usr/local/lib/node_modules/npm/` mit. NPM tragt eine eigene
# `picomatch@4.0.3`-Kopie in seinen internen `node_modules`, die
# Trivy als HIGH meldet (CVE-2026-33671). Im Runtime-Image starten
# wir mit `node dist/main.cjs` und brauchen npm nicht — entfernen
# schliesst diese CVE-Quelle ohne Workarounds (analog Dashboard-
# Dockerfile, plan-0.8.5 Tranche 1 Closeout).
RUN rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx

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

ENV NODE_ENV=production \
    ANALYZER_LISTEN_HOST=0.0.0.0 \
    ANALYZER_LISTEN_PORT=7000

COPY --from=build /deploy ./

USER node
EXPOSE 7000

CMD ["node", "dist/main.cjs"]
