# CSS Implementation TODO

---

# CSS1 (W3C Recommendation — https://www.w3.org/TR/CSS1/)

## §1 Basic Concepts

### §1.3 Inheritance
Inheritable properties fall back to the parent's value when no rule applies.
Works via **render-level inheritance**: `paintLayoutBox` passes `currentStyle TextStyle` by value down the tree (`render/paint.go:284-313`), only overriding when a box has an explicit non-nil style. This means it affects paint only — layout-affecting properties still need CSS-level inheritance.

- [x] `color` - inherited via `currentStyle.Color` in render pass
- [x] `font-family` - inherited via `currentStyle.FontFamily` in render pass
- [x] `font-weight` - inherited via `currentStyle.Bold` in render pass
- [x] `font-style` - inherited via `currentStyle.Italic` in render pass
- [x] `font-size` - inherited via `currentStyle.Size` in render pass
- [x] `text-decoration` - inherited via `currentStyle.TextDecoration` in render pass
- [x] `text-transform` - inherited via `currentStyle.TextTransform` in render pass
- [x] `text-align` - inherited in layout tree style propagation before line layout
- [x] `line-height` - NOT in `currentStyle`; affects layout, needs CSS-level inheritance
- [x] `letter-spacing` - NOT in `currentStyle`; parsed and applied in render
- [x] `word-spacing` - NOT in `currentStyle`
- [ ] `list-style` inheritance - list-style properties should inherit to nested list items
- [x] `white-space` inheritance - inherited in layout tree build (`layout/layout.go`), affects wrap decisions in `compute.go`

### §1.7 CSS Parsing
- [x] CSS comments `/* */` - `skipWhitespace()` now skips `/* ... */` blocks (§1.7: "a comment is equivalent to whitespace")
- [x] Forward-compatible parsing (§7.1) - unknown at-rules skipped via `skipAtRule()` (handles both statement and block forms); unknown properties/values silently ignored by switch fallthrough

### §2 Selectors
- [x] Type selectors - `H1`, `P`, `DIV`
- [x] Class selectors - `.class`
- [x] ID selectors - `#id`
- [x] Descendant selectors - `div p`, `ul li` (via `Selector.Ancestor` chain in `css/parser.go`)
- [x] Grouping - `H1, H2, H3` (comma-separated selectors)

### §2.1 Pseudo-classes
- [~] `A:link` - parsed and matched (`css/css.go`)
- [~] `A:visited` - parsed and matched (`css/css.go`)
- [ ] `A:active` - parsed but not matched; code says "not yet supported" (`css/css.go:569`)

