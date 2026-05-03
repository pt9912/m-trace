// Shared ESLint base for the m-trace workspace (Schritt 30).
//
// Bezug:
//   - spec/lastenheft.md §10.2 (Frontend), §10.3 (Player-SDK), §10.4
//     (Stream Analyzer) — alle drei verlangen "SOLID-nahe Boundary-/
//     Komplexitätsregeln" als Qualitätsanteil neben tsc/svelte-check.
//   - docs/user/quality.md §1.1 — gemeinsames SOLID-nahes Profil:
//     Import-Boundaries, verbotene Deep Imports, Komplexität,
//     Funktionslänge, Verschachtelung, stabile Public APIs.
//
// Pfad B aus der Schritt-30-Planung: ESLint deckt Komplexität +
// TS-Idiome + (paketspezifisch) Svelte ab. Boundary-Checks und
// Public-API-Snapshots bleiben als bestehende custom Scripts pro
// Paket — sie funktionieren und sind projektspezifisch knapp.
//
// Per-Paket-Configs in `apps/*/eslint.config.mjs` und
// `packages/*/eslint.config.mjs` importieren diese Base und ergänzen
// (a) `parserOptions.project` für die paket-eigene tsconfig.json und
// (b) optionale paket-spezifische Carveouts (Browser-Globals,
// Svelte-Plugin, generated-code-Pfade).
//
// `//eslint-disable`-Pragmas bleiben ausgeschlossen, analog zur
// Go-Linie in apps/api/.golangci.yml — Carveouts sind Scope-
// Definitionen, keine Suppressions.

import tseslint from 'typescript-eslint';
import globals from 'globals';

// SOLID-nahe Schwellen (Iter-1, analog zur Go-Linie). Werden nach
// erstem Lauf justiert, falls sie unrealistisch sind.
const COMPLEXITY = 15;
const MAX_LINES_PER_FUNCTION = 100;
const MAX_STATEMENTS_PER_FUNCTION = 30;
const MAX_DEPTH = 4;
const MAX_NESTED_CALLBACKS = 3;

/**
 * Liefert die geteilte Base-Konfig. Aufrufer übergeben die Wurzel-
 * Verzeichnisse ihres Pakets, damit `parserOptions.project` und
 * `parserOptions.tsconfigRootDir` korrekt aufgelöst werden.
 *
 * Alle Regel-Blöcke werden auf `tsFiles` (Default `**\/*.ts`)
 * eingeschränkt, damit der typescript-eslint-Parser nicht versucht,
 * Svelte-/Vue-/sonstige Non-TS-Files zu parsen. SvelteKit-Apps
 * ergänzen ihre eigenen `**\/*.svelte`-Blöcke nach dem Spread.
 *
 * @param {object} opts
 * @param {string} opts.tsconfigRootDir — Absoluter Pfad zum Paket
 *   (`import.meta.dirname` im Aufrufer).
 * @param {string[]} [opts.tsconfigProject=['./tsconfig.json']] —
 *   Project-Files für typescript-eslint.
 * @param {string[]} [opts.tsFiles=['**\/*.ts']] — Glob-Filter, auf
 *   den die TS-spezifischen Regeln eingeschränkt werden.
 * @returns {import('eslint').Linter.Config[]}
 */
