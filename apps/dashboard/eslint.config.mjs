// ESLint-Konfig für apps/dashboard.
//
// Bezug: spec/lastenheft.md §10.2 (Frontend SvelteKit); siehe
// eslint.config.shared.mjs am Repo-Root für die geteilte Base.
//
// SvelteKit-Setup:
//  - sharedConfig deckt **/*.ts ab (typescript-eslint type-aware).
//  - eslint-plugin-svelte recommended-Set deckt **/*.svelte ab.
//  - Per-File-Override für **/*.svelte setzt svelte-eslint-parser
//    mit tseslint.parser als script-parser, damit <script lang="ts">-
//    Blöcke type-aware geparst werden können.
//
// Voraussetzung: `svelte-kit sync` muss vor `eslint .` gelaufen sein,
// damit `.svelte-kit/tsconfig.json` existiert (wird vom paketeigenen
// tsconfig.json extended). Das lint-Script erzwingt diese Reihenfolge.

import { sharedConfig } from '../../eslint.config.shared.mjs';
import globals from 'globals';
import sveltePlugin from 'eslint-plugin-svelte';
import svelteParser from 'svelte-eslint-parser';
import tseslint from 'typescript-eslint';

export default [
  ...sharedConfig({ tsconfigRootDir: import.meta.dirname }),
  ...sveltePlugin.configs.recommended,
  {
    files: ['**/*.svelte'],
    languageOptions: {
      parser: svelteParser,
      parserOptions: {
        parser: tseslint.parser,
        project: ['./tsconfig.json'],
        tsconfigRootDir: import.meta.dirname,
        extraFileExtensions: ['.svelte'],
      },
      globals: globals.browser,
    },
  },
  // Browser-globals für TS-Files (z. B. src/lib/api.ts nutzt fetch()).
  // Überschreibt nicht den Test-Carveout aus shared, der für
  // `**/*.test.ts`/`tests/**` Vitest- und Node-globals einsetzt.
  {
    files: ['**/*.ts'],
    languageOptions: {
      globals: globals.browser,
    },
  },

  // Carveout: SvelteKit-2.21+ hat eine type-safe Routing-API mit
  // `resolve('/sessions/[id]', { id })` aus `$app/paths`. Die
  // Migration aller heute existierenden Routen-Anchors ist eigene
  // Folge-Arbeit (Roadmap-Item TBD); aktuell deployment-irrelevant,
  // weil die App ohne base-path läuft.
  //
  // Bewusst eng auf `routes/**` gefasst: neue Svelte-Komponenten
  // außerhalb der Routen-Hierarchie (z. B. künftige shared
  // Components in src/lib/) bleiben default-strict, damit der
  // Migration-Backlog nicht stillschweigend weiterwächst.
  {
    files: ['**/routes/**/*.svelte'],
    rules: {
      'svelte/no-navigation-without-resolve': 'off',
    },
  },

];
