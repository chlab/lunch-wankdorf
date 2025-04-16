<script setup>
import { ref, onMounted, computed } from 'vue';
import MenuItem from './components/MenuItem.vue';
import Skeleton from './components/Skeleton.vue';
import DateNavigator from './components/DateNavigator.vue';
import RestaurantFilter from './components/RestaurantFilter.vue';
import { getISOWeekNumber } from './util/date'

const baseUrl = 'https://pub-201cbf927f0b4c8991d32485a57b9d40.r2.dev';
const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

// Get menu filenames based on current week number and year
const getMenuFiles = () => {
  // Get current week number and year
  const now = new Date();
  const startOfYear = new Date(now.getFullYear(), 0, 1);
  const weekNumber = getISOWeekNumber();
  const year = now.getFullYear();

  // Return menu filenames with week number and year format
  return {
    gira: `gira_${weekNumber}_${year}.json`,
    luna: `luna_${weekNumber}_${year}.json`,
    sole: `sole_${weekNumber}_${year}.json`
  };
};

const menuFiles = getMenuFiles();

// Current date and derived values
const currentDate = ref(new Date());
const selectedDay = ref(days[currentDate.value.getDay()]);
const menu = ref({});
const loading = ref(true);
const error = ref(null);
const selectedRestaurant = ref('');
const availableRestaurants = ref([]);

// Format the current date for display
const formattedDate = computed(() => {
  return new Intl.DateTimeFormat('de-CH', { 
    weekday: 'long'
  }).format(currentDate.value);
});

// Navigate to previous day
const goToPreviousDay = () => {
  if (currentDate.value.getDay() == 1) {
    return;
  }
  const newDate = new Date(currentDate.value);
  newDate.setDate(newDate.getDate() - 1);
  currentDate.value = newDate;
  selectedDay.value = days[currentDate.value.getDay()];
};

// Navigate to next day
const goToNextDay = () => {
  if (currentDate.value.getDay() == 5) {
    return;
  }
  const newDate = new Date(currentDate.value);
  newDate.setDate(newDate.getDate() + 1);
  currentDate.value = newDate;
  selectedDay.value = days[currentDate.value.getDay()];
};

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
    
    // Process the results and combine menus
    results.forEach(result => {
      if (!result) return;
      
      const { restaurant, data: menuData } = result;
      
      // For each day in the menu
      Object.keys(menuData).forEach(day => {
        if (!combinedMenu[day]) {
          combinedMenu[day] = [];
        }
        
        // Add restaurant name to each menu item
        const itemsWithRestaurant = menuData[day].map(item => ({
          ...item,
          restaurant: restaurant
        }));
        
        // Add items to combined menu
        combinedMenu[day] = [...combinedMenu[day], ...itemsWithRestaurant];
      });
      
      // Extract restaurant name to add to available restaurants list
      if (!availableRestaurants.value.includes(restaurant)) {
        availableRestaurants.value.push(restaurant);
      }
    });
    
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

// Filtered menu items based on selected restaurant
const filteredMenuItems = computed(() => {
  if (!menu.value[selectedDay.value]) {
    return [];
  }
  
  if (selectedRestaurant.value) {
    // Filter items for the selected restaurant
    return menu.value[selectedDay.value].filter(
      item => item.restaurant === selectedRestaurant.value
    );
  }
  
  // Group items by restaurant
  const restaurantGroups = {};
  menu.value[selectedDay.value].forEach(item => {
    if (!restaurantGroups[item.restaurant]) {
      restaurantGroups[item.restaurant] = [];
    }
    restaurantGroups[item.restaurant].push(item);
  });
  
  // Get restaurant names and shuffle their order
  const restaurants = Object.keys(restaurantGroups).sort(() => Math.random() - 0.5);
  
  // Concatenate all items by restaurant in the shuffled order
  return restaurants.flatMap(restaurant => restaurantGroups[restaurant]);
});

onMounted(() => {
  loadMenus();
});
</script>
<template>
  <div class="min-h-screen flex flex-col">
    <!-- Sticky Header -->
    <header class="top-0 pt-4">
      <DateNavigator 
        :formatted-date="formattedDate" 
        @date-back="goToPreviousDay" 
        @date-forward="goToNextDay" 
      />
    </header>

    <!-- Main Content -->
    <main class="flex-grow container mx-auto py-6 px-4 max-w-4xl">
      <!-- Restaurant filter -->
      <RestaurantFilter 
        :restaurants="availableRestaurants" 
        :selected-restaurant="selectedRestaurant"
        @select-restaurant="handleRestaurantSelect"
      />

      <!-- loader -->
      <div v-if="loading" class="space-y-4">
        <Skeleton v-for="i in 2" :key="i" />
      </div>
      <div v-else-if="error" class="text-center py-8 text-red-500">
        Error loading menu: {{ error }}
      </div>
      <!-- menu items list -->
      <div v-else-if="menu[selectedDay]" class="space-y-4">
        <MenuItem 
          v-for="(item, index) in filteredMenuItems" 
          :key="index" 
          :name="item.name" 
          :description="item.description" 
          :type="item.type"
          :link="item.link || ''"
          :restaurant="item.restaurant || ''"
        />
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