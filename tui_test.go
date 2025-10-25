package main

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
)

func initialModelAdvanced() model {
	tableStyle := table.DefaultStyles()
	tableStyle.Selected = lipgloss.NewStyle().Background(Color.Highlight)

	processTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "PID", Width: 10},
			{Title: "Name", Width: 25},
			{Title: "CPU", Width: 12},
			{Title: "MEM", Width: 12},
			{Title: "Username", Width: 12},
			{Title: "Time", Width: 12},
		}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false), // Start with the table unfocused
		table.WithHeight(20),
		table.WithStyles(tableStyle),
	)

	return model{
		processTable: processTable,
		tableStyle:   tableStyle,
		baseStyle:    lipgloss.NewStyle(),
		viewStyle:    lipgloss.NewStyle(),
	}
}

func TestLatestView(t *testing.T) {
	m := initialModelAdvanced()
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 200))
	time.Sleep(3 * time.Second)
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t)

	finalModel := tm.FinalModel(t).(model)
	latestView := finalModel.View()

	if len(latestView) == 0 {
		t.Errorf("latest view is empty")
	}

	if !strings.Contains(latestView, "CPU") {
		t.Errorf("output should contain 'CPU', but it doesn't")
	}
	if !strings.Contains(latestView, "MEM") {
		t.Errorf("output should contain 'MEM', but it doesn't")
	}
	if !strings.Contains(latestView, "PID") {
		t.Errorf("output should contain 'PID', but it doesn't")
	}
}

func TestTableInteraction(t *testing.T) {
	m := initialModelAdvanced()
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 200))
	time.Sleep(2 * time.Second)

	// Press 'esc' to focus the table
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	time.Sleep(1 * time.Second)

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(1 * time.Second)

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t)

	finalModel := tm.FinalModel(t).(model)
	selectedIndex := finalModel.processTable.Cursor()

	if selectedIndex != 2 {
		t.Errorf("expected selected index to be 2, but got %d", selectedIndex)
	}
}
