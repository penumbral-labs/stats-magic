# Stats Magic

[![CI](https://github.com/penumbral-labs/stats-magic/actions/workflows/ci.yml/badge.svg)](https://github.com/penumbral-labs/stats-magic/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/penumbral-labs/stats-magic)](https://github.com/penumbral-labs/stats-magic/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/penumbral-labs/stats-magic)](go.mod)
[![License](https://img.shields.io/github/license/penumbral-labs/stats-magic?cacheSeconds=3600)](LICENSE)

Terminal UI for calculating and comparing PF2e spell damage distributions.

![Demo](https://vhs.charm.sh/vhs-4PEsZ5TmsRXuELqFi55Q0g.gif)

## What It Does

Model encounter parameters against official PF2e creature building tables,
browse a library of ~40 Remaster spells, compare expected damage across
degrees of success, and visualize probability distributions — all from
your terminal.

- **Full degree-of-success engine** — crit/hit/miss/fumble with nat 1/20
  step adjustments, basic save multipliers, attack roll vs AC
- **CLT-based damage modeling** — mixture PDFs weighted by degree
  probabilities, not just averages
- **Braille charts** — area charts and sparklines rendered in Unicode braille
  for high-density visualization in minimal space
- **Heightening** — per-rank damage tables with support for every-N-ranks
  scaling (e.g., Ice Storm: +1d8 per 2 ranks)
- **Spell comparison** — select multiple spells to see expected damage,
  ±1σ range, and delta from best side-by-side
- **Encounter tuning** — PC level drives spell DC, attack mod, and enemy
  stats from GM Core Tables 2-5/2-6. Override any value manually, or
  cycle enemy save profiles (Low/Med/High) per save type
- **AoN import** — paste an Archives of Nethys URL to auto-import spell
  data (dice, save type, heightening) via the AoN Elasticsearch API
- **Persistence** — save/load configurations to
  `~/.config/stats-magic/spells.json`

## Install

Download a prebuilt binary from the
[latest release](https://github.com/penumbral-labs/stats-magic/releases/latest):

```bash
# Linux (amd64)
curl -Lo stats-magic \
  https://github.com/penumbral-labs/stats-magic/releases/latest/download/stats-magic-linux-amd64
chmod +x stats-magic
./stats-magic

# macOS (Apple Silicon)
curl -Lo stats-magic \
  https://github.com/penumbral-labs/stats-magic/releases/latest/download/stats-magic-darwin-arm64
chmod +x stats-magic
./stats-magic
```

Or install with Go:

```bash
go install github.com/penumbral-labs/stats-magic@latest
```

Or build from source:

```bash
git clone https://github.com/penumbral-labs/stats-magic.git
cd stats-magic
go build -o stats-magic .
./stats-magic
```

## Usage

| Key | Action |
| --- | --- |
| `j`/`k` | Navigate spell list |
| `Space` | Toggle spell for comparison |
| `a` | Add spell from preset library |
| `n` | New spell (name or AoN URL) |
| `e`/`Enter` | Edit selected spell |
| `d` | Remove selected spell |
| `+`/`-` | Heighten/lower cast rank |
| `Tab` | Edit encounter parameters |
| `Space` | Cycle save profile (Low/Med/High) when editing saves |
| `Ctrl+S` | Save configuration |
| `q` | Quit |

## How It Works

The engine models each degree of success as a separate damage distribution,
then combines them into a mixture PDF weighted by the probability of each
degree occurring.

For a save spell against DC 28 with enemy Reflex +19:

```text
Target number = DC - save mod = 28 - 19 = 9
P(crit fail) = P(roll ≤ -1)  + P(nat 1 step-down)
P(fail)      = P(roll 0..9)
P(success)   = P(roll 10..19) + P(nat 20 step-up)
P(crit succ) = P(roll ≥ 20)
```

Each degree applies its multiplier (e.g., basic save: crit fail 2x,
fail 1x, success 0.5x, crit success 0x) to the base dice distribution.
The per-degree PDFs are computed via Central Limit Theorem approximation
and summed into a mixture distribution for visualization.

## Data Sources

- **Creature AC**: GM Core Table 2-5 (Moderate tier)
- **Creature Saves**: GM Core Table 2-6 (Low/Moderate/High tiers)
- **PC Spell DC**: Remaster primary caster progression
  (Trained → Expert → Master → Legendary)
- **Spell Data**: Archives of Nethys Elasticsearch API

## License

GPL-3.0. See [LICENSE](LICENSE).
