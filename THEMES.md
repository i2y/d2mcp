# D2 Themes

This document lists all available themes in D2MCP. The D2 library provides 20 themes total - 18 light themes and 2 dark themes.

## How to Use Themes

When using the `d2_render` or `d2_render_to_file` tools, you can specify a theme by its ID:

```json
{
  "content": "a -> b: Hello",
  "format": "svg",
  "theme": 4  // Cool Classics theme
}
```

## Available Themes

### Light Themes

| ID | Name | Description |
|----|------|-------------|
| 0 | Neutral Default | The default D2 theme with neutral colors |
| 1 | Neutral Grey | A greyscale theme |
| 3 | Flagship Terrastruct | Terrastruct's flagship theme |
| 4 | Cool Classics | Classic cool color palette |
| 5 | Mixed Berry Blue | Berry-inspired blue tones |
| 6 | Grape Soda | Purple grape-inspired colors |
| 7 | Aubergine | Deep purple theme |
| 8 | Colorblind Clear | Optimized for colorblind accessibility |
| 100 | Vanilla Nitro Cola | Vanilla and cola inspired |
| 101 | Orange Creamsicle | Orange cream colors |
| 102 | Shirley Temple | Pink and red tones |
| 103 | Earth Tones | Natural earth colors |
| 104 | Everglade Green | Green nature theme |
| 105 | Buttered Toast | Warm yellow/brown tones |
| 300 | Terminal | Terminal/console style |
| 301 | Terminal Grayscale | Grayscale terminal style |
| 302 | Origami | Paper-inspired theme |
| 303 | C4 | C4 architecture diagram style |

### Dark Themes

| ID | Name | Description |
|----|------|-------------|
| 200 | Dark Mauve | Dark theme with mauve accents |
| 201 | Dark Flagship Terrastruct | Dark version of the flagship theme |

## Theme Characteristics

- **IDs 0-99**: Standard light themes
- **IDs 100-199**: Special light themes  
- **IDs 200-299**: Dark themes
- **IDs 300+**: Terminal and special style themes

## Implementation Details

The theme system is implemented using the D2 library's built-in theme catalog (`d2themes/d2themescatalog`). Each theme defines:

- A color palette with neutrals (N1-N7)
- Base colors (B1-B6) for containers
- Alternative colors (AA2, AA4, AA5, AB4, AB5)
- Special rules for certain themes (e.g., Terminal themes have special rendering)

When a theme ID is provided, it's passed to the D2 rendering engine which applies the appropriate color scheme and styling rules to the diagram.
