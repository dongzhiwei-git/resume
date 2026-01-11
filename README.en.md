# ResumeToJob (Go Version)

Author: ricardo

A resume builder powered by Go + Gin and plain HTML/CSS. It provides a Chinese UI by default and now includes an English README. Features include real‑time preview, multiple experience/education entries, theme templates, color/font size/paper size configuration, and PDF printing/export.

## Features
- Real‑time preview (split view: editor on the left, preview on the right)
- Multiple entries: add/remove work experiences and education records
- Theme configuration: classic/modern/minimal templates, theme color, font size, paper size (A4/Letter)
- Print/export: one‑click printing or save as PDF in the preview page
- Simple frontend templates: easy to customize structure and styles

## Project Structure
- `main.go`: server entry, routing, form parsing, template rendering
- `templates/`: HTML templates
  - `index.html`: landing page
  - `editor.html`: editor (split view + live preview)
  - `view.html`: full preview page for printing
  - `resume_content.html`: resume content fragment for live preview API
  - `header.html`, `footer.html`: shared layout
- `static/css/style.css`: base styles
- `go.mod`: dependencies and versions

## Getting Started
- Environment: Go 1.19 or newer
- Install deps:
  - `go mod tidy`
- Run:
  - `go run main.go`
- Access:
  - `http://localhost:8080`
  - Editor: `http://localhost:8080/editor`

## Usage
- Fill personal info, experiences, and education; use “Add” buttons to create multiple entries
- The right preview updates automatically
- Click “Generate Full Preview / Print” to open a full screen preview and use the browser print dialog to save as PDF

## API
- `POST /api/preview`: returns `resume_content.html` based on form data (used for live preview)
- `POST /preview`: returns the full preview page for printing

## Form Fields
- Basic: `name`, `email`, `phone`, `summary`
- Theme (dot notation): `config.template`, `config.color`, `config.font_size`, `config.paper_size`
- Experience (array, dot notation): `experience[0].title`, `experience[0].company`, `experience[0].date`, `experience[0].description`
- Education (array, dot notation): `education[0].degree`, `education[0].school`, `education[0].date`

Note: To ensure compatibility with older Go/Gin environments, the backend parses `PostForm` directly and maps complex array fields robustly.

## Customize & Extend
- Modify `resume_content.html` and styles; use `--theme-color`, `template-xxx`, `font-size-xxx`
- Add new `config.template` options and corresponding styles
- Fonts & i18n: you may add Web Fonts; Chinese‑safe default fonts are used to avoid garbling

## Troubleshooting
- Experience/Education not shown: ensure dot notation (e.g., `experience[0].title`) is used; the project auto reindexes indices and parses on backend
- Slow deps: set `GOPROXY` (e.g., `go env -w GOPROXY=https://goproxy.cn,direct`) then run `go mod tidy`

## License
- Non‑Commercial License: Attribution must be “ricardo”; commercial use is prohibited.
