<script setup>
import MenuIcon from './MenuIcon.vue'

const props = defineProps({
  name: {
    type: String,
    required: true
  },
  description: {
    type: String,
    required: true
  },
  type: {
    type: String,
    default: "meat"
  },
  link: {
    type: String
  },
  icon: {
    type: String
  },
  restaurant: {
    type: String,
    default: ""
  },
  foodtruck: {
    type: String,
    default: ""
  }
});


</script>

<style scoped>
@media (max-width: 768px) {
  .is-restaurant-selected .restaurant-label {
    display: none;
  }
}
</style>

<template>
  <div class="max-w-lg rounded-lg shadow-lg bg-white mx-auto">
    <div class="px-6 py-4">
      <div class="flex items-center">
        <h3 class="font-medium capitalize">{{ name }}</h3>
      </div>
      <div class="flex items-start gap-3 mt-1">
        <MenuIcon v-if="icon" :icon="icon" />
        <p class="text-gray-600">{{ description }}</p>
      </div>
      <div class="mt-4 flex gap-2">
        <div v-if="restaurant" class="flex px-3 py-1 bg-gray-200 rounded-full">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />
          </svg>
          <span class="text-xs ml-2 restaurant-label">{{ restaurant }}</span>
        </div>
        
        <div v-if="foodtruck" class="flex px-3 py-1 bg-amber-100 text-amber-800 rounded-full">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 18.75a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 0 1-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 0 0-3.213-9.193 2.056 2.056 0 0 0-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 0 0-10.026 0 1.106 1.106 0 0 0-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12" />
          </svg>
          <span class="text-xs ml-2">{{ foodtruck }}</span>
        </div>

        <span v-if="type === 'vegetarian'" class="px-3 py-1 text-xs bg-green-100 text-green-800 rounded-full">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"  class="size-4" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
            <path d="M11 20A7 7 0 0 1 9.8 6.1C15.5 5 17 4.48 19 2c1 2 2 4.18 2 8 0 5.5-4.78 10-10 10Z"></path><path d="M2 21c0-3 1.85-5.36 5.08-6C9.5 14.52 12 13 13 12"></path>
          </svg>
        </span>

        <a v-if="link" :href="link" class="bg-gray-200 rounded-full py-1 px-1 ml-auto hover:bg-gray-300" target="_blank">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
          </svg>
        </a>
      </div>
    </div>
  </div>
</template>