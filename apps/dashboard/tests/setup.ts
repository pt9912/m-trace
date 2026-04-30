import { vi } from "vitest";

vi.mock("$env/dynamic/public", () => ({
  env: {}
}));
