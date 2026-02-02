# HTML Implementation TODO

## Completed Tags

### Document Structure
- [x] `<html>` - root element
- [x] `<head>` - document head
- [x] `<body>` - document body
- [x] `<title>` - page title
- [x] `<link>` - external resources (stylesheets)
- [x] `<style>` - embedded CSS (WHATWG 4.2.6: disabled property support)
- [x] `<script>` - JavaScript

### Text Content
- [x] `<h1>` - `<h6>` - headings
- [x] `<p>` - paragraph
- [x] `<br>` - line break
- [x] `<hr>` - horizontal rule (WHATWG 4.4.2 compliant)
- [x] `<pre>` - preformatted text (WHATWG 4.4.3: whitespace/tabs preserved, multi-line in nested elements)
- [x] `<blockquote>` - block quotation
- [x] `<div>` - division
- [x] `<span>` - inline container

### Text Formatting
- [x] `<strong>` / `<b>` - bold
- [x] `<em>` / `<i>` - italic
- [x] `<u>` - underline
- [x] `<small>` - smaller text
- [x] `<del>` - deleted text (strikethrough)
- [x] `<ins>` - inserted text (underline)
- [x] `<q>` - inline quotation

### Links & Media
- [x] `<a>` - hyperlink (href, target, rel attributes supported)
  - [x] `download` attribute - downloads to ~/Downloads with random filename
  - [ ] `ping` attribute - URLs to ping on click
  - [ ] `hreflang` attribute - language of linked resource
  - [ ] `type` attribute - MIME type hint
  - [ ] `referrerpolicy` attribute - referrer policy
- [x] `<img>` - image

### Lists
- [x] `<ul>` - unordered list (WHATWG 4.4.6 compliant)
- [x] `<ol>` - ordered list (WHATWG 4.4.5: start/reversed/type attributes)
- [x] `<menu>` - toolbar menu (WHATWG 4.4.7 compliant, semantic ul)
- [x] `<li>` - list item (WHATWG 4.4.8: value attribute, works in ol/ul/menu)
- [x] `<dl>` - description list (WHATWG 4.4.9: compliance tests added, div wrapper pattern supported)
- [x] `<dt>` - description term (WHATWG 4.4.10: compliance tests added)
- [x] `<dd>` - description details (40px indent)

### Tables
- [x] `<table>` - table
- [x] `<thead>` - table header group
- [x] `<tbody>` - table body group
- [x] `<tfoot>` - table footer group
- [x] `<tr>` - table row
- [x] `<th>` - table header cell
- [x] `<td>` - table data cell

### Forms
- [x] `<form>` - form container
- [x] `<input>` - text/password/email/number/checkbox/radio/file
- [x] `<button>` - button
- [x] `<textarea>` - multiline text
- [x] `<select>` - dropdown
- [x] `<option>` - select option
- [x] `<label>` - form label
- [x] `<fieldset>` - form group (basic implementation, see Known Issues)
- [x] `<legend>` - fieldset caption (basic implementation, see Known Issues)

### Semantic
- [x] `<header>` - header section
- [x] `<footer>` - footer section
- [x] `<main>` - main content
- [x] `<nav>` - navigation
- [x] `<section>` - section
- [x] `<article>` - article
- [x] `<center>` - centered content (deprecated but still used)
- [x] `<search>` - search section (WHATWG 4.4.15 compliant)

---

## Missing Tags

### Text Formatting
- [ ] `<code>` - inline code
- [ ] `<kbd>` - keyboard input
- [ ] `<samp>` - sample output
- [ ] `<var>` - variable
- [ ] `<abbr>` - abbreviation
- [ ] `<cite>` - citation
- [ ] `<mark>` - highlighted text
- [ ] `<sub>` - subscript
- [ ] `<sup>` - superscript
- [ ] `<time>` - date/time
- [ ] `<dfn>` - definition term

### Media
- [ ] `<video>` - video player
- [ ] `<audio>` - audio player
- [ ] `<source>` - media source
- [ ] `<picture>` - responsive images
- [x] `<figure>` - figure container (WHATWG 4.4.12: block element, 40px indent, compliance tests)
- [x] `<figcaption>` - figure caption (block element)
- [ ] `<canvas>` - drawing canvas
- [ ] `<svg>` - vector graphics
- [ ] `<iframe>` - embedded frame

### Tables
- [x] `<caption>` - table caption (centered text)
- [x] Nested tables inside cells
- [x] Inline elements (links, spans, formatting) inside cells
- [ ] `<colgroup>` - column group
- [ ] `<col>` - column properties
- [ ] `colspan` / `rowspan` attributes
- [ ] Content-based column width calculation
- [ ] CSS `width` style on table cells (currently all cells get equal width)

