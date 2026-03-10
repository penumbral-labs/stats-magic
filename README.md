# Stats Magic

Terminal UI for calculating and comparing PF2e spell damage distributions.

Model encounter parameters (DC, saves, AC), browse a library of ~40 Remaster spells,
compare expected damage across degrees of success, and visualize probability
distributions with braille charts.

## Features

- Full PF2e degree-of-success system (crit/hit/miss/fumble with nat 1/20 step adjustments)
- CLT-based damage distribution modeling with mixture PDFs
- Braille area charts and sparklines for probability visualization
- Heightening support with per-rank damage tables
- Side-by-side spell comparison (expected damage, ±1σ range, delta from best)
- Editable encounter parameters (spell DC, attack mod, enemy saves, AC)
- Save/load spell configurations to `~/.config/stats-magic/spells.json`
- Spell preset picker with fuzzy search

## Install

Download a prebuilt binary from the [latest release](https://github.com/penumbral-labs/stats-magic/releases/latest):

```bash
# Linux (amd64)
curl -Lo stats-magic https://github.com/penumbral-labs/stats-magic/releases/latest/download/stats-magic-linux-amd64
chmod +x stats-magic
./stats-magic
```

Or install with Go:

```bash
go install github.com/penumbral-labs/stats-magic@latest
```

Or build from source:

```bash
go build -o stats-magic .
./stats-magic
```

## Usage

| Key | Action |
| --- | --- |
| `j/k` | Navigate spell list |
| `Space` | Toggle spell for comparison |
| `a` | Add spell from preset library |
| `n` | Create new custom spell |
| `e` / `Enter` | Edit selected spell |
| `d` | Remove selected spell |
| `+/-` | Heighten/lower spell rank |
| `Tab` | Edit encounter parameters |
| `Ctrl+S` | Save configuration |
| `q` | Quit |

## License

GPL-3.0. See [LICENSE](LICENSE).
