<script setup>
import { computed, ref } from 'vue';
import MenuIcon from './MenuIcon.vue';
import { dishTitle } from '../util/text';

const props = defineProps({
  /** @type {import('../util/menu').MenuItem} */
  item: {
    type: Object,
    required: true,
  },
  compact: {
    type: Boolean,
    default: false,
  },
  // The recommendation card shows the restaurant, the grouped lists have a heading
  showRestaurant: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['showPhoto']);

// The photos are hotlinked from the restaurants, so a dead one falls back to the
// icon rather than leaving a broken image behind
const photoFailed = ref(false);

const photo = computed(() => (photoFailed.value ? '' : props.item.photo));
const title = computed(() => dishTitle(props.item.name));
</script>

<template>
  <div
    :class="[
      'bg-white relative',
      compact
        ? 'px-2 py-2'
        : 'max-w-md rounded-lg shadow-md mx-auto transition-transform duration-200 hover:scale-[1.02]',
    ]"
  >
    <div :class="compact ? '' : 'px-6 py-4 pb-6'">
      <!-- Title row with vegi icon -->
      <div class="flex items-center gap-2">
        <h3 :class="[compact ? 'text-sm' : 'font-medium']">{{ title }}</h3>
        <span
          v-if="item.type === 'vegetarian'"
          class="text-green-600"
          role="img"
          aria-label="Vegetarisch"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            :class="compact ? 'size-3' : 'size-4'"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path
              d="M11 20A7 7 0 0 1 9.8 6.1C15.5 5 17 4.48 19 2c1 2 2 4.18 2 8 0 5.5-4.78 10-10 10Z"
            ></path>
            <path d="M2 21c0-3 1.85-5.36 5.08-6C9.5 14.52 12 13 13 12"></path>
          </svg>
        </span>
      </div>
      <!-- Description row, with the dish photo where the icon would otherwise go -->
      <div :class="['flex items-start gap-3', compact ? 'mt-1' : 'mt-3']">
        <button
          v-if="photo && !compact"
          class="size-16 flex-shrink-0 overflow-hidden rounded-full border-2 border-gray-300 shadow-sm cursor-pointer transition-transform duration-200 hover:scale-105"
          :aria-label="`Foto von ${title} vergrössern`"
          @click="emit('showPhoto', item)"
        >
          <img
            :src="photo"
            alt=""
            loading="lazy"
            class="size-full object-cover"
            @error="photoFailed = true"
          />
        </button>
        <MenuIcon v-else-if="item.icon && !compact" :icon="item.icon" />
        <p :class="[compact ? 'text-xs text-gray-500' : 'text-gray-600']">{{ item.description }}</p>
      </div>
      <!-- Badges row (the foodtruck badge needs the room a card gives it) -->
      <div
        v-if="(item.restaurant && showRestaurant) || (item.foodtruck && !compact)"
        class="mt-3 flex gap-2"
      >
        <div
          v-if="item.restaurant && showRestaurant"
          class="flex items-center px-3 py-1 bg-gray-200 rounded-full"
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
          <span class="text-xs ml-2">{{ item.restaurant }}</span>
        </div>
        <div
          v-if="item.foodtruck && !compact"
          class="flex items-center px-3 py-1 bg-amber-100 text-amber-800 rounded-full"
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
              d="M8.25 18.75a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 0 1-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 0 0-3.213-9.193 2.056 2.056 0 0 0-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 0 0-10.026 0 1.106 1.106 0 0 0-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12"
            />
          </svg>
          <span class="text-xs ml-2">{{ item.foodtruck }}</span>
        </div>
      </div>
    </div>
    <!-- Floating link button (hidden in compact mode) -->
    <a
      v-if="item.link && !compact"
      :href="item.link"
      target="_blank"
      rel="noopener noreferrer"
      :aria-label="`${title} auf der Website von ${item.restaurant} ansehen`"
      class="absolute bottom-3 right-3 bg-gray-200 rounded-full p-1.5 hover:bg-gray-300"
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
    </a>
  </div>
</template>
