<script setup>
defineProps({
  restaurants: {
    type: Array,
    default: () => [],
  },
});

const selected = defineModel({ type: String, default: '' });

const select = (restaurant) => {
  // Clicking the selected restaurant again clears the filter
  selected.value = restaurant === selected.value ? '' : restaurant;
};
</script>

<template>
  <div v-if="restaurants.length > 0" class="flex space-x-2">
    <button
      v-for="restaurant in restaurants"
      :key="restaurant"
      class="flex-shrink-0 flex px-3 py-1 rounded-full transition-colors cursor-pointer"
      :class="{
        'bg-gray-300 hover:bg-gray-400 hover:text-white': restaurant !== selected,
        'bg-rose-400 text-white': restaurant === selected,
      }"
      :aria-pressed="restaurant === selected"
      @click="select(restaurant)"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke-width="1.5"
        stroke="currentColor"
        class="size-4"
        aria-hidden="true"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z"
        />
      </svg>
      <span class="text-xs ml-2">{{ restaurant }}</span>
    </button>
  </div>
</template>
