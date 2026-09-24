import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import { detectCssFacts, extractCssFacts } from './css.mjs';
import { aggregateFindings, compareDebt, validateManifest } from './debt.mjs';
import { detectTailwindFacts } from './tailwind.mjs';

const file = 'app/components/PolicyFixture.vue';
const fact = (kind, value, extra = {}) => ({ kind, value, file, line: 10, ...extra });
const values = (findings, rule) => findings.filter(finding => finding.rule === rule).map(finding => finding.value);

test('all ten retired visual-guard dark entries remain forbidden as classes and selectors', () => {
  const names = ['games-page--dark', 'search-results--dark', 'is-dark-theme',
    'spotlight-panels--dark', 'about-page--dark', 'legal-page--dark', 'updates-page--dark',
    'nav-home-page--dark', 'gf-static-page--dark', 'lottery-page--dark'];
  for (const name of names) {
    assert.equal(values(detectCssFacts([fact('class', name)]), 'legacy-dark-entry').length, 1, name);
    const css = extractCssFacts(`.${name} { display: block; }`, { file });
    assert.equal(values(detectCssFacts(css), 'legacy-dark-entry').length, 1, name);
  }
});

test('Tailwind compiler distinguishes actual appearance from structure and semantic classes', async () => {
  const allowed = 'flex grid gap-4 px-4 text-center truncate max-w-[2080px] z-[120] min-h-[calc(100vh-1rem)] -translate-x-1/2 bg-cover bg-center bg-no-repeat from-10% border-collapse divide-x-reverse blur-wrapper';
  assert.deepEqual(await detectTailwindFacts([fact('class', allowed)]), []);
  const visual = 'text-sm font-semibold rounded-lg bg-white opacity-70 dark:bg-slate-900 hover:scale-105 group-hover/menu:translate-x-1 peer-focus:rotate-3 bg-[#07111f] hover:scale-[1.05] text-(--color)';
  const findings = await detectTailwindFacts([fact('class', visual)]);
  assert.deepEqual(values(findings, 'tailwind-appearance'), visual.split(' '));
  assert.deepEqual(values(findings, 'tailwind-arbitrary-appearance'), ['bg-[#07111f]', 'hover:scale-[1.05]', 'text-(--color)']);
});

test('Tailwind keeps authored repeats, modifier chains, arbitrary colons and source lines', async () => {
  const findings = await detectTailwindFacts([fact('class', 'md:hover:!bg-white\nbg-white! bg-white [data-mode=dark]:bg-[color:#ffffff]')]);
  assert.equal(values(findings, 'tailwind-appearance').length, 4);
  assert.deepEqual(findings.filter(finding => finding.rule === 'tailwind-appearance').map(finding => finding.line), [10, 11, 11, 11]);
  assert.equal(values(findings, 'tailwind-arbitrary-appearance').length, 1);
});

test('arbitrary CSS properties follow appearance ownership, including contextual transforms', async () => {
  const structural = '[width:10px] [position:absolute] [background-size:cover] [text-align:center] [transform:translateX(10px)]';
  assert.deepEqual(await detectTailwindFacts([fact('class', structural)]), []);
  const visual = '[color:#fff] [background:#fff] [border-color:#123] [box-shadow:0_0_2px_#000] [font-size:12px] hover:[transform:scale(1.05)]';
  const findings = await detectTailwindFacts([fact('class', visual)]);
  assert.deepEqual(values(findings, 'tailwind-appearance'), visual.split(' '));
  assert.deepEqual(values(findings, 'tailwind-arbitrary-appearance'), visual.split(' '));
});

test('detectors consume only style facts; selectors and message text are not class use', async () => {
  const facts = [fact('message', 'bg-white #ffffff !important :deep(.x)'), fact('selector', ':deep(.rounded-xl)')];
  assert.deepEqual(await detectTailwindFacts(facts), []);
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), []);
  assert.equal(values(detectCssFacts(facts), 'deep-selector').length, 1);
});

