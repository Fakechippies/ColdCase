# Design System Master File

> **LOGIC:** When building a specific page, first check `design-system/pages/[page-name].md`.
> If that file exists, its rules **override** this Master file.
> If not, strictly follow the rules below.

---

**Project:** ColdCase Console
**Updated:** 2026-11-18
**Category:** Local forensics operations console
**Pattern:** Real-Time / Operations (read-only telemetry) + dense dashboard
**Design Dials:** Variance 5/10 (balanced) · Motion 3/10 (subtle) · Density 8/10 (dashboard)

---

## Global Rules

### Color Palette

Black surfaces, ember-orange primary, cool-gray neutrals, status accents.

| Role | Hex | CSS Variable |
|------|-----|--------------|
| Background | `#080809` | `--bg` |
| Background soft | `#0B0B0D` | `--bg-soft` |
| Surface / card | `#111114` | `--card` |
| Surface raised | `#16161A` | `--card-2` |
| Surface hover | `#1B1B20` | `--card-3` |
| Inset (inputs, terminal) | `#050506` | `--inset` |
| Border | `#24242B` | `--border` |
| Border strong | `#35353E` | `--border-strong` |
| Foreground | `#F4F4F5` | `--fg` |
| Muted foreground | `#A1A1AA` | `--muted` |
| Faint foreground | `#71717A` | `--faint` |
| Primary / accent | `#F97316` | `--orange` |
| Primary bright | `#FB923C` | `--orange-bright` |
| Primary soft fill | `rgba(249,115,22,.13)` | `--orange-soft` |
| Primary hairline | `rgba(249,115,22,.40)` | `--orange-line` |
| Success | `#22C55E` | `--green` |
| Info | `#38BDF8` | `--blue` |
| Warning | `#F59E0B` | `--amber` |
| Danger | `#EF4444` | `--red` |

**Color notes:** Orange is the single accent for primary actions, active nav,
focus rings, chart highlights and links. Status colors are semantic only
(available/missing, running/failed, unlocked/locked/sealed) and are always
paired with a label or icon so color is never the only indicator.

### Typography

- **Heading / body:** Inter, system-ui fallback stack
- **Mono:** JetBrains Mono, ui-monospace fallback stack (tool names, commands, hashes, terminal)
- **Scale:** 11 / 13 / 14 / 16 / 20 / 26 / 34 px
- **Mood:** technical, precise, high-contrast, ops-room calm

No webfont imports: the console is local-only and must render offline. Use the
system stack already declared in `web/static/app.css`.

### Spacing

4/8 px rhythm, dense dashboard density.

| Token | Value | Usage |
|-------|-------|-------|
| `--s1` | 4px | Icon gaps, tight inline |
| `--s2` | 8px | Standard inline gaps |
| `--s3` | 12px | Control padding, list gaps |
| `--s4` | 16px | Card internals, grid gaps |
| `--s5` | 24px | Card padding, section rhythm |
| `--s6` | 32px | Page gutters |
| `--s7` | 48px | Hero / empty-state space |

### Shape, elevation, motion

- Radius: `12px` cards, `8px` controls, `6px` small inset, `999px` badges/chips.
- Shadow: `0 1px 0 rgba(255,255,255,.02) inset, 0 10px 30px rgba(0,0,0,.28)`.
- Motion: `180ms cubic-bezier(.4,0,.2,1)` for background, border, color, transform.
- Press state: `translateY(1px)` only — never layout-shifting scale.
- `prefers-reduced-motion: reduce` disables transitions and animations.

---

## Layout Shell

- Fixed left sidebar (`248px`): brand, primary nav, runtime + key status footer.
- Sticky topbar: eyebrow, page title, subtitle, page actions.
- Single scroll column of cards inside `main`; `--s5` vertical rhythm.
- Nav tabs: **Dashboard** (read-only telemetry) · **Run Tool** · **Tool Catalog** ·
  **Availability** · **Sessions**. Actions never live on the dashboard.
- Collapse to a stacked layout below `860px`; sidebar becomes a wrapping nav bar.

---

## Component Specs

### Buttons

```css
.btn {
  min-height: 36px;
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--card-2);
  color: var(--fg);
  font: 600 13px/1 var(--font-sans);
  cursor: pointer;
  transition: background 180ms, border-color 180ms, color 180ms, transform 180ms;
}
.btn:hover { background: var(--card-3); border-color: #4a4a55; }
.btn:active { transform: translateY(1px); }
.btn:disabled { opacity: .4; cursor: not-allowed; }

.btn--primary {
  background: linear-gradient(180deg, var(--orange-bright), var(--orange));
  border-color: var(--orange);
  color: #140800;
  box-shadow: 0 8px 20px rgba(249, 115, 22, .18);
}
.btn--danger { background: var(--red-soft); border-color: rgba(239,68,68,.5); color: #fca5a5; }
.btn--ghost  { background: transparent; border-color: var(--border); color: var(--muted); }
```

