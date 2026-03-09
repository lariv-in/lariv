package lago

import (
	"log"

	tea "charm.land/bubbletea/v2"
		"charm.land/lipgloss/v2"
)

type Model struct {
	ActiveTab int
	Styles *Styles
}


type Styles struct {
	doc         lipgloss.Style
	highlight   lipgloss.Style
	inactiveTab lipgloss.Style
	activeTab   lipgloss.Style
	window      lipgloss.Style
}

func NewModel() Model {
	return Model {
		ActiveTab: 0,
		Styles: newStyles(true),
	}
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func newStyles(bgIsDark bool) *Styles {
	lightDark := lipgloss.LightDark(bgIsDark)

	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")
	highlightColor := lightDark(lipgloss.Color("#874BFD"), lipgloss.Color("#7D56F4"))

	s := new(Styles)
	s.doc = lipgloss.NewStyle().
		Padding(1, 2, 1, 2)
	s.inactiveTab = lipgloss.NewStyle().
		Border(inactiveTabBorder, true).
		BorderForeground(highlightColor).
		Padding(0, 1)
	s.activeTab = s.inactiveTab.
		Border(activeTabBorder, true)
	s.window = lipgloss.NewStyle().
		BorderForeground(highlightColor).
		Padding(2, 0).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop()
	return s
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l", "n", "tab":
			m.ActiveTab = min(m.ActiveTab+1, len(tabs)-1)
			return m, nil
		case "left", "h", "p", "shift+tab":
			m.ActiveTab = max(m.ActiveTab-1, 0)
			return m, nil
		}
	}
	return m, nil
}


func (m Model) View() tea.View {
	// TODO: Render the tabs
	return tea.NewView("")
}



func RunTui() {
	model := NewModel()
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Panicf("Error running tui: %e", err)
	}
}

var tabs []tea.Model = []tea.Model{}
var tabNames []string = []string {}

func AddTab(name string, model tea.Model) {
	tabs = append(tabs, model)
	tabNames = append(tabNames, name)
}
