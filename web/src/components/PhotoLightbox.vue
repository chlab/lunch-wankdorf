<script setup>
import { onMounted, onUnmounted, ref, useTemplateRef } from 'vue';

const props = defineProps({
  photo: {
    type: String,
    required: true,
  },
  name: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(['close']);

const closeButton = useTemplateRef('closeButton');
const dialog = useTemplateRef('dialog');
const loading = ref(true);

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
    ref="dialog"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
    role="dialog"
    aria-modal="true"
    :aria-label="`Foto von ${props.name}`"
    @click.self="emit('close')"
  >
    <figure class="relative max-h-full max-w-lg overflow-hidden rounded-lg bg-white shadow-xl">
      <button
        ref="closeButton"
        class="absolute top-2 right-2 rounded-full bg-white/90 p-1.5 text-gray-700 hover:bg-white cursor-pointer"
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

      <div v-if="loading" class="flex h-64 w-full items-center justify-center bg-gray-100">
        <span class="text-sm text-gray-400">Foto wird geladen …</span>
      </div>

      <img
        :src="props.photo"
        :alt="props.name"
        class="max-h-[70vh] w-full object-contain"
        :class="loading ? 'hidden' : ''"
        @load="loading = false"
        @error="emit('close')"
      />

      <figcaption class="px-4 py-3 text-sm font-medium text-gray-700">
        {{ props.name }}
      </figcaption>
    </figure>
  </div>
</template>
