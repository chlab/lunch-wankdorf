import { describe, expect, test } from "vitest";

import { dispatchWorkflow } from "../src/github.js";

const env = {
  GH_OWNER: "chlab",
  GH_REPO: "lunch-wankdorf",
  GH_REF: "main",
  GH_DISPATCH_TOKEN: "test-token",
};

describe("dispatchWorkflow", () => {
  test("posts a dispatch for the workflow on the configured ref", async () => {
    const calls = [];
    const fetchImpl = async (url, init) => {
      calls.push({ url, init });
      return new Response(null, { status: 204 });
    };

    await dispatchWorkflow(env, "weekly-menu-fetch.yml", fetchImpl);

    expect(calls).toHaveLength(1);
    expect(calls[0].url).toBe(
      "https://api.github.com/repos/chlab/lunch-wankdorf/actions/workflows/weekly-menu-fetch.yml/dispatches",
    );
    expect(calls[0].init.method).toBe("POST");
    expect(JSON.parse(calls[0].init.body)).toEqual({ ref: "main" });
  });

  test("authenticates with the dispatch token and identifies itself", async () => {
    let headers;
    const fetchImpl = async (_url, init) => {
      headers = init.headers;
      return new Response(null, { status: 204 });
    };

    await dispatchWorkflow(env, "weekly-menu-fetch.yml", fetchImpl);

    expect(headers.Authorization).toBe("Bearer test-token");
    expect(headers["User-Agent"]).toBeTruthy();
  });

  test("throws with the status and body when GitHub rejects the dispatch", async () => {
    const fetchImpl = async () =>
      new Response("Bad credentials", { status: 401 });

    await expect(
      dispatchWorkflow(env, "weekly-menu-fetch.yml", fetchImpl),
    ).rejects.toThrow(/401.*Bad credentials/s);
  });
});
