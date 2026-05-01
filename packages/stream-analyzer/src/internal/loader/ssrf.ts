/**
 * SSRF-Schutzschicht für den Manifest-Loader. Tranche-2-DoD verlangt
 * scheme-Whitelist, Credentials-Block und harte Sperren für lokale,
 * private, link-local, loopback, reservierte und Multicast-Bereiche
 * (IPv4 und IPv6). Die Funktionen sind reine Datenfunktionen, damit
 * sie ohne Netzwerk testbar sind.
 *
 * DNS-Rebinding-Entscheidung (siehe `docs/user/stream-analyzer.md`
 * §6): der Loader klassifiziert den Hostnamen einmal — IP-Literale
 * werden direkt gegen die Sperrlisten geprüft, Domain-Hostnamen über
 * `LoaderRuntime.resolveHost` aufgelöst — und delegiert die eigent-
 * liche Verbindung anschließend an die globale `fetch`-Implemen-
 * tierung. Ein TCP-Pin gegen die validierte IP wäre Tranche-6-Arbeit
 * (custom Dispatcher); 0.3.0 verzichtet bewusst darauf und verlässt
 * sich für vollen Rebinding-Schutz auf eine zusätzliche Egress-/
 * Firewall-Schicht.
 */

const ALLOWED_SCHEMES: ReadonlySet<string> = new Set(["http:", "https:"]);

export type SsrfBlockReason =
  | "scheme_not_allowed"
  | "credentials_in_url"
  | "host_missing"
  | "ip_blocked"
  | "ip_unparseable";

export interface SsrfDecision {
  readonly ok: boolean;
  readonly reason?: SsrfBlockReason;
  readonly detail?: Readonly<Record<string, unknown>>;
}

const OK: SsrfDecision = { ok: true };

/** Validiert URL-Schema, Credentials und Host (nicht die IP). */
export function validateUrl(url: URL): SsrfDecision {
  if (!ALLOWED_SCHEMES.has(url.protocol)) {
    return {
      ok: false,
      reason: "scheme_not_allowed",
      detail: { protocol: url.protocol }
    };
  }
  if (url.username !== "" || url.password !== "") {
    return { ok: false, reason: "credentials_in_url" };
  }
  if (url.hostname === "") {
    return { ok: false, reason: "host_missing" };
  }
  return OK;
}

/**
 * Prüft eine aufgelöste IP-Adresse gegen die Sperrlisten.
 * `family` ist 4 oder 6 (analog Node `dns.lookup`-Result).
 */
export function validateResolvedIp(address: string, family: 4 | 6): SsrfDecision {
  if (family === 4) {
    const numeric = parseIPv4(address);
    if (numeric === null) {
      return { ok: false, reason: "ip_unparseable", detail: { address, family } };
    }
    if (isBlockedIpv4(numeric)) {
      return { ok: false, reason: "ip_blocked", detail: { address, family } };
    }
    return OK;
  }
  const parsed = parseIPv6(address);
  if (parsed === null) {
    return { ok: false, reason: "ip_unparseable", detail: { address, family } };
  }
  if (isBlockedIpv6(parsed)) {
    return { ok: false, reason: "ip_blocked", detail: { address, family } };
  }
  return OK;
}

interface CidrV4 {
  readonly base: number;
  readonly prefix: number;
}

const BLOCKED_IPV4: ReadonlyArray<CidrV4> = [
  { base: ipv4(0, 0, 0, 0), prefix: 8 }, // current network / unspecified
  { base: ipv4(10, 0, 0, 0), prefix: 8 }, // private
  { base: ipv4(100, 64, 0, 0), prefix: 10 }, // CGN
  { base: ipv4(127, 0, 0, 0), prefix: 8 }, // loopback
  { base: ipv4(169, 254, 0, 0), prefix: 16 }, // link-local
  { base: ipv4(172, 16, 0, 0), prefix: 12 }, // private
  { base: ipv4(192, 0, 0, 0), prefix: 24 }, // IETF protocol assignment
  { base: ipv4(192, 0, 2, 0), prefix: 24 }, // TEST-NET-1
  { base: ipv4(192, 88, 99, 0), prefix: 24 }, // 6to4 anycast
  { base: ipv4(192, 168, 0, 0), prefix: 16 }, // private
  { base: ipv4(198, 18, 0, 0), prefix: 15 }, // benchmarking
  { base: ipv4(198, 51, 100, 0), prefix: 24 }, // TEST-NET-2
  { base: ipv4(203, 0, 113, 0), prefix: 24 }, // TEST-NET-3
  { base: ipv4(224, 0, 0, 0), prefix: 4 }, // multicast
  { base: ipv4(240, 0, 0, 0), prefix: 4 } // reserved (incl. 255.255.255.255)
];

function isBlockedIpv4(numeric: number): boolean {
  for (const cidr of BLOCKED_IPV4) {
    if (matchesCidrV4(numeric, cidr)) return true;
  }
  return false;
}

function matchesCidrV4(numeric: number, cidr: CidrV4): boolean {
  const mask = cidr.prefix === 0 ? 0 : (0xffffffff << (32 - cidr.prefix)) >>> 0;
  return ((numeric & mask) >>> 0) === ((cidr.base & mask) >>> 0);
}

