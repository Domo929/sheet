package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Domo929/sheet/internal/ui"
)

func main() {
	// Create the application model
	model, err := ui.NewModel()
	if err != nil {
		fmt.Printf("Error initializing application: %v\n", err)
		os.Exit(1)
	}

	// Create and run the Bubble Tea program
	// Note: Not using WithAltScreen to avoid cursor positioning issues on exit
	p := tea.NewProgram(model, tea.WithMouseCellMotion())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
