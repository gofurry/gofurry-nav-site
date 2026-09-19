import postcss from 'postcss'
import postcssHtml from 'postcss-html'
import postcssLess from 'postcss-less'

export default {
  extends: ['stylelint-config-recommended', 'stylelint-config-recommended-vue'],
  overrides: [
    {
      files: ['**/*.less'],
      customSyntax: postcssLess,
      rules: {
        // CSS value grammar cannot resolve Less mixin variables before compilation.
        'declaration-property-value-no-unknown': [true, { ignoreProperties: { '/.*/': '/@/' } }],
      },
    },
    {
      files: ['**/*.vue'],
      // Explicit strict CSS/Less parsers; do not use HTML's forgiving CSS fallback.
      customSyntax: postcssHtml({ css: postcss, less: postcssLess }),
      rules: {
        // SFC styles support both Vue v-bind() and Less mixin variable references.
        'declaration-property-value-no-unknown': [true, { ignoreProperties: { '/.*/': '/v-bind\\(.+\\)|@/' } }],
      },
    },
    {
      files: ['app/components/common/MobileBottomTabBar.vue'],
      // The existing accessible visually-hidden utility retains its clip fallback.
      rules: { 'property-no-deprecated': [true, { ignoreProperties: ['clip'] }] },
    },
    {
      files: ['app/assets/styles/pages/nav.less'],
      // Existing wrapping compatibility is a later migration, not P1 correctness.
      rules: { 'declaration-property-value-keyword-no-deprecated': [true, { ignoreKeywords: ['break-word'] }] },
    },
    {
      files: [
        'app/assets/styles/pages/games.less',
        'app/assets/styles/pages/nav.less',
      ],
      // These owners deliberately reopen selectors for accumulated state/layout rules.
      rules: { 'no-duplicate-selectors': null },
    },
  ],
  rules: {
    // Cascade ordering across nested domains is migration work, not P1 correctness.
    'no-descending-specificity': null,
    // Tailwind v4 directives are compiled by the existing Vite integration.
    'at-rule-no-unknown': [true, { ignoreAtRules: ['theme', 'source', 'utility', 'variant', 'custom-variant', 'apply', 'reference', 'config', 'plugin'] }],
  },
}
