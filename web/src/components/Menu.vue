<script setup>
import MenuItemFilter from './MenuItemFilter.vue';
import RestaurantFilter from './RestaurantFilter.vue';

// Props
defineProps({
  restaurants: {
    type: Array,
    required: true
  },
  selectedRestaurant: {
    type: String,
    required: true
  },
  vegetarianFilter: {
    type: Boolean,
    required: true
  }
});

// Emits
const emit = defineEmits(['select-restaurant', 'toggle-vegetarian']);

// Event handlers
const handleRestaurantSelect = (restaurant) => {
  emit('select-restaurant', restaurant);
};

const handleVegetarianToggle = (isVegetarian) => {
  emit('toggle-vegetarian', isVegetarian);
};
</script>
<style scoped>
.scrollbar-hide {
  -ms-overflow-style: none;  /* IE and Edge */
  scrollbar-width: none;  /* Firefox */
}

.scrollbar-hide::-webkit-scrollbar {
  display: none;  /* Chrome, Safari and Opera */
}
</style>
<template>
  <div class="max-w-full mb-6 overflow-hidden">
    <div class="flex md:justify-center space-x-2 overflow-x-auto scrollbar-hide">
      <!-- Menu item filter -->
      <MenuItemFilter 
        :vegetarian-filter="vegetarianFilter"
        @toggle-vegetarian="handleVegetarianToggle"
      />
      
      <!-- Restaurant filter -->
      <RestaurantFilter 
        :restaurants="restaurants" 
        :selected-restaurant="selectedRestaurant"
        @select-restaurant="handleRestaurantSelect"
      />
    </div>
  </div>
</template>

