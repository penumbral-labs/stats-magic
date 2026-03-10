package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// presetItem wraps a SpellPreset for the list.Model interface.
type presetItem struct {
	preset SpellPreset
}

func (i presetItem) Title() string {
	return fmt.Sprintf("[%d] %s", i.preset.Rank, i.preset.Name)
}

func (i presetItem) Description() string {
	parts := []string{i.preset.Dice}
	if i.preset.Type == SpellTypeSave {
		parts = append(parts, i.preset.SaveType+" save")
	} else {
		parts = append(parts, "attack roll")
	}
	if i.preset.HeightenDie > 0 {
		parts = append(parts, fmt.Sprintf("+%dd/rank", i.preset.HeightenDie))
	}
	parts = append(parts, i.preset.Description)
	return strings.Join(parts, " | ")
}

func (i presetItem) FilterValue() string {
	return i.preset.PresetFilterString()
}

// pickerMsg is sent when a preset is selected from the picker.
type pickerMsg struct {
	preset SpellPreset
}

// pickerCancelMsg is sent when the picker is dismissed.
type pickerCancelMsg struct{}

// pickerModel wraps a list.Model for spell preset selection.
type pickerModel struct {
	list   list.Model
	active bool
}

func newPickerModel(width, height int) pickerModel {
	presets := AllPresets()
	items := make([]list.Item, len(presets))
	for i, p := range presets {
		items[i] = presetItem{preset: p}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#e94560")).
		BorderLeftForeground(lipgloss.Color("#e94560"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#a8a8b3")).
		BorderLeftForeground(lipgloss.Color("#e94560"))

	pickerWidth := width - 8
	if pickerWidth < 50 {
		pickerWidth = 50
	}
	if pickerWidth > 90 {
		pickerWidth = 90
	}
	pickerHeight := height - 6
	if pickerHeight < 10 {
		pickerHeight = 10
	}
	if pickerHeight > 30 {
		pickerHeight = 30
	}

	l := list.New(items, delegate, pickerWidth, pickerHeight)
	l.Title = "Select a Spell Preset"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e94560")).
		Bold(true).
		Padding(0, 1)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#e94560"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#e94560"))

	return pickerModel{list: l, active: false}
}

func (p pickerModel) Update(msg tea.Msg) (pickerModel, tea.Cmd) {
	if !p.active {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't intercept if the list is filtering
		if p.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			p.list, cmd = p.list.Update(msg)
			return p, cmd
		}

		switch msg.String() {
		case "enter":
			if item, ok := p.list.SelectedItem().(presetItem); ok {
				p.active = false
				return p, func() tea.Msg { return pickerMsg{preset: item.preset} }
			}
		case "esc", "q":
			p.active = false
			return p, func() tea.Msg { return pickerCancelMsg{} }
		}
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p pickerModel) View() string {
	if !p.active {
		return ""
	}

	overlay := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#e94560")).
		Padding(1, 2).
		Render(p.list.View())

	return overlay
}