export function sharedConfig({
  tsconfigRootDir,
  tsconfigProject = ['./tsconfig.json'],
  tsFiles = ['**/*.ts'],
}) {
  // typescript-eslint recommendedTypeChecked liefert eine Liste von
  // Configs; wir spread sie und ergänzen jeden Block, der Regeln
  // enthält, um den `files`-Filter, damit .svelte/.vue/.mjs-Files
  // nicht versuchsweise typescript-aware geparst werden.
  const tseslintConfigs = tseslint.configs.recommendedTypeChecked.map((cfg) =>
    cfg.rules ? { ...cfg, files: tsFiles } : cfg
  );

  return [
    // Globale Ignores für alle Aufrufer. Per-Paket-Overrides können
    // weitere ignores hinzufügen. Config-Files (eslint, vitest, vite,
    // tsup, svelte) liegen in der Regel außerhalb des tsconfig
    // `include` — sie zu type-aware zu linten verlangt eine separate
    // tsconfig.eslint.json je Paket; da Config-Files trivial bleiben,
    // ignorieren wir sie hier.
    {
      ignores: [
        '**/dist/**',
        '**/build/**',
        '**/coverage/**',
        '**/node_modules/**',
        '**/.svelte-kit/**',
        '**/*.d.ts',
        '**/eslint.config.*',
        '**/vitest.config.*',
        '**/vite.config.*',
        '**/tsup.config.*',
        '**/svelte.config.*',
        '**/playwright.config.*',
        // Build-/Boundary-/Snapshot-Scripts liegen häufig als .mjs
        // vor; typescript-eslint kann sie auch bei Aufnahme in
        // tsconfig.include nicht zuverlässig type-aware parsen.
        // tsc --noEmit prüft sie weiterhin, ESLint bleibt draußen.
        '**/scripts/**/*.{js,mjs,cjs}',
      ],
    },

    // typescript-eslint mit Type-Checked-Regeln. Liefert no-floating-
    // promises, no-misused-promises, prefer-readonly, no-unsafe-*,
    // restrict-template-expressions usw. Auf tsFiles eingeschränkt.
    ...tseslintConfigs,

    {
      files: tsFiles,
      languageOptions: {
        parserOptions: {
          project: tsconfigProject,
          tsconfigRootDir,
          ecmaVersion: 'latest',
          sourceType: 'module',
        },
      },
      rules: {
        // SOLID-nahe Komplexitäts-Schwellen (siehe oben).
        complexity: ['error', { max: COMPLEXITY }],
        'max-lines-per-function': [
          'error',
          {
            max: MAX_LINES_PER_FUNCTION,
            skipBlankLines: true,
            skipComments: true,
          },
        ],
        'max-statements': ['error', MAX_STATEMENTS_PER_FUNCTION],
        'max-depth': ['error', MAX_DEPTH],
        'max-nested-callbacks': ['error', MAX_NESTED_CALLBACKS],

        // Designsignal: Mutationen über Parameter sind ein
        // Boundary-Problem (Caller vs. Callee).
        'no-param-reassign': ['error', { props: false }],

        // Konsistente Toolchain-Hygiene. typescript-eslint kennt das
        // präzisere Pendant — wir aktivieren beides nicht doppelt.
        'no-unused-vars': 'off',
        '@typescript-eslint/no-unused-vars': [
          'error',
          { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
        ],
      },
    },

    // ----- Test-Carveout (analog zur Go-Linie) -----------------------
    // Tests dürfen Setup-/Match-/Iterations-Logik haben — Komplexitäts-
    // signale wirken designseitig auf Produktionscode, nicht auf
    // Test-Fixtures. tsc-/typecheck-Strenge bleibt voll erhalten.
    {
      files: [
        '**/*.test.ts',
        '**/*.spec.ts',
        '**/tests/**/*.ts',
        '**/test/**/*.ts',
      ],
      languageOptions: {
        globals: {
          ...globals.node,
          // Vitest stellt describe/it/expect bereit (vitest/globals
          // in tsconfig.types).
          describe: 'readonly',
          it: 'readonly',
          test: 'readonly',
          expect: 'readonly',
          beforeAll: 'readonly',
          beforeEach: 'readonly',
          afterAll: 'readonly',
          afterEach: 'readonly',
          vi: 'readonly',
        },
      },
      rules: {
        complexity: 'off',
        'max-lines-per-function': 'off',
        'max-statements': 'off',
        'max-depth': 'off',
        'max-nested-callbacks': 'off',

        // Vitest-Test-Bodies werden idiomatisch oft als
        // `async () => { ... }` geschrieben (Konsistenz über
        // alle Test-Cases hinweg, auch wenn das einzelne Body
        // synchron arbeitet). require-await wäre hier reines
        // Code-Style-Geräusch, kein Korrektheitssignal.
        '@typescript-eslint/require-await': 'off',

        // Mocks und Stubs werden oft als `any` getypt oder mit
        // type-assertions versehen, damit Test-Setup nicht den
        // gesamten echten Typ nachbauen muss. Die no-unsafe-*-
        // Regeln-Familie ist auf Produktionscode gemünzt; in Tests
        // verlieren sie den Nutzen, weil Test-Setup-Code bewusst
        // Typgrenzen umgeht.
        '@typescript-eslint/no-unsafe-assignment': 'off',
        '@typescript-eslint/no-unsafe-argument': 'off',
        '@typescript-eslint/no-unsafe-call': 'off',
        '@typescript-eslint/no-unsafe-member-access': 'off',
        '@typescript-eslint/no-unsafe-return': 'off',

        // Tests werfen bewusst Non-Error-Objekte (Strings, Numbers,
        // plain Objects) als Fixtures, um Error-Handling-Pfade auf
        // Robustheit gegen unerwartete Throw-Werte zu prüfen.
        // Produktions-Throws bleiben strikt auf Error-Klassen
        // eingeschränkt.
        '@typescript-eslint/only-throw-error': 'off',
      },
    },

    // ----- Skript-Zone (analog zur Go-Linie) -------------------------
    // Internes Build-/Tooling, kein Produkt-Code. Komplexität +
    // Funktionslänge gelockert, weil Tools (CLI-Builder, Snapshot-
    // Generatoren) andere Lesbarkeitsoptima haben.
    {
      files: ['**/scripts/**/*.{js,mjs,cjs,ts}'],
      languageOptions: {
        globals: globals.node,
      },
      rules: {
        complexity: 'off',
        'max-lines-per-function': 'off',
        'max-statements': 'off',
      },
    },
  ];
}