test('CSS declarations count colors individually and ignore comments and dynamic colors', () => {
  const facts = extractCssFacts(`/* #fff !important :deep(.x) */
.card {
  color: #1234;
  box-shadow: 0 0 2px rgba(1, 2, 3, .5), 0 0 4px #11223344;
  background: linear-gradient(rgb(1 2 3 / 50%), #123456);
  border-color: rgba(var(--gf-color), 0.5);
  content: '!important';
  outline-color: var(--gf-border); /* #000 */
}`, { file, line: 10 });
  const findings = detectCssFacts(facts);
  assert.deepEqual(values(findings, 'raw-visual-value'), ['#1234', 'rgba(1, 2, 3, .5)', '#11223344', 'rgb(1 2 3 / 50%)', '#123456']);
  assert.deepEqual(values(findings, 'important'), []);
  assert.deepEqual(values(findings, 'deep-selector'), []);
  assert.equal(findings[0].line, 12);
});

test('Less variables, mixins, nested selectors, apply and spaced important parse as facts', async () => {
  const facts = extractCssFacts(`@ink: #fff;
.mixin(@space: 1rem) { padding: @space; }
.card {
  .mixin();
  color: #123 ! important;
  &:deep(.child):deep (.other) { @apply flex bg-white rounded-lg; }
}`, { file, less: true });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#fff', '#123']);
  assert.equal(values(detectCssFacts(facts), 'important').length, 1);
  assert.equal(values(detectCssFacts(facts), 'deep-selector').length, 2);
  assert.deepEqual(values(await detectTailwindFacts(facts), 'tailwind-appearance'), ['bg-white', 'rounded-lg']);
});

test('multiline values and apply annotations report their actual authored lines', async () => {
  const facts = extractCssFacts(`.card {
  background:
    linear-gradient(#fff,
      #000)
    !important;
  @apply
    bg-white
    !important;
}`, { file, less: true, line: 10 });
  const css = detectCssFacts(facts);
  assert.deepEqual(css.filter(finding => finding.rule === 'raw-visual-value').map(finding => finding.line), [12, 13]);
  assert.deepEqual(css.filter(finding => finding.rule === 'important').map(finding => finding.line), [14, 17]);
  const utilities = await detectTailwindFacts(facts);
  assert.equal(utilities.length, 1);
  assert.equal(utilities[0].line, 16);
});

test('approved tokens require the exact file, selector and custom-property prefix', () => {
  const approvedFile = 'app/assets/styles/pages/games.less';
  const facts = extractCssFacts(`.games-page { --games-bg: #fff; background: #222; --other-bg: #333; }
html.dark .games-page { --games-bg: #000; }
.games-page .child { --games-bg: #444; }
.parent { .games-page { --games-bg: #555; } }`, { file: approvedFile, less: true });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#222', '#333', '#444', '#555']);
  const outsideFile = extractCssFacts('.games-page { --games-bg: #fff; }', { file, less: true });
  assert.deepEqual(values(detectCssFacts(outsideFile), 'raw-visual-value'), ['#fff']);
});

test('global and page tokens share the canonical tokens.less root and dark owner', () => {
  const facts = extractCssFacts(`:root {
  --gf-surface: rgba(1, 2, 3, .5);
  --gf-page-background: #123456;
  --gf-page-pattern-color: #ABCDEF;
}
html.dark {
  --gf-surface: rgba(4, 5, 6, .5);
  --gf-page-background: #654321;
  --gf-page-pattern-color: #FEDCBA;
}`, { file: 'app/assets/styles/tokens.less', less: true });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), []);
});

test('Site Groups tokens require the exact Nav file, theme root and domain prefix', () => {
  const approvedFile = 'app/assets/styles/pages/nav.less';
  const approved = `.site-group-page { --nav-site-group-card-bg: #fff; }
html.dark .site-group-page { --nav-site-group-card-bg: #000; }`;
  assert.deepEqual(detectCssFacts(extractCssFacts(approved, { file: approvedFile, less: true })), []);
  const rejected = `.site-group-page { background: #111; --nav-home-card-bg: #222; }
html.dark .site-group-page { --games-bg: #333; }
.site-group-page .child { --nav-site-group-card-bg: #444; }
html.dark .site-group-page .child { --nav-site-group-card-bg: #555; }
.parent { .site-group-page { --nav-site-group-card-bg: #666; } }
:root { --nav-site-group-card-bg: #777; }`;
  assert.deepEqual(values(detectCssFacts(extractCssFacts(rejected, { file: approvedFile, less: true })), 'raw-visual-value'),
    ['#111', '#222', '#333', '#444', '#555', '#666', '#777']);
  for (const file of ['app/assets/styles/pages/games.less', 'app/assets/styles/pages/site-groups.less']) {
    assert.deepEqual(values(detectCssFacts(extractCssFacts(approved, { file, less: true })), 'raw-visual-value'), ['#fff', '#000']);
  }
});

