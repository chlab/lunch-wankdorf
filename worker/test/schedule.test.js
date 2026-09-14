import { describe, expect, test } from "vitest";

import { dueJobs } from "../src/schedule.js";

describe("dueJobs", () => {
  test("dispatches the menu fetch at 06:00 Zurich time in summer", () => {
    expect(dueJobs(new Date("2026-06-15T04:00:00Z"))).toEqual([
      "weekly-menu-fetch.yml",
    ]);
  });

  test("dispatches the menu fetch at 06:00 Zurich time in winter", () => {
    expect(dueJobs(new Date("2026-01-12T05:00:00Z"))).toEqual([
      "weekly-menu-fetch.yml",
    ]);
  });

  test("skips the UTC hour that is 06:00 local only in the other DST offset", () => {
    expect(dueJobs(new Date("2026-06-15T03:00:00Z"))).toEqual([]);
    expect(dueJobs(new Date("2026-01-12T04:00:00Z"))).toEqual([]);
  });

  test("dispatches the photo fetch on weekdays at 08:00 Zurich time", () => {
    expect(dueJobs(new Date("2026-06-16T06:00:00Z"))).toEqual([
      "daily-photo-fetch.yml",
    ]);
  });

  test("does not dispatch the photo fetch at the weekend", () => {
    expect(dueJobs(new Date("2026-06-20T06:00:00Z"))).toEqual([]);
    expect(dueJobs(new Date("2026-06-21T06:00:00Z"))).toEqual([]);
  });

  test("dispatches both Monday jobs at their own hours, never together", () => {
    expect(dueJobs(new Date("2026-06-15T04:00:00Z"))).toEqual([
      "weekly-menu-fetch.yml",
    ]);
    expect(dueJobs(new Date("2026-06-15T06:00:00Z"))).toEqual([
      "daily-photo-fetch.yml",
    ]);
  });

  test("dispatches the prune at 23:00 Zurich time on Sunday", () => {
    expect(dueJobs(new Date("2026-06-21T21:00:00Z"))).toEqual([
      "weekly-menu-prune.yml",
    ]);
    expect(dueJobs(new Date("2026-01-11T22:00:00Z"))).toEqual([
      "weekly-menu-prune.yml",
    ]);
  });

  test("dispatches the prune on the days the clocks change", () => {
    expect(dueJobs(new Date("2026-03-29T21:00:00Z"))).toEqual([
      "weekly-menu-prune.yml",
    ]);
    expect(dueJobs(new Date("2026-10-25T22:00:00Z"))).toEqual([
      "weekly-menu-prune.yml",
    ]);
  });

  test("stays quiet on an hour nothing is scheduled for", () => {
    expect(dueJobs(new Date("2026-06-17T12:00:00Z"))).toEqual([]);
  });

  test("treats local midnight as hour 0, not hour 24", () => {
    expect(dueJobs(new Date("2026-06-14T22:00:00Z"))).toEqual([]);
  });
});
