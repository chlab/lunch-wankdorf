<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import MenuItem from './components/MenuItem.vue';
import Skeleton from './components/Skeleton.vue';
import DateNavigator from './components/DateNavigator.vue';
import Menu from './components/Menu.vue';
import { getISOWeekNumber } from './util/date';
import foodtrucksMenu from './foodtrucks.json';

const baseUrl = 'https://pub-201cbf927f0b4c8991d32485a57b9d40.r2.dev';
const days = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'];

// Get menu filenames based on current week number and year
const getMenuFiles = () => {
  const now = new Date();
  const weekNumber = getISOWeekNumber();
  const year = now.getFullYear();

  // Return menu filenames with week number and year format
  return {
    gira: `gira_${weekNumber}_${year}.json`,
    luna: `luna_${weekNumber}_${year}.json`,
    sole: `sole_${weekNumber}_${year}.json`,
    espace: `espace_${weekNumber}_${year}.json`,
    freibank: `freibank_${weekNumber}_${year}.json`,
    turbolama: `turbolama_${weekNumber}_${year}.json`,
  };
};

const menuFiles = getMenuFiles();

// Get current day name
const getCurrentDay = () => days[new Date().getDay()];

// Get day from URL (only used for browser back/forward navigation)
const getDayFromURL = () => {
  const urlParams = new URLSearchParams(window.location.search);
  const dayParam = urlParams.get('day');
  if (dayParam && days.includes(dayParam.toLowerCase())) {
    return dayParam.toLowerCase();
  }
  return getCurrentDay();
};

// Function to update URL with current day
const updateURLWithDay = (day) => {
  const url = new URL(window.location);
  url.searchParams.set('day', day);
  window.history.pushState({}, '', url);
};

// Get a date adjusted to a specific day of the week
const getDateForDay = (dayName) => {
  const targetIndex = days.indexOf(dayName);
  const now = new Date();
  const diff = targetIndex - now.getDay();
  const date = new Date(now);
  date.setDate(date.getDate() + diff);
  return date;
};

// Current date and derived values (always start with today, ignore URL param on initial load)
const selectedDay = ref(getCurrentDay());
const currentDate = ref(getDateForDay(selectedDay.value));

const menu = ref({});
const loading = ref(true);
const error = ref(null);
const selectedRestaurant = ref('');
const availableRestaurants = ref(['Foodtrucks']);
const vegetarianFilter = ref(false);

// Format the current date for display
const formattedDate = computed(() => {
  return new Intl.DateTimeFormat('de-CH', {
    weekday: 'long'
  }).format(currentDate.value);
});

// Navigation bounds (Monday = 1, Friday = 5)
const canGoBack = computed(() => currentDate.value.getDay() !== 1);
const canGoForward = computed(() => currentDate.value.getDay() !== 5);

// Navigate by delta days (-1 for previous, +1 for next)
const navigateDay = (delta) => {
  const currentDayOfWeek = currentDate.value.getDay();
  if ((delta < 0 && currentDayOfWeek === 1) || (delta > 0 && currentDayOfWeek === 5)) {
    return;
  }
  const newDate = new Date(currentDate.value);
  newDate.setDate(newDate.getDate() + delta);
  currentDate.value = newDate;
  selectedDay.value = days[currentDate.value.getDay()];
  updateURLWithDay(selectedDay.value);
};

const goToPreviousDay = () => navigateDay(-1);
const goToNextDay = () => navigateDay(1);

