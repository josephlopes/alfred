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

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	pomodoroDuration = 25 * time.Minute
	tickInterval     = time.Second
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Padding(0, 1)
	cursorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	doneStyle2   = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("240"))
	itemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	inputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	catBadge     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	dueBadge     = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	filterActive = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	filterInact  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	pomodoroRun  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	pomodoroIdle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// ── Modes ─────────────────────────────────────────────────────────────────────

type tuiMode int

const (
	modeNormal tuiMode = iota
	modeInput
	modeCategoryInput
)

// ── Messages ──────────────────────────────────────────────────────────────────

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ── Model ─────────────────────────────────────────────────────────────────────

type tuiModel struct {
	tasks      []model.Task
	categories []string

	cursor      int
	catFilter   int
	visibleIdxs []int

	mode  tuiMode
	input string

	pomActive    bool
	pomTaskID    int
	pomRemaining time.Duration
	pomStartedAt time.Time

	err error
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (m tuiModel) filterLabel(i int) string {
	if i == 0 {
		return "All"
	}
	return m.categories[i-1]
}

func (m tuiModel) filterCount() int {
	return 1 + len(m.categories)
}

func (m *tuiModel) rebuildVisible() {
	m.visibleIdxs = nil
	for i, t := range m.tasks {
		if m.catFilter == 0 {
			m.visibleIdxs = append(m.visibleIdxs, i)
			continue
		}
		cat := m.filterLabel(m.catFilter)
		if strings.EqualFold(t.Category, cat) {
			m.visibleIdxs = append(m.visibleIdxs, i)
		}
	}
	if m.cursor >= len(m.visibleIdxs) {
		m.cursor = intMax(0, len(m.visibleIdxs)-1)
	}
}

func (m tuiModel) selectedTaskIdx() int {
	if len(m.visibleIdxs) == 0 || m.cursor >= len(m.visibleIdxs) {
		return -1
	}
	return m.visibleIdxs[m.cursor]
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m tuiModel) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tickMsg:
		if !m.pomActive {
			return m, nil
		}
		elapsed := time.Since(m.pomStartedAt)
		m.pomRemaining = pomodoroDuration - elapsed
		if m.pomRemaining <= 0 {
			m.pomActive = false
			m.pomRemaining = 0
			for i, t := range m.tasks {
				if t.ID == m.pomTaskID {
					m.tasks[i].Pomodoros++
					_ = store.Save(m.tasks)
					break
				}
			}
			return m, tea.Printf("\a") // terminal bell
		}
		return m, doTick()

	case tea.KeyMsg:
		switch m.mode {
		case modeNormal:
			return m.updateNormal(msg)
		case modeInput:
			return m.updateInput(msg)
		case modeCategoryInput:
			return m.updateCategoryInput(msg)
		}
	}
	return m, nil
}

// ── Normal mode ───────────────────────────────────────────────────────────────

func (m tuiModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {

	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.visibleIdxs)-1 {
			m.cursor++
		}

	case " ", "enter":
		if idx := m.selectedTaskIdx(); idx >= 0 {
			m.tasks[idx].Done = !m.tasks[idx].Done
			_ = store.Save(m.tasks)
		}

	case "a":
		m.mode = modeInput
		m.input = ""

	case "d":
		if idx := m.selectedTaskIdx(); idx >= 0 {
			m.tasks = append(m.tasks[:idx], m.tasks[idx+1:]...)
			_ = store.Save(m.tasks)
			m.rebuildVisible()
		}

	case "[", "left", "h":
		if m.catFilter > 0 {
			m.catFilter--
			m.rebuildVisible()
		}

	case "]", "right", "l":
		if m.catFilter < m.filterCount()-1 {
			m.catFilter++
			m.rebuildVisible()
		}

	case "n":
		m.mode = modeCategoryInput
		m.input = ""

	case "p":
		if idx := m.selectedTaskIdx(); idx >= 0 {
			if m.pomActive && m.pomTaskID == m.tasks[idx].ID {
				m.pomActive = false
			} else {
				m.pomActive = true
				m.pomTaskID = m.tasks[idx].ID
				m.pomStartedAt = time.Now()
				m.pomRemaining = pomodoroDuration
				return m, doTick()
			}
		}
	}
	return m, nil
}

