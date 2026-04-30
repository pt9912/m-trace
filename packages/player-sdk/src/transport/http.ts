import type { Transport } from "../types/config";
import type { PlaybackEventBatch } from "../types/events";

export class HttpTransport implements Transport {
  constructor(
    private readonly endpoint: string,
    private readonly token: string
  ) {}

  async send(batch: PlaybackEventBatch): Promise<void> {
    const response = await fetch(this.endpoint, {
      method: "POST",
      credentials: "omit",
      headers: {
        "Content-Type": "application/json",
        "X-MTrace-Token": this.token
      },
      body: JSON.stringify(batch)
    });

    if (!response.ok) {
      throw new Error(`m-trace transport failed: ${response.status}`);
    }
  }
}