// Load all menus and combine them
const loadMenus = async () => {
  try {
    loading.value = true;
    const combinedMenu = {};
    
    // Create an array of promises for all fetch requests
    const fetchPromises = Object.entries(menuFiles).map(async ([restaurant, filename]) => {
      const url = `${baseUrl}/${filename}`;
      
      try {
        const response = await fetch(url);
        if (!response.ok) {
          console.error(`Failed to fetch menu for ${restaurant}: ${response.status} ${response.statusText}`);
          return null;
        }
        
        const menuData = await response.json();
        return { restaurant: restaurant.charAt(0).toUpperCase() + restaurant.slice(1), data: menuData };
      } catch (err) {
        console.error(`Error fetching menu for ${restaurant}:`, err);
        return null;
      }
    });
    
    // Wait for all fetches to complete
    const results = await Promise.all(fetchPromises);
    
    // Keep track of restaurant types for ordering
    const dailyRestaurants = [];
    const tempWeeklyRestaurants = [];
    
    // Process the results and combine menus
    results.forEach(result => {
      if (!result) return;
      
      const { restaurant, data: menuData } = result;
      
      // Track restaurants by type
      if (menuData.type === 'daily') {
        dailyRestaurants.push(restaurant);
      } else if (menuData.type === 'weekly') {
        tempWeeklyRestaurants.push(restaurant);
      }
      
      // Process based on menu type
      if (menuData.type === 'daily' && menuData.menu) {
        // For daily menus (organized by day)
        Object.keys(menuData.menu).forEach(day => {
          // Normalize day name to lowercase format
          const normalizedDay = day.toLowerCase();
          
          if (!combinedMenu[normalizedDay]) {
            combinedMenu[normalizedDay] = [];
          }
          
          // Add restaurant name to each menu item
          const itemsWithRestaurant = menuData.menu[day].map(item => ({
            ...item,
            restaurant: restaurant
          }));
          
          // Add items to combined menu
          combinedMenu[normalizedDay] = [...combinedMenu[normalizedDay], ...itemsWithRestaurant];
        });
      } else if (menuData.type === 'weekly' && menuData.menu) {
        // For weekly menus (same items all week)
        // Add weekly menu items to all weekdays (Monday-Friday)
        ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'].forEach(day => {
          if (!combinedMenu[day]) {
            combinedMenu[day] = [];
          }
          
          // Add restaurant name to each menu item and add to the combined menu
          const itemsWithRestaurant = menuData.menu.map(item => ({
            ...item,
            restaurant: restaurant
          }));
          
          // Add to combined menu (we'll sort later in the filteredMenuItems computed)
          combinedMenu[day] = [...combinedMenu[day], ...itemsWithRestaurant];
        });
      }
    });
    
    // Add static foodtrucks menu to combined menu
    if (foodtrucksMenu.type === 'daily' && foodtrucksMenu.menu) {
      Object.keys(foodtrucksMenu.menu).forEach(day => {
        // Normalize day name to lowercase format
        const normalizedDay = day.toLowerCase();
        
        if (!combinedMenu[normalizedDay]) {
          combinedMenu[normalizedDay] = [];
        }
        
        const foodtruckDailyItems = foodtrucksMenu.menu[day].filter(item => item.enabled);

        // Add foodtrucks items to combined menu
        combinedMenu[normalizedDay] = [...combinedMenu[normalizedDay], ...foodtruckDailyItems];
      });
    }
    
    // Set restaurants in the correct order: daily first, then foodtrucks, then weekly
    availableRestaurants.value = [
      ...dailyRestaurants,
      'Foodtrucks',
      ...tempWeeklyRestaurants
    ];
    
    menu.value = combinedMenu;
  } catch (err) {
    console.error('Error loading menus:', err);
    error.value = err.message;
  } finally {
    loading.value = false;
  }
};

// Handle restaurant selection from RestaurantFilter component
const handleRestaurantSelect = (restaurant) => {
  selectedRestaurant.value = restaurant;
};

// Handle vegetarian filter toggle from MenuItemFilter component
const handleVegetarianToggle = (isVegetarian) => {
  vegetarianFilter.value = isVegetarian;
};

// Find the menu key for the selected day (case-insensitive)
const menuDayKey = computed(() => {
  return Object.keys(menu.value).find(
    key => key.toLowerCase() === selectedDay.value.toLowerCase()
  );
});

// Check if menu exists for the selected day
const hasMenuForSelectedDay = computed(() => !!menuDayKey.value);

// Filtered menu items based on selected restaurant and vegetarian filter
const filteredMenuItems = computed(() => {
  if (!menuDayKey.value || !menu.value[menuDayKey.value]) {
    return [];
  }

  let items = menu.value[menuDayKey.value];

  if (selectedRestaurant.value) {
    items = items.filter(item => item.restaurant === selectedRestaurant.value);
  }

  if (vegetarianFilter.value) {
    items = items.filter(item => item.type === 'vegan' || item.type === 'vegetarian');
  }

  return items;
});

