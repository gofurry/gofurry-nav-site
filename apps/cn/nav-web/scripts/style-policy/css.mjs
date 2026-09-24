import postcss from 'postcss';
import lessSyntax from 'postcss-less';

// These are precise declaration locations, not file-level style exemptions.
const tokenLocations = [
  ['app/assets/styles/tokens.less', [':root', 'html.dark'], '--gf-'],
  ['app/assets/styles/components/nav.less', ['.gf-nav', 'html.dark .gf-nav', '.mobile-bottom-tabs', 'html.dark .mobile-bottom-tabs'], '--gf-nav-'],
  ['app/assets/styles/components/footer.less', ['.gf-footer-shell', 'html.dark .gf-footer-shell'], '--gf-footer-'],
  ['app/assets/styles/components/preferences.less', ['.gf-preferences-modal', 'html.dark .gf-preferences-modal'], '--gf-preferences-'],
  ['app/assets/styles/components/game-review-dialog.less', ['.review-dialog-backdrop', 'html.dark .review-dialog-backdrop'], '--games-review-'],
  ['app/assets/styles/components/error.less', ['.error-page', 'html.dark .error-page'], '--gf-error-'],
  ['app/assets/styles/components/page-scroll-dock.less', ['.page-scroll-dock'], '--gf-scroll-dock-'],
  ['app/assets/styles/primitives/rating.less', ['.gf-rating', 'html.dark .gf-rating'], '--gf-rating-'],
  ['app/assets/styles/pages/games.less', ['.games-page', 'html.dark .games-page'], '--games-'],
  ['app/assets/styles/pages/games.less', ['.game-detail-lightbox', 'html.dark .game-detail-lightbox'], '--games-detail-'],
  ['app/assets/styles/pages/games-search.less', ['.games-search-page', 'html.dark .games-search-page'], '--games-search-'],
  // Filter is body-mounted; share only the existing Search token declarations.
  ['app/assets/styles/pages/games-search.less', ['.games-search-page,.games-search-overlay-scope', 'html.dark .games-search-page,html.dark .games-search-overlay-scope'], '--games-search-'],
  ['app/assets/styles/pages/nav.less', ['.nav-home-page', 'html.dark .nav-home-page'], '--nav-home-'],
  ['app/assets/styles/pages/nav.less', ['.site-group-page', 'html.dark .site-group-page'], '--nav-site-group-'],
  // These states also serve Site Groups outside the Nav Home root.
  ['app/assets/styles/pages/nav.less', ['.nav-site-card', 'html.dark .nav-site-card'], '--nav-home-card-hover-'],
  ['app/assets/styles/pages/nav.less', ['.nav-group-toggle', 'html.dark .nav-group-toggle'], '--nav-home-group-toggle-'],
  // Body Teleports cannot inherit Nav Home's page-root declarations.
  ['app/assets/styles/pages/nav.less', ['.site-popover', 'html.dark .site-popover'], '--nav-home-popover-'],
  ['app/assets/styles/pages/nav.less', ['.group-popover', 'html.dark .group-popover'], '--nav-home-group-popover-'],
  ['app/assets/styles/pages/nav.less', ['.nav-transition-bar__author', 'html.dark .nav-transition-bar__author'], '--nav-home-transition-author-'],
  ['app/assets/styles/pages/updates.less', ['.updates-page', 'html.dark .updates-page'], '--updates-'],
  ['app/assets/styles/pages/lottery.less', ['.lottery-page,.lottery-activation-page,.lottery-modal', 'html.dark .lottery-page,html.dark .lottery-activation-page,html.dark .lottery-modal'], '--lottery-'],
  ['app/assets/styles/pages/static.less', ['.about-page', 'html.dark .about-page'], '--about-'],
  ['app/assets/styles/pages/static.less', ['.legal-page', 'html.dark .legal-page'], '--legal-'],
  ['app/assets/styles/pages/insights/foundation.less', ['.insights-page'], '--insights-'],
];

const legacyClasses = [
  'games-page--dark', 'search-results--dark', 'is-dark-theme',
  'spotlight-panels--dark', 'about-page--dark', 'legal-page--dark',
  'updates-page--dark', 'nav-home-page--dark', 'gf-static-page--dark',
  'lottery-page--dark',
];
const legacySet = new Set(legacyClasses);
const legacySelector = new RegExp(`\\.(?:${legacyClasses.join('|')})(?![\\w-])`, 'g');
const rawColor = /#[\da-fA-F]{8}(?![\da-fA-F])|#[\da-fA-F]{6}(?![\da-fA-F])|#[\da-fA-F]{4}(?![\da-fA-F])|#[\da-fA-F]{3}(?![\da-fA-F])|\brgba?\([^()]*\)/g;
const numericRgb = /^rgba?\(\s*[-+.\d][\d\s.,%/+-]*\)$/;