test('main.css root and dark page-token declarations are no longer approved owners', () => {
  const facts = extractCssFacts(`:root {
  --gf-page-background: #123456;
  --gf-page-pattern-color: #ABCDEF;
}
html.dark {
  --gf-page-background: #654321;
  --gf-page-pattern-color: #FEDCBA;
}`, { file: 'app/assets/css/main.css' });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#123456', '#ABCDEF', '#654321', '#FEDCBA']);
});

test('shared Review tokens require the exact body-Teleport root, theme and compound file', () => {
  const file = 'app/assets/styles/components/game-review-dialog.less';
  const approved = `.review-dialog-backdrop { --games-review-panel-bg: #fff; }
html.dark .review-dialog-backdrop { --games-review-panel-bg: #000; }`;
  assert.deepEqual(detectCssFacts(extractCssFacts(approved, { file, less: true })), []);
  const rejected = `.review-dialog-backdrop { background: #111; --games-home-bg: #222; }
.review-dialog { --games-review-panel-bg: #333; }
.review-dialog-backdrop .child { --games-review-panel-bg: #444; }
html.dark .review-dialog { --games-review-panel-bg: #555; }
.parent { .review-dialog-backdrop { --games-review-panel-bg: #666; } }
:root { --games-review-panel-bg: #777; }`;
  assert.deepEqual(values(detectCssFacts(extractCssFacts(rejected, { file, less: true })), 'raw-visual-value'),
    ['#111', '#222', '#333', '#444', '#555', '#666', '#777']);
  for (const file of ['app/assets/styles/pages/games.less', 'app/assets/styles/components/other.less']) {
    assert.deepEqual(values(detectCssFacts(extractCssFacts(approved, { file, less: true })), 'raw-visual-value'), ['#fff', '#000']);
  }
});

test('tokens.less approval does not cover ordinary properties, wrong prefixes or other selectors', () => {
  const facts = extractCssFacts(`:root { background: #111; --other-surface: #222; }
html.dark { color: #333; --other-surface: #444; }
.card { --gf-surface: #555; }
html.dark .card { --gf-page-background: #666; }
.parent { :root { --gf-surface: #777; } html.dark { --gf-page-background: #888; } }`, {
    file: 'app/assets/styles/tokens.less', less: true,
  });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#111', '#222', '#333', '#444', '#555', '#666', '#777', '#888']);
});

test('Rating tokens are approved at the exact primitive owner in light and dark', () => {
  const facts = extractCssFacts(`.gf-rating {
  --gf-rating-empty: rgba(1, 2, 3, .5);
  --gf-rating-fill: #123456;
}
html.dark .gf-rating { --gf-rating-empty: rgba(4, 5, 6, .5); }`, {
    file: 'app/assets/styles/primitives/rating.less', less: true,
  });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), []);
});

test('the former components Rating path no longer approves token declarations', () => {
  const facts = extractCssFacts(`.gf-rating { --gf-rating-fill: #123456; }
html.dark .gf-rating { --gf-rating-empty: rgba(4, 5, 6, .5); }`, {
    file: 'app/assets/styles/components/rating.less', less: true,
  });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#123456', 'rgba(4, 5, 6, .5)']);
});

test('Rating approval stays limited to exact selectors, prefix and primitive file', () => {
  const facts = extractCssFacts(`.gf-rating { color: #111; --other-fill: #222; }
html.dark .gf-rating { background: #333; --gf-fill: #444; }
.gf-rating .child { --gf-rating-fill: #555; }
html.dark .gf-rating .child { --gf-rating-fill: #666; }
.parent { .gf-rating { --gf-rating-fill: #777; } }`, {
    file: 'app/assets/styles/primitives/rating.less', less: true,
  });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#111', '#222', '#333', '#444', '#555', '#666', '#777']);
  const outsideFile = extractCssFacts('.gf-rating { --gf-rating-fill: #888; }', {
    file: 'app/assets/styles/primitives/new-rating.less', less: true,
  });
  assert.deepEqual(values(detectCssFacts(outsideFile), 'raw-visual-value'), ['#888']);
});

