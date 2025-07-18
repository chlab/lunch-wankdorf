<script setup>
const props = defineProps({
  vegetarianFilter: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['toggleVegetarian']);

const toggleVegetarian = () => {
  emit('toggleVegetarian', !props.vegetarianFilter);
};
</script>
<template>
  <div class="flex space-x-2 border-r border-dotted border-gray-300 pr-2">
    <button 
      class="flex-shrink-0 flex px-3 py-1 rounded-full transition-colors cursor-pointer"
      :class="{
        'bg-gray-300 hover:bg-gray-400 hover:text-white': !vegetarianFilter,
        'bg-green-500 text-white': vegetarianFilter
      }"
      @click="toggleVegetarian"
    >
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor"  class="size-4" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
        <path d="M11 20A7 7 0 0 1 9.8 6.1C15.5 5 17 4.48 19 2c1 2 2 4.18 2 8 0 5.5-4.78 10-10 10Z"></path><path d="M2 21c0-3 1.85-5.36 5.08-6C9.5 14.52 12 13 13 12"></path>
      </svg>
      <span class="text-xs ml-2">Vegi</span>
    </button>
  </div>
</template>