### Cards

```css
.card {
  background: linear-gradient(180deg, var(--card) 0%, #0f0f12 100%);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px;
  box-shadow: var(--shadow);
}
.card__head { display: flex; justify-content: space-between; gap: 16px; margin-bottom: 16px; }
.card__sub  { color: var(--faint); font-size: 11px; }
```

### Inputs

```css
input, select, textarea {
  background: var(--inset);
  color: var(--fg);
  border: 1px solid #2d2d35;
  border-radius: 8px;
  padding: 9px 11px;
  font-size: 13px;
}
input:focus, select:focus, textarea:focus {
  outline: none;
  border-color: var(--orange);
  box-shadow: 0 0 0 3px var(--orange-soft);
}
```

### File dropzone

- Dashed `#3b3b45` border on inset surface; orange border + 5% orange wash on hover/drag.
- Icon tile, filename + hint, pill-shaped "Browse" affordance.
- Real `<input type="file">` visually hidden but focusable via the wrapper label.

### Badges, chips, dots

- Badges: pill, 10px uppercase-ish lowercase text, semantic tinted background + border.
- Chips: filter controls; active chip is solid orange with near-black text.
- Status dots always accompany a text label; `dot--ok/info/warn/danger/off` plus
  `dot--unlocked/locked/sealed` for session state.

### Data visualization (server-rendered SVG)

No chart library, no inline styles, CSP-safe.

- **Donut:** 140×140 viewBox, r=54, stroke width 13, track `#1e1e24`.
  Segments: native = orange, container = blue, missing = `#3f3f46`.
  Center label shows available percentage.
- **Coverage meter:** 100×8 viewBox with `preserveAspectRatio="none"`, stacked
  rounded segments per category (native / container / missing).
- **Bar chart:** 100×82 viewBox, 7 daily bars, value labels above bars, day labels below.
- **Usage bars:** 100×6 meter per tool, orange fill at 85% opacity.
- Charts always carry `role="img"` + `aria-label`; decorative SVG is `aria-hidden`.

---

## Style Guidelines

**Style:** Technical Ops / Terminal Noir

**Keywords:** black, ember orange, high contrast, mono data, dense, scannable, precise

**Key effects:** thin hairline borders, subtle top-lit gradients on cards,
8px glow ring on status dots, orange left rail on the active nav item,
180ms hover transitions. No glow-heavy cyberpunk clutter, no gradients on text.

### Page patterns

- **Dashboard:** intro + environment pills → 4 KPI stats → 2×2 chart grid →
  category coverage bars → sessions table + latest-command feed. Read-only.
- **Availability:** coverage donut + execution-path box → per-category coverage →
  filterable/searchable tool table with status indicators.
- **Tool Catalog:** category chips with icons → per-category card grids with tool
  icons, binary, availability badge and root marker.
- **Run Tool:** tool picker → attach evidence + args + session → live SSE terminal.
- **Sessions:** create form + lifecycle legend → session table; detail page has
  verification banner, command history and evidence hashes.

---

## Anti-Patterns (Do NOT Use)

- ❌ Actions or forms on the dashboard
- ❌ Color-only status indicators
- ❌ Emojis as icons — use inline SVG (Lucide-style 24×24 stroke paths)
- ❌ Inline `style` attributes or inline `<script>` (CSP: `style-src 'self'`)
- ❌ Low-contrast muted text below 4.5:1
- ❌ Layout-shifting hover transforms
- ❌ Instant state changes — use the 180ms token
- ❌ Invisible focus states

---

## Pre-Delivery Checklist

- [ ] No emojis; every icon is inline SVG from `pkg/server/icons.go`
- [ ] Orange reserved for primary action, active state, focus, and chart accent
- [ ] Status color always paired with text or icon
- [ ] Focus visible on all interactive elements (orange ring)
- [ ] `prefers-reduced-motion` respected
- [ ] Charts render without JS and carry accessible labels
- [ ] Responsive at 375 / 768 / 1024 / 1440 px; no horizontal scroll
- [ ] Dashboard contains no mutating controls
- [ ] Touch/click targets ≥ 36px (desktop) with expanded hit areas on compact controls