// Group menu items by restaurant for display
const groupedMenuItems = computed(() => {
  const items = filteredMenuItems.value;

  // If a restaurant is already selected, just return one group
  if (selectedRestaurant.value) {
    return items.length > 0 ? [{ restaurant: selectedRestaurant.value, items }] : [];
  }

  // Group items by restaurant
  const groupsMap = {};
  items.forEach(item => {
    if (!groupsMap[item.restaurant]) {
      groupsMap[item.restaurant] = [];
    }
    groupsMap[item.restaurant].push(item);
  });

  // Restaurants that should always be at the end
  const appendedRestaurants = ['Foodtrucks', 'Turbolama'];

  // Separate main restaurants from appended ones
  const mainGroups = [];
  const appendedGroups = [];

  Object.keys(groupsMap).forEach(restaurant => {
    const group = { restaurant, items: groupsMap[restaurant] };
    if (appendedRestaurants.includes(restaurant)) {
      appendedGroups.push(group);
    } else {
      mainGroups.push(group);
    }
  });

  // Randomize main restaurants, keep appended in fixed order
  mainGroups.sort(() => Math.random() - 0.5);
  appendedGroups.sort((a, b) =>
    appendedRestaurants.indexOf(a.restaurant) - appendedRestaurants.indexOf(b.restaurant)
  );

  return [...mainGroups, ...appendedGroups];
});

// Handle popstate events (browser back/forward navigation)
const handlePopState = () => {
  const newDay = getDayFromURL();
  if (newDay !== selectedDay.value) {
    selectedDay.value = newDay;
    currentDate.value = getDateForDay(newDay);
  }
};

onMounted(() => {
  loadMenus();
  
  // Add event listener for browser navigation
  window.addEventListener('popstate', handlePopState);
});

// Clean up event listener when component is unmounted
onUnmounted(() => {
  window.removeEventListener('popstate', handlePopState);
});
</script>
<template>
  <div class="min-h-screen flex flex-col">
    <!-- Sticky Header -->
    <header class="top-0 pt-4">
      <DateNavigator
        :formatted-date="formattedDate"
        :can-go-back="canGoBack"
        :can-go-forward="canGoForward"
        @date-back="goToPreviousDay"
        @date-forward="goToNextDay"
      />
    </header>

    <!-- Main Content -->
    <main class="flex-grow container mx-auto py-6 px-4 max-w-4xl">
      <!-- Filters -->
      <Menu 
        :restaurants="availableRestaurants" 
        :selected-restaurant="selectedRestaurant"
        :vegetarian-filter="vegetarianFilter"
        @select-restaurant="handleRestaurantSelect"
        @toggle-vegetarian="handleVegetarianToggle"
      /> 

      <!-- loader -->
      <div v-if="loading" class="space-y-4">
        <Skeleton v-for="i in 2" :key="i" />
      </div>
      <div v-else-if="error" class="text-center py-8 text-red-500">
        Error loading menu: {{ error }}
      </div>
      <!-- menu items list grouped by restaurant -->
      <div v-else-if="hasMenuForSelectedDay" class="space-y-6">
        <div v-for="group in groupedMenuItems" :key="group.restaurant">
          <h2 class="text-lg font-semibold text-gray-700 mb-3 max-w-lg mx-auto">{{ group.restaurant }}</h2>
          <div class="space-y-4">
            <MenuItem
              v-for="(item, index) in group.items"
              :key="index"
              :name="item.name"
              :description="item.description"
              :type="item.type"
              :link="item.link || ''"
              :icon="item.icon || ''"
              :restaurant="item.restaurant || ''"
              :foodtruck="item.foodtruck || ''"
              :class="{ 'is-restaurant-selected': !!selectedRestaurant }"
            />
          </div>
        </div>
      </div>
      <!-- no items today -->
      <div v-else class="text-center py-8 text-gray-500">
        No menu available for today.
      </div>
    </main>

    <!-- Footer -->
    <footer class="bg-gray-100 py-4 mt-auto">
      <div class="container space-x-10 mx-auto text-center text-gray-400">
        <a href="https://github.com/chlab/lunch-wankdorf" target="_blank" class="inline-flex items-center hover:text-gray-900 transition">
          <svg class="h-5 w-5 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
          </svg>
          <span>View on GitHub</span>
        </a>
      </div>
    </footer>
  </div>
</template>