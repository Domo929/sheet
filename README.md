# D&D 5e Character Sheet TUI

A Terminal User Interface (TUI) application for managing D&D 5th Edition character sheets, built with Go and Bubble Tea.

## Features

- Full D&D 5e character creation and management
- Interactive terminal interface
- Character progression tracking
- Inventory and equipment management
- Spellcasting support
- Combat features (HP, death saves, conditions)
- Rest mechanics (short/long rest)

See [DESIGN.md](DESIGN.md) for complete feature list and [IMPLEMENTATION.md](IMPLEMENTATION.md) for development plan.

## Requirements

- Go 1.24 or higher

## Installation

```bash
go build -o sheet ./cmd/sheet
```

## Usage

```bash
./sheet
```

## Development

### Project Structure

```
sheet/
├── cmd/
│   └── sheet/          # Main application entry point
├── internal/
│   ├── models/         # Data models
│   ├── ui/            # UI components
│   ├── engine/        # Calculation engine
│   ├── data/          # Data loading
│   └── storage/       # File management
├── data/              # External JSON data files
├── DESIGN.md          # Design documentation
└── IMPLEMENTATION.md  # Implementation plan
```

### Building

```bash
go build -o sheet ./cmd/sheet
```

### Running

```bash
go run ./cmd/sheet
```

## Contributing

See [IMPLEMENTATION.md](IMPLEMENTATION.md) for git workflow and development guidelines.

## License

TBD