<script setup>
const compactView = defineModel('compactView', { type: Boolean, default: false });
</script>

<template>
  <div class="flex gap-1">
    <button
      @click="compactView = false"
      :class="['p-1.5 rounded', !compactView ? 'bg-gray-100' : 'hover:bg-gray-50']"
      :aria-pressed="!compactView"
      aria-label="Kartenansicht"
      title="Kartenansicht"
    >
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 3A2.25 2.25 0 0 0 3 5.25v13.5A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V5.25A2.25 2.25 0 0 0 18.75 3H5.25Z" />
      </svg>
    </button>
    <button
      @click="compactView = true"
      :class="['p-1.5 rounded', compactView ? 'bg-gray-100' : 'hover:bg-gray-50']"
      :aria-pressed="compactView"
      aria-label="Listenansicht"
      title="Listenansicht"
    >
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0ZM3.75 12h.007v.008H3.75V12Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm-.375 5.25h.007v.008H3.75v-.008Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
      </svg>
    </button>
  </div>
</template>
