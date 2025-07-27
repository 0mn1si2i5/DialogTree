// Path: ./service/tea_service/placeholder_dialog.go

package tea_service

import tea "github.com/charmbracelet/bubbletea"

type DialogModel struct {
	lines  []string
	goBack bool
}

func NewDialogModel(convID int64) DialogModel {
	return DialogModel{
		lines: []string{
			"💬 You: 月亮是什么？",
			"🤖 AI: 月亮是地球的天然卫星。",
		},
	}
}

func (m DialogModel) Update(msg tea.Msg) (DialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "backspace" {
			m.goBack = true
		}
	}
	return m, nil
}

func (m DialogModel) View() string {
	view := "对话内容：\n\n"
	for _, line := range m.lines {
		view += line + "\n"
	}
	view += "\n[按 ESC 返回]"
	return view
}

func (m DialogModel) ShouldGoBack() bool { return m.goBack }
