# Browser-Go (Educational Browser Engine)

Browser-Go is a learning-focused web browser engine written in Go. It fetches HTML over HTTP, parses it into a DOM tree, computes layout, and renders the result in a GUI window using the Fyne toolkit. The project is intentionally simple and readable so that each stage of a browser pipeline can be explored and modified.

## Project Goals

- Demonstrate a minimal, end-to-end browser pipeline.
- Explore core web platform concepts (DOM, layout, rendering, events).
- Track ongoing research against WHATWG HTML, DOM, and CSS specifications.
- Provide a foundation for incremental additions (CSS, JavaScript, layout).

## High-Level Architecture

```
URL → HTTP Fetch → DOM Tree → Layout Tree → Display Commands → GUI
        (main.go)    (dom/)     (layout/)      (render/)       (render/)
```

### Pipeline Stages

1. **Network (main.go)**: Performs HTTP GET/POST to fetch HTML and assets.
2. **HTML Parsing (dom/)**: Builds a DOM tree from HTML tokens.
3. **Layout (layout/)**: Builds a layout tree, computes box sizes/positions.
4. **Painting (render/)**: Generates display commands.
5. **GUI (render/)**: Renders to a Fyne canvas with click handling.

## Repository Structure

| Path | Purpose |
|------|---------|
| `main.go` | Entry point and pipeline orchestration |
| `dom/` | DOM tree representation and HTML parsing |
| `css/` | CSS parsing (partial application) |
| `layout/` | Box model and layout computation |
| `render/` | Fyne window, painting, and interaction |
| `js/` | JavaScript runtime bindings (Goja) |
| `testpage/` | Manual testing HTML page |
| `serve.sh` | Convenience script for a local HTTP server |

## Features Snapshot

### HTML
- Core document structure (`html`, `head`, `body`, `title`, `style`, `script`, `link`).
- Text content elements (`h1`-`h6`, `p`, `div`, `span`, `blockquote`, `pre`).
- Inline formatting (`strong`, `em`, `u`, `small`, `del`, `ins`, `q`).
- Lists (`ul`, `ol`, `li`, `dl`, `dt`, `dd`).
- Tables (`table`, `thead`, `tbody`, `tfoot`, `tr`, `th`, `td`, `caption`).
- Forms (`form`, `input`, `textarea`, `select`, `option`, `label`, `button`).

### CSS (Partial)
- Color, background color, borders, margins, padding.
- Typography basics: font size, weight, style, text-align, text-transform.
- Box sizing: width/height and min/max constraints.
- Positioning: static + absolute (with known limitations).
- User-agent defaults for common block elements.

### JavaScript (Partial)
- Goja runtime integration, script execution.
- Document and element wrappers with core DOM access.
- Event registration (addEventListener) and click dispatch.
- Basic dialog APIs (`alert`, `confirm`, `prompt`).

## Build and Run

### Prerequisites

- Go 1.21+.
- Fyne runtime dependencies (OpenGL + windowing libraries). On Linux, this typically includes X11 and OpenGL headers (`libx11`, `libgl`, etc.).

### Commands

```bash
# Build the browser
go build -o browser

# Run with a URL
go run . https://example.com

# Serve the test page locally
./serve.sh
go run . http://localhost:8080
```

## Testing

```bash
# Run all tests
go test ./...

# Package-specific tests
go test ./dom
go test ./layout
```

> Note: GUI packages may require OS-level dependencies to build in headless environments.

## Research Notes & References

This project tracks a subset of browser standards and implementation status in the TODO files:

- `TODO.md` — General roadmap, bugs, and feature ideas.
- `TODO-HTML.md` — HTML tag coverage and missing features.
- `TODO-CSS.md` — CSS properties and selector support.
- `TODO-JS.md` — JavaScript runtime and DOM API coverage.

Relevant specifications:

- HTML: https://html.spec.whatwg.org/
- DOM: https://dom.spec.whatwg.org/
- CSS: https://www.w3.org/TR/CSS/
- ECMAScript: https://tc39.es/ecma262/

## Known Limitations

- No full CSS cascade or selector specificity.
- Limited JavaScript DOM mutation APIs.
- Rendering gaps for some positioned elements and inline whitespace.
- Headless builds may fail without OpenGL/X11 dependencies.

## Contributing

This repository is intended for learning and experimentation. Contributions should favor small, incremental improvements and clear documentation of browser concepts.
