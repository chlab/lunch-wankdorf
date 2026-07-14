<script setup>
defineProps({
  formattedDate: {
    type: String,
    required: true,
  },
  canGoBack: {
    type: Boolean,
    default: true,
  },
  canGoForward: {
    type: Boolean,
    default: true,
  },
});

defineEmits(['dateBack', 'dateForward']);
</script>
<template>
  <div class="flex mx-auto max-w-md text-center justify-between items-center">
    <button
      :disabled="!canGoBack"
      aria-label="Vorheriger Tag"
      :class="[
        'rounded-full py-4 px-4',
        canGoBack ? 'hover:bg-gray-300 cursor-pointer' : 'opacity-40 cursor-default',
      ]"
      @click="$emit('dateBack')"
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
          d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18"
        />
      </svg>
    </button>
    <div>
      <h1 class="text-lg font-bold">Lunch Wankdorf</h1>
      <p class="text-gray-600" aria-live="polite">{{ formattedDate }}</p>
    </div>
    <button
      :disabled="!canGoForward"
      aria-label="Nächster Tag"
      :class="[
        'rounded-full py-4 px-4',
        canGoForward ? 'hover:bg-gray-300 cursor-pointer' : 'opacity-40 cursor-default',
      ]"
      @click="$emit('dateForward')"
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
          d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3"
        />
      </svg>
    </button>
  </div>
</template>
