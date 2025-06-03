<script setup>
import { computed } from 'vue'

const props = defineProps({
  icon: {
    type: String,
    required: true
  },
  size: {
    type: String,
    default: '12'
  }
});

// map commonly misued icon names to the actual name
const iconMappings = {
  'bowl': 'rice-bowl',
  'burger': 'hamburger',
  'pasta': 'spaghetti',
};

const actualIcon = computed(() => {
  return iconMappings[props.icon] || props.icon;
});

const iconCdnUrl = computed(() => {
  return `https://img.icons8.com/plasticine/200/${actualIcon.value}.png`;
})

const sizeClasses = computed(() => {
  return `w-${props.size} h-${props.size}`;
})
</script>

<template>
  <div :class="sizeClasses" class="bg-gray-200 rounded-full flex-shrink-0 flex items-center justify-center">
    <img :src="iconCdnUrl" :alt="`${actualIcon} icon`" />
  </div>
</template>