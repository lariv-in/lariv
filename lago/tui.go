package lago

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lariv-in/lago/registry"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"gorm.io/gorm"
)

type focus int

const (
	focusSidebar focus = iota
	focusTable
	focusForm
	focusCreateForm
	focusImport
	focusExport
)

const sidebarWidth = 20

type model struct {
	currentTab int
	tabs       []registry.Pair[string, AdminPanelInterface]
	db         *gorm.DB
	rows       []map[string]any
	columns    []string
	focus      focus
	currentRow int
	form       *huh.Form
	formValues map[string]*string
	importPath *string
	exportPath *string
	formErr    string
	width      int
	height     int
	tabScroll  int
	rowScroll  int
}

func initialModel(db *gorm.DB) model {
	stable := RegistryAdmin.AllStable(registry.AlphabeticalByKey[AdminPanelInterface]{})
	m := model{
		currentTab: 0,
		tabs:       *stable,
		db:         db,
		width:      80,
		height:     24,
	}
	m.loadRows()
	return m
}

func (m *model) loadRows() {
	m.currentRow = 0
	m.rowScroll = 0
	if len(m.tabs) == 0 || m.db == nil {
		m.rows = nil
		m.columns = nil
		return
	}
	panel := m.tabs[m.currentTab].Value
	rows, err := panel.List(m.db, 1, 20)
	if err != nil || len(rows) == 0 {
		m.rows = nil
		m.columns = nil
		return
	}
	m.rows = rows

	if fields := panel.GetListFields(); len(fields) > 0 {
		m.columns = fields
	} else {
		cols := make([]string, 0, len(rows[0]))
		for k := range rows[0] {
			cols = append(cols, k)
		}
		sort.Strings(cols)
		m.columns = cols
	}
}

func (m *model) buildForm() {
	row := m.rows[m.currentRow]
	m.formValues = make(map[string]*string, len(row))

	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fields := make([]huh.Field, 0, len(keys))
	for _, k := range keys {
		val := fmt.Sprintf("%v", row[k])
		m.formValues[k] = &val
		fields = append(fields, huh.NewInput().Title(k).Value(m.formValues[k]))
	}

	m.form = huh.NewForm(huh.NewGroup(fields...)).
		WithWidth(m.contentWidth() - 4).
		WithHeight(m.contentHeight())
}

func (m *model) buildCreateForm() {
	panel := m.tabs[m.currentTab].Value
	editableFields := panel.EditableFields()
	m.formValues = make(map[string]*string, len(editableFields))

	fields := make([]huh.Field, 0, len(editableFields))
	for _, k := range editableFields {
		val := ""
		m.formValues[k] = &val
		fields = append(fields, huh.NewInput().Title(k).Value(m.formValues[k]))
	}

	m.form = huh.NewForm(huh.NewGroup(fields...)).
		WithWidth(m.contentWidth() - 4).
		WithHeight(m.contentHeight())
}

func (m *model) buildImportForm() {
	path := ""
	m.importPath = &path
	m.form = huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("CSV File Path").Value(m.importPath),
	)).
		WithWidth(m.contentWidth() - 4).
		WithHeight(m.contentHeight())
}

func (m *model) buildExportForm() {
	path := ""
	m.exportPath = &path
	m.form = huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Export CSV File Path").Value(m.exportPath),
	)).
		WithWidth(m.contentWidth() - 4).
		WithHeight(m.contentHeight())
}

func (m model) contentWidth() int {
	return m.width - sidebarWidth
}

// contentHeight returns available lines for content (excluding help line and borders).
func (m model) contentHeight() int {
	return m.height - 3
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) nextTab() model {
	if len(m.tabs) == 0 {
		return m
	}
	m.currentTab = (m.currentTab + 1) % len(m.tabs)
	m.ensureTabVisible()
	m.loadRows()
	return m
}

func (m model) prevTab() model {
	if len(m.tabs) == 0 {
		return m
	}
	m.currentTab = (m.currentTab - 1 + len(m.tabs)) % len(m.tabs)
	m.ensureTabVisible()
	m.loadRows()
	return m
}

func (m *model) ensureTabVisible() {
	visible := max(m.contentHeight(), 1)
	if m.currentTab < m.tabScroll {
		m.tabScroll = m.currentTab
	} else if m.currentTab >= m.tabScroll+visible {
		m.tabScroll = m.currentTab - visible + 1
	}
}

