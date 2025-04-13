<script setup>
import { ref, onMounted } from 'vue';
import MenuItem from './components/MenuItem.vue';
import Skeleton from './components/Skeleton.vue';
import DateNavigator from './components/DateNavigator.vue';

const menuUrl = 'https://pub-201cbf927f0b4c8991d32485a57b9d40.r2.dev/gira_20250413_171330.json';

const menu = ref({});
const loading = ref(true);
const error = ref(null);

const loadMenu = async () => {
  try {
    loading.value = true;
    const response = await fetch(menuUrl);
    if (!response.ok) {
      throw new Error(`Failed to fetch menu: ${response.status} ${response.statusText}`);
    }
    menu.value = await response.json();
  } catch (err) {
    console.error('Error fetching menu:', err);
    error.value = err.message;
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadMenu();
});

// TODO extract from menu items
const restaurants = [
  'Luna',
  'Gira',
  'Sole',
  'Espace'
];
</script>
<template>
  <div class="min-h-screen flex flex-col">
    <!-- Sticky Header -->
    <header class="top-0 py-4 shadow-md z-10">
      <DateNavigator />
    </header>

    <!-- Main Content -->
    <main class="flex-grow container mx-auto py-6 px-4 max-w-4xl">
      <!-- restaurant filter -->
      <div class="flex max-w-md justify-between mx-auto mb-6" v-if="restaurants.length > 0">
        <span></span>
        <div class="flex space-x-2">
          <button class="flex px-3 py-1 bg-gray-300 rounded-full hover:bg-gray-400 hover:text-white" v-for="restaurant in restaurants">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />
            </svg>
            <span class="text-xs ml-2">{{ restaurant }}</span>
          </button>
        </div>
        <span></span>
      </div>

      <!-- loader -->
      <div v-if="loading" class="space-y-4">
        <Skeleton v-for="i in 2" />
      </div>
      <div v-else-if="error" class="text-center py-8 text-red-500">
        Error loading menu: {{ error }}
      </div>
      <!-- menu items list -->
      <div v-else-if="menu[selectedDay]" class="space-y-4">
        <MenuItem 
          v-for="(item, index) in menu[selectedDay]" 
          :key="index" 
          :name="item.name" 
          :description="item.description" 
          :type="item.type"
          :link="item.link || ''"
        />
      </div>
      <!-- no items today -->
      <div v-else class="text-center py-8 text-gray-500">
        No menu available for today.
      </div>
    </main>

    <!-- Footer -->
    <footer class="bg-gray-100 py-4 mt-auto">
      <div class="container mx-auto text-center text-gray-600">
        <a href="https://github.com/leuenbergerc/lunch-wankdorf" target="_blank" class="inline-flex items-center hover:text-gray-900 transition">
          <svg class="h-5 w-5 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
          </svg>
          <span>View on GitHub</span>
        </a>
      </div>
    </footer>
  </div>
</template>