// ESLint-Konfig für packages/stream-analyzer.
//
// Bezug: spec/lastenheft.md §10.4 (Stream Analyzer); siehe
// eslint.config.shared.mjs am Repo-Root für die geteilte Base.
//
// Boundary-Check (Public ⇸ internal/) und Public-API-Snapshot
// laufen weiterhin über scripts/check-boundaries.mjs und
// scripts/check-public-api.mjs — projekteigene Scripts, die nicht
// durch ESLint-import-Plugin ersetzt werden (siehe Schritt-30-
// Planung, Pfad B).

import { sharedConfig } from '../../eslint.config.shared.mjs';
import globals from 'globals';

export default [
  ...sharedConfig({
    tsconfigRootDir: import.meta.dirname,
  }),
  {
    // Node.js-Runtime — Stream-Analyzer läuft im Backend
    // (analyzer-service ruft das Paket auf) und in der CLI.
    languageOptions: {
      globals: globals.node,
    },
  },
];
