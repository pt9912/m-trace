#!/usr/bin/env node
// Einzel-Ton-Detektor per Goertzel-Algorithmus (Einzel-Bin-DFT).
//
// Liest rohe Mono-PCM-Samples als little-endian float32 von stdin (das
// Format, das `ffmpeg ... -f f32le -` liefert) und prueft, ob ein
// dominanter Sinuston bei der Zielfrequenz vorliegt. Eingesetzt vom
// WebRTC-Ton-Smoke (`scripts/smoke-webrtc-tone.sh`): der Lab-Publisher
// erzeugt `sine=frequency=1000` (examples/webrtc/ffmpeg-rtsp-loop.sh),
// dieser Check verifiziert, dass die Medien-Pipeline diesen Ton sauber
// bis zum Egress traegt — komplementaer zum getStats-Drift-Smoke, der
// nur `bytesReceived>0` (Medien fliessen), nicht die Tonqualitaet prueft.
//
// Methode: Goertzel liefert die DFT-Bin-Leistung |X_k|^2 bei einer
// bekannten Frequenz dependency-frei (keine FFT-Lib). Diskriminierend
// ist der Energie-Anteil des Ziel-Bands an der Gesamtenergie
// (Parseval): fuer einen reinen, exakt auf dem Bin liegenden Ton
// konzentriert sich ~die Haelfte der Gesamtenergie im positiven
// Frequenz-Bin (die andere Haelfte im konjugierten -f-Bin), also
// Anteil ~0.5; eine abwesende Frequenz oder breitbandiges Rauschen
// liefert ~0. Ein schmales Band um die Zielfrequenz faengt Leckage bei
// nicht-exakter Bin-Lage (reale, codec-/resample-behaftete Toene) ab.
//
// Verdict (beide muessen gelten):
//   1. RMS des Signals >= --min-rms → keine Stille / kein Decode-Fehler.
//   2. Ziel-Band-Energie / Gesamtenergie >= --min-fraction → sauberer Ton an
//      der Zielfrequenz, kein Rauschen, keine falsche Frequenz.
//
// Exit 0 = PASS, Exit 1 = FAIL (Verdict verletzt), Exit 2 = Nutzungsfehler.

function parseArgs(argv) {
  const opts = {
    rate: 48000,
    freq: 1000,
    bandHz: 3,
    minFraction: 0.2,
    minRms: 0.01,
  };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    const next = () => {
      const v = argv[i + 1];
      if (v === undefined) throw new Error(`missing value for ${arg}`);
      i += 1;
      return v;
    };
    switch (arg) {
      case "--rate": opts.rate = Number(next()); break;
      case "--freq": opts.freq = Number(next()); break;
      case "--band-hz": opts.bandHz = Number(next()); break;
      case "--min-fraction": opts.minFraction = Number(next()); break;
      case "--min-rms": opts.minRms = Number(next()); break;
      default: throw new Error(`unknown argument: ${arg}`);
    }
  }
  if (!(opts.rate > 0) || !(opts.freq > 0)) {
    throw new Error("--rate and --freq must be positive");
  }
  if (!(opts.minFraction > 0 && opts.minFraction <= 1)) {
    throw new Error("--min-fraction must be in (0, 1]");
  }
  return opts;
}

// Goertzel-Leistung (|X_k|^2, unnormiert) bei Frequenz `freq` ueber `samples`.
function goertzelPower(samples, rate, freq) {
  const omega = (2 * Math.PI * freq) / rate;
  const coeff = 2 * Math.cos(omega);
  let s1 = 0;
  let s2 = 0;
  for (let n = 0; n < samples.length; n += 1) {
    const s0 = samples[n] + coeff * s1 - s2;
    s2 = s1;
    s1 = s0;
  }
  return s1 * s1 + s2 * s2 - coeff * s1 * s2;
}

function readStdin() {
  return new Promise((resolve, reject) => {
    const chunks = [];
    process.stdin.on("data", (c) => chunks.push(c));
    process.stdin.on("end", () => resolve(Buffer.concat(chunks)));
    process.stdin.on("error", reject);
  });
}

async function main() {
  let opts;
  try {
    opts = parseArgs(process.argv.slice(2));
  } catch (err) {
    console.error(`[check-tone] usage error: ${err.message}`);
    console.error(
      "[check-tone] usage: ffmpeg ... -f f32le - | node scripts/check-tone.mjs " +
        "[--rate 48000] [--freq 1000] [--band-hz 3] [--min-fraction 0.2] [--min-rms 0.01]",
    );
    process.exit(2);
  }

  const raw = await readStdin();
  const count = Math.floor(raw.length / 4);
  // readFloatLE statt Float32Array-View: robust gegen unausgerichtete
  // Buffer-Offsets aus Buffer.concat.
  const samples = new Float32Array(count);
  for (let i = 0; i < count; i += 1) samples[i] = raw.readFloatLE(i * 4);

  if (samples.length < opts.rate / 10) {
    console.error(
      `[check-tone] FAIL: too few samples (${samples.length}); need >= ${opts.rate / 10} ` +
        "(>=0.1s) for a stable single-bin estimate.",
    );
    process.exit(1);
  }

  let sumSquares = 0;
  for (let n = 0; n < samples.length; n += 1) sumSquares += samples[n] * samples[n];
  const rms = Math.sqrt(sumSquares / samples.length);

  // Ziel-Band [freq-bandHz .. freq+bandHz] in 1-Hz-Schritten aufsummieren,
  // um Leckage bei nicht-exakter Bin-Lage einzufangen.
  let bandPower = 0;
  for (let f = opts.freq - opts.bandHz; f <= opts.freq + opts.bandHz; f += 1) {
    if (f > 0 && f < opts.rate / 2) bandPower += goertzelPower(samples, opts.rate, f);
  }
  // Parseval: Summe |X_k|^2 ueber alle N Bins = N * Summe(x^2).
  const totalEnergy = samples.length * sumSquares;
  const fraction = totalEnergy > 0 ? bandPower / totalEnergy : 0;

  console.log(
    `[check-tone] samples=${samples.length} rms=${rms.toFixed(4)} ` +
      `freq=${opts.freq}Hz(±${opts.bandHz}) energy-fraction=${fraction.toFixed(3)} ` +
      `(min-rms=${opts.minRms} min-fraction=${opts.minFraction})`,
  );

  const failures = [];
  if (rms < opts.minRms) {
    failures.push(`rms ${rms.toFixed(4)} < ${opts.minRms} (silence/decode failure)`);
  }
  if (fraction < opts.minFraction) {
    failures.push(
      `energy fraction ${fraction.toFixed(3)} < ${opts.minFraction} ` +
        `(no clean ${opts.freq}Hz tone: noise or wrong frequency)`,
    );
  }

  if (failures.length) {
    console.error(`[check-tone] FAIL: ${failures.join("; ")}`);
    process.exit(1);
  }
  console.log(`[check-tone] OK -- clean ${opts.freq}Hz tone detected.`);
}

main().catch((err) => {
  console.error(`[check-tone] unexpected error: ${err.stack || err}`);
  process.exit(2);
});