test('moved zero-debt primitive needs no budget and new primitive debt defaults to zero', async () => {
  const manifest = validateManifest(JSON.parse(await readFile(new URL('../../frontend-style-debt.json', import.meta.url), 'utf8')));
  const movedFile = 'app/assets/styles/primitives/rating.less';
  const movedSource = await readFile(new URL(`../../${movedFile}`, import.meta.url), 'utf8');
  const movedFacts = extractCssFacts(movedSource, { file: movedFile, less: true });
  const movedFindings = [...detectCssFacts(movedFacts), ...await detectTailwindFacts(movedFacts)];
  assert.deepEqual(movedFindings, []);

  const newFile = 'app/assets/styles/primitives/new.less';
  const newFindings = detectCssFacts(extractCssFacts('.new-primitive { color: #123456; }', { file: newFile, less: true }));
  const actual = aggregateFindings([...movedFindings, ...newFindings], manifest);
  assert.equal(manifest.baseline['raw-visual-value'][movedFile], undefined);
  assert.equal(manifest.baseline['raw-visual-value'][newFile], undefined);
  // This fixture scans only the two primitive paths; other files' stale entries
  // belong to the separate full-tree parity test, not this local regression.
  const differences = compareDebt(actual, manifest.baseline).filter(item => [movedFile, newFile].includes(item.file));
  assert.deepEqual(differences, [{
    rule: 'raw-visual-value', file: newFile, baseline: 0, actual: 1, kind: 'regression',
  }]);
});

test('approved lottery group normalizes whitespace but does not approve individual roots', () => {
  const facts = extractCssFacts(`.lottery-page,\n .lottery-activation-page , .lottery-modal { --lottery-bg: #fff; }
html.dark .lottery-page, html.dark .lottery-activation-page, html.dark .lottery-modal { --lottery-bg: #000; }
.lottery-page { --lottery-bg: #123; }`, { file: 'app/assets/styles/pages/lottery.less', less: true });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#123']);
});

test('Search Filter shares only the exact page/Teleport token roots and Search prefix', () => {
  const file = 'app/assets/styles/pages/games-search.less';
  const source = `.games-search-page, .games-search-overlay-scope { --games-search-border: #111; }
html.dark .games-search-page, html.dark .games-search-overlay-scope { --games-search-border: #222; }`;
  assert.deepEqual(detectCssFacts(extractCssFacts(source, { file, less: true })), []);
  const rejected = `.games-search-overlay-scope { --games-search-border: #333; }
.games-search-page, .games-search-overlay-scope { color: #444; --other-color: #555; }
.games-search-overlay-scope .child { --games-search-border: #666; }
.parent { .games-search-page, .games-search-overlay-scope { --games-search-border: #777; } }`;
  assert.deepEqual(values(detectCssFacts(extractCssFacts(rejected, { file, less: true })), 'raw-visual-value'),
    ['#333', '#444', '#555', '#666', '#777']);
  assert.deepEqual(values(detectCssFacts(extractCssFacts(source, { file: 'app/assets/styles/pages/games.less', less: true })), 'raw-visual-value'), ['#111', '#222']);
});

test('Preferences tokens require the exact compound owner and both theme roots', () => {
  const source = `.gf-preferences-modal { --gf-preferences-input-surface: #123456; }
html.dark .gf-preferences-modal { --gf-preferences-toggle-thumb: rgba(1, 2, 3, .5); }`;
  const approvedFile = 'app/assets/styles/components/preferences.less';
  assert.deepEqual(detectCssFacts(extractCssFacts(source, { file: approvedFile, less: true })), []);
  for (const file of ['app/assets/styles/primitives/modal.less', 'app/assets/styles/components/modal.less', 'app/components/Preferences.vue']) {
    assert.equal(values(detectCssFacts(extractCssFacts(source, { file, less: true })), 'raw-visual-value').length, 2);
  }
});

test('Preferences approval never covers ordinary properties, other prefixes or nested/other selectors', () => {
  const facts = extractCssFacts(`.gf-preferences-modal { background: #111; --other-surface: #222; }
html.dark .gf-preferences-modal { color: #333; --gf-input-surface: #444; }
.preferences-toggle { --gf-preferences-toggle-surface: #555; }
.gf-preferences-modal .child { --gf-preferences-input-surface: #666; }
.parent { .gf-preferences-modal { --gf-preferences-input-surface: #777; } }
:root { --gf-preferences-input-surface: #888; }`, {
    file: 'app/assets/styles/components/preferences.less', less: true,
  });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#111', '#222', '#333', '#444', '#555', '#666', '#777', '#888']);
});

