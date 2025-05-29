package main

import (
	"fmt"
	"os"

	"slices"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO: this is for easier dev, will be changed
const ConfigDir = "./config"

var (
	focusedColStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			Padding(1).
			BorderForeground(lipgloss.Color("69"))

	blurredColStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			Padding(1).
			BorderForeground(lipgloss.Color("253"))

	selectedVarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")). // green
				Bold(true)

	unselectedVarStyle = lipgloss.NewStyle().
				Faint(true) // dim gray
)

type model struct {
	Vars              Vars
	Collections       Collections
	varsList          list.Model
	colList           list.Model
	filePicker        filepicker.Model
	focusedList       int // 0 = vars, 1 = collections, 2 = file picker
	focusedCollection int
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func UserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting user home directory: %v\n", err)
		return ""
	}
	return homeDir
}

func newModel(vars Vars, cols Collections) model {
	// build vars list
	var varItems []list.Item
	for _, v := range vars.All() {
		varItems = append(varItems, item{title: v.Key, desc: v.Val})
	}
	vl := list.New(varItems, list.NewDefaultDelegate(), 30, 14)
	vl.Title = "Vars"

	// build collections list
	var colItems []list.Item
	for _, c := range cols {
		colItems = append(colItems, item{title: c.Name, desc: c.Filename})
	}
	cl := list.New(colItems, list.NewDefaultDelegate(), 30, 14)
	cl.Title = "Collections"

	// set up the file picker
	fp := filepicker.New()
	fp.ShowPermissions = false
	fp.ShowSize = false
	fp.ShowHidden = true
	fp.CurrentDirectory = UserHomeDir()
	fp.SetHeight(14)

	return model{
		Vars:        vars,
		Collections: cols,
		varsList:    vl,
		colList:     cl,
		filePicker:  fp,
		focusedList: 0,
	}
}

func (m model) Init() tea.Cmd {
	// This is correct - keep this to start loading files
	return m.filePicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		w, h := ws.Width, ws.Height
		colWidth := w/3 - 2

		m.varsList.SetSize(colWidth, h-5)
		m.colList.SetSize(colWidth, h-5)
		m.filePicker.SetHeight(h - 8)

		focusedColStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			Padding(1).
			BorderForeground(lipgloss.Color("69")).
			Width(colWidth)

		blurredColStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			Padding(1).
			BorderForeground(lipgloss.Color("253")).
			Width(colWidth)

		_, cmd := m.filePicker.Update(ws)
		m.applyVarHighlights(m.focusedCollection)
		cmds = append(cmds, cmd)
	} else if key, ok := msg.(tea.KeyMsg); ok {
		// Handle global keys first
		switch key.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "shift+tab":
			m.focusedList = (m.focusedList + 1) % 3
			return m, nil
		}

		if m.focusedList == 2 {
			// Add your "w" key handling here
			if key.String() == "w" {
				c := m.Collections[m.colList.Index()]
				c.WriteToEnvFile(m.filePicker.CurrentDirectory, c.Filename)
			}

			if key.String() == "s" {
				err := m.Collections[m.colList.Index()].WriteToSymlink(m.filePicker.CurrentDirectory)
				if err != nil {
					fmt.Printf("Error creating symlink: %v\n", err)
				}
			}

			var cmd tea.Cmd
			m.filePicker, cmd = m.filePicker.Update(key)
			cmds = append(cmds, cmd)
		}
	} else {
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)
	}

	var cmdSub tea.Cmd
	switch m.focusedList {
	case 0:
		m.varsList, cmdSub = m.varsList.Update(msg)

		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			vid := m.varsList.Index()
			ci := m.focusedCollection
			ids := m.Collections[ci].GetVarIds()
			found := slices.Contains(ids, vid)
			if found {
				m.Collections[ci].RemoveVar(vid) // you'll need a RemoveVar method
				m.Collections.Save()
			} else {
				m.Collections[ci].AddVar(vid)
				m.Collections.Save()
			}
			m = m.applyVarHighlights(m.focusedCollection)
		}
	case 1:
		m.colList, cmdSub = m.colList.Update(msg)

		m.focusedCollection = m.colList.Index()
		m = m.applyVarHighlights(m.focusedCollection)
	case 2:
	}

	if cmdSub != nil {
		cmds = append(cmds, cmdSub)
	}

	return m, tea.Batch(cmds...)
}

