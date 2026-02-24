package main

import (
	"fmt"
	"os"

	"github.com/Domo929/sheet/internal/ui"
	tea "charm.land/bubbletea/v2"
)

func main() {
	// Create the application model
	model, err := ui.NewModel()
	if err != nil {
		fmt.Printf("Error initializing application: %v\n", err)
		os.Exit(1)
	}

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
