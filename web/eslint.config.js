import js from '@eslint/js';
import pluginVue from 'eslint-plugin-vue';
import prettier from 'eslint-config-prettier';
import globals from 'globals';

export default [
  { ignores: ['dist/**'] },
  js.configs.recommended,
  ...pluginVue.configs['flat/recommended'],
  {
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: globals.browser,
    },
    rules: {
      // Single-word component names (App, Menu, Skeleton) are fine in a project this size
      'vue/multi-word-component-names': 'off',
    },
  },
  // Formatting is prettier's job, so switch off the rules that would fight it
  prettier,
];
