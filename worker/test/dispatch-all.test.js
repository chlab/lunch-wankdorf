import { describe, expect, test } from "vitest";

import { dispatchAll } from "../src/dispatch-all.js";

const silent = { log: () => {}, error: () => {} };

describe("dispatchAll", () => {
  test("dispatches every workflow it is given", async () => {
    const dispatched = [];
    const dispatch = async (workflow) => dispatched.push(workflow);

    await dispatchAll(["a.yml", "b.yml"], {}, { dispatch, log: silent });

    expect(dispatched).toEqual(["a.yml", "b.yml"]);
  });

  test("dispatches nothing when the run is a dry run", async () => {
    const dispatched = [];
    const dispatch = async (workflow) => dispatched.push(workflow);

    await dispatchAll(
      ["a.yml"],
      { DRY_RUN: "true" },
      { dispatch, log: silent },
    );

    expect(dispatched).toEqual([]);
  });

  test("dispatches the remaining workflows when one fails", async () => {
    const dispatched = [];
    const dispatch = async (workflow) => {
      if (workflow === "a.yml") throw new Error("boom");
      dispatched.push(workflow);
    };

    await expect(
      dispatchAll(["a.yml", "b.yml"], {}, { dispatch, log: silent }),
    ).rejects.toThrow("boom");

    expect(dispatched).toEqual(["b.yml"]);
  });

  test("resolves quietly when nothing is due", async () => {
    const dispatch = async () => {
      throw new Error("should not be called");
    };

    await expect(
      dispatchAll([], {}, { dispatch, log: silent }),
    ).resolves.toBeUndefined();
  });
});
