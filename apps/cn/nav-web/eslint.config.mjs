import { createConfigForNuxt } from '@nuxt/eslint-config'

// Static Nuxt 4/Vue/TypeScript checks: no runtime module or formatter integration.
export default createConfigForNuxt({
  dirs: { src: ['./app'], servers: ['./server'] },
  features: { stylistic: false, formatters: false, tooling: false },
}).append({
  name: 'gofurry/correctness-only',
  rules: {
    // These presets are conventions, not correctness checks for the P1 boundary.
    'import/first': 'off',
    '@typescript-eslint/consistent-type-imports': 'off',
    '@typescript-eslint/unified-signatures': 'off',
    'vue/attributes-order': 'off',
    'vue/attribute-hyphenation': 'off',
    'vue/block-order': 'off',
    'vue/component-definition-name-casing': 'off',
    'vue/html-self-closing': 'off',
    'vue/first-attribute-linebreak': 'off',
    'vue/multi-word-component-names': 'off',
    'vue/order-in-components': 'off',
    'vue/one-component-per-file': 'off',
    'vue/prop-name-casing': 'off',
    'vue/this-in-template': 'off',
    'vue/v-bind-style': 'off',
    'vue/v-on-style': 'off',
    'vue/v-on-event-hyphenation': 'off',
    'vue/v-slot-style': 'off',
    // Bulk suppressions track errors only; new correctness findings must fail CI.
    'vue/require-default-prop': 'error',
    'vue/no-template-shadow': 'error',
    'vue/no-v-html': 'error',
    'vue/html-end-tags': 'error',
    'vue/require-explicit-emits': 'error',
    'vue/require-prop-types': 'error',
    'vue/no-lone-template': 'error',
    'vue/no-multiple-slot-args': 'error',
    'vue/no-required-prop-with-default': 'error',
  },
}, {
  name: 'gofurry/geographic-data-amd',
  files: ['app/assets/js/china.js'],
  // Existing geographic data uses a UMD wrapper with an optional AMD loader.
  languageOptions: { globals: { define: 'readonly' } },
})
