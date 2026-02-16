# Browser TODO


csswg-drafts

tc39-ecma262 for JavaScript

whatwg-dom

whatwg-html
https://html.spec.whatwg.org/


# HTML SPECS

[x] - <p>
[x] - <html>
[x] - <head> (WHATWG 4.2.1 compliance)
[x] - <title> (WHATWG 4.2.2 compliance - document.title getter/setter, titleElement.text)
[x] - <base> (WHATWG 4.2.3 compliance - element.href with URL resolution, element.children)
[x] - <style> (WHATWG 4.2.6 compliance - disabled property)
[x] - <body onbeforeunload> - Navigation warning event
[x] - <hr> (WHATWG 4.4.2 compliance - thematic break)
[x] - <pre> (WHATWG 4.4.3 compliance - whitespace preservation, tabs, multi-line in nested elements)
[x] - <blockquote> (WHATWG 4.4.4 compliance - cite property)
[x] - <ol> (WHATWG 4.4.5)
[x] - <em> (WHATWG 4.5.2 compliance - italic styling, inline box)
[x] - <strong> (WHATWG 4.5.3 compliance - bold styling, inline box)
[x] - <s> (WHATWG 4.5.5 compliance - strikethrough styling, inline box)
[x] - <q> (WHATWG 4.5.7 partial - curly quotes added, missing: cite attr, nested quote styles, HTMLQuoteElement)
[x] - <cite> (WHATWG 4.5.6 compliance - italic styling, inline box)
[x] - <dfn> (WHATWG 4.5.8 compliance - italic styling, inline box)
[x] - <abbr> (WHATWG 4.5.9 compliance - dotted underline, title attribute for expansion)
[x] - <data> (WHATWG 4.5.13 compliance - value attribute, HTMLDataElement.value property)
[x] - <time> (WHATWG 4.5.14 compliance - datetime attribute, HTMLTimeElement.dateTime property)
[x] - <mark> (WHATWG 4.5.23 compliance - yellow background, black text)
[x] - <ins> (WHATWG 4.7.1 compliance - underline, HTMLModElement.cite/dateTime, transparent content model)
[] - fix navigation for hash-only URLs (e.g., "#section") - scroll to element with ID

### <a> Missing / non-compliant
- Enforce content model (WHATWG 4.5.1):
  - [x] No nested `<a>` (auto-fixed by x/net/html parser)
  - [ ] No interactive content descendants (button, input, select, textarea, audio/video controls, details, label, embed, iframe)
  - [ ] No tabindex descendants
- [x] Enforce attribute omission rules when href is absent (FindLinkInfo returns nil)
- [x] Implement download behavior
- [x] Implement ping behavior
- [x] Implement hreflang metadata
- [x] Implement type metadata
- [x] Implement referrerpolicy behavior + reflection
- [x] Proper HTMLAnchorElement interface (text, relList, HTMLHyperlinkElementUtils)
- [x] Placeholder <a> behavior when href is absent (non-link, no link styling, no pointer cursor)
- [ ] Keyboard focus/activation (tab/enter/space), focus ring
- [x] Allow preventDefault() on link clicks
- [ ] Named target handling (frames/contexts) beyond _blank
- [x] Visited/unvisited link styling (visitedURLs map, LinkStyler)

### HTMLBodyElement (WHATWG 4.3.1)
- [ ] `document.body` setter - Allow setting body element
- [ ] `onload` event - Window load event on body (High priority)
- [ ] `ononline` / `onoffline` - Network status events
- [ ] `onhashchange` - URL hash navigation
- [ ] `onpopstate` - History API
- [ ] `onmessage` - postMessage API
- [ ] `onstorage` - localStorage events
- [ ] `onpagehide` / `onpageshow` - Page visibility events
- [ ] `HTMLBodyElement` interface - Proper DOM interface
### HTMLStyleElement (WHATWG 4.2.6)
- [x] `styleElement.disabled` - Getter/setter to enable/disable stylesheet
- [ ] `styleElement.sheet` - Get associated CSSStyleSheet object (LinkStyle interface)
- [ ] `styleElement.media` - Get/set media query string
- [ ] `styleElement.type` - Validate type attribute (only "text/css" or empty)
- [ ] `styleElement.title` - Style sheet set name for alternate stylesheets
- [ ] `load` event - Fire when style processing completes
- [ ] `error` event - Fire when style loading fails

