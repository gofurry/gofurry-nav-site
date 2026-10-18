import { readFile } from 'node:fs/promises';
import { __unstable__loadDesignSystem } from 'tailwindcss';

const palette = /^(?:black|white|transparent|current|inherit|slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)(?:-|\/|$)/;
const interaction = /^(?:group-|peer-)?(?:hover|focus|focus-visible|focus-within|active)(?:\/[^:]+)?$/;
let compiler;

function designSystem() {
  // A rejected initialization stays rejected: never silently fall back to regex validity.
  compiler ??= readFile(new URL(import.meta.resolve('tailwindcss/theme.css')), 'utf8')
    .then(theme => __unstable__loadDesignSystem(`${theme}\n@tailwind utilities;`));
  return compiler;
}

function candidateParts(candidate) {
  let depth = 0;
  let start = 0;
  const parts = [];
  for (let index = 0; index < candidate.length; index++) {
    const char = candidate[index];
    if (char === '\\') { index++; continue; }
    if ('[('.includes(char)) depth++;
    if ('])'.includes(char)) depth--;
    if (char === ':' && depth === 0) {
      parts.push(candidate.slice(start, index));
      start = index + 1;
    }
  }
  return { base: candidate.slice(start).replace(/^!|!$/g, ''), variants: parts };
}

function isVisualProperty(property, interactive) {
  if (/^(?:transform|scale|translate|rotate)$/.test(property)) return interactive;
  if (/^(?:color|opacity|box-shadow|filter|backdrop-filter|line-height|letter-spacing|accent-color|caret-color|fill|stroke(?:-.+)?|font(?:-.+)?|transition(?:-.+)?|animation(?:-.+)?|outline(?:-.+)?|text-(?:shadow|transform|decoration(?:-.+)?|underline-offset))$/.test(property)) return true;
  if (/^border(?:-|$)/.test(property)) return !/^(?:border-collapse|border-spacing)$/.test(property);
  if (/^background(?:-|$)/.test(property)) return !/^background-(?:attachment|clip|origin|size|position(?:-[xy])?|repeat)$/.test(property);
  return false;
}

/** The P0 taxonomy; validity comes separately from the actual Tailwind compiler. */
function isAppearance({ base, variants }) {
  const interactive = variants.some(variant => interaction.test(variant));
  const property = base.match(/^\[([a-z-]+):/);
  if (property) return isVisualProperty(property[1], interactive);
  if (/^-?(?:scale|translate|rotate|skew|transform)(?:-|$)/.test(base) && interactive) return true;
  if (/^(?:rounded|shadow|inset-shadow|drop-shadow|text-shadow|ring|inset-ring)(?:-|$)/.test(base)) return true;
  if (/^(?:opacity|leading|tracking|font|blur|brightness|contrast|grayscale|hue-rotate|invert|saturate|sepia|backdrop-(?:blur|brightness|contrast|grayscale|hue-rotate|invert|opacity|saturate|sepia)|transition|duration|delay|ease|animate)(?:-|$)/.test(base)) return true;
  if (/^(?:italic|not-italic|antialiased|subpixel-antialiased|uppercase|lowercase|capitalize|normal-case|underline|overline|line-through|no-underline)$/.test(base)) return true;
  if (/^(?:decoration|underline-offset|accent|caret|fill|stroke)(?:-|$)/.test(base)) return true;
  if (/^text-/.test(base)) return !/^text-(?:left|center|right|justify|start|end|wrap|nowrap|balance|pretty|ellipsis|clip)$/.test(base);
  if (/^bg-/.test(base)) return !/^bg-(?:fixed|local|scroll|clip-.+|origin-.+|auto|cover|contain|center|top|right|bottom|left|repeat(?:-.+)?|no-repeat)$/.test(base);
  if (/^(?:from|via|to)-/.test(base)) return palette.test(base.replace(/^[^-]+-/, '')) || /-\[|-\(/.test(base);
  if (/^(?:border|divide|outline)(?:-|$)/.test(base)) return !/^(?:border-(?:collapse|separate|spacing(?:-.+)?)|divide-(?:x|y)-reverse)$/.test(base);
  return false;
}

export async function detectTailwindFacts(facts) {
  const system = await designSystem();
  const validity = new Map();
  const findings = [];
  for (const fact of facts) {
    if (fact.kind !== 'class') continue;
    for (const match of fact.value.matchAll(/\S+/g)) {
      const value = match[0];
      if (!validity.has(value)) validity.set(value, system.candidatesToCss([value])[0] !== null);
      if (!validity.get(value)) continue;
      const parts = candidateParts(value);
      if (!isAppearance(parts)) continue;
      const finding = {
        file: fact.file,
        offset: fact.offset === undefined ? undefined : fact.offset + match.index,
        line: fact.line + (fact.value.slice(0, match.index).match(/\n/g) || []).length,
        value,
        message: 'Tailwind appearance belongs in Less and the shared token/primitive layer.',
      };
      findings.push({ rule: 'tailwind-appearance', ...finding });
      if (/[[(]/.test(parts.base)) findings.push({ rule: 'tailwind-arbitrary-appearance', ...finding });
    }
  }
  return findings;
}
