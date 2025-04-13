<script setup>
import MenuItem from './components/MenuItem.vue';

const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
const today = days[new Date().getDay()];

const menu = {
  "Sunday": [
    {
      "name": "Insalata di tonno",
      "description": "mit Mischsalat, Tomaten, Peperoni, Radieschen, Bohnen, Kartoffeln, Thon und Oliven, Italienne Dressing",
      "type": "vegetarian"
    },
    {
      "name": "Hausgemachte Pasta Primavera",
      "description": "con asparagi verdi e aglio orsino, mit grünen Spargeln, Bärlauch, Gemüsestreifen, Frühlingszwiebeln, Reibkäse und Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Pizza al diavolo",
      "description": "Tomatensauce, Mozzarella, scharfe Salami-Chorizo, Peperoni, Zwiebeln, Knoblauch, Oregano und Menusalat",
      "type": "meat"
    },
    {
      "name": "Insalata Caesar",
      "description": "Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caesar Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta alla fiamma",
      "description": "mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza asparagi verdi",
      "description": "mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ],
  "Tuesday": [
    {
      "name": "Insalata Mykonos",
      "description": "mit farbigem Peperoni-, Auberginen-Antipasti, Fetawürfeli und Oliven, Italienne Dressing",
      "type": "vegetarian"
    },
    {
      "name": "Hausgemachte Pasta, all'amatriciana",
      "description": "mit Speck-Tomatensauce, Zwergtomaten und Petersilie, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza Lopes",
      "description": "mit Tomatensauce, Mozzarella, Raclettekäse, junger Spinat, rosa Pfefferkörner und Oregano, Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Insalata Caesar",
      "description": "Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caesar Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta alla fiamma",
      "description": "mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza asparagi verdi",
      "description": "mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ],
  "Wednesday": [
    {
      "name": "Insalata Mexicana",
      "description": "Mischsalat, Avocados, Tomaten, Peperoni, Mais, rote Bohnen, Nachos, Chunky-Salsa",
      "type": "vegetarian"
    },
    {
      "name": "Hausgemachte Pasta all`arrabiata",
      "description": "rassige Tomatensauce, Peperoncini, Petersilie, Tomatenwürfel, Reibkäse, Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Pizza salmone affumicato",
      "description": "Tomatensauce, Mozzarella, Rauchlachs, Kapern, rote Zwiebeln, Oregano, Menusalat",
      "type": "meat"
    },
    {
      "name": "Insalata Caesar",
      "description": "Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caesar Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta alla fiamma",
      "description": "mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza asparagi verdi",
      "description": "mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ],
  "Thursday": [
    {
      "name": "Insalata Greca",
      "description": "Blattsalat, Rucola, Feta, Tomaten, Peperoni, Gurken, schwarze Oliven und Zwiebeln, Balsamico-Oreganodressing",
      "type": "vegetarian"
    },
    {
      "name": "Hausgemachte Pasta con feta",
      "description": "aglio, pomodori secci e olive nere, mit Feta, Dörrtomaten, Petersilie, schwarzen Oliven, Knoblauch, Reibkäse und Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Pizza Dolce Vita",
      "description": "mit Tomatensauce, Mozzarella, Hinterschinken, Ananas, Oregano und Menusalat",
      "type": "meat"
    },
    {
      "name": "Insalata Caesar",
      "description": "Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caesar Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta alla fiamma",
      "description": "mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza asparagi verdi",
      "description": "mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ],
  "Friday": [
    {
      "name": "Insalata Ankara",
      "description": "Mischsalat, Bulgur, Gurken, Oliven, Tomaten, Zwiebeln, Feta, Kabisduett, Karotten, Ajvar-Italienne Dressing",
      "type": "vegetarian"
    },
    {
      "name": "Hausgemachte Pasta Alfredo",
      "description": "mit Rahmsauce, Pouletstreifen, Gartenerbsen, Petersilie, Reibkäse und Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza funghi",
      "description": "mit Tomatensauce, Mozzarella, Zweierlei Champignons, Oregano und Menusalat",
      "type": "vegetarian"
    },
    {
      "name": "Insalata Caesar",
      "description": "Mischsalat, junger Lattich, Speckwürfeli, Poulet, Ei, Cherrytomaten, Grana Padano, Croùtons & Caesar Dressing",
      "type": "meat"
    },
    {
      "name": "Pasta alla fiamma",
      "description": "mit Rindfleischstreifen, Tomatenrahmsauce, rote Zwiebeln, Peperoni, Grappa, Menusalat",
      "type": "meat"
    },
    {
      "name": "Pizza asparagi verdi",
      "description": "mit Tomatensauce, Mozzarella, grünen Spargeln, Bärlauchpesto, Frühlingszwiebeln und Oregano, Menusalat",
      "type": "vegetarian"
    }
  ]
};
</script>
<template>
  <div class="container mx-auto px-4 py-8 max-w-4xl">
    <h1 class="text-3xl font-bold text-center mb-6">Lunch Wankdorf</h1>
    <div class="bg-white rounded-lg shadow-md p-6">
      <h2 class="text-xl font-semibold mb-4">{{ today }}'s Menu</h2>
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
  </div>
</template>