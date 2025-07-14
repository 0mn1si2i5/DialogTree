// Path: ./service/tea_service/list.go

package tea_service

import (
	"dialogTree/global"
	"dialogTree/models"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type Model struct {
	List list.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return docStyle.Render(m.List.View())
}

func teaList(title string, itemList []list.Item) *tea.Program {
	var items []list.Item
	for _, item := range itemList {
		items = append(items, item)
	}

	m := Model{List: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.List.Title = title

	p := tea.NewProgram(m, tea.WithAltScreen())

	return p
}

func ShowAllSessions() (p *tea.Program, err error) {
	var sessionList []models.SessionModel
	err = global.DB.Preload("CategoryModel").Order("updated_at DESC").Find(&sessionList).Error
	if err != nil {
		return
	}
	var tlist []list.Item
	for _, session := range sessionList {
		var d string
		if session.CategoryModel == nil {
			d = fmt.Sprintf("%s|%s", session.CreatedAt.Format("01-02 15:04"), session.Summary)
		} else {
			d = fmt.Sprintf("%s|%s|%s", session.UpdatedAt.Format("01-02 15:04"), session.CategoryModel.Name, session.Summary)
		}
		itm := item{
			title: fmt.Sprintf("%03d.%s", session.ID, session.Tittle),
			desc:  d,
		}
		tlist = append(tlist, itm)
	}
	p = teaList("会话列表", tlist)
	return
}
