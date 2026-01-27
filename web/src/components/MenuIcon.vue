<script setup>
import { computed } from 'vue'

const props = defineProps({
  icon: {
    type: String,
    required: true
  }
});

// map commonly misued icon names to the actual name
const iconMappings = {
  'bowl': 'rice-bowl',
  'bowl-vegi': 'rice-bowl',
  'bowl-fleisch': 'rice-bowl',
  'burger': 'hamburger',
  'pasta': 'spaghetti',
  'meat': 'steak-rare',
  'grilled-food': 'steak'
};

const actualIcon = computed(() => {
  return iconMappings[props.icon] || props.icon;
});

const iconCdnUrl = computed(() => {
  return `https://img.icons8.com/plasticine/200/${actualIcon.value}.png`;
})
</script>

<template>
  <div class="w-12 h-12 p-1 bg-gray-200 rounded-full flex-shrink-0 flex items-center justify-center">
    <img :src="iconCdnUrl" :alt="`${actualIcon} icon`" />
  </div>
</template>