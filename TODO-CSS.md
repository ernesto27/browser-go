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
- [ ] CSS comments `/* */` - parser does not skip comment blocks; breaks or corrupts parsing (`css/parser.go` — `skipWhitespace()` only skips unicode spaces)

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
- [ ] `letter-spacing` - NOT in `currentStyle`; also parsed but not applied in render
- [ ] `word-spacing` - NOT in `currentStyle`

### §5.2 Font Properties
- [ ] `font-variant` - `normal | small-caps`
- [ ] `font` - shorthand (font-style/variant/weight/size/line-height/family)
- [ ] `font-size` keyword values - `xx-small`, `x-small`, `small`, `medium`, `large`, `x-large`, `xx-large`, `larger`, `smaller`

### §5.3 Background Properties
- [ ] `background-attachment` - `scroll | fixed`

### §4.1.1 Box Model
- [ ] Margin collapsing - adjacent vertical margins should collapse to the maximum value

---

## Missing Properties

### Background
- [x] `background` - shorthand (color and url)
- [ ] `background-position` - position
- [ ] `background-size` - cover/contain/size
- [ ] `background-repeat` - repeat/no-repeat

### Border
- [x] `border-top-left-radius` - individual corner (pixel-raster path in `render/rounded.go`)
- [x] `border-top-right-radius` - individual corner (pixel-raster path in `render/rounded.go`)
- [x] `border-bottom-left-radius` - individual corner (pixel-raster path in `render/rounded.go`)
- [x] `border-bottom-right-radius` - individual corner (pixel-raster path in `render/rounded.go`)

### Box Model
- [ ] `box-sizing` - border-box/content-box
- [ ] `overflow` - visible/hidden/scroll/auto
- [ ] `overflow-x` - horizontal overflow
- [ ] `overflow-y` - vertical overflow
- [ ] `clear` - clear floats

### Typography
- [x] `line-height` - line spacing
- [ ] `letter-spacing` - character spacing
- [ ] `word-spacing` - word spacing
- [ ] `text-shadow` - text shadow
- [ ] `white-space` - whitespace handling
- [ ] `text-overflow` - ellipsis/clip
- [ ] `text-indent` - first line indent
- [ ] `vertical-align` - inline alignment

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
- [ ] `list-style` - shorthand
- [ ] `list-style-type` - disc/circle/square/decimal/none
- [ ] `list-style-position` - inside/outside
- [ ] `list-style-image` - custom marker

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
- [x] `em` - relative to font size
- [ ] `rem` - relative to root font size
- [ ] `%` - percentage (partial: works for table cell width)
- [x] `vw` - viewport width
- [x] `vh` - viewport height
- [ ] `calc()` - calculations
- [ ] `rgb()` / `rgba()` - color functions (`css/css.go:204` — ParseColor handles named colors and hex; ~80 CSS named colors now included)
- [ ] `hsl()` / `hsla()` - color functions
- [ ] `transparent` keyword — not recognized by ParseColor; returns nil (no background drawn)

### Selectors
- [ ] Child selectors - `ul > li`
- [ ] Pseudo-classes - `:hover`, `:focus`, `:active`, `:first-child`, `:last-child`
- [ ] Pseudo-elements - `::before`, `::after`
- [ ] Attribute selectors - `[type="text"]`, `[href^="https"]`
- [ ] Sibling selectors - `h1 + p`, `h1 ~ p`

### Cascade & Specificity
- [x] Specificity calculation - proper weighting via `[3]int` (ID, class, tag) in `css/css.go`
- [x] `!important` - override rules

### At-Rules (needed for WHATWG 4.2.6 style element compliance)
- [ ] `@media` - media queries (screen, print, width conditions)
- [ ] `@import` - import external stylesheets
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
- [ ] `letter-spacing` - parsed into `Style.LetterSpacing` but never read by layout or render
- [ ] `display: block/inline` - only `none` actually works
- [ ] `display: inline-block` - not recognized; treated as `InlineBox` without proper inline-block sizing (`layout/layout.go:178`)
- [ ] `position: relative/fixed/sticky` - only `absolute` works
  - `position: fixed` bug: not extracted from normal flow like `absolute` is (`layout/compute.go:72-81`)
- [ ] `z-index` - parsed but stacking order not enforced (all elements paint in DOM order)

---

## Known Issues
- [ ] `position: absolute` - text/color inside positioned elements not rendering
- [ ] `top/left/right/bottom: 0` silently ignored — `> 0` guard in `computeBlockLayout` and `mergeStyles` drops zero and negative offsets (`layout/compute.go:485-497`)
