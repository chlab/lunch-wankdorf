import { getISOWeekNumber, getISOWeekYear, WEEKDAYS } from './date';
import foodtrucksMenu from '../foodtrucks.json';

/**
 * A single dish as rendered by MenuItem.
 *
 * @typedef {Object} MenuItem
 * @property {string} name
 * @property {string} description
 * @property {string} [type] - 'meat', 'vegetarian' or 'vegan'
 * @property {string} [icon] - icons8 icon name, see MenuIcon
 * @property {string} [link] - absolute http(s) URL, '' when there is none
 * @property {string} restaurant - display name, e.g. 'Gira'
 * @property {string} [foodtruck] - name of the truck, foodtruck items only
 */

// The R2 bucket doesn't allow localhost, so local dev can point at a mirror instead
const BASE_URL =
  import.meta.env.VITE_MENU_BASE_URL ?? 'https://pub-201cbf927f0b4c8991d32485a57b9d40.r2.dev';

// Restaurants we fetch a weekly file for. Freibank and Turbolama are disabled: they
// kept moving their menu around, so nothing is scraped for them and the file 404s.
export const RESTAURANTS = ['gira', 'luna', 'sole', 'espace'];

export const FOODTRUCKS = 'Foodtrucks';
export const FOODTRUCKS_ENABLED = false;

// Restaurants pinned to the end of the list, in this order
export const APPENDED_RESTAURANTS = [FOODTRUCKS, 'Turbolama'];

// Espace only publishes a weekly PDF, so its links never point at a dish
const RESTAURANTS_WITHOUT_LINKS = ['Espace'];

// 'gira' -> 'Gira'
const displayName = (restaurant) => restaurant.charAt(0).toUpperCase() + restaurant.slice(1);

// Menu data is scraped, so a link is only trusted once it parses as an absolute http(s) URL
const safeLink = (link) => {
  try {
    const url = new URL(link);
    return url.protocol === 'http:' || url.protocol === 'https:' ? url.href : '';
  } catch {
    return '';
  }
};

/** @returns {MenuItem} */
const toMenuItem = (item, restaurant) => ({
  ...item,
  restaurant,
  link: RESTAURANTS_WITHOUT_LINKS.includes(restaurant) ? '' : safeLink(item.link ?? ''),
  icon: item.icon ?? '',
  foodtruck: item.foodtruck ?? '',
});

const addItems = (combined, day, items) => {
  combined[day] = [...(combined[day] ?? []), ...items];
};

/**
 * Fetch one restaurant's menu file for the ISO week `date` falls into.
 * Rejects on network errors and non-2xx responses.
 */
export async function fetchMenu(restaurant, date) {
  const filename = `${restaurant}_${getISOWeekNumber(date)}_${getISOWeekYear(date)}.json`;
  const response = await fetch(`${BASE_URL}/${filename}`);
  if (!response.ok) {
    throw new Error(`${response.status} ${response.statusText}`);
  }
  return response.json();
}

/**
 * Merge the fetched restaurant menus (plus the static foodtrucks menu) into a
 * map of weekday -> items. A malformed file only drops its own restaurant.
 *
 * @param {{restaurant: string, data: Object}[]} menus
 * @returns {{menu: Object<string, MenuItem[]>, restaurants: string[]}}
 */
export function combineMenus(menus) {
  const combined = {};
  const daily = [];
  const weekly = [];

  menus.forEach(({ restaurant, data }) => {
    const name = displayName(restaurant);
    try {
      if (data.type === 'daily' && data.menu) {
        Object.entries(data.menu).forEach(([day, items]) => {
          addItems(
            combined,
            day.toLowerCase(),
            items.map((item) => toMenuItem(item, name))
          );
        });
        daily.push(name);
      } else if (data.type === 'weekly' && Array.isArray(data.menu)) {
        const items = data.menu.map((item) => toMenuItem(item, name));
        WEEKDAYS.forEach((day) => addItems(combined, day, items));
        weekly.push(name);
      } else {
        console.error(`Unexpected menu format for ${name}:`, data.type);
      }
    } catch (err) {
      console.error(`Skipping malformed menu for ${name}:`, err);
    }
  });

  if (FOODTRUCKS_ENABLED && foodtrucksMenu.type === 'daily' && foodtrucksMenu.menu) {
    Object.entries(foodtrucksMenu.menu).forEach(([day, items]) => {
      addItems(
        combined,
        day.toLowerCase(),
        items.filter((item) => item.enabled).map((item) => toMenuItem(item, FOODTRUCKS))
      );
    });
  }

  return {
    menu: combined,
    restaurants: [...daily, ...(FOODTRUCKS_ENABLED ? [FOODTRUCKS] : []), ...weekly],
  };
}