function ipv4(a: number, b: number, c: number, d: number): number {
  return ((a << 24) | (b << 16) | (c << 8) | d) >>> 0;
}

export function parseIPv4(address: string): number | null {
  const parts = address.split(".");
  if (parts.length !== 4) return null;
  let result = 0;
  for (const part of parts) {
    if (!/^[0-9]+$/.test(part)) return null;
    if (part.length > 1 && part.startsWith("0")) return null; // no leading zeros
    const n = Number(part);
    if (n < 0 || n > 255) return null;
    result = (result << 8) | n;
  }
  return result >>> 0;
}

interface CidrV6 {
  /** 8 hextets, big-endian. */
  readonly base: Uint16Array;
  /** Prefix in bits (0..128). */
  readonly prefix: number;
}

const BLOCKED_IPV6: ReadonlyArray<CidrV6> = [
  { base: hextets("0:0:0:0:0:0:0:0"), prefix: 128 }, // unspecified
  { base: hextets("0:0:0:0:0:0:0:1"), prefix: 128 }, // loopback
  { base: hextets("0:0:0:0:0:ffff:0:0"), prefix: 96 }, // IPv4-mapped
  { base: hextets("64:ff9b:0:0:0:0:0:0"), prefix: 96 }, // IPv4/IPv6 translation
  { base: hextets("100:0:0:0:0:0:0:0"), prefix: 64 }, // discard-only
  { base: hextets("2001:db8:0:0:0:0:0:0"), prefix: 32 }, // documentation
  { base: hextets("fc00:0:0:0:0:0:0:0"), prefix: 7 }, // unique local
  { base: hextets("fe80:0:0:0:0:0:0:0"), prefix: 10 }, // link-local
  { base: hextets("ff00:0:0:0:0:0:0:0"), prefix: 8 } // multicast
];

function isBlockedIpv6(parsed: Uint16Array): boolean {
  for (const cidr of BLOCKED_IPV6) {
    if (matchesCidrV6(parsed, cidr)) return true;
  }
  // IPv4-mapped IPv6 (::ffff:a.b.c.d) hit BLOCKED_IPV6[2] above; for
  // any other address that decodes to an embedded IPv4 (e.g. SIIT
  // 64:ff9b::a.b.c.d), the v4 subnet matches and is already blocked.
  return false;
}

function matchesCidrV6(addr: Uint16Array, cidr: CidrV6): boolean {
  let bitsLeft = cidr.prefix;
  for (let i = 0; i < 8; i++) {
    if (bitsLeft <= 0) return true;
    if (bitsLeft >= 16) {
      if (addr[i] !== cidr.base[i]) return false;
      bitsLeft -= 16;
      continue;
    }
    const mask = (0xffff << (16 - bitsLeft)) & 0xffff;
    if ((addr[i] & mask) !== (cidr.base[i] & mask)) return false;
    return true;
  }
  return true;
}

function hextets(canonical: string): Uint16Array {
  const parts = canonical.split(":");
  const result = new Uint16Array(8);
  for (let i = 0; i < 8; i++) {
    result[i] = parseInt(parts[i], 16);
  }
  return result;
}

export function parseIPv6(address: string): Uint16Array | null {
  // Strip zone id ("fe80::1%eth0") — we do not honour it.
  const zoneIdx = address.indexOf("%");
  const cleaned = zoneIdx === -1 ? address : address.slice(0, zoneIdx);

  // Embedded IPv4 form: "::ffff:a.b.c.d" or "1:2:3:4:5:6:a.b.c.d".
  const dotIdx = cleaned.lastIndexOf(".");
  let head = cleaned;
  let tailHextets: number[] | null = null;
  if (dotIdx !== -1) {
    const colonIdx = cleaned.lastIndexOf(":");
    if (colonIdx === -1 || colonIdx > dotIdx) return null;
    const v4 = parseIPv4(cleaned.slice(colonIdx + 1));
    if (v4 === null) return null;
    head = cleaned.slice(0, colonIdx);
    tailHextets = [(v4 >>> 16) & 0xffff, v4 & 0xffff];
  }

  // "::" expansion.
  let groups: string[];
  const dcIdx = head.indexOf("::");
  if (dcIdx === -1) {
    groups = head === "" ? [] : head.split(":");
  } else {
    if (head.indexOf("::", dcIdx + 1) !== -1) return null;
    const left = head.slice(0, dcIdx);
    const right = head.slice(dcIdx + 2);
    const leftGroups = left === "" ? [] : left.split(":");
    const rightGroups = right === "" ? [] : right.split(":");
    const expectedTail = tailHextets ? 2 : 0;
    const fillCount = 8 - expectedTail - leftGroups.length - rightGroups.length;
    if (fillCount < 0) return null;
    groups = [...leftGroups, ...Array.from({ length: fillCount }, () => "0"), ...rightGroups];
  }

  const expected = tailHextets ? 6 : 8;
  if (groups.length !== expected) return null;

  const out = new Uint16Array(8);
  for (let i = 0; i < groups.length; i++) {
    const g = groups[i];
    if (!/^[0-9a-fA-F]{1,4}$/.test(g)) return null;
    out[i] = parseInt(g, 16);
  }
  if (tailHextets) {
    out[6] = tailHextets[0];
    out[7] = tailHextets[1];
  }
  return out;
}
