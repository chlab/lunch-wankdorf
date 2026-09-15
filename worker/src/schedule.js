const TIME_ZONE = "Europe/Zurich";

const WEEKDAYS = { Sun: 0, Mon: 1, Tue: 2, Wed: 3, Thu: 4, Fri: 5, Sat: 6 };

// Local times in Europe/Zurich. Keep in sync with the healthchecks.io schedules.
export const SCHEDULE = [
  { workflow: "weekly-menu-fetch.yml", weekdays: [1], hour: 6 },
  { workflow: "daily-photo-fetch.yml", weekdays: [1, 2, 3, 4, 5], hour: 11 },
  { workflow: "weekly-menu-prune.yml", weekdays: [0], hour: 23 },
];

// h23 keeps local midnight at 0; hour12:false reports it as 24 in some runtimes.
const formatter = new Intl.DateTimeFormat("en-US", {
  timeZone: TIME_ZONE,
  weekday: "short",
  hour: "numeric",
  hourCycle: "h23",
});

function localWeekdayAndHour(date) {
  const parts = Object.fromEntries(
    formatter.formatToParts(date).map(({ type, value }) => [type, value]),
  );
  return { weekday: WEEKDAYS[parts.weekday], hour: Number(parts.hour) };
}

export function dueJobs(date) {
  const { weekday, hour } = localWeekdayAndHour(date);
  return SCHEDULE.filter(
    (job) => job.hour === hour && job.weekdays.includes(weekday),
  ).map((job) => job.workflow);
}
