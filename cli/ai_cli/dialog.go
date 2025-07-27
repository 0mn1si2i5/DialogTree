// Path: ./cli/ai_cli/dialog.go

package ai_cli

import (
	"context"
	"dialogTree/core"
	"dialogTree/service/dialog_service"
	"dialogTree/service/tea_service"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"

	"github.com/urfave/cli/v3"
)

func ShowDialogs(ctx context.Context, c *cli.Command) error {
	core.InitWithVector() // 使用带向量服务的初始化
	p, err := tea_service.ShowAllSessions()
	if err != nil {
		return err
	}
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func EnterDialog(ctx context.Context, c *cli.Command) error {
	core.InitWithVector()

	// 获取会话列表
	sessions, err := dialog_service.CliDialogServiceInstance.GetSessionList()
	if err != nil {
		fmt.Printf("获取会话列表失败: %v\n", err)
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("暂无会话，请先创建一个会话")
		return nil
	}

	// 显示会话列表
	fmt.Println("=== 会话列表 ===")
	for i, session := range sessions {
		fmt.Printf("%d. %s (摘要: %s)\n", i+1, session.Tittle, session.Summary)
	}

	// 让用户选择
	fmt.Print("请选择会话编号（输入数字）: ")
	var choice int
	_, err = fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(sessions) {
		fmt.Println("输入无效")
		return nil
	}

	selectedSession := sessions[choice-1]
	fmt.Printf("进入会话: %s\n", selectedSession.Tittle)

	// 开始对话
	return dialog_service.CliDialogServiceInstance.StartDialogChat(selectedSession.ID, nil)
}

func EnterRecent(ctx context.Context, c *cli.Command) error {
	core.InitWithVector()

	// 获取最近的会话
	session, err := dialog_service.CliDialogServiceInstance.GetRecentSession()
	if err != nil {
		fmt.Println("暂无最近会话，创建新会话...")
		session, err = dialog_service.CliDialogServiceInstance.CreateQuickSession("CLI快速会话")
		if err != nil {
			fmt.Printf("创建会话失败: %v\n", err)
			return err
		}
	}

	fmt.Printf("进入最近会话: %s\n", session.Tittle)

	// 开始对话
	return dialog_service.CliDialogServiceInstance.StartDialogChat(session.ID, nil)
}

func EnterDialogUI(ctx context.Context, c *cli.Command) error {
	core.InitWithVector()
	model := tea_service.NewMainModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Bubbletea UI error: %v", err)
	}
	return nil
}
