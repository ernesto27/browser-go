# CSS Implementation TODO

## Completed
- [x] `color` - text color
- [x] `background-color` - background color
- [x] `background-image` - url() images (remote)
- [x] `background-image` - local files (file:// and absolute paths)
- [x] `font-size` - text size
- [x] `font-weight` - bold
- [x] `font-style` - italic
- [x] `font-family` - font stack
- [x] `margin` - all sides
- [x] `margin-top/right/bottom/left` - individual margins
- [x] `padding` - all sides
- [x] `padding-top/right/bottom/left` - individual padding
- [x] `text-align` - left/center/right
- [x] `text-decoration` - underline/line-through
- [x] `text-transform` - uppercase/lowercase/capitalize
- [x] `border` - shorthand
- [x] `border-width/color/style` - border properties
- [x] `border-top/right/bottom/left` - individual borders
- [x] `border-radius` - rounded corners
- [x] `width` - element width
- [x] `height` - element height
- [x] `min-width` - minimum width
- [x] `max-width` - maximum width
- [x] `min-height` - minimum height
- [x] `max-height` - maximum height
- [x] `display` - block/inline/none
- [x] `position` - static/relative/absolute
- [x] `top/left/right/bottom` - position offsets
- [x] `z-index` - stacking order
- [x] `float` - left/right
- [x] `opacity` - transparency
- [x] `visibility` - visible/hidden
- [x] `cursor` - pointer/text/crosshair
- [x] `em` unit - relative to parent font size
- [x] User-agent default styles (margins for p, h1-h6, ul, ol, blockquote, hr)
- [x] `line-height` - line spacing (unitless, px, normal keyword)
- [x] Word wrapping for long text - text wraps within container width
- [x] `margin: auto` - horizontal centering
- [x] `margin` multi-value shorthand - `margin: 10px 20px` (2/3/4 values)
- [x] Descendant selectors - `div p`, `ul li` (implemented via `Selector.Ancestor` chain in `css/parser.go`)
- [x] Specificity calculation - proper weighting via `[3]int` (ID, class, tag) in `css/css.go`

---

# BASIC CONCEPTS (CSS1 §1)

### §1.7 CSS Parsing
- [x] CSS comments `/* */` - `skipWhitespace()` now skips `/* ... */` blocks (CSS1 §1.7: "a comment is equivalent to whitespace")
- [ ] Forward-compatible parsing (§7.1) - unknown at-rules should be skipped gracefully; unknown properties/values already ignored

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
- [x] `opacity` - inherited via `currentStyle.Opacity` in render pass
- [x] `visibility` - inherited via `currentStyle.Visibility` in render pass
- [x] `text-align` - inherited in layout tree style propagation before line layout
- [x] `line-height` - NOT in `currentStyle`; affects layout, needs CSS-level inheritance
- [x] `letter-spacing` - NOT in `currentStyle`; also parsed but not applied in render
- [x] `word-spacing` - NOT in `currentStyle`

### §5.2 Font Properties
- [x] `font-variant` - `normal | small-caps`
- [x] `font` - shorthand (font-style/variant/weight/size/line-height/family)
- [x] `font-size` keyword values - `xx-small`, `x-small`, `small`, `medium`, `large`, `x-large`, `xx-large`, `larger`, `smaller`

### §2 Pseudo-classes & Pseudo-elements (CSS1)
- [~] `A:link` - parsed and matched (`css/css.go`)
- [~] `A:visited` - parsed and matched (`css/css.go`)
- [ ] `A:active` - parsed but not matched; code says "not yet supported" (`css/css.go:565`)
- [ ] `:first-line` pseudo-element (§2.3) - apply styles to first formatted line of a block element
- [ ] `:first-letter` pseudo-element (§2.4) - apply styles to first letter (drop caps, initial caps)

### §5.3 Background Properties
- [ ] `background-attachment` - `scroll | fixed`

### §4 Formatting Model
- [ ] Margin collapsing (§4.1.1) - adjacent vertical margins should collapse to the maximum value
- [ ] Horizontal formatting 7-property constraint (§4.1.2) - sum of margin-left + border-left + padding-left + width + padding-right + border-right + margin-right must equal parent width
- [ ] `display: list-item` (§4.1.3/§5.6.1) - formatted as block with list-item marker; currently only `none` display value works

---

## Missing Properties

### Background
- [x] `background` - shorthand (color and url)
- [ ] `background-position` - position
- [~] `background-size` - cover/contain/size (partial: parsing + storage done, Fyne rendering has sizing issues)
- [ ] `background-repeat` - repeat/no-repeat

### Border
- [x] `border-top-left-radius` - individual corner (pixel-raster path in `render/rounded.go`)
- [x] `border-top-right-radius` - individual corner (pixel-raster path in `render/rounded.go`)
- [x] `border-bottom-left-radius` - individual corner (pixel-raster path in `render/rounded.go`)
- [x] `border-bottom-right-radius` - individual corner (pixel-raster path in `render/rounded.go`)
- [ ] `border-style` rendering variations (CSS1 §5.5.17) - `dotted`, `dashed`, `double`, `groove`, `ridge`, `inset`, `outset` all parsed but rendered as solid (`render/paint.go`)

### Box Model
- [x] `box-sizing` - border-box/content-box
- [~] `overflow` - partial: `visible|hidden|scroll|auto` parsed; used as fallback for overflow-x/overflow-y; no real scrollbars on shorthand alone
- [x] `overflow-x` - horizontal clipping, scrollbar rendering (track + thumb), drag interaction, scroll offset tracking for `auto`/`scroll`
- [x] `overflow-y` - vertical clipping, scrollbar rendering (track + thumb), drag interaction, scroll offset tracking for `auto`/`scroll`
- [x] `clear` - clear floats

### Typography
- [x] `line-height` - line spacing
- [ ] `letter-spacing` - character spacing (CSS1 §5.4.2)
- [ ] `word-spacing` - word spacing (CSS1 §5.4.2)
- [ ] `text-shadow` - text shadow
- [~] `white-space` - partial: `normal` and `nowrap` supported; `pre` (CSS1 §5.6.2), `pre-wrap`, and `pre-line` not yet implemented
- [x] `text-overflow` - ellipsis/clip
- [ ] `text-indent` - first line indent (CSS1 §5.4.7 — inheritable, applies to block-level elements)
- [ ] `vertical-align` - inline alignment (CSS1 §5.4.4 — baseline/sub/super/top/middle/bottom/text-top/text-bottom/percentage)
- [ ] `text-decoration: overline` - CSS1 §5.4.3 specifies `overline` and `blink` in addition to `underline`/`line-through`

### Flexbox
- [ ] `display: flex` - flex container
- [ ] `flex-direction` - row/column
- [ ] `justify-content` - main axis alignment
- [ ] `align-items` - cross axis alignment
- [ ] `align-content` - multi-line alignment
- [ ] `flex-wrap` - wrap items
- [ ] `gap` - spacing between items
- [ ] `flex-grow` - grow factor
- [ ] `flex-shrink` - shrink factor
- [ ] `flex-basis` - initial size
- [ ] `order` - item order
- [ ] `align-self` - individual alignment

### Grid
- [ ] `display: grid` - grid container
- [ ] `grid-template-columns` - column definitions
- [ ] `grid-template-rows` - row definitions
- [ ] `grid-gap` / `gap` - grid spacing
- [ ] `grid-column` - column span
- [ ] `grid-row` - row span

### Effects
- [ ] `box-shadow` - drop shadow
- [ ] `transform` - rotate/scale/translate
- [ ] `transform-origin` - transform center
- [ ] `transition` - animated changes
- [ ] `animation` - keyframe animations
- [ ] `filter` - blur/brightness/etc

### List
- [x] `list-style` - shorthand
- [x] `list-style-type` - disc/circle/square/decimal/none
- [ ] `list-style-type` extended values (CSS1 §5.6.3) - `lower-roman`, `upper-roman`, `lower-alpha`, `upper-alpha`
- [ ] `list-style-position` - inside/outside (CSS1 §5.6.5)
- [ ] `list-style-image` - custom marker (CSS1 §5.6.4)

### Table
- [x] `width` on table cells - px and % widths on td/th
- [ ] `border-collapse` - collapse/separate
- [ ] `border-spacing` - cell spacing
- [ ] `table-layout` - auto/fixed

### Other
- [ ] `outline` - focus outline
- [ ] `content` - generated content
- [ ] `pointer-events` - click behavior
- [ ] `user-select` - text selection

### Units (parsing)
- [x] `em` - relative to font size (CSS1 §6.1)
- [ ] `ex` - relative to x-height of font (CSS1 §6.1 — typically ~0.5em)
- [ ] `rem` - relative to root font size
- [ ] `%` - percentage (CSS1 §6.2 — partial: works for table cell width)
- [x] `vw` - viewport width
- [x] `vh` - viewport height
- [ ] `pt` - points, 1pt = 1/72in (CSS1 §6.1 — very common in print-era stylesheets)
- [ ] `pc` - picas, 1pc = 12pt (CSS1 §6.1)
- [ ] `in` - inches (CSS1 §6.1)
- [ ] `cm` - centimeters (CSS1 §6.1)
- [ ] `mm` - millimeters (CSS1 §6.1)
- [ ] `calc()` - calculations
- [ ] `rgb()` / `rgba()` - color functions (CSS1 §6.3 — `css/css.go` ParseColor handles named colors and hex only)
- [ ] `hsl()` / `hsla()` - color functions
- [ ] `transparent` keyword (CSS1 §5.3.2) — not recognized by ParseColor; returns nil (no background drawn)

### Selectors
- [ ] Child selectors - `ul > li`
- [ ] Pseudo-classes - `:hover`, `:focus`, `:active`, `:first-child`, `:last-child`
- [ ] Pseudo-elements - `::before`, `::after`
- [ ] Pseudo-elements (CSS1) - `:first-line` (§2.3), `:first-letter` (§2.4) — CSS1 used single-colon syntax
- [ ] Attribute selectors - `[type="text"]`, `[href^="https"]`
- [ ] Sibling selectors - `h1 + p`, `h1 ~ p`

### Cascade & Specificity
- [x] Specificity calculation - proper weighting via `[3]int` (ID, class, tag) in `css/css.go`
- [x] `!important` - override rules

### At-Rules (needed for WHATWG 4.2.6 style element compliance)
- [x] `@import` - import external stylesheets (CSS1 §3 — must occur at start of stylesheet, before any declarations)
- [ ] `@media` - media queries (screen, print, width conditions)
- [ ] `@charset` - character encoding declaration
- [ ] `@font-face` - custom font definitions
- [ ] `@keyframes` - animation keyframes
- [ ] `@supports` - feature queries

### Shorthand Expansion
- [x] `margin` multi-value - `margin: 10px 20px`, `margin: 10px 20px 30px 40px`
- [ ] `padding` multi-value - `padding: 10px 20px`
- [ ] `border-radius` multi-value - per-corner values

---

## Parsed But Not Applied
- [ ] `cursor` - parsed but not applied in render
- [ ] `letter-spacing` - parsed into `Style.LetterSpacing` but never read by layout or render (CSS1 §5.4.2)
- [ ] `word-spacing` - parsed into `Style.WordSpacing` but never read by layout or render (CSS1 §5.4.1)
- [ ] `vertical-align` - parsed into `Style.VerticalAlign` but not used in inline layout (CSS1 §5.4.4)
- [ ] `display: block/inline` - only `none` actually works (CSS1 §5.6.1)
- [ ] `display: inline-block` - not recognized; treated as `InlineBox` without proper inline-block sizing (`layout/layout.go:178`)
- [ ] `position: relative/fixed/sticky` - only `absolute` works
  - `position: fixed` bug: not extracted from normal flow like `absolute` is (`layout/compute.go:72-81`)
- [ ] `z-index` - parsed but stacking order not enforced (all elements paint in DOM order)
- [ ] `border-style` variations - `dotted`/`dashed`/`double`/`groove`/`ridge`/`inset`/`outset` parsed but all render as solid (CSS1 §5.5.17)

---

## Known Issues
- [ ] `position: absolute` - text/color inside positioned elements not rendering
- [ ] `top/left/right/bottom: 0` silently ignored — `> 0` guard in `computeBlockLayout` and `mergeStyles` drops zero and negative offsets (`layout/compute.go:485-497`)