test('Modal primitive and Preferences retire the old raw debt without transferring a budget', async () => {
  const manifest = validateManifest(JSON.parse(await readFile(new URL('../../frontend-style-debt.json', import.meta.url), 'utf8')));
  const oldFile = 'app/assets/styles/components/modal.less';
  await assert.rejects(readFile(new URL(`../../${oldFile}`, import.meta.url)), { code: 'ENOENT' });
  for (const file of ['app/assets/styles/primitives/modal.less', 'app/assets/styles/components/preferences.less']) {
    const source = await readFile(new URL(`../../${file}`, import.meta.url), 'utf8');
    assert.deepEqual(detectCssFacts(extractCssFacts(source, { file, less: true })), []);
    assert.equal(manifest.baseline['raw-visual-value'][file], undefined);
    assert(!source.includes('gf-modal__toggle'));
  }
  assert.equal(manifest.baseline['raw-visual-value'][oldFile], undefined);
});

test('inline/script visual values count raw literals without counting Tailwind a third time', async () => {
  const facts = [fact('visual-value', 'color: #abc; fill: #112233; border-color: rgb(2 3 4 / .2)'), fact('class', 'bg-[#fff]')];
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#abc', '#112233', 'rgb(2 3 4 / .2)']);
  assert.equal((await detectTailwindFacts(facts)).length, 2);
});

test('legacy entries match actual classes and selector entries, not canonical dark or text', () => {
  const facts = [
    fact('class', 'games-page--dark games-page--dark-extra html dark'),
    fact('selector', '.games-page--dark, :global(.dark .child), html.dark .x, :global(html.dark .y), [title=".games-page--dark :deep(.x)"]'),
    fact('message', 'games-page--dark'),
  ];
  assert.deepEqual(values(detectCssFacts(facts), 'legacy-dark-entry'), ['games-page--dark', 'games-page--dark', ':global(.dark']);
  assert.deepEqual(values(detectCssFacts(facts), 'deep-selector'), []);
});

test('CSS and Less parser failures fail closed', () => {
  for (const less of [false, true]) {
    assert.throws(() => extractCssFacts('.card { color: #fff', { file, less }), /Unclosed block/);
    assert.throws(() => extractCssFacts('.card { color #fff; }', { file, less }), /Unknown word/);
  }
});

test('Detail Lightbox owns only its exact body-mounted token roots', () => {
  const file = 'app/assets/styles/pages/games.less';
  const source = '.game-detail-lightbox { --games-detail-overlay: rgba(0,0,0,.86); } html.dark .game-detail-lightbox { --games-detail-overlay: rgba(0,0,0,.90); }';
  assert.deepEqual(detectCssFacts(extractCssFacts(source, { file, less: true })), []);
  assert.equal(detectCssFacts(extractCssFacts(source, { file: 'app/assets/styles/pages/insights.less', less: true })).length, 2);
  const invalid = '.game-detail-lightbox { background: #111; --other-overlay: #222; } .game-detail-lightbox .child { --games-detail-overlay: #333; } .parent { .game-detail-lightbox { --games-detail-overlay: #444; } }';
  assert.deepEqual(values(detectCssFacts(extractCssFacts(invalid, { file, less: true })), 'raw-visual-value'), ['#111', '#222', '#333', '#444']);
});

test('Detail page palette approval is limited to its exact file, roots and namespace', () => {
  const file = 'app/assets/styles/pages/games.less';
  const source = '.game-detail-page { --games-detail-media-bg: #123456; } html.dark .game-detail-page { --games-detail-chart-axis: rgba(1, 2, 3, .5); }';
  assert.deepEqual(detectCssFacts(extractCssFacts(source, { file, less: true })), []);
  assert.equal(detectCssFacts(extractCssFacts(source, { file: 'app/assets/styles/pages/insights.less', less: true })).length, 2);
  const invalid = '.game-detail-page { color: #111; --games-home-bg: #222; } .game-detail-page .child { --games-detail-media-bg: #333; } .parent { .game-detail-page { --games-detail-media-bg: #444; } } :root { --games-detail-media-bg: #555; }';
  assert.deepEqual(values(detectCssFacts(extractCssFacts(invalid, { file, less: true })), 'raw-visual-value'), ['#111', '#222', '#333', '#444', '#555']);
});
