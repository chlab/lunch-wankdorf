export function getISOWeekNumber(date = new Date()) {
    const target = new Date(date.valueOf());
  
    // Set to nearest Thursday: current date + 4 - current day number
    // Make Sunday (0) into 7
    const dayNumber = (date.getDay() + 6) % 7; // Monday = 0, Sunday = 6
    target.setDate(target.getDate() - dayNumber + 3);
  
    // January 4 is always in week 1
    const firstThursday = new Date(target.getFullYear(), 0, 4);
    const firstDayNumber = (firstThursday.getDay() + 6) % 7;
    firstThursday.setDate(firstThursday.getDate() - firstDayNumber + 3);
  
    // Calculate week number
    const weekNumber = 1 + Math.round((target - firstThursday) / (7 * 24 * 60 * 60 * 1000));
  
    return weekNumber;
  }
  