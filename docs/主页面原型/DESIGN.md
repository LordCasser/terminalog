```markdown
# Design System Document: Terminal Editorial

## 1. Overview & Creative North Star
**Creative North Star: "The Brutalist Compiler"**

This design system moves beyond the cliché "hacker terminal" aesthetic to create a high-end, editorial experience for developers. It marries the rigid, technical precision of a CLI with the sophisticated layout logic of a premium digital journal. We are not just building a blog; we are building a high-fidelity documentation of thought.

The system breaks the standard "boxed" template by utilizing **intentional asymmetry** and **monolithic layering**. By embracing a 0px border-radius across the entire system, we lean into a sharp, uncompromising Brutalism. The "Compiler" aspect comes from our use of high-contrast "syntax highlighting" colors to guide the eye through complex information architecture, transforming raw text into a curated visual narrative.

---

## 2. Colors: The Dracula Spectrum
Our palette is rooted in the Dracula theme but expanded into functional tiers. We rely on chromatic depth rather than structural lines to define our space.

### The "No-Line" Rule
**Explicit Instruction:** 1px solid borders are strictly prohibited for sectioning. Boundaries must be defined solely through background color shifts. Use `surface-container-low` for secondary sections and `surface-container-high` for elevated interactive zones. The eye should perceive change through tonal transition, not wireframe lines.

### Surface Hierarchy & Nesting
*   **Surface Lowest (#0b0e18):** Used for the "void"—the deepest background level behind the main content.
*   **Surface (#11131e):** The primary canvas for terminal output and blog reading.
*   **Surface Container (#1d1f2b):** Used for sidebar navigation or secondary meta-information.
*   **Surface Container Highest (#323440):** Reserved for "active" terminal lines or highlighted code blocks.

### The "Glass & Gradient" Rule
To add visual "soul," we implement **"Terminal Fog"**:
*   **Glassmorphism:** Floating elements (like command palettes or tooltips) should use `surface-container` with a `backdrop-blur` of 12px and 80% opacity.
*   **Signature Textures:** For Hero sections or CTAs, use a subtle linear gradient from `primary` (#d7baff) to `primary_container` (#bd93f9) at a 45-degree angle. This provides a professional polish that offsets the flat terminal aesthetic.

---

## 3. Typography: Monolithic Precision
We employ a dual-font strategy to balance technical utility with long-form readability.

*   **Display & Headlines (Space Grotesk):** This is our "Editorial" voice. It is wide, geometric, and assertive. Use `display-lg` for post titles to create high-contrast impact against the monospace environment.
*   **Terminal & UI (JetBrains Mono / Fira Code):** Every piece of UI metadata—dates, tags, "lines of code"—must use the monospace stack. This reinforces the "Compiler" identity.
*   **Body Content (Inter):** For the markdown reading experience, we switch to a clean sans-serif. This ensures that 2,000-word deep dives are legible and reduce cognitive load compared to full-page monospace text.

---

## 4. Elevation & Depth
In this system, elevation is a product of **Tonal Layering**, not light sources.

*   **The Layering Principle:** Stack `surface-container` tiers to create depth. A `surface-container-lowest` card placed on a `surface` background creates a "inset" look, perfect for code snippets.
*   **Ambient Shadows:** If an element must "float" (e.g., a modal), use an ultra-diffused shadow: `box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5)`. The shadow should feel like a dark glow, mimicking a CRT screen's depth.
*   **The "Ghost Border" Fallback:** If a container lacks contrast against a background, use the `outline-variant` token at **15% opacity**. This "Ghost Border" provides an accessibility hint without creating a "boxed-in" feeling.

---

## 5. Components

### Buttons
*   **Primary:** Solid `primary` (#d7baff) with `on-primary` (#411478) text. 0px radius.
*   **Secondary:** Ghost style. No background, `outline` color text, with a `surface-variant` background on hover.
*   **Terminal Variant:** Looks like a command input: `> [Action]`. Text-only with a blinking cursor suffix (`_`) on hover.

### Input Fields
*   **Style:** Minimalist underlines only. Use `outline` (#968e9c) for the inactive state and `tertiary` (#31e368) for the active state to mimic a successful terminal command. 
*   **Error State:** Use `error` (#ffb4ab) for both text and the underline.

### Cards & Lists
*   **Rule:** No dividers. 
*   **Execution:** Use `48px` of vertical whitespace (from the spacing scale) to separate list items. For blog cards, use a `surface-container-low` background that shifts to `surface-container-high` on hover.

### Terminal Breadcrumbs
A custom component for this system. Instead of `Home / Blog / Title`, use a directory path style: `~/root/blog/the-brutalist-compiler.md`. Use `secondary` (#ffafd7) for the directory and `foreground` for the file.

---

## 6. Do's and Don'ts

### Do
*   **Do** use extreme typographic scale. A `display-lg` headline next to a `label-sm` date creates the "Editorial" feel.
*   **Do** use `tertiary` (#31e368) for all "Success" or "Live" indicators to mimic a green terminal prompt.
*   **Do** embrace the 0px radius. Every corner must be sharp.

### Don't
*   **Don't** use standard "Grey." Use the Dracula-tinted neutrals (`surface-variant`, `outline`) to maintain the purple/blue undertone.
*   **Don't** use shadows to define cards. Use background color shifts.
*   **Don't** center-align long-form text. All terminal logic is left-aligned; keep the editorial content left-aligned to match the "Brutalist" grid.

---

## 7. Accessibility Note
While the theme is high-contrast, always ensure that text placed on `primary_container` or `secondary_container` meets WCAG AA standards by using the corresponding `on-` tokens (`on_primary_container`, etc.). The Dracula palette is naturally vibrant, but maintain the 4.5:1 ratio for all body text.```