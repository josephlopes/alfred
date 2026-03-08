package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/alfred/alfred/internal/model"
	"github.com/alfred/alfred/internal/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Padding(0, 1)
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	doneStyle2  = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("240"))
	itemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	inputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// ── Modes ─────────────────────────────────────────────────────────────────────

type mode int

const (
	modeNormal mode = iota
	modeInput
)

// ── Model ─────────────────────────────────────────────────────────────────────

type tuiModel struct {
	tasks  []model.Task
	cursor int
	mode   mode
	input  string
	err    error
}

// ── Messages ──────────────────────────────────────────────────────────────────

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// ── Init ──────────────────────────────────────────────────────────────────────

func (m tuiModel) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case modeNormal:
			return m.updateNormal(msg)
		case modeInput:
			return m.updateInput(msg)
		}
	case errMsg:
		m.err = msg.err
	}
	return m, nil
}

func (m tuiModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
	case " ", "enter":
		if len(m.tasks) > 0 {
			m.tasks[m.cursor].Done = !m.tasks[m.cursor].Done
			if err := store.Save(m.tasks); err != nil {
				m.err = err
			}
		}
	case "a":
		m.mode = modeInput
		m.input = ""
	case "d":
		if len(m.tasks) > 0 {
			m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...)
			if m.cursor > 0 && m.cursor >= len(m.tasks) {
				m.cursor = len(m.tasks) - 1
			}
			if err := store.Save(m.tasks); err != nil {
				m.err = err
			}
		}
	}
	return m, nil
}

func (m tuiModel) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = modeNormal
		m.input = ""
	case "enter":
		title := strings.TrimSpace(m.input)
		if title != "" {
			t := model.Task{
				ID:        store.NextID(m.tasks),
				Title:     title,
				Done:      false,
				CreatedAt: time.Now(),
			}
			m.tasks = append(m.tasks, t)
			if err := store.Save(m.tasks); err != nil {
				m.err = err
			}
			m.cursor = len(m.tasks) - 1
		}
		m.mode = modeNormal
		m.input = ""
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m tuiModel) View() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Alfred") + "\n\n")

	if len(m.tasks) == 0 && m.mode == modeNormal {
		sb.WriteString(helpStyle.Render("No tasks yet. Press 'a' to add one.") + "\n")
	}

	for i, t := range m.tasks {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}

		checkbox := "[ ]"
		title := itemStyle.Render(t.Title)
		if t.Done {
			checkbox = "[x]"
			title = doneStyle2.Render(t.Title)
		}

		sb.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, title))
	}

	if m.mode == modeInput {
		sb.WriteString("\n" + inputStyle.Render("New task: "+m.input+"_") + "\n")
		sb.WriteString(helpStyle.Render("enter: save • esc: cancel") + "\n")
	} else {
		sb.WriteString("\n" + helpStyle.Render("a: add • d: delete • space/enter: toggle • q: quit") + "\n")
	}

	if m.err != nil {
		sb.WriteString(errStyle.Render("Error: "+m.err.Error()) + "\n")
	}

	return sb.String()
}

// ── Entry point ───────────────────────────────────────────────────────────────

func runTUI() error {
	tasks, err := store.Load()
	if err != nil {
		return err
	}

	m := tuiModel{tasks: tasks}
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
