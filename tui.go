package main

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type model struct {
	width      int
	height     int
	lastUpdate time.Time

	processTable table.Model
	tableStyle   table.Styles
	baseStyle    lipgloss.Style
	viewStyle    lipgloss.Style

	CpuUsage cpu.TimesStat
	MemUsage mem.VirtualMemoryStat
}

type TickMsg time.Time

type Theme struct {
	Primary   lipgloss.AdaptiveColor
	Secondary lipgloss.AdaptiveColor
	Highlight lipgloss.AdaptiveColor
	Border    lipgloss.AdaptiveColor
	Green     lipgloss.AdaptiveColor
	Red       lipgloss.AdaptiveColor
}

var Color = Theme{
	Primary:   lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"},
	Secondary: lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"},
	Highlight: lipgloss.AdaptiveColor{Light: "#8b2def", Dark: "#8b2def"},
	Border:    lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"},
	Green:     lipgloss.AdaptiveColor{Light: "#00FF00", Dark: "#00FF00"},
	Red:       lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"},
}

// Calls the tickEvery function to set up a command that sends a TickMsg every second.
// This command will be executed immediately when the program starts, initiating the periodic updates.
func (m model) Init() tea.Cmd {
	return tickEvery()
}

func tickEvery() tea.Cmd {
	// tea.Every function is a helper function from the Bubble Tea framework
	// that schedules a command to run at regular intervals.
	return tea.Every(time.Second,
		// Callback function that takes the current time (t time.Time) as a parameter and returns a message (tea.Msg).
		// This callback is invoked every second.
		func(t time.Time) tea.Msg {
			return TickMsg(t)
		})
}

func (m model) View() string {
	// Sets the width of the column to the width of the terminal (m.width) and adds padding of 1 unit on the top.
	// Render is a method from the lipgloss package that applies the defined style and returns a function that can render styled content.
	column := m.baseStyle.Width(m.width).Padding(1, 0, 0, 0).Render
	// Set the content to match the terminal dimensions (m.width and m.height).
	content := m.baseStyle.
		Width(m.width).
		Height(m.height).
		Render(
			// Vertically join multiple elements aligned to the left.
			lipgloss.JoinVertical(lipgloss.Left,
				column(m.viewHeader()),
				column(m.viewProcess()),
			),
		)

	return content
}

// Takes a tea.Msg as input and uses a type switch to handle different types of messages.
// Each case in the switch statement corresponds to a specific message type.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// message is sent when the window size changes
	// save to reflect the new dimensions of the terminal window.
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	// message is sent when a key is pressed.
	case tea.KeyMsg:
		switch msg.String() {
		// Toggles the focus state of the process table
		case "esc":
			if m.processTable.Focused() {
				m.tableStyle.Selected = m.baseStyle
				m.processTable.SetStyles(m.tableStyle)
				m.processTable.Blur()
			} else {
				m.tableStyle.Selected = m.tableStyle.Selected.Background(Color.Highlight)
				m.processTable.SetStyles(m.tableStyle)
				m.processTable.Focus()
			}
		// Moves the focus up in the process table if the table is focused.
		case "up", "k":
			if m.processTable.Focused() {
				m.processTable.MoveUp(1)
			}
		// Moves the focus down in the process table if the table is focused.
		case "down", "j":
			if m.processTable.Focused() {
				m.processTable.MoveDown(1)
			}
		// Quits the program by returning the tea.Quit command.
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	// This custom message is sent periodically by the tickEvery function.
	// The model's lastUpdate field is updated to the current time.
	// Fetching CPU Stats, Memory Stats & Processes
	// Returning Command: The tickEvery command is returned to ensure that the TickMsg continues to be sent periodically.
	case TickMsg:
		m.lastUpdate = time.Time(msg)
		cpuStats, err := GetCPUStats()
		if err != nil {
			slog.Error("Could not get CPU info", "error", err)
		} else {
			m.CpuUsage = cpuStats
		}

		memStats, err := GetMEMStats()
		if err != nil {
			slog.Error("Could not get memory info", "error", err)
		} else {
			m.MemUsage = memStats
		}

		procs, err := GetProcesses(5)
		if err != nil {
			slog.Error("Could not get processes", "error", err)
		} else {
			rows := []table.Row{}
			for _, p := range procs {
				memString, memUnit := convertBytes(p.Memory)
				rows = append(rows, table.Row{
					fmt.Sprintf("%d", p.PID),
					p.Name,
					fmt.Sprintf("%.2f%%", p.CPUPercent),
					fmt.Sprintf("%s %s", memString, memUnit),
					p.Username,
					p.RunningTime,
				})
			}
			m.processTable.SetRows(rows)
		}

		return m, tickEvery()
	}
	// If the message type does not match any of the handled cases, the model is returned unchanged, and no new command is issued.
	return m, nil
}

