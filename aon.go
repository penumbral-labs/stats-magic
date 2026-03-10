package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// fetchAoNSpell returns a tea.Cmd that fetches and parses a spell from Archives of Nethys.
func fetchAoNSpell(rawURL string) tea.Cmd {
	return func() tea.Msg {
		spell, err := parseAoNSpell(rawURL)
		if err != nil {
			return aonErrorMsg{err}
		}
		return aonSpellMsg{spell}
	}
}

// parseAoNSpell extracts the spell ID from a URL and fetches spell data from the AoN API.
func parseAoNSpell(rawURL string) (Spell, error) {
	id := extractSpellID(rawURL)
	if id == "" {
		return Spell{}, fmt.Errorf("could not extract spell ID from URL")
	}

	data, err := fetchAoNAPI(id)
	if err != nil {
		return Spell{}, err
	}

	return buildSpellFromAoN(data)
}

// extractSpellID pulls the numeric ID from an AoN URL like "...Spells.aspx?ID=1312".
func extractSpellID(rawURL string) string {
	re := regexp.MustCompile(`(?i)[?&]ID=(\d+)`)
	if m := re.FindStringSubmatch(rawURL); m != nil {
		return m[1]
	}
	return ""
}

// aonSpellData holds the relevant fields from the AoN Elasticsearch response.
type aonSpellData struct {
	Name       string   `json:"name"`
	Level      int      `json:"level"`
	SavingThow string   `json:"saving_throw"`
	Markdown   string   `json:"markdown"`
	Heighten   []string `json:"heighten"`
}

// aonSearchResponse represents the Elasticsearch search API response.
type aonSearchResponse struct {
	Hits struct {
		Hits []struct {
			Source aonSpellData `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// fetchAoNAPI queries the AoN Elasticsearch API for a spell by ID.
func fetchAoNAPI(spellID string) (aonSpellData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := fmt.Sprintf(`{"query":{"term":{"_id":"spell-%s"}},"_source":["name","level","saving_throw","markdown","heighten"]}`, spellID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://elasticsearch.aonprd.com/aon/_search",
		strings.NewReader(query))
	if err != nil {
		return aonSpellData{}, fmt.Errorf("request build failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return aonSpellData{}, fmt.Errorf("API fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return aonSpellData{}, fmt.Errorf("API returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return aonSpellData{}, fmt.Errorf("read failed: %w", err)
	}

	var result aonSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return aonSpellData{}, fmt.Errorf("JSON parse failed: %w", err)
	}

	if len(result.Hits.Hits) == 0 {
		return aonSpellData{}, fmt.Errorf("spell not found (ID %s)", spellID)
	}

	return result.Hits.Hits[0].Source, nil
}

// buildSpellFromAoN constructs a Spell from AoN API data.
func buildSpellFromAoN(data aonSpellData) (Spell, error) {
	sp := Spell{
		Name:        data.Name,
		BaseRank:    data.Level,
		Multipliers: DefaultSaveMultipliers(),
		Type:        SpellTypeSave,
		SaveType:    "Reflex",
	}

	// Parse saving throw field (e.g. "basic  Reflex", "basic Fortitude", "")
	saveLower := strings.ToLower(data.SavingThow)
	switch {
	case strings.Contains(saveLower, "fortitude"):
		sp.SaveType = "Fortitude"
	case strings.Contains(saveLower, "will"):
		sp.SaveType = "Will"
	case strings.Contains(saveLower, "reflex"):
		sp.SaveType = "Reflex"
	default:
		// Check markdown for "spell attack" pattern
		if strings.Contains(strings.ToLower(data.Markdown), "spell attack") {
			sp.Type = SpellTypeAttack
			sp.SaveType = ""
			sp.Multipliers = DefaultAttackMultipliers()
		}
	}

	// Extract dice from markdown
	sp.Dice = extractDice(data.Markdown)

	// Extract heightening from the heighten field and markdown
	sp.HeightenDie, sp.HeightenStep = extractHeighten(data)

	return sp, nil
}

// --- Extraction Helpers ---

var (
	// Matches dice like "6d6", "2d12+4", "2d8-1"
	reDice = regexp.MustCompile(`(\d+)d(\d+)(?:\s*([+-])\s*(\d+))?`)
	// Matches "increases by NdN" or "+NdN" in heighten text
	reHeightenDice = regexp.MustCompile(`(?:increases?\s+by\s+|add(?:itional)?\s+|\+)(\d+)d(\d+)`)
)

func extractDice(markdown string) DiceFormula {
	// Strip markdown links and tags for cleaner matching
	text := stripMarkdown(markdown)

	matches := reDice.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		count, _ := strconv.Atoi(m[1])
		sides, _ := strconv.Atoi(m[2])
		bonus := 0
		if m[4] != "" {
			bonus, _ = strconv.Atoi(m[4])
			if m[3] == "-" {
				bonus = -bonus
			}
		}
		// Skip d20 rolls and tiny dice
		if sides == 20 {
			continue
		}
		if count >= 1 && sides >= 4 {
			return DiceFormula{Count: count, Sides: sides, Bonus: bonus}
		}
	}
	return DiceFormula{}
}

func extractHeighten(data aonSpellData) (int, int) {
	// The heighten field contains entries like "+1", "+2"
	// The markdown contains the actual text like "the damage increases by 2d6"
	if len(data.Heighten) == 0 {
		return 0, 0
	}

	// Find the heighten section in markdown
	heightenRe := regexp.MustCompile(`(?i)\*\*Heightened\s*\(\+(\d+)\)\*\*\s*(.+?)(?:\*\*Heightened|\z)`)
	matches := heightenRe.FindAllStringSubmatch(data.Markdown, -1)
	for _, m := range matches {
		increment, _ := strconv.Atoi(m[1])
		if increment != 1 && increment != 2 {
			continue
		}
		body := m[2]
		if dm := reHeightenDice.FindStringSubmatch(body); dm != nil {
			count, _ := strconv.Atoi(dm[1])
			if count > 0 {
				return count, increment
			}
		}
	}

	return 0, 0
}

// stripMarkdown removes markdown links and XML-like tags from AoN markdown content.
func stripMarkdown(md string) string {
	// Remove [text](url) → text
	reLink := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	md = reLink.ReplaceAllString(md, "$1")
	// Remove XML-like tags
	reTag := regexp.MustCompile(`<[^>]+>`)
	md = reTag.ReplaceAllString(md, " ")
	// Remove ** bold markers
	md = strings.ReplaceAll(md, "**", "")
	return md
}
