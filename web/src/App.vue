<script setup>
import MenuItem from './components/MenuItem.vue';

const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
// const today = days[new Date().getDay()];
const today = 'Monday';

const menu = {
  "Monday": [
    {
      "name": "Insalata Del Giorno",
      "description": "«Insalata di tonno», mit Mischsalat, Tomaten, Peperoni, Radieschen, Bohnen, Kartoffeln, Thon und Oliven, Italienne Dressing",
      "type": "vegetarian"
    },
    {
      "name": "Pasta Del Giorno",
      "description": "«Hausgemachte Pasta Primavera, con asparagi verdi e aglio orsino», mit grünen Spargeln, Bärlauch, Gemüsestreifen, Frühlingszwiebeln, Reibkäse und Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Pizza Del Giorno",
      "description": "«Pizza al diavolo», Tomatensauce, Mozzarella, scharfe Salami-Chorizo, Peperoni, Zwiebeln, Knoblauch, Oregano und Menusalat",
      "type": "meat"
    },
    {
      "name": "Insalata Settimanale",
      "description": "«Insalata Caesar», Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caeser Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta Settimanale",
      "description": "«Pasta alla fiamma», mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza Settimanale",
      "description": "«Pizza asparagi verdi, con di orsino e cipolline», mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ],
  "Tuesday": [
    {
      "name": "Insalata Del Giorno",
      "description": "«Insalata Mykonos», mit farbigem Peperoni-, Auberginen-Antipasti, Fetawürfeli und Oliven, Italienne Dressing",
      "type": "vegetarian"
    },
    {
      "name": "Pasta Del Giorno",
      "description": "«Hausgemachte Pasta, all'amatriciana», mit Speck-Tomatensauce, Zwergtomaten und Petersilie, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza Del Giorno",
      "description": "«Pizza Lopes», mit Tomatensauce, Mozzarella, Raclettekäse, junger Spinat, rosa Pfefferkörner und Oregano, Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Insalata Settimanale",
      "description": "«Insalata Caesar», Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caeser Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta Settimanale",
      "description": "«Pasta alla fiamma», mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza Settimanale",
      "description": "«Pizza asparagi verdi, con di orsino e cipolline», mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ],
  "Wednesday": [
    {
      "name": "Insalata Del Giorno",
      "description": "«Insalata Mexicana», Mischsalat, Avocados, Tomaten, Peperoni, Mais, rote Bohnen, Nachos, Chunky-Salsa",
      "type": "vegetarian"
    },
    {
      "name": "Pasta Del Giorno",
      "description": "«Hausgemachte Pasta all`arrabiata», rassige Tomatensauce, Peperoncini, Petersilie, Tomatenwürfel, Reibkäse, Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Pizza Del Giorno",
      "description": "«Pizza salmone affumicato», Tomatensauce, Mozzarella, Rauchlachs, Kapern, rote Zwiebeln, Oregano, Menusalat",
      "type": "meat"
    },
    {
      "name": "Insalata Settimanale",
      "description": "«Insalata Caesar», Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caeser Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta Settimanale",
      "description": "«Pasta alla fiamma», mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza Settimanale",
      "description": "«Pizza asparagi verdi, con di orsino e cipolline», mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ],
  "Thursday": [
    {
      "name": "Insalata Del Giorno",
      "description": "«Insalata Greca», Blattsalat, Rucola, Feta, Tomaten, Peperoni, Gurken, schwarze Oliven und Zwiebeln, Balsamico-Oreganodressing",
      "type": "vegetarian"
    },
    {
      "name": "Pasta Del Giorno",
      "description": "«Hausgemachte Pasta con feta, aglio, pomodori secci e olive nere», mit Feta, Dörrtomaten, Petersilie, schwarzen Oliven, Knoblauch, Reibkäse und Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Pizza Del Giorno",
      "description": "«Pizza Dolce Vita», mit Tomatensauce, Mozzarella, Hinterschinken, Ananas, Oregano und Menusalat",
      "type": "meat"
    },
    {
      "name": "Insalata Settimanale",
      "description": "«Insalata Caesar», Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caeser Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta Settimanale",
      "description": "«Pasta alla fiamma», mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza Settimanale",
      "description": "«Pizza asparagi verdi, con di orsino e cipolline», mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ],
  "Friday": [
    {
      "name": "Insalata Del Giorno",
      "description": "«Insalata Ankara», Mischsalat, Bulgur, Gurken, Oliven, Tomaten, Zwiebeln, Feta, Kabisduett, Karotten, Ajvar-Italienne Dressing",
      "type": "vegetarian"
    },
    {
      "name": "Pasta Del Giorno",
      "description": "«Hausgemachte Pasta Alfredo», mit Rahmsauce, Pouletstreifen, Gartenerbsen, Petersilie, Reibkäse und Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza Del Giorno",
      "description": "«Pizza funghi», mit Tomatensauce, Mozzarella, Zweierlei Champignons, Oregano und Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Insalata Settimanale",
      "description": "«Insalata Caesar», Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caeser Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta Settimanale",
      "description": "«Pasta alla fiamma», mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza Settimanale",
      "description": "«Pizza asparagi verdi, con di orsino e cipolline», mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ]
};

// TODO extract from menu items
const restaurants = [
  'Luna',
  'Gira',
  'Sole',
  'Espace'
];
</script>
<template>
  <div class="container mx-auto py-6 max-w-4xl">
    
    <div class="flex max-w-md justify-between mx-auto mb-6" v-if="restaurants.length > 0">
      <span>Filter by:</span>

      <div class="flex space-x-2">
        <button class="flex px-3 py-1 bg-gray-300 rounded-full hover:bg-gray-400 hover:text-white" v-for="restaurant in restaurants">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z" />
          </svg>
          <span class="text-xs ml-2">{{ restaurant }}</span>
        </button>
      </div>
    </div>

    <div v-if="menu[today]" class="space-y-4">
      <MenuItem 
        v-for="(item, index) in menu[today]" 
        :key="index" 
        :name="item.name" 
        :description="item.description" 
        :type="item.type"
      />
    </div>
    <div v-else class="text-center py-8 text-gray-500">
      No menu available for today.
    </div>
  </div>
</template>