func (m *model) ensureRowVisible() {
	// 2 lines for header+separator, 2 for title+blank
	visible := max(m.contentHeight()-4, 1)
	if m.currentRow < m.rowScroll {
		m.rowScroll = m.currentRow
	} else if m.currentRow >= m.rowScroll+visible {
		m.rowScroll = m.currentRow - visible + 1
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureTabVisible()
		m.ensureRowVisible()
		return m, nil

	case tea.KeyPressMsg:
		switch m.focus {
		case focusSidebar:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "j", "tab":
				m = m.nextTab()
			case "k", "shift+tab":
				m = m.prevTab()
			case "enter", "l":
				if len(m.rows) > 0 {
					m.focus = focusTable
					m.currentRow = 0
					m.rowScroll = 0
				}
			}
		case focusTable:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "j":
				if m.currentRow < len(m.rows)-1 {
					m.currentRow++
					m.ensureRowVisible()
				}
			case "k":
				if m.currentRow > 0 {
					m.currentRow--
					m.ensureRowVisible()
				}
			case "escape", "h":
				m.focus = focusSidebar
			case "enter":
				m.buildForm()
				m.focus = focusForm
				return m, m.form.Init()
			case "n":
				m.buildCreateForm()
				m.focus = focusCreateForm
				return m, m.form.Init()
			case "i":
				m.buildImportForm()
				m.focus = focusImport
				return m, m.form.Init()
			case "e":
				m.buildExportForm()
				m.focus = focusExport
				return m, m.form.Init()
			}
		case focusForm:
			if msg.String() == "escape" {
				m.form = nil
				m.formValues = nil
				m.focus = focusTable
				return m, nil
			}
		case focusCreateForm, focusImport, focusExport:
			if msg.String() == "escape" {
				m.form = nil
				m.formValues = nil
				m.importPath = nil
				m.exportPath = nil
				m.focus = focusTable
				return m, nil
			}
		}
	}

	if m.form != nil && (m.focus == focusForm || m.focus == focusCreateForm || m.focus == focusImport || m.focus == focusExport) {
		f, cmd := m.form.Update(msg)
		m.form = f.(*huh.Form)
		if m.form.State == huh.StateCompleted {
			panel := m.tabs[m.currentTab].Value
			var err error
			switch m.focus {
			case focusForm:
				id := fmt.Sprintf("%v", m.rows[m.currentRow]["ID"])
				err = panel.Save(m.db, id, m.formValues)
			case focusCreateForm:
				err = panel.Create(m.db, m.formValues)
			case focusImport:
				var count int
				count, err = panel.ImportCSV(m.db, *m.importPath)
				if err == nil {
					m.formErr = fmt.Sprintf("Imported %d rows", count)
				}
			case focusExport:
				var count int
				count, err = panel.ExportCSV(m.db, *m.exportPath)
				if err == nil {
					m.formErr = fmt.Sprintf("Exported %d rows to %s", count, *m.exportPath)
				}
			}
			m.form = nil
			m.formValues = nil
			m.importPath = nil
			m.exportPath = nil
			if err != nil {
				m.formErr = err.Error()
			} else if m.focus != focusImport && m.focus != focusExport {
				m.formErr = ""
			}
			m.loadRows()
			m.focus = focusTable
			return m, nil
		}
		return m, cmd
	}

	return m, nil
}

var (
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			PaddingLeft(1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				PaddingLeft(1)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))
)

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return s[:max-1] + "…"
}

func (m model) renderSidebar() string {
	visible := max(m.contentHeight(), 1)

	end := min(m.tabScroll+visible, len(m.tabs))

	var lines []string
	for i := m.tabScroll; i < end; i++ {
		label := truncate(m.tabs[i].Key, sidebarWidth-4)
		if i == m.currentTab {
			lines = append(lines, activeTabStyle.Render("▸ "+label))
		} else {
			lines = append(lines, inactiveTabStyle.Render("  "+label))
		}
	}

	// Scroll indicators
	if m.tabScroll > 0 {
		lines = append([]string{inactiveTabStyle.Render("  ▲")}, lines...)
	}
	if end < len(m.tabs) {
		lines = append(lines, inactiveTabStyle.Render("  ▼"))
	}

	sidebar := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().Width(sidebarWidth).Render(sidebar)
}

