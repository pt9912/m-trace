import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

const rootPackage = readJSON("package.json");
const sdkPackage = readJSON("packages/player-sdk/package.json");
const eventSchema = readJSON("contracts/event-schema.json");
const sdkCompat = readJSON("contracts/sdk-compat.json");
const sdkVersionSource = readText("packages/player-sdk/src/version.ts");
const apiSource = readText("apps/api/hexagon/application/register_playback_event_batch.go");
const telemetryModel = readText("docs/telemetry-model.md");
const playerSDKDoc = readText("docs/player-sdk.md");

const exportedSDKName = matchConst(sdkVersionSource, "PLAYER_SDK_NAME");
const exportedSDKVersion = matchConst(sdkVersionSource, "PLAYER_SDK_VERSION");
const exportedSchemaVersion = matchConst(sdkVersionSource, "EVENT_SCHEMA_VERSION");
const apiSupportedSchemaVersion = matchGoConst(apiSource, "SupportedSchemaVersion");

assert(rootPackage.version === sdkPackage.version, "root package version and player-sdk package version must match");
assert(sdkPackage.name === sdkCompat.package_name, "player-sdk package name must match contracts/sdk-compat.json");
assert(sdkPackage.version === sdkCompat.sdk_version, "player-sdk package version must match contracts/sdk-compat.json");
assert(exportedSDKName === sdkCompat.package_name, "exported PLAYER_SDK_NAME must match contracts/sdk-compat.json");
assert(exportedSDKVersion === sdkCompat.sdk_version, "exported PLAYER_SDK_VERSION must match contracts/sdk-compat.json");
assert(exportedSchemaVersion === eventSchema.schema_version, "exported EVENT_SCHEMA_VERSION must match contracts/event-schema.json");
assert(sdkCompat.wire_schema_version === eventSchema.schema_version, "sdk compat wire schema must match event schema");
assert(sdkCompat.api_supported_schema_version === apiSupportedSchemaVersion, "sdk compat API schema must match apps/api SupportedSchemaVersion");
assert(apiSupportedSchemaVersion === eventSchema.schema_version, "API SupportedSchemaVersion must match contracts/event-schema.json");

for (const doc of [
  ["docs/telemetry-model.md", telemetryModel],
  ["docs/player-sdk.md", playerSDKDoc]
]) {
  assert(doc[1].includes("contracts/event-schema.json"), `${doc[0]} must reference contracts/event-schema.json`);
  assert(doc[1].includes("contracts/sdk-compat.json"), `${doc[0]} must reference contracts/sdk-compat.json`);
}

function readJSON(relativePath) {
  return JSON.parse(readText(relativePath));
}

function readText(relativePath) {
  return readFileSync(path.join(repoRoot, relativePath), "utf8");
}

function matchConst(source, name) {
  const match = source.match(new RegExp(`export const ${name} = "([^"]+)"`));
  assert(match, `missing exported const ${name}`);
  return match[1];
}

function matchGoConst(source, name) {
  const match = source.match(new RegExp(`const ${name} = "([^"]+)"`));
  assert(match, `missing Go const ${name}`);
  return match[1];
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}