const countLines = value => (value.match(/\n/g) || []).length;
const removeComments = value => value.replace(/\/\*[\s\S]*?\*\//g, comment => comment.replace(/[^\r\n]/g, ' '));
const normalizeSelector = value => value.trim().replace(/\s+/g, ' ').replace(/\s*,\s*/g, ',');

function containingRule(node) {
  for (let parent = node.parent; parent; parent = parent.parent) {
    if (parent.type === 'rule') return parent;
  }
  return undefined;
}

/** Parse authored declarations/selectors once; malformed CSS/Less propagates. */
export function extractCssFacts(value, { file, line = 1, less = false }) {
  const root = less ? lessSyntax.parse(value, { from: file }) : postcss.parse(value, { from: file });
  const facts = [];
  const add = (kind, node, text, extra = {}) => {
    facts.push({ file, line: line + node.source.start.line - 1, kind, value: text, ...extra });
  };
  root.walk(node => {
    if (node.type === 'rule') add('selector', node, node.selector);
    if (node.type === 'decl' || (node.type === 'atrule' && node.variable)) {
      const rule = containingRule(node);
      const valueLine = line + node.source.start.line - 1
        + countLines(node.variable ? node.raws.afterName || '' : node.raws.between || '');
      add('css-declaration', node, removeComments(node.value), {
        line: valueLine,
        property: node.prop || `@${node.name}`,
        selector: rule?.selector || '',
        nested: Boolean(rule && containingRule(rule)),
      });
      if (node.important) {
        const annotation = node.raws.important || '!important';
        add('important', node, annotation.trim(), {
          line: valueLine + countLines(node.value) + countLines(annotation.slice(0, annotation.indexOf('!'))),
        });
      }
      if (node.variable) {
        const annotation = removeComments(node.value).match(/!\s*important\s*$/);
        if (annotation) add('important', node, annotation[0].trim(), { line: valueLine + countLines(node.value.slice(0, annotation.index)) });
      }
    }
    if (node.type === 'atrule' && node.name === 'apply') {
      const params = removeComments(node.params);
      const paramsLine = line + node.source.start.line - 1 + countLines(node.raws.afterName || '');
      const annotation = params.match(/!\s*important\s*$/);
      add('class', node, annotation ? params.slice(0, annotation.index) : params, { line: paramsLine });
      if (annotation) add('important', node, annotation[0].trim(), { line: paramsLine + countLines(params.slice(0, annotation.index)) });
    }
  });
  return facts;
}

function isApproved(fact) {
  return !fact.nested && tokenLocations.some(([file, selectors, prefix]) => (
    fact.file === file && fact.property.startsWith(prefix)
    && selectors.includes(normalizeSelector(fact.selector))
  ));
}

function finding(rule, fact, value, index, message) {
  return { rule, file: fact.file, line: fact.line + countLines(fact.value.slice(0, index)), value, message,
    offset: fact.offset === undefined ? undefined : fact.offset + index };
}

/** Detect only normalized style facts, never utility-looking arbitrary source text. */
export function detectCssFacts(facts) {
  const findings = [];
  for (const fact of facts) {
    if (fact.kind === 'important') {
      findings.push(finding('important', fact, fact.value, 0, 'CSS !important remains tracked style debt.'));
    }
    if (fact.kind === 'visual-value' || (fact.kind === 'css-declaration' && !isApproved(fact))) {
      for (const match of removeComments(fact.value).matchAll(rawColor)) {
        if (match[0].startsWith('rgb') && !numericRgb.test(match[0])) continue;
        findings.push(finding('raw-visual-value', fact, match[0], match.index, 'Raw color belongs in an approved token declaration.'));
      }
    }
    if (fact.kind === 'selector') {
      // Ignore quoted attribute values while retaining authored selector positions.
      const selector = removeComments(fact.value).replace(/"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'/g, text => text.replace(/[^\r\n]/g, ' '));
      for (const match of selector.matchAll(/:deep\s*\(/g)) {
        findings.push(finding('deep-selector', fact, match[0], match.index, 'Deep selector crosses a component styling boundary.'));
      }
      for (const match of selector.matchAll(legacySelector)) {
        findings.push(finding('legacy-dark-entry', fact, match[0].slice(1), match.index, 'Use the canonical html.dark theme entry.'));
      }
      for (const match of selector.matchAll(/:global\(\s*\.dark\b/g)) {
        findings.push(finding('legacy-dark-entry', fact, match[0], match.index, 'Use :global(html.dark ...) instead of a separate dark entry.'));
      }
    }
    if (fact.kind === 'class') {
      for (const match of fact.value.matchAll(/\S+/g)) {
        if (legacySet.has(match[0])) findings.push(finding('legacy-dark-entry', fact, match[0], match.index, 'Use the canonical html.dark theme entry.'));
      }
    }
  }
  return findings;
}
