import { describe, expect, test } from "vitest";

import { withRetry } from "../src/retry.js";

const noSleep = async () => {};

describe("withRetry", () => {
  test("returns the result without retrying when the call succeeds", async () => {
    let attempts = 0;
    const result = await withRetry(
      async () => {
        attempts++;
        return "ok";
      },
      { sleep: noSleep },
    );

    expect(result).toBe("ok");
    expect(attempts).toBe(1);
  });

  test("retries until the call succeeds", async () => {
    let attempts = 0;
    const result = await withRetry(
      async () => {
        attempts++;
        if (attempts < 3) throw new Error("transient");
        return "ok";
      },
      { sleep: noSleep },
    );

    expect(result).toBe("ok");
    expect(attempts).toBe(3);
  });

  test("gives up after three attempts and rethrows the last error", async () => {
    let attempts = 0;
    const failing = async () => {
      attempts++;
      throw new Error(`attempt ${attempts} failed`);
    };

    await expect(withRetry(failing, { sleep: noSleep })).rejects.toThrow(
      "attempt 3 failed",
    );
    expect(attempts).toBe(3);
  });

  test("backs off for longer between each attempt", async () => {
    const waits = [];
    const failing = async () => {
      throw new Error("nope");
    };

    await expect(
      withRetry(failing, { sleep: async (ms) => waits.push(ms) }),
    ).rejects.toThrow("nope");

    expect(waits).toHaveLength(2);
    expect(waits[1]).toBeGreaterThan(waits[0]);
  });
});