### §2.3–§2.4 Pseudo-elements
- [~] `:first-line` (§2.3) - apply styles to first formatted line of a block element (render-time only; font-size won't affect line breaking; no inheritance into nested inline elements)
- [ ] `:first-letter` (§2.4) - apply styles to first letter (drop caps, initial caps)

### §3 At-Rules
- [x] `@import` - import external stylesheets (must occur at start of stylesheet, before any declarations)

### §3 Cascade & Specificity
- [x] Specificity calculation - proper weighting via `[3]int` (ID, class, tag) in `css/css.go`
- [x] `!important` - override rules

### §4 Formatting Model
- [~] Margin collapsing (§4.1.1) - adjacent positive vertical margins between sibling block elements now collapse to max; parent/child, empty-block, and full negative-margin behavior still pending
- [ ] Horizontal formatting 7-property constraint (§4.1.2) - sum of margin-left + border-left + padding-left + width + padding-right + border-right + margin-right must equal parent width
- [ ] `display: list-item` (§4.1.3/§5.6.1) - formatted as block with list-item marker
- [ ] Replaced elements (§4.4) - images, form elements: intrinsic width/height, auto sizing

## §5 Properties

### §5.2 Font Properties
- [x] `font-family` - font stack
- [x] `font-style` - `normal | italic | oblique`
- [x] `font-variant` - `normal | small-caps`
- [x] `font-weight` - `normal | bold | bolder | lighter | 100-900`
- [x] `font-size` - text size
- [x] `font-size` keyword values - `xx-small`, `x-small`, `small`, `medium`, `large`, `x-large`, `xx-large`, `larger`, `smaller`
- [x] `font` - shorthand (font-style/variant/weight/size/line-height/family)

### §5.3 Color & Background Properties
- [x] `color` - text color
- [x] `background-color` - background color
- [x] `transparent` keyword (§5.3.2) — added as `RGBA{0,0,0,0}` in ParseColor named colors map
- [x] `background-image` - url() images (remote + local files)
- [ ] `background-repeat` - `repeat | repeat-x | repeat-y | no-repeat` (§5.3.4)
- [ ] `background-attachment` - `scroll | fixed` (§5.3.5)
- [ ] `background-position` - position (§5.3.6)
- [x] `background` - shorthand (color and url)

### §5.4 Text Properties
- [x] `word-spacing` - extra spacing applied per word in render (§5.4.1)
- [x] `letter-spacing` - character spacing (§5.4.2)
- [x] `text-decoration` - `underline | line-through` (§5.4.3)
- [x] `text-decoration: overline` - §5.4.3 specifies `overline` and `blink` in addition to `underline`/`line-through`
- [~] `vertical-align` - parsed into `Style.VerticalAlign` but not used in inline layout (§5.4.4 — baseline/sub/super/top/text-top/middle/bottom/text-bottom/percentage)
- [x] `text-transform` - `uppercase | lowercase | capitalize` (§5.4.5)
- [x] `text-align` - `left | center | right` (§5.4.6)
- [x] `text-align: justify` (§5.4.6)
- [~] `text-indent` - first line indent (§5.4.7 — parsed, wrapping-aware, render offset; inheritance not yet implemented)
- [x] `line-height` - line spacing, unitless/px/normal keyword (§5.4.8)

### §5.5 Box Properties
- [x] `margin-top/right/bottom/left` - individual margins (§5.5.1–§5.5.4)
- [x] `margin` - shorthand with 1-4 values (§5.5.5)
- [x] `margin: auto` - horizontal centering
- [x] `padding-top/right/bottom/left` - individual padding (§5.5.6–§5.5.9)
- [x] `padding` - shorthand with 1-4 values (§5.5.10)
- [x] `border-top-width/right-width/bottom-width/left-width` (§5.5.11–§5.5.14)
- [x] `border-width` - shorthand (§5.5.15)
- [x] `border-color` (§5.5.16)
- [x] `border-style` - parsed (§5.5.17)
- [ ] `border-style` rendering variations (§5.5.17) - `dotted`, `dashed`, `double`, `groove`, `ridge`, `inset`, `outset` all parsed but rendered as solid (`render/paint.go`)
- [x] `border-width` keywords (§5.5.11) - `thin | medium | thick`
- [x] `border-top/right/bottom/left` - individual borders (§5.5.18–§5.5.21)
- [x] `border` - shorthand (§5.5.22)
- [x] `width` (§5.5.23)
- [x] `height` (§5.5.24)
- [x] `float` - `left | right | none` (§5.5.25)
- [x] `clear` - `none | left | right | both` (§5.5.26)

### §5.6 Classification Properties
- [x] `display: block | inline | none` (§5.6.1)
- [~] `display: block/inline` - only `none` actually works; block/inline parsed but not enforced (§5.6.1)
- [ ] `display: list-item` (§5.6.1)
- [~] `white-space` - `normal` and `nowrap` supported; `pre` not yet implemented (§5.6.2)
- [x] `list-style-type` - disc/circle/square/decimal/none (§5.6.3)
- [ ] `list-style-type` extended values (§5.6.3) - `lower-roman`, `upper-roman`, `lower-alpha`, `upper-alpha`
- [ ] `list-style-image` - custom marker (§5.6.4)
- [ ] `list-style-position` - inside/outside (§5.6.5)
- [x] `list-style` - shorthand (§5.6.6)

## §6 Units

- [x] `px` - pixels (§6.1)
- [x] `em` - relative to font size (§6.1)
- [ ] `ex` - relative to x-height of font (§6.1 — typically ~0.5em)
- [ ] `pt` - points, 1pt = 1/72in (§6.1)
- [ ] `pc` - picas, 1pc = 12pt (§6.1)
- [ ] `in` - inches (§6.1)
- [ ] `cm` - centimeters (§6.1)
- [ ] `mm` - millimeters (§6.1)
- [ ] `%` - percentage (§6.2 — partial: works for table cell width only)
- [ ] `rgb()` - color function (§6.3 — `css/css.go` ParseColor handles named colors and hex only)
- [x] Named colors - standard CSS1 color keywords (§6.3)
- [x] `#hex` colors - 3 and 6 digit hex notation (§6.3)

## CSS1 — User-Agent Defaults
- [x] User-agent default styles (margins for p, h1-h6, ul, ol, blockquote, hr)
- [x] Word wrapping for long text - text wraps within container width
