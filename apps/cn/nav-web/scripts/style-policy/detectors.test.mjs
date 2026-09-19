import assert from 'node:assert/strict';
import test from 'node:test';
import { detectCssFacts, extractCssFacts } from './css.mjs';
import { detectTailwindFacts } from './tailwind.mjs';

const file = 'app/components/PolicyFixture.vue';
const fact = (kind, value, extra = {}) => ({ kind, value, file, line: 10, ...extra });
const values = (findings, rule) => findings.filter(finding => finding.rule === rule).map(finding => finding.value);

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

test('approved lottery group normalizes whitespace but does not approve individual roots', () => {
  const facts = extractCssFacts(`.lottery-page,\n .lottery-activation-page , .lottery-modal { --lottery-bg: #fff; }
html.dark .lottery-page, html.dark .lottery-activation-page, html.dark .lottery-modal { --lottery-bg: #000; }
.lottery-page { --lottery-bg: #123; }`, { file: 'app/assets/styles/pages/lottery.less', less: true });
  assert.deepEqual(values(detectCssFacts(facts), 'raw-visual-value'), ['#123']);
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
