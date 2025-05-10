package main

import (
	"fmt"
	"os"

	"slices"

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
	focusedList       int // 0 = vars, 1 = collections
	focusedCollection int
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func newModel(vars Vars, cols Collections) model {
	// build vars list
	var varItems []list.Item
	for _, v := range vars.All() {
		varItems = append(varItems, item{title: v.Key, desc: v.Description})
	}
	vl := list.New(varItems, list.NewDefaultDelegate(), 30, 14)
	vl.Title = "Vars"

	// build collections list
	var colItems []list.Item
	for _, c := range cols {
		colItems = append(colItems, item{title: c.Name, desc: c.Description})
	}
	cl := list.New(colItems, list.NewDefaultDelegate(), 30, 14)
	cl.Title = "Collections"

	return model{
		Vars:        vars,
		Collections: cols,
		varsList:    vl,
		colList:     cl,
		focusedList: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
		// resize the lists to fit the window
		width, height := wsMsg.Width, wsMsg.Height
		m.varsList.SetSize(width/2-1, height-5)
		m.colList.SetSize(width/2-1, height-5)
	}
	// allow quitting or focus‐switching
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "shift+tab":
			m.focusedList = (m.focusedList + 1) % 2
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.focusedList == 0 {
		// update the vars list…
		m.varsList, cmd = m.varsList.Update(msg)

		// if Enter pressed while varsList is focused, toggle membership
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			vid := m.varsList.Index() // selected var index
			ci := m.focusedCollection // current collection index
			ids := m.Collections[ci].GetVarIds()
			// see if vid is already in this collection
			found := slices.Contains(ids, vid)
			if found {
				m.Collections[ci].RemoveVar(vid) // you'll need a RemoveVar method
				m.Collections.Save()
			} else {
				m.Collections[ci].AddVar(vid)
				m.Collections.Save()
			}
			// **re‐apply highlighting immediately**
			m = m.applyVarHighlights(m.focusedCollection)
		}
	} else {
		// update the collections list…
		m.colList, cmd = m.colList.Update(msg)

		// remember which collection is selected
		m.focusedCollection = m.colList.Index()
		// **apply highlighting when you move the collection**
		m = m.applyVarHighlights(m.focusedCollection)
	}

	return m, cmd
}

// add this helper *below* your Update:
func (m model) applyVarHighlights(ci int) model {
	raw := m.Vars.All()
	ids := m.Collections[ci].GetVarIds()
	sel := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		sel[id] = struct{}{}
	}

	// rebuild every item with a checkbox + color
	for i := range raw {
		var style lipgloss.Style
		if _, ok := sel[i]; ok {
			style = selectedVarStyle
		} else {
			style = unselectedVarStyle
		}

		m.varsList.SetItem(i, item{
			title: style.Render(raw[i].Key),
			desc:  raw[i].Description,
		})
	}
	return m
}

func (m model) View() string {
	var left, right string
	if m.focusedList == 0 {
		left = focusedColStyle.Render(m.varsList.View())
		right = blurredColStyle.Render(m.colList.View())
	} else {
		left = blurredColStyle.Render(m.varsList.View())
		right = focusedColStyle.Render(m.colList.View())
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		right,
	)
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

	collections := Collections{
		NewCollection("Development Environment", "Local development variables"),
		NewCollection("Staging Environment", "Staging server configuration"),
		NewCollection("Production Environment", "Production server settings"),
		NewCollection("Dev PostgreSQL", "Vars for Dev Postgres DB"),
		NewCollection("Monitoring & Logging", "Log levels / telemetry endpoints"),
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
