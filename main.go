package main

import (
	"log"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	tableStyle := table.DefaultStyles()
	tableStyle.Selected = lipgloss.NewStyle().Background(Color.Highlight)

	// Creates a new table with specified columns and initial empty rows.
	processTable := table.New(
		// We use this to define our table "header"
		table.WithColumns([]table.Column{
			{Title: "PID", Width: 10},
			{Title: "Name", Width: 25},
			{Title: "CPU", Width: 12},
			{Title: "MEM", Width: 12},
			{Title: "Username", Width: 12},
			{Title: "Time", Width: 12},
		}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(20),
		table.WithStyles(tableStyle),
	)

	m := model{
		processTable: processTable,
		tableStyle:   tableStyle,
		baseStyle:    lipgloss.NewStyle(),
		viewStyle:    lipgloss.NewStyle(),
	}

	// Create a new Bubble Tea program with the model and enable alternate screen
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program and handle any errors
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
