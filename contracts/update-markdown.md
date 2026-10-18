# Update Markdown contract — version 1

This is the semantic/security contract for Issue #132 Release Notes Markdown.
P2 Admin preview and P3 Nav Web article rendering install matching parser and
sanitizer versions and run the same
[fixture](fixtures/update-markdown.json). The applications remain independent
pnpm projects. This contract does not prescribe preview/public-page pixels.

## Parser and source

Both applications use `markdown-it` **15.0.2** and `sanitize-html` **2.17.7**. The single rendering
entry point in each application is `renderUpdateMarkdown()`: Admin owns
`apps/cn/admin/react/src/features/release-notes/markdown.ts`, and Nav Web owns
`apps/cn/web/app/utils/updateMarkdown.ts`. Each always performs
Markdown parsing followed by final sanitization. No component calls the parser
directly or inserts untrusted HTML into `dangerouslySetInnerHTML` / `v-html`.

Parser options: `html: false`, `breaks: true`, `linkify: true`,
`typographer: false`; the table rule is disabled. Preserve stored Markdown exactly;
render-time normalization must never rewrite the editor's source.

Supported semantics are paragraphs, h2–h6, strong/em/del, unordered/ordered lists,
list items, blockquotes, links, images, inline/fenced code, horizontal rules and
line breaks. Normalize strikethrough to `del`. Per the P2 heading contract, all
Markdown heading levels move down one: `# → h2`, `## → h3`, through `##### → h6`;
`######` also remains h6. The public article owns its own h1.

Raw HTML is literal text. There is no MDX, table rendering, Mermaid diagram
rendering or syntax highlighting. A `mermaid` fence is ordinary escaped code.

## URLs and generated attributes

- Links accept absolute `https://` / `http://`, `mailto:`, site-root `/path`, and
  `#anchor`. HTTP(S) links receive `target="_blank"` and
  `rel="noopener noreferrer"`. Site-root, anchor and mailto links remain same-tab.
- Images accept only absolute `https://` or site-root `/path`. They receive
  `loading="lazy"`, `decoding="async"`, `referrerpolicy="no-referrer"`, preserving
  alt text and optional title. No image upload or asset picker is involved.
- Reject protocol-relative URLs (`//`), bare/`./`/`../` relatives,
  javascript/vbscript/data/file/blob schemes, embedded whitespace/control
  characters and backslashes. Image HTTP is also rejected. The sanitizer applies
  this policy again to decoded HTML attributes, including encoded-scheme attacks.
- Rejected links have no navigable href (parser-rejected Markdown may remain
  literal source text). Rejected images are omitted. Do not repair unsafe URLs.

## Final sanitizer

Only these elements may survive:

```text
p h2 h3 h4 h5 h6 strong em del ul ol li blockquote a img code pre hr br
```

Only these attributes may survive:

| Element | Attributes |
| --- | --- |
| a | href, target, rel — rebuilt by the renderer |
| img | src, alt, title, loading, decoding, referrerpolicy |
| code | class — tokens matching `^language-[a-zA-Z0-9_-]+$` only |

All other attributes, including event handlers, styles, IDs and data attributes,
are forbidden. No iframe, video/audio, SVG, script/style or form controls survive.
Author-provided target/rel/loading/referrerpolicy cannot override generated policy.

## Shared fixture

`update-markdown.json` contains semantic DOM checks, not serialized snapshots:
`elements` specify selectors, counts, text and attributes; `absentSelectors`
specify forbidden output; `textContains` checks escaped/literal content.
`kind=markdown` runs the public rendering entry point; `kind=html` probes its final
sanitizer directly with adversarial input. The latter is a test boundary, never a
second preview path. Both Admin and P3 Nav Web must satisfy every case and the
global attribute allowlist. Future extensions require an explicit contract change.
