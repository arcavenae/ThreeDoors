# Themes

ThreeDoors includes multiple door themes that change the visual appearance of the doors view. Each theme has a distinct aesthetic that affects door frames, borders, and text styling.

---

## Available Themes

| Theme | Description |
|-------|-------------|
| `classic` | Traditional door styling with standard box-drawing characters |
| `modern` | Contemporary, clean design with improved contrast |
| `scifi` | Sci-fi / cyberpunk aesthetic with angular frames |
| `shoji` | Japanese minimalist sliding doors with large panes and thin frames |

### Seasonal Variants

Each theme has seasonal variants that activate automatically based on the date:

- **Spring** — Fresh, lighter color palette
- **Summer** — Warm, vibrant tones
- **Autumn** — Rich, earthy colors
- **Winter** — Cool, muted palette

Seasonal auto-switching happens transparently. The base theme determines the overall frame structure; the season modifies the color palette.

---

## Switching Themes

### In the TUI

Run the `:theme` command to open the theme picker. Navigate through available themes and press Enter to select.

### In config.yaml

```yaml
theme: modern
```

### Via CLI

```bash
threedoors config set theme scifi
```

Your choice is saved to `config.yaml` and persists across sessions.

---

## Theme Selection During Onboarding

The onboarding wizard includes a theme picker step where you can preview and select your preferred theme before using ThreeDoors for the first time. You can always change it later.
