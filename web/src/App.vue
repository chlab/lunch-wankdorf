<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue';
import MenuItem from './components/MenuItem.vue';
import MenuItemFilter from './components/MenuItemFilter.vue';
import RestaurantFilter from './components/RestaurantFilter.vue';
import Skeleton from './components/Skeleton.vue';
import DateNavigator from './components/DateNavigator.vue';
import ViewToggle from './components/ViewToggle.vue';
import { useMenus } from './composables/useMenus';
import { APPENDED_RESTAURANTS, FOODTRUCKS } from './util/menu';
import { WEEKDAYS, getDateForDay, getSelectableDay } from './util/date';
import { dateSeed, pick, shuffle } from './util/random';

// The day the menu was loaded for. Held in state so a tab left open overnight can
// notice it went stale (see refreshIfStale).
const today = ref(new Date());

// ?day= makes a day linkable; anything outside Monday-Friday falls back to today
const dayFromURL = () => {
  const day = new URLSearchParams(window.location.search).get('day')?.toLowerCase();
  return WEEKDAYS.includes(day) ? day : getSelectableDay(today.value);
};

const selectedDay = ref(dayFromURL());
const currentDate = computed(() => getDateForDay(selectedDay.value, today.value));

const selectDay = (day, { replace = false } = {}) => {
  selectedDay.value = day;
  const url = new URL(window.location);
  url.searchParams.set('day', day);
  window.history[replace ? 'replaceState' : 'pushState']({}, '', url);
};

const { menu, availableRestaurants, loading, error, loadMenus } = useMenus();

const selectedRestaurant = ref('');
const vegetarianFilter = ref(false);
const compactView = ref(localStorage.getItem('compactView') === 'true');

watch(compactView, (value) => {
  localStorage.setItem('compactView', value);
});

const clearFilters = () => {
  selectedRestaurant.value = '';
  vegetarianFilter.value = false;
};

const formattedDate = computed(() =>
  new Intl.DateTimeFormat('de-CH', { weekday: 'long' }).format(currentDate.value)
);

// Navigation is bounded by the published days (Monday-Friday)
const canGoBack = computed(() => WEEKDAYS.indexOf(selectedDay.value) > 0);
const canGoForward = computed(() => WEEKDAYS.indexOf(selectedDay.value) < WEEKDAYS.length - 1);

// Navigate by delta days (-1 for previous, +1 for next)
const navigateDay = (delta) => {
  const day = WEEKDAYS[WEEKDAYS.indexOf(selectedDay.value) + delta];
  if (day) {
    selectDay(day);
  }
};

const goToPreviousDay = () => navigateDay(-1);
const goToNextDay = () => navigateDay(1);

const hasMenuForSelectedDay = computed(() => selectedDay.value in menu.value);

const filteredMenuItems = computed(() => {
  let items = menu.value[selectedDay.value] ?? [];

  if (selectedRestaurant.value) {
    items = items.filter((item) => item.restaurant === selectedRestaurant.value);
  }

  if (vegetarianFilter.value) {
    items = items.filter((item) => item.type === 'vegan' || item.type === 'vegetarian');
  }

  return items;
});

// Daily recommendation - a random menu item, but the same one all day
const dailyRecommendation = computed(() => {
  // Exclude foodtrucks as they aren't consistently there
  const items = filteredMenuItems.value.filter((item) => item.restaurant !== FOODTRUCKS);
  if (items.length === 0) return null;

  return pick(items, dateSeed(currentDate.value));
});

// Group menu items by restaurant for display
const groupedMenuItems = computed(() => {
  const items = filteredMenuItems.value;

  if (selectedRestaurant.value) {
    return items.length > 0 ? [{ restaurant: selectedRestaurant.value, items }] : [];
  }

  const groups = {};
  items.forEach((item) => {
    groups[item.restaurant] = [...(groups[item.restaurant] ?? []), item];
  });

  const isAppended = (restaurant) => APPENDED_RESTAURANTS.includes(restaurant);
  const toGroup = (restaurant) => ({ restaurant, items: groups[restaurant] });
  const names = Object.keys(groups);

  // Main restaurants are shuffled (stable for the day), the rest keeps a fixed order
  return [
    ...shuffle(names.filter((name) => !isAppended(name)), dateSeed(currentDate.value)).map(toGroup),
    ...APPENDED_RESTAURANTS.filter((name) => names.includes(name)).map(toGroup),
  ];
});

