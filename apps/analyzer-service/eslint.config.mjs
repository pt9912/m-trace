// ESLint-Konfig für apps/analyzer-service.
//
// Bezug: spec/lastenheft.md §10.4 (Stream Analyzer); siehe
// eslint.config.shared.mjs am Repo-Root für die geteilte Base und
// die Begründungen der SOLID-nahen Schwellen + Carveouts.

import { sharedConfig } from '../../eslint.config.shared.mjs';
import globals from 'globals';

export default [
  ...sharedConfig({
    tsconfigRootDir: import.meta.dirname,
  }),
  {
    // Node.js-Runtime — globals sind Node-Standard.
    languageOptions: {
      globals: globals.node,
    },
  },
];
