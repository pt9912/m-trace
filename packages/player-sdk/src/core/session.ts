export function createSessionId(random: () => number = Math.random): string {
  const now = Date.now().toString(36);
  const entropy = Array.from({ length: 12 }, () => Math.floor(random() * 36).toString(36)).join("");
  return `${now}-${entropy}`;
}