// Browser back/forward navigation
const handlePopState = () => {
  selectedDay.value = dayFromURL();
};

// Menu files are per ISO week, so a tab left open across midnight or into the next
// week keeps showing what it loaded. Re-check whenever the tab becomes visible.
const refreshIfStale = () => {
  const now = new Date();
  if (document.visibilityState !== 'visible' || now.toDateString() === today.value.toDateString()) {
    return;
  }
  today.value = now;
  selectDay(getSelectableDay(now), { replace: true });
  loadMenus(now);
};

onMounted(() => {
  loadMenus(today.value);
  window.addEventListener('popstate', handlePopState);
  document.addEventListener('visibilitychange', refreshIfStale);
});

onUnmounted(() => {
  window.removeEventListener('popstate', handlePopState);
  document.removeEventListener('visibilitychange', refreshIfStale);
});
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="top-0 pt-4">
      <DateNavigator
        :formatted-date="formattedDate"
        :can-go-back="canGoBack"
        :can-go-forward="canGoForward"
        @date-back="goToPreviousDay"
        @date-forward="goToNextDay"
      />
    </header>

    <main class="flex-grow container mx-auto py-6 px-4 max-w-4xl">
      <!-- Filters -->
      <div class="max-w-full mb-6 overflow-hidden">
        <div class="flex md:justify-center space-x-2 overflow-x-auto scrollbar-hide">
          <MenuItemFilter v-model="vegetarianFilter" />
          <RestaurantFilter v-model="selectedRestaurant" :restaurants="availableRestaurants" />
        </div>
      </div>

      <!-- loader -->
      <div v-if="loading" class="space-y-4">
        <Skeleton v-for="i in 2" :key="i" />
      </div>
      <div v-else-if="error" class="text-center py-8 text-red-500">
        {{ error }}
      </div>
      <!-- no menu published for this day -->
      <div v-else-if="!hasMenuForSelectedDay" class="text-center py-8 text-gray-500">
        Für diesen Tag ist kein Menü verfügbar.
      </div>
      <!-- everything got filtered away -->
      <div v-else-if="filteredMenuItems.length === 0" class="text-center py-8 text-gray-500">
        <p>Keine Menüs für die gewählten Filter.</p>
        <button
          @click="clearFilters"
          class="mt-3 px-3 py-1 rounded-full bg-gray-300 hover:bg-gray-400 hover:text-white transition-colors cursor-pointer text-xs"
        >
          Filter zurücksetzen
        </button>
      </div>
      <!-- menu items grouped by restaurant -->
      <div v-else class="space-y-6">
        <!-- Daily recommendation -->
        <div v-if="dailyRecommendation" class="mb-6">
          <div class="flex items-center justify-between mb-3 max-w-md mx-auto">
            <h2 class="text-lg font-semibold text-gray-700">Tagesempfehlung</h2>
            <ViewToggle v-model:compactView="compactView" />
          </div>
          <div :class="compactView ? 'max-w-md mx-auto rounded-lg shadow-md overflow-hidden bg-white p-4' : ''">
            <MenuItem :item="dailyRecommendation" :compact="compactView" show-restaurant />
          </div>
        </div>

        <div v-for="group in groupedMenuItems" :key="group.restaurant">
          <h2 class="text-lg font-semibold text-gray-700 mb-3 max-w-md mx-auto">{{ group.restaurant }}</h2>
          <div :class="compactView ? 'max-w-md mx-auto rounded-lg shadow-md overflow-hidden bg-white p-4' : 'space-y-4'">
            <MenuItem
              v-for="(item, index) in group.items"
              :key="index"
              :item="item"
              :compact="compactView"
            />
          </div>
        </div>
      </div>
    </main>

    <footer class="bg-gray-100 py-4 mt-auto">
      <div class="container space-x-10 mx-auto text-center text-gray-400">
        <a
          href="https://github.com/chlab/lunch-wankdorf"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center hover:text-gray-900 transition"
        >
          <svg class="h-5 w-5 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
          </svg>
          <span>View on GitHub</span>
        </a>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.scrollbar-hide {
  -ms-overflow-style: none; /* IE and Edge */
  scrollbar-width: none; /* Firefox */
}

.scrollbar-hide::-webkit-scrollbar {
  display: none; /* Chrome, Safari and Opera */
}
</style>