// Uses lipgloss.JoinVertical and lipgloss.JoinHorizontal to arrange the header content.
// It displays the last update time and various system statistics (CPU and memory usage) in a structured format.
func (m model) viewHeader() string {
	// defines the style for list items, including borders, border color, height, and padding.
	list := m.baseStyle.
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(Color.Border).
		Height(4).
		Padding(0, 1)

	// applies bold styling to the text.
	listHeader := m.baseStyle.Bold(true).Render

	// helper function that formats a key-value pair with an optional suffix. It aligns the value to the right and renders it with the specified style.
	listItem := func(key string, value string, suffix ...string) string {
		finalSuffix := ""
		if len(suffix) > 0 {
			finalSuffix = suffix[0]
		}

		listItemValue := m.baseStyle.Align(lipgloss.Right).Render(fmt.Sprintf("%s%s", value, finalSuffix))

		listItemKey := func(key string) string {
			return m.baseStyle.Render(key + ":")
		}

		return fmt.Sprintf("%s %s", listItemKey(key), listItemValue)
	}

	var lastUpdate string
	if m.lastUpdate.IsZero() {
		lastUpdate = "updating..."
	} else {
		duration := time.Since(m.lastUpdate)
		if duration < time.Second {
			lastUpdate = fmt.Sprintf("%dms ago", duration.Milliseconds())
		} else {
			lastUpdate = fmt.Sprintf("%.1fs ago", duration.Seconds())
		}
	}

	return m.viewStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			fmt.Sprintf("Last update: %s\n", lastUpdate),
			lipgloss.JoinHorizontal(lipgloss.Top,
				// Progress Bars
				list.Render(
					lipgloss.JoinVertical(lipgloss.Left,
						listHeader("% Usage"),
						listItem("CPU", fmt.Sprintf("%s %.1f", progressBar(100-m.CpuUsage.Idle, m.baseStyle), 100-m.CpuUsage.Idle), "%"),
						listItem("MEM", fmt.Sprintf("%s %.1f", progressBar(m.MemUsage.UsedPercent, m.baseStyle), m.MemUsage.UsedPercent), "%"),
					),
				),

				// CPU
				list.Border(lipgloss.NormalBorder(), false).Render(
					lipgloss.JoinVertical(lipgloss.Left,
						listHeader("CPU"),
						listItem("user", fmt.Sprintf("%.1f", m.CpuUsage.User), "%"),
						listItem("sys", fmt.Sprintf("%.1f", m.CpuUsage.System), "%"),
						listItem("idle", fmt.Sprintf("%.1f", m.CpuUsage.Idle), "%"),
					),
				),
				list.Border(lipgloss.NormalBorder(), false).Render(
					lipgloss.JoinVertical(lipgloss.Left,
						listHeader(""),
						listItem("nice", fmt.Sprintf("%.1f", m.CpuUsage.Nice), "%"),
						listItem("iowait", fmt.Sprintf("%.1f", m.CpuUsage.Iowait), "%"),
						listItem("irq", fmt.Sprintf("%.1f", m.CpuUsage.Irq), "%"),
					),
				),
				list.Render(
					lipgloss.JoinVertical(lipgloss.Left,
						listHeader(""),
						listItem("softirq", fmt.Sprintf("%.1f", m.CpuUsage.Softirq), "%"),
						listItem("steal", fmt.Sprintf("%.1f", m.CpuUsage.Steal), "%"),
						listItem("guest", fmt.Sprintf("%.1f", m.CpuUsage.Guest), "%"),
					),
				),

				// MEM
				list.Border(lipgloss.NormalBorder(), false).Render(
					lipgloss.JoinVertical(lipgloss.Left,
						listHeader("MEM"),
						func() string {
							value, unit := convertBytes(m.MemUsage.Total)
							return listItem("total", value, unit)
						}(),
						func() string {
							value, unit := convertBytes(m.MemUsage.Used)
							return listItem("used", value, unit)
						}(),
						func() string {
							value, unit := convertBytes(m.MemUsage.Available)
							return listItem("free", value, unit)
						}(),
					),
				),
				list.Render(
					lipgloss.JoinVertical(lipgloss.Left,
						listHeader(""),
						func() string {
							value, unit := convertBytes(m.MemUsage.Active)
							return listItem("active", value, unit)
						}(),
						func() string {
							value, unit := convertBytes(m.MemUsage.Buffers)
							return listItem("buffers", value, unit)
						}(),
						func() string {
							value, unit := convertBytes(m.MemUsage.Cached)
							return listItem("cached", value, unit)
						}(),
					),
				),
			),
		),
	)
}

func (m model) viewProcess() string {
	return m.viewStyle.Render(m.processTable.View())
}

// creates a visual representation of a percentage as a progress bar.
func progressBar(percentage float64, baseStyle lipgloss.Style) string {
	totalBars := 20
	fillBars := int(percentage / 100 * float64(totalBars))
	// renders the filled part of the progress bar with a green color.
	filled := baseStyle.
		Foreground(Color.Green).
		Render(strings.Repeat("|", fillBars))
	// renders the empty part of the progress bar with a secondary color.
	empty := baseStyle.
		Foreground(Color.Secondary).
		Render(strings.Repeat("|", totalBars-fillBars))

	return baseStyle.Render(fmt.Sprintf("%s%s%s%s", "[", filled, empty, "]"))
}