func (m model) applyVarHighlights(ci int) model {
	raw := m.Vars.All()
	ids := m.Collections[ci].GetVarIds()
	sel := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		sel[id] = struct{}{}
	}

	for i := range raw {
		var style lipgloss.Style
		if _, ok := sel[i]; ok {
			style = selectedVarStyle
		} else {
			style = unselectedVarStyle
		}

		m.varsList.SetItem(i, item{
			title: style.Render(raw[i].Key),
			desc:  raw[i].Val,
		})
	}
	return m
}

func (m model) View() string {
	var left, middle, right string

	// Create a wrapper style that truncates content
	wrapStyle := lipgloss.NewStyle().MaxWidth(m.varsList.Width() - 4)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MarginTop(1)
	helpText := "w: write file | s: write symlink"

	if m.focusedList == 0 {
		left = focusedColStyle.Render(m.varsList.View())
		middle = blurredColStyle.Render(m.colList.View())
		rightContent := wrapStyle.Render(m.filePicker.View()) + "\n" + helpStyle.Render(helpText)
		right = blurredColStyle.Render(wrapStyle.Render(rightContent))
	} else if m.focusedList == 1 {
		left = blurredColStyle.Render(m.varsList.View())
		middle = focusedColStyle.Render(m.colList.View())
		rightContent := wrapStyle.Render(m.filePicker.View()) + "\n" + helpStyle.Render(helpText)
		right = blurredColStyle.Render(wrapStyle.Render(rightContent))
	} else {
		left = blurredColStyle.Render(m.varsList.View())
		middle = blurredColStyle.Render(m.colList.View())
		rightContent := wrapStyle.Render(m.filePicker.View()) + "\n" + helpStyle.Render(helpText)
		right = focusedColStyle.Render(rightContent)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)
}

// main is the entry point of the program.
func main() {
	// Create mock Vars and Collections
	vars := NewVars()
	vars.Add("AWS_ACCOUNT", "123456789", "dev environment")
	vars.Add("DB_HOST", "db-prod.example.com", "Primary database host")
	vars.Add("DB_PORT", "5432", "Database port number")
	vars.Add("DB_USER", "prod_user", "Database username")
	vars.Add("DB_PASSWORD", "prod_password", "Production database password")
	vars.Add("DB_NAME", "prod_db", "Production database name")
	vars.Add("AWS_ACCOUNT", "234567890", "Production AWS account ID")
	vars.Add("S3_BUCKET", "myapp-prod-bucket", "Primary S3 bucket name")
	vars.Add("REDIS_URL", "redis://cache.prod:6379", "Redis cache connection URL")
	vars.Add("LOG_LEVEL", "info", "Application log verbosity level")
	vars.Add("TELEMETRY_URL", "https://telemetry.example.com", "Telemetry endpoint URL")
	vars.Add("TELEMETRY_TOKEN", "abc123", "Telemetry authentication token")
	vars.Add("TELEMETRY_ENABLED", "true", "Enable telemetry collection")
	vars.Add("TELEMETRY_INTERVAL", "60", "Telemetry data collection interval in seconds")
	vars.Add("TELEMETRY_DEBUG", "false", "Enable debug mode for telemetry")
	vars.Add("TELEMETRY_LOG_LEVEL", "debug", "Log level for telemetry data")
	vars.Add("TELEMETRY_LOG_FILE", "/var/log/telemetry.log", "File path for telemetry logs")

	collections := Collections{
		NewCollection("Dev_Env", ".env"),
		NewCollection("Staging_Env", ".env.staging"),
		NewCollection("Production_Env", ".env"),
		NewCollection("Dev_PostgreSQL", ".env"),
		NewCollection("Monitoring_Logging", ".env"),
	}
	// Add Vars to Collections
	collections[0].AddVar(0)
	collections[0].AddVar(1)
	collections[1].AddVar(2)
	collections[1].AddVar(3)
	collections[2].AddVar(0)
	collections[2].AddVar(2)
	collections[3].AddVar(1)
	collections[3].AddVar(3)
	collections[3].AddVar(4)
	collections[3].AddVar(6)
	collections[3].AddVar(9)
	collections[4].AddVar(5)
	collections[4].AddVar(7)
	collections[4].AddVar(8)
	collections[4].AddVar(9)
	collections[4].AddVar(10)
	collections[4].AddVar(11)
	collections[4].AddVar(12)
	collections[4].AddVar(13)
	collections[4].AddVar(14)
	collections[4].AddVar(15)

	// Start Bubble Tea with both lists
	p := tea.NewProgram(newModel(vars, collections))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}

	// On quit, persist to disk
	if err := vars.Save(); err != nil {
		fmt.Printf("Error saving vars: %v\n", err)
		os.Exit(1)
	}
	if err := collections.Save(); err != nil {
		fmt.Printf("Error saving collections: %v\n", err)
		os.Exit(1)
	}
}
