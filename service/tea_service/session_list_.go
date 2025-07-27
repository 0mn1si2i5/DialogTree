// Path: ./service/tea_service/session_list_.go

package tea_service

import (
	"dialogTree/global"
	"dialogTree/models"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionItem struct {
	id    int64
	title string
	desc  string
}

func (i sessionItem) Title() string       { return i.title }
func (i sessionItem) Description() string { return i.desc }
func (i sessionItem) FilterValue() string { return i.title }

type SessionListModel struct {
	list     list.Model
	selected bool
}

func NewSessionListModel() SessionListModel {
	var sessionList []models.SessionModel
	_ = global.DB.Preload("CategoryModel").Order("updated_at DESC").Find(&sessionList)

	var items []list.Item
	for _, s := range sessionList {
		desc := fmt.Sprintf("%s|%s", s.UpdatedAt.Format("01-02 15:04"), s.Summary)
		if s.CategoryModel != nil {
			desc = fmt.Sprintf("%s|%s|%s", s.UpdatedAt.Format("01-02 15:04"), s.CategoryModel.Name, s.Summary)
		}
		items = append(items, sessionItem{
			id:    s.ID,
			title: fmt.Sprintf("%03d.%s", s.ID, s.Tittle),
			desc:  desc,
		})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "会话列表"

	return SessionListModel{
		list:     l,
		selected: false,
	}
}

func (m SessionListModel) Update(msg tea.Msg) (SessionListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.selected = true
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SessionListModel) View() string {
	return docStyle.Render(m.list.View())
}

func (m SessionListModel) ShouldEnter() bool {
	return m.selected
}

func (m SessionListModel) SelectedID() int64 {
	if itm, ok := m.list.SelectedItem().(sessionItem); ok {
		return itm.id
	}
	return 0
}
