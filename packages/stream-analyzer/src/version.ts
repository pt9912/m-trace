import packageJson from "../package.json" with { type: "json" };

export const STREAM_ANALYZER_NAME: string = packageJson.name;
export const STREAM_ANALYZER_VERSION: string = packageJson.version;
