package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the data structure that will hold the state of our program.
type model struct {
	// Add any fields you need for your program's state here.
}

// Init initializes the program's state.
func (m model) Init() tea.Cmd {
	// This function is called when the program starts.
	// You can use it to initialize your program's state.
	return nil
}

// Update updates the program's state based on user input.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// This function is called whenever a message is received.
	// You can use it to update your program's state based on user input.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the program's state to the terminal.
func (m model) View() string {
	// You can use it to display the current state of your program.
	return "Press Ctrl+C or 'q' to quit.\n"
}

// main is the entry point of the program.
func main() {
	fmt.Println("Hello, Bubble Tea!")
	// Create a new Bubble Tea program
	p := tea.NewProgram(model{})
	// Start the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
		os.Exit(1)
	}
	// The program will run until it is stopped by the user
	// or an error occurs. The program will exit with a status code of 0
	// if it completes successfully, or a non-zero status code if an error
	// occurs.
	os.Exit(0)
}