func (m model) renderTable() string {
	currentPanel := m.tabs[m.currentTab].Value
	title := fmt.Sprintf("Model: %s", currentPanel.ModelName())

	if len(m.rows) == 0 {
		return title + "\n\nNo records found."
	}

	contentW := m.contentWidth() - 4 // padding/border

	// Calculate column widths
	widths := make(map[string]int, len(m.columns))
	for _, col := range m.columns {
		widths[col] = len(col)
	}
	for _, row := range m.rows {
		for _, col := range m.columns {
			val := fmt.Sprintf("%v", row[col])
			if len(val) > widths[col] {
				widths[col] = len(val)
			}
		}
	}

	// Shrink columns to fit width if needed
	separatorCost := (len(m.columns) - 1) * 3 // " │ "
	totalWidth := separatorCost
	for _, col := range m.columns {
		totalWidth += widths[col]
	}
	if totalWidth > contentW && contentW > separatorCost {
		available := contentW - separatorCost
		for _, col := range m.columns {
			max := available / len(m.columns)
			if max < 3 {
				max = 3
			}
			if widths[col] > max {
				widths[col] = max
			}
		}
	}

	// Header
	var header []string
	var separator []string
	for _, col := range m.columns {
		header = append(header, fmt.Sprintf("%-*s", widths[col], truncate(col, widths[col])))
		separator = append(separator, strings.Repeat("─", widths[col]))
	}

	var b strings.Builder
	b.WriteString(title + "\n\n")
	b.WriteString(strings.Join(header, " │ ") + "\n")
	b.WriteString(strings.Join(separator, "─┼─") + "\n")

	// Visible rows
	visibleRows := max(
		// title, blank, header, separator
		m.contentHeight()-4, 1)
	end := min(m.rowScroll+visibleRows, len(m.rows))

	if m.rowScroll > 0 {
		b.WriteString("  ▲\n")
	}

	for i := m.rowScroll; i < end; i++ {
		row := m.rows[i]
		var cells []string
		for _, col := range m.columns {
			val := fmt.Sprintf("%v", row[col])
			cells = append(cells, fmt.Sprintf("%-*s", widths[col], truncate(val, widths[col])))
		}
		line := strings.Join(cells, " │ ")
		if m.focus == focusTable && i == m.currentRow {
			line = selectedRowStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	if end < len(m.rows) {
		b.WriteString("  ▼\n")
	}

	return b.String()
}

func (m model) View() tea.View {
	if len(m.tabs) == 0 {
		return tea.NewView("No admin panels registered.\n")
	}

	sidebar := m.renderSidebar()

	currentPanel := m.tabs[m.currentTab].Value
	var content string
	switch {
	case m.focus == focusForm && m.form != nil:
		content = fmt.Sprintf("Editing: %s\n\n", currentPanel.ModelName()) + m.form.View()
	case m.focus == focusCreateForm && m.form != nil:
		content = fmt.Sprintf("New: %s\n\n", currentPanel.ModelName()) + m.form.View()
	case m.focus == focusImport && m.form != nil:
		content = fmt.Sprintf("Import CSV: %s\n\n", currentPanel.ModelName()) + m.form.View()
	case m.focus == focusExport && m.form != nil:
		content = fmt.Sprintf("Export CSV: %s\n\n", currentPanel.ModelName()) + m.form.View()
	default:
		content = m.renderTable()
	}

	contentStyled := lipgloss.NewStyle().
		Width(m.contentWidth()).
		Height(m.contentHeight()).
		PaddingLeft(2).
		Render(content)

	layout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, contentStyled)

	var help string
	switch m.focus {
	case focusSidebar:
		help = "j/k: navigate tabs • enter/l: focus table • q: quit"
	case focusTable:
		if m.formErr != "" {
			help = "Error: " + m.formErr + "\n  "
		}
		help += "j/k: navigate rows • enter: edit • n: new • i: import • e: export • esc/h: back • q: quit"
	case focusForm, focusCreateForm, focusImport, focusExport:
		help = "enter: submit • esc: cancel"
	}

	return tea.NewView(layout + "\n" + help)
}
