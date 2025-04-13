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
  <div class="border-b pb-4 last:border-0 last:pb-0">
    <div class="flex items-center">
      <h3 class="font-medium">{{ name }}</h3>
      <span v-if="type === 'vegetarian'" class="ml-2 px-2 py-1 text-xs bg-green-100 text-green-800 rounded-full">Vegetarian</span>
    </div>
    <p class="text-gray-600 mt-1">{{ description }}</p>
  </div>
</template>