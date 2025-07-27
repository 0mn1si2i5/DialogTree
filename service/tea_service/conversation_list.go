// Path: ./service/tea_service/placeholder_conv.go

package tea_service

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ConversationListModel struct {
	list     list.Model
	enter    bool
	goBack   bool
	selected int64
}

type listItem struct {
	title string
	desc  string
}

func NewConvListModel(sessionID int64) ConversationListModel {
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Conversations"
	return ConversationListModel{list: l}
}

func (m ConversationListModel) Update(msg tea.Msg) (ConversationListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.enter = true
			m.selected = int64(m.list.Index())
		case "esc", "backspace":
			m.goBack = true
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ConversationListModel) View() string {
	return docStyle.Render(m.list.View())
}

func (m ConversationListModel) ShouldEnter() bool  { return m.enter }
func (m ConversationListModel) ShouldGoBack() bool { return m.goBack }
func (m ConversationListModel) SelectedID() int64  { return m.selected }