### Forms
- [ ] `<datalist>` - input suggestions
- [ ] `<output>` - calculation result
- [ ] `<progress>` - progress bar
- [ ] `<meter>` - gauge/meter
- [ ] `<optgroup>` - option group

### Interactive
- [ ] `<details>` - collapsible content
- [ ] `<summary>` - details summary
- [ ] `<dialog>` - modal dialog

### Semantic
- [x] `<aside>` - sidebar content (block element)
- [ ] `<address>` - contact info
- [ ] `<hgroup>` - heading group

### Document Metadata
- [x] `<base>` - base URL for relative links
- [ ] `<meta>` - document metadata (viewport, charset, etc.)
- [ ] `<noscript>` - fallback for no JavaScript

### Embedded Content
- [ ] `<embed>` - external content plugin
- [ ] `<object>` - embedded object
- [ ] `<param>` - object parameter
- [ ] `<map>` - image map container
- [ ] `<area>` - image map clickable area

### Ruby Annotations (East Asian text)
- [ ] `<ruby>` - ruby annotation container
- [ ] `<rt>` - ruby text (pronunciation)
- [ ] `<rp>` - ruby fallback parenthesis

### Text Direction & Breaks
- [ ] `<wbr>` - word break opportunity
- [ ] `<bdi>` - bidirectional isolation
- [ ] `<bdo>` - bidirectional override

### Web Components
- [ ] `<template>` - content template (not rendered)
- [ ] `<slot>` - web component slot

---

## Missing Input Types
- [ ] `type="date"` - date picker
- [ ] `type="time"` - time picker
- [ ] `type="datetime-local"` - date and time picker
- [ ] `type="month"` - month picker
- [ ] `type="week"` - week picker
- [ ] `type="color"` - color picker
- [ ] `type="range"` - slider control
- [ ] `type="search"` - search field
- [ ] `type="tel"` - telephone input
- [ ] `type="url"` - URL input
- [ ] `type="hidden"` - hidden field

---

## Missing Attributes & Features

### Form Validation
- [x] `required` - required field validation (red border + prevents submit)
- [ ] `pattern` - regex validation
- [ ] `min` / `max` - number range validation
- [ ] `minlength` / `maxlength` - text length validation
- [ ] `step` - number increment
- [ ] `:valid` / `:invalid` pseudo-classes

### Table Features
- [ ] `colspan` - cell column span
- [ ] `rowspan` - cell row span
- [ ] `scope` - header cell scope

### Link Features
- [x] `target="_blank"` - open in new window (basic implementation)
- [ ] `rel="noopener"` - security for external links (parsed but not enforced)
- [ ] `download` - download link

### Image Features
- [ ] `srcset` - responsive image sources
- [ ] `sizes` - responsive image sizes
- [ ] `loading="lazy"` - lazy loading
- [ ] `alt` text display on error

### Accessibility (ARIA)
- [ ] `role` - element role
- [ ] `aria-label` - accessible label
- [ ] `aria-hidden` - hide from screen readers
- [ ] `aria-expanded` - expandable state
- [ ] `aria-describedby` - description reference
- [ ] `tabindex` - keyboard navigation order

### Global Attributes
- [ ] `contenteditable` - editable content
- [ ] `draggable` - drag and drop
- [ ] `hidden` - hide element
- [ ] `title` - tooltip on hover
- [ ] `lang` - language specification
- [ ] `data-*` - custom data attributes

---

## Known Issues
- [x] Whitespace between inline elements missing (e.g., `<strong>`, `<em>`)
- [ ] Text inside `position: absolute` elements not rendering
- [x] `<main>`, `<nav>`, `<section>`, `<article>`, `<aside>` added to blockElements map
- [ ] No keyboard navigation between form elements (Tab key)
- [x] No form validation feedback UI (implemented red border for required fields)
- [ ] Images don't show alt text on load failure
- [ ] `<fieldset>` legend spacing needs fine-tuning (gap between legend text and border)
- [ ] `<fieldset>` without legend shows no top border (basic fieldset case)

---

## Future Enhancements

### Performance
- [ ] Incremental layout (don't recompute entire tree)
- [ ] Virtual scrolling for long pages
- [ ] Image caching to disk
- [ ] Lazy image loading

### User Experience
- [ ] Text selection and copy
- [ ] Find in page (Ctrl+F)
- [ ] Zoom in/out
- [ ] Print page
- [ ] View page source
- [ ] Developer tools panel

### Standards Compliance
- [ ] DOCTYPE handling
- [ ] Character encoding detection
- [ ] Quirks mode vs standards mode
- [ ] HTML entity decoding (`&nbsp;`, `&amp;`, etc.)
