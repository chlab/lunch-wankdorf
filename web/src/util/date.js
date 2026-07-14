// ISO order: the menu week runs Monday to Sunday, matching getISOWeekNumber
const DAYS = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];

// Days a menu is published for
export const WEEKDAYS = DAYS.slice(0, 5);

// Date.getDay() is Sunday-based, the ISO week is Monday-based
const isoDayIndex = (date) => (date.getDay() + 6) % 7;

// The ISO week is the one containing its Thursday, which is what decides the year
// a week belongs to: Dec 29 2025 is week 1 of 2026, Jan 1 2027 is week 53 of 2026.
const thursdayOf = (date) => {
  const thursday = new Date(date.valueOf());
  thursday.setDate(thursday.getDate() - isoDayIndex(thursday) + 3);
  return thursday;
};

export function getISOWeekNumber(date = new Date()) {
  const target = thursdayOf(date);

  // January 4 is always in week 1
  const firstThursday = thursdayOf(new Date(target.getFullYear(), 0, 4));

  return 1 + Math.round((target - firstThursday) / (7 * 24 * 60 * 60 * 1000));
}

/** The year the ISO week belongs to, which is what the backend names files with. */
export function getISOWeekYear(date = new Date()) {
  return thursdayOf(date).getFullYear();
}

/**
 * The weekday to show for `date`. On the weekend we fall back to Friday: the
 * menus for the coming week are only published on Monday, so the newest data we
 * have is the one for the week that just ended.
 */
export function getSelectableDay(date = new Date()) {
  const day = DAYS[isoDayIndex(date)];
  return WEEKDAYS.includes(day) ? day : 'friday';
}

/** The date `dayName` falls on within the ISO week containing `from`. */
export function getDateForDay(dayName, from = new Date()) {
  const date = new Date(from);
  date.setDate(date.getDate() + (DAYS.indexOf(dayName) - isoDayIndex(from)));
  return date;
}
