<script>
import { ref, computed } from 'vue';

const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
const currentDate = ref(new Date());
const selectedDay = computed(() => {
  return days[currentDate.value.getDay()];
});

// Format selected date for display using Swiss German locale
const formattedDate = computed(() => {
  return new Intl.DateTimeFormat('de-CH', { 
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  }).format(currentDate.value);
});

// Navigate to previous day
const goToPreviousDay = () => {
  const newDate = new Date(currentDate.value);
  newDate.setDate(newDate.getDate() - 1);
  currentDate.value = newDate;
};

// Navigate to next day
const goToNextDay = () => {
  const newDate = new Date(currentDate.value);
  newDate.setDate(newDate.getDate() + 1);
  currentDate.value = newDate;
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
      <p class="text-gray-600">{{ formattedDate }}</p>
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