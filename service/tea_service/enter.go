// Path: ./service/tea_service/enter.go

package tea_service

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ViewState int

const (
	SessionListView ViewState = iota
	ConversationListView
	DialogView
)

type MainModel struct {
	state       ViewState
	sessionList SessionListModel
	convList    ConversationListModel
	dialog      DialogModel
	selectedID  int64 // 选中 Session 或 Dialog 的 ID
}

func NewMainModel() MainModel {
	return MainModel{
		state:       SessionListView,
		sessionList: NewSessionListModel(),
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case SessionListView:
		newSessionModel, cmd := m.sessionList.Update(msg)
		m.sessionList = newSessionModel

		if m.sessionList.ShouldEnter() {
			m.selectedID = m.sessionList.SelectedID()
			m.convList = NewConvListModel(m.selectedID)
			m.state = ConversationListView
		}
		return m, cmd

	case ConversationListView:
		newConvModel, cmd := m.convList.Update(msg)
		m.convList = newConvModel

		if m.convList.ShouldEnter() {
			m.selectedID = m.convList.SelectedID()
			m.dialog = NewDialogModel(m.selectedID)
			m.state = DialogView
		}
		if m.convList.ShouldGoBack() {
			m.state = SessionListView
		}
		return m, cmd

	case DialogView:
		newDialogModel, cmd := m.dialog.Update(msg)
		m.dialog = newDialogModel

		if m.dialog.ShouldGoBack() {
			m.state = ConversationListView
		}
		return m, cmd
	}

	return m, nil
}

func (m MainModel) View() string {
	switch m.state {
	case SessionListView:
		return m.sessionList.View()
	case ConversationListView:
		return m.convList.View()
	case DialogView:
		return m.dialog.View()
	default:
		return "未知视图"
	}
}
