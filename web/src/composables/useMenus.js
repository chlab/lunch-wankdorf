import { ref } from 'vue';
import { RESTAURANTS, combineMenus, fetchMenu } from '../util/menu';

/**
 * Loads every restaurant's menu for the week `date` falls into and exposes the
 * combined result. Restaurants fail independently: one unreachable or malformed
 * file drops that restaurant, and only a complete failure surfaces an error.
 */
export function useMenus() {
  const menu = ref({});
  const availableRestaurants = ref([]);
  const loading = ref(true);
  const error = ref(null);

  // Guards against an earlier load (e.g. a slow refresh) overwriting a later one
  let latestRequest = 0;

  const loadMenus = async (date = new Date()) => {
    const request = ++latestRequest;
    loading.value = true;
    error.value = null;

    const results = await Promise.all(
      RESTAURANTS.map(async (restaurant) => {
        try {
          return { restaurant, data: await fetchMenu(restaurant, date) };
        } catch (err) {
          console.error(`Failed to load menu for ${restaurant}:`, err);
          return null;
        }
      })
    );

    if (request !== latestRequest) return;

    const loaded = results.filter(Boolean);
    if (loaded.length === 0) {
      menu.value = {};
      availableRestaurants.value = [];
      error.value = 'Die Menüs konnten nicht geladen werden.';
    } else {
      const { menu: combined, restaurants } = combineMenus(loaded);
      menu.value = combined;
      availableRestaurants.value = restaurants;
    }

    loading.value = false;
  };

  return { menu, availableRestaurants, loading, error, loadMenus };
}
