<script setup>
defineProps({
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
  }
});
</script>

<template>
  <div class="max-w-sm rounded-lg shadow-lg bg-white mx-auto">
    <div class="px-6 py-4">
      <div class="flex items-center">
        <h3 class="font-medium">{{ name }}</h3>
      </div>
      <p class="text-gray-600 mt-1">{{ description }}</p>
      <div class="mt-4 flex gap-2">
        <div class="flex px-3 py-1 bg-gray-200 rounded-full">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />
          </svg>
          <span class="text-xs ml-2">Luna</span>
        </div>
        
        <span v-if="type === 'vegetarian'" class="px-3 py-1 text-xs bg-green-100 text-green-800 rounded-full">Vegetarian</span>
      </div>
    </div>
  </div>
</template>