// ── Input mode (new task) ─────────────────────────────────────────────────────

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
			cat := ""
			if m.catFilter > 0 {
				cat = m.filterLabel(m.catFilter)
			}
			t := model.Task{
				ID:        store.NextID(m.tasks),
				Title:     title,
				CreatedAt: time.Now(),
				Category:  cat,
			}
			m.tasks = append(m.tasks, t)
			_ = store.Save(m.tasks)
			m.rebuildVisible()
			m.cursor = len(m.visibleIdxs) - 1
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

// ── Category input mode ───────────────────────────────────────────────────────

func (m tuiModel) updateCategoryInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = modeNormal
		m.input = ""
	case "enter":
		name := strings.TrimSpace(m.input)
		if name != "" {
			found := false
			for _, c := range m.categories {
				if strings.EqualFold(c, name) {
					found = true
					break
				}
			}
			if !found {
				m.categories = append(m.categories, name)
				_ = store.SaveCategories(m.categories)
				m.catFilter = len(m.categories) // switch to new tab
				m.rebuildVisible()
			}
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

	// Title
	sb.WriteString(titleStyle.Render("Alfred") + "\n")

	// Category filter bar
	sb.WriteString(m.viewFilterBar())
	sb.WriteString("\n\n")

	// Task list
	if len(m.visibleIdxs) == 0 {
		sb.WriteString(helpStyle.Render("No tasks here. Press 'a' to add one.") + "\n")
	} else {
		for i, taskIdx := range m.visibleIdxs {
			t := m.tasks[taskIdx]

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

			line := fmt.Sprintf("%s%s %s", cursor, checkbox, title)
			if t.Category != "" {
				line += "  " + catBadge.Render("["+t.Category+"]")
			}
			if t.DueDay != "" {
				line += "  " + dueBadge.Render(t.DueDay)
			}
			if t.Pomodoros > 0 {
				line += fmt.Sprintf("  🍅×%d", t.Pomodoros)
			}
			sb.WriteString(line + "\n")
		}
	}

	// Pomodoro status bar
	sb.WriteString("\n" + m.viewPomodoro())

	// Input prompt / help bar
	switch m.mode {
	case modeInput:
		sb.WriteString("\n" + inputStyle.Render("New task: "+m.input+"_") + "\n")
		sb.WriteString(helpStyle.Render("enter: save  •  esc: cancel") + "\n")
	case modeCategoryInput:
		sb.WriteString("\n" + inputStyle.Render("New category: "+m.input+"_") + "\n")
		sb.WriteString(helpStyle.Render("enter: save  •  esc: cancel") + "\n")
	default:
		sb.WriteString("\n" + helpStyle.Render(
			"a: add  d: delete  space: toggle  p: pomodoro  [/]: filter  n: new category  q: quit",
		) + "\n")
	}

	if m.err != nil {
		sb.WriteString(errStyle.Render("Error: "+m.err.Error()) + "\n")
	}

	return sb.String()
}

func (m tuiModel) viewFilterBar() string {
	tabs := make([]string, m.filterCount())
	for i := range tabs {
		label := m.filterLabel(i)
		if i == m.catFilter {
			tabs[i] = filterActive.Render("[ " + label + " ]")
		} else {
			tabs[i] = filterInact.Render("  " + label + "  ")
		}
	}
	return strings.Join(tabs, "")
}

func (m tuiModel) viewPomodoro() string {
	if !m.pomActive {
		return pomodoroIdle.Render("🍅 No active pomodoro  (press 'p' on a task to start)") + "\n"
	}

	taskTitle := fmt.Sprintf("#%d", m.pomTaskID)
	for _, t := range m.tasks {
		if t.ID == m.pomTaskID {
			taskTitle = t.Title
			break
		}
	}

	mins := int(m.pomRemaining.Minutes())
	secs := int(m.pomRemaining.Seconds()) % 60
	return pomodoroRun.Render(
		fmt.Sprintf("🍅 %02d:%02d  %s  (p: stop)", mins, secs, taskTitle),
	) + "\n"
}

// ── Entry point ───────────────────────────────────────────────────────────────

func runTUI() error {
	tasks, err := store.Load()
	if err != nil {
		return err
	}

	cats, err := store.LoadCategories()
	if err != nil {
		return err
	}

	m := tuiModel{
		tasks:      tasks,
		categories: cats,
	}
	m.rebuildVisible()

	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
