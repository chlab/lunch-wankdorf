<script setup>
import { ref, onMounted } from 'vue';
import MenuItem from './components/MenuItem.vue';
import Skeleton from './components/Skeleton.vue';

const menuUrl = 'https://pub-201cbf927f0b4c8991d32485a57b9d40.r2.dev/gira_20250413_171330.json';
const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
// const today = days[new Date().getDay()];
const today = 'Monday'; // For testing a specific day

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
  <div class="container mx-auto py-6 max-w-4xl min-h-screen">
    
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
    <div v-else-if="menu[today]" class="space-y-4">
      <MenuItem 
        v-for="(item, index) in menu[today]" 
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
  </div>
</template>