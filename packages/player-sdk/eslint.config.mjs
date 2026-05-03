// ESLint-Konfig für packages/player-sdk.
//
// Bezug: spec/lastenheft.md §10.3 (Player-SDK); siehe
// eslint.config.shared.mjs am Repo-Root für die geteilte Base.
//
// Browser-SDK: globals enthalten Window/DOM-APIs zusätzlich zu den
// minimalen Defaults. Boundary-Check (`core/` ⇸ Browser-Adapter,
// hls.js) und Public-API-Snapshot bleiben über die projekteigenen
// scripts/check-*.mjs erhalten — siehe Schritt-30-Planung Pfad B.

import { sharedConfig } from '../../eslint.config.shared.mjs';
import globals from 'globals';

export default [
  ...sharedConfig({
    tsconfigRootDir: import.meta.dirname,
  }),
  {
    languageOptions: {
      globals: globals.browser,
    },
  },
];
