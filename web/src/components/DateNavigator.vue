<script setup>
const props = defineProps({
  formattedDate: {
    type: String,
    required: true
  }
});

const emit = defineEmits(['dateBack', 'dateForward']);

// Simple navigation handlers that just emit events
const goToPreviousDay = () => {
  emit('dateBack');
};

const goToNextDay = () => {
  emit('dateForward');
};
</script>
<template>
  <div class="flex mx-auto max-w-md text-center justify-between items-center">
    <button 
      @click="goToPreviousDay" 
      class="hover:bg-gray-300 rounded-full py-4 px-4"
    >
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
        <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
      </svg>
    </button>
    <div>
      <h1 class="text-lg font-bold">Lunch Wankdorf</h1>
      <p class="text-gray-600">{{ props.formattedDate }}</p>
    </div>
    <button 
      @click="goToNextDay" 
      class="hover:bg-gray-300 rounded-full py-4 px-4" 
    >
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
        <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
      </svg>
    </button>
  </div>
</template>