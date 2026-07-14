<script setup>
import { computed, onMounted, onUnmounted, ref, useTemplateRef } from 'vue';
import { dishTitle } from '../util/text';
import { RESTAURANT_URLS } from '../util/menu';

const props = defineProps({
  /** @type {import('../util/menu').MenuItem} */
  item: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(['close']);

const closeButton = useTemplateRef('closeButton');
const loading = ref(true);

const title = computed(() => dishTitle(props.item.name));
const photo = computed(() => props.item.photoLarge || props.item.photo);

// The photos are the restaurants', so credit them. Espace's dishes carry no link
// of their own, so that falls back to the restaurant's menu page.
const source = computed(() => props.item.link || RESTAURANT_URLS[props.item.restaurant] || '');

// Whatever was focused before we opened, so it can be handed focus back on close
const previouslyFocused = document.activeElement;

const onKeydown = (event) => {
  if (event.key === 'Escape') {
    emit('close');
    return;
  }

  // Keep tabbing inside the dialog: there is nothing behind it worth reaching
  if (event.key === 'Tab') {
    event.preventDefault();
    closeButton.value?.focus();
  }
};

onMounted(() => {
  closeButton.value?.focus();
  document.addEventListener('keydown', onKeydown);
  // The page behind must not scroll away under the photo
  document.body.style.overflow = 'hidden';
});

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown);
  document.body.style.overflow = '';
  previouslyFocused?.focus?.();
});
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
    role="dialog"
    aria-modal="true"
    :aria-label="`Foto von ${title}`"
    @click.self="emit('close')"
  >
    <figure class="relative max-h-full max-w-lg overflow-hidden rounded-lg bg-white shadow-xl">
      <button
        ref="closeButton"
        class="absolute top-2 right-2 rounded-full bg-white/90 p-1.5 text-gray-700 hover:bg-white cursor-pointer transition-transform duration-200 hover:scale-110"
        aria-label="Foto schliessen"
        @click="emit('close')"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          class="size-5"
          aria-hidden="true"
        >
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
        </svg>
      </button>

      <!-- Same pulse the menu skeletons use -->
      <div v-if="loading" class="h-64 w-full animate-pulse bg-gray-200"></div>

      <img
        :src="photo"
        :alt="title"
        class="max-h-[70vh] w-full object-contain"
        :class="loading ? 'hidden' : ''"
        @load="loading = false"
        @error="emit('close')"
      />

      <figcaption class="px-4 py-3 text-center">
        <p class="text-sm font-medium text-gray-700">{{ title }}</p>
        <p class="mt-0.5 text-xs text-gray-400">
          Foto von
          <a
            v-if="source"
            :href="source"
            target="_blank"
            rel="noopener noreferrer"
            class="underline hover:text-gray-600"
          >
            {{ item.restaurant }}
          </a>
          <template v-else>{{ item.restaurant }}</template>
        </p>
      </figcaption>
    </figure>
  </div>
</template>