### HTMLTableElement (WHATWG 4.9.1)
- [x] `table.caption` getter/setter - Get/set caption element
- [x] `table.createCaption()` / `table.deleteCaption()` - Create/remove caption
- [x] `table.tHead` getter/setter - Get/set thead element (inserts after caption/colgroup per spec)
- [x] `table.createTHead()` / `table.deleteTHead()` - Create/remove thead
- [x] `table.tFoot` getter/setter - Get/set tfoot element (appends at end per spec)
- [x] `table.createTFoot()` / `table.deleteTFoot()` - Create/remove tfoot
- [x] `table.tBodies` - HTMLCollection of tbody elements
- [x] `table.createTBody()` - Create and insert tbody
- [x] `table.rows` - HTMLCollection of all tr elements (ordered: thead, tbody, tfoot)
- [x] `table.insertRow(index)` - Insert row at index (WHATWG 4.9.1)
- [x] `table.deleteRow(index)` - Remove row at index

### HTMLTableRowElement (WHATWG 4.9.8)
- [x] `tr.cells` - HTMLCollection of td/th elements
- [x] `tr.rowIndex` - Row position in table
- [x] `tr.sectionRowIndex` - Row position in section
- [x] `tr.insertCell(index)` / `tr.deleteCell(index)` - Insert/remove cells

### HTMLTableCellElement (WHATWG 4.9.11)
- [x] `td.colSpan` / `td.rowSpan` - Span attributes (getter/setter with clamping per spec)
- [ ] `td.cellIndex` - Cell position in row
- [ ] `td.headers` - Associated header cells
- [ ] `td.scope` - Header cell scope (th only)

### Table Layout Gaps
- [x] `colspan` attribute support in layout
- [x] `rowspan` attribute support in layout
- [ ] Content-based column width calculation
- [x] CSS `width` on table cells (px and % widths on td/th)
- [ ] `thead`/`tfoot` ordering (thead first, tfoot last per spec)
- [ ] Text wrapping inside cells
- [ ] `vertical-align` in cells
- [ ] `<colgroup>` / `<col>` elements

### CSSStyleSheet (CSSOM)
- [ ] `sheet.cssRules` - Get list of CSS rules
- [ ] `sheet.insertRule(rule, index)` - Add a CSS rule
- [ ] `sheet.deleteRule(index)` - Remove a CSS rule
- [ ] `sheet.disabled` - Enable/disable the stylesheet




## Related TODO Files

| File | Purpose |
|------|---------|
| `TODO-HTML.md` | HTML tags implementation |
| `TODO-CSS.md` | CSS properties implementation |
| `TODO-JS.md` | JavaScript implementation |
| `test-todo.md` | Test coverage tracking |

---


# BUGS 

[] - Text overlapp with bold and normal text
[x] - allow select text
[] - Partial text selection (select characters within a line, not entire text boxes)
[] - show indication of CTRL-C copied text 

## In Progress
- [ ] Word wrapping for long text that exceeds container width
- [x] Tooltip on hover for `title` attribute (partial - positioning needs fixes for scroll offset)

---

## Known Issues
- [x] Whitespace between inline elements is missing (e.g., "Here is**bold**and" instead of "Here is **bold** and")
  - Spaces between inline elements like `<strong>`, `<em>`, `<small>` are not rendering
  - Need to debug DOM parser to see if whitespace text nodes are preserved
- [ ] `position: absolute` - text/color inside positioned elements not rendering
  - Background colors and borders of positioned elements work
  - Text inside positioned elements is missing
  - Children of positioned elements are not being painted correctly
- [ ] Margin collapsing not implemented
  - Adjacent vertical margins should collapse (larger wins, not add up)
  - Causes excessive spacing in nested block elements (e.g., `<p>` inside `<blockquote>`)
  - CSS spec: https://www.w3.org/TR/CSS2/box.html#collapsing-margins
- [ ] Text decoration not inherited by nested inline elements
  - `<s>` strikethrough does not show on nested `<a>` links
  - `<a>` underline overwrites parent's text-decoration instead of combining
  - Example: `<s>Visit <a href="#">link</a></s>` - link has no strikethrough

---

## Future Features
- [ ] Forward navigation button
- [ ] Keyboard shortcuts (Ctrl+R refresh, Alt+Left back)
- [ ] Browser history (back/forward)
- [ ] Bookmarks
- [ ] Multiple tabs

---

## Refactoring
- [ ] Refactor Rect usage pattern:
  ```go
  X:     box.Rect.X,
  Y:     box.Rect.Y,
  Width: box.Rect.Width,
  ```

---

## Testing Resources
- [ ] http://acid1.acidtests.org
- [ ] http://acid2.acidtests.org
- https://github.com/web-platform-tests/wpt